package zte

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Constants for ZTE API
const (
	// Login constants
	DefaultUsername = "admin"
	LoginSuccess    = "0"

	// Time durations
	CacheTTL    = 30 * time.Second
	LoginExpiry = 5 * time.Minute

	// API endpoints and paths
	BaseAPIPath      = "/goform/goform_get_cmd_process"
	SetCmdPath       = "/goform/goform_set_cmd_process"
	IndexHTMLPath    = "/index.html"
	LDTimestampParam = "LD"

	// HTTP headers
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded; charset=UTF-8"
	XRequestedWith            = "XMLHttpRequest"
	AcceptHeader              = "application/json, text/javascript, */*; q=0.01"
	AcceptLanguageHeader      = "zh-CN,zh;q=0.9"

	// API parameter values
	IsTestFalse   = "false"
	GoFormIDLogin = "LOGIN"

	// Critical fields for validation
	NetworkTypeField = "network_type"
	ProviderField    = "network_provider"
	ErrorField       = "Error"
	ResultField      = "result"
)

// Config 客户端配置
type Config struct {
	Host       string
	Password   string
	UserAgent  string
	HTTPClient *http.Client
}

// DeviceData 设备数据
type DeviceData struct {
	// 设备信息
	FirmwareVersion string // 固件版本

	// 流量统计
	MonthlyTxBytes     int64
	MonthlyRxBytes     int64
	MonthlyTime        int64 // 月在线时长(秒)
	MonthlyQuotaBytes  int64 // 月流量配额上限(字节)，设备未设置时为 0
	DateMonth          int

	// 实时速率
	RealtimeTxThrpt int64 // 实时上传速率 (bytes/s)
	RealtimeRxThrpt int64 // 实时下载速率 (bytes/s)

	// 信号状态
	SignalBar   int    // 信号格数 (0-5)
	NetworkType string // 网络类型 (5G/LTE等)
	Provider    string // 运营商
	PPPStatus   string // 连接状态
	RSRP5G      int    // 5G RSRP (dBm)
	RSSI        int    // RSSI 信号

	// 5G NR 详细信号
	NrSNR   int // 5G SNR 信噪比
	NrRSRP  int // NR RSRP (dBm)
	NrRSRQ  int // NR RSRQ (dB)
	NrBands int // 5G 频段

	// WiFi 状态
	WifiStaNum int  // WiFi 连接设备数
	WifiOnOff  bool // WiFi 开关状态

	Success bool
}

// Client ZTE 设备客户端
type Client struct {
	config      Config
	cacheMu     sync.RWMutex
	cacheData   *DeviceData
	cacheTime   time.Time
	cacheTTL    time.Duration
	loginExpiry time.Time
}

// ZTEResponse ZTE API 响应
type ZTEResponse map[string]interface{}

// NewClient 创建新客户端
func NewClient(config Config) *Client {
	return &Client{
		config:   config,
		cacheTTL: CacheTTL,
	}
}

// GetDeviceData 获取设备数据（带缓存）
func (c *Client) GetDeviceData() *DeviceData {
	c.cacheMu.RLock()
	if c.cacheData != nil && time.Since(c.cacheTime) < c.cacheTTL {
		data := c.cacheData
		c.cacheMu.RUnlock()
		return data
	}
	c.cacheMu.RUnlock()

	// 需要刷新缓存
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	// 双重检查
	if c.cacheData != nil && time.Since(c.cacheTime) < c.cacheTTL {
		return c.cacheData
	}

	data := c.fetchDeviceData()
	c.cacheData = data
	c.cacheTime = time.Now()
	return data
}

// fetchDeviceData 从 ZTE 设备获取数据
func (c *Client) fetchDeviceData() *DeviceData {
	// 获取设备数据
	data, err := c.getDeviceStatus()
	if err != nil {
		log.Printf("Failed to get device data: %v", err)
		// 获取数据失败，尝试重新登录
		log.Println("Attempting to re-login due to fetch failure...")
		if err := c.login(); err != nil {
			log.Printf("Re-login failed: %v", err)
			return &DeviceData{Success: false}
		}
		// 重新登录成功，更新登录过期时间
		c.loginExpiry = time.Now().Add(LoginExpiry)

		// 再次尝试获取设备数据
		data, err := c.getDeviceStatus()
		if err != nil {
			log.Printf("Failed to get device data after re-login: %v", err)
			return &DeviceData{Success: false}
		}
		return data
	}

	return data
}

// getLD 获取登录所需的 LD token
func (c *Client) getLD() (string, error) {
	baseURL := fmt.Sprintf("http://%s", c.config.Host)
	timestamp := time.Now().UnixMilli()
	reqURL := fmt.Sprintf("%s%s?isTest=%s&cmd=%s&_=%d", baseURL, BaseAPIPath, IsTestFalse, LDTimestampParam, timestamp)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", err
	}
	c.setCommonHeaders(req, baseURL)

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result ZTEResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse LD response: %w", err)
	}

	ld, ok := result[LDTimestampParam].(string)
	if !ok || ld == "" {
		return "", fmt.Errorf("%s not found in response", LDTimestampParam)
	}

	return ld, nil
}

// login 登录 ZTE 设备
func (c *Client) login() error {
	baseURL := fmt.Sprintf("http://%s", c.config.Host)

	// 1. 获取 LD token
	ld, err := c.getLD()
	if err != nil {
		return fmt.Errorf("failed to get %s: %w", LDTimestampParam, err)
	}

	// 2. 计算加密密码: sha256(sha256(password) + LD)
	encryptedPassword := encryptPasswordWithLD(c.config.Password, ld)

	// 3. 发送登录请求
	loginURL := fmt.Sprintf("%s%s", baseURL, SetCmdPath)
	formData := url.Values{}
	formData.Set("isTest", IsTestFalse)
	formData.Set("goformId", GoFormIDLogin)
	formData.Set("user", DefaultUsername)
	formData.Set("password", encryptedPassword)

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	c.setCommonHeaders(req, baseURL)
	req.Header.Set("Content-Type", ContentTypeFormURLEncoded)
	req.Header.Set("Origin", baseURL)

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result ZTEResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	// 检查登录结果: result=0 表示成功, result=3 表示已登录或失败
	resultVal := fmt.Sprintf("%v", result[ResultField])
	if resultVal != LoginSuccess {
		return fmt.Errorf("login failed: result=%s", resultVal)
	}

	log.Printf("Login successful to %s", c.config.Host)
	return nil
}

// getDeviceStatus 获取设备状态数据
func (c *Client) getDeviceStatus() (*DeviceData, error) {
	baseURL := fmt.Sprintf("http://%s", c.config.Host)

	// 使用完整的查询列表以确保所有字段都能返回 (某些字段需要作为一组查询才会返回)
	cmd := "usb_port_switch,battery_charging,sms_received_flag,sms_unread_num,sms_sim_unread_num," +
		"sim_msisdn,data_volume_limit_switch,battery_value,battery_vol_percent,network_signalbar," +
		"network_rssi,cr_version,iccid,imei,imsi,ipv6_wan_ipaddr,lan_ipaddr,mac_address,msisdn," +
		"network_information,Lte_ca_status,rssi,Z5g_rsrp,lte_rsrp,wifi_access_sta_num,loginfo," +
		"data_volume_alert_percent,data_volume_limit_size,data_volume_limit_unit,realtime_rx_thrpt,realtime_tx_thrpt," +
		"realtime_time,monthly_tx_bytes,monthly_rx_bytes,monthly_time,network_type,network_provider," +
		"ppp_status,signalbar,wifi_onoff_state,date_month"

	timestamp := time.Now().UnixMilli()
	reqURL := fmt.Sprintf("%s%s?isTest=%s&cmd=%s&multi_data=1&_=%d", baseURL, BaseAPIPath, IsTestFalse, cmd, timestamp)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	c.setCommonHeaders(req, baseURL)

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result ZTEResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查是否有错误
	if errMsg, ok := result[ErrorField].(string); ok && errMsg != "" {
		return nil, fmt.Errorf("API error: %s", errMsg)
	}

	data := &DeviceData{Success: true}

	// 设备信息
	data.FirmwareVersion = parseStringFromJSON(result["cr_version"])

	// 流量统计
	data.MonthlyTxBytes = parseIntFromJSON(result["monthly_tx_bytes"])
	data.MonthlyRxBytes = parseIntFromJSON(result["monthly_rx_bytes"])
	data.MonthlyTime = parseIntFromJSON(result["monthly_time"])
	data.DateMonth = int(parseIntFromJSON(result["date_month"]))
	data.MonthlyQuotaBytes = parseQuotaBytes(result["data_volume_limit_switch"], result["data_volume_limit_size"], result["data_volume_limit_unit"])

	// 实时速率
	data.RealtimeTxThrpt = parseIntFromJSON(result["realtime_tx_thrpt"])
	data.RealtimeRxThrpt = parseIntFromJSON(result["realtime_rx_thrpt"])

	// 信号状态
	data.SignalBar = int(parseIntFromJSON(result["signalbar"]))
	data.NetworkType = translateNetworkType(parseStringFromJSON(result["network_type"]))
	data.Provider = parseStringFromJSON(result["network_provider"])
	data.PPPStatus = parseStringFromJSON(result["ppp_status"])
	data.RSRP5G = int(parseIntFromJSON(result["Z5g_rsrp"]))
	data.RSSI = int(parseIntFromJSON(result["rssi"]))

	// 5G 详细信号
	data.NrSNR = int(parseIntFromJSON(result["Nr_snr"]))
	data.NrRSRP = int(parseIntFromJSON(result["nr_rsrp"]))
	data.NrRSRQ = int(parseIntFromJSON(result["nr_rsrq"]))
	data.NrBands = int(parseIntFromJSON(result["Nr_bands"]))

	// WiFi 状态
	data.WifiStaNum = int(parseIntFromJSON(result["wifi_access_sta_num"]))
	data.WifiOnOff = parseStringFromJSON(result["wifi_onoff_state"]) == "1"

	// 检查关键字段是否存在
	if data.NetworkType == "" {
		return nil, fmt.Errorf("critical field '%s' is missing or empty", NetworkTypeField)
	}
	if data.Provider == "" {
		return nil, fmt.Errorf("critical field '%s' is missing or empty", ProviderField)
	}

	log.Printf("Device data: TX=%d, RX=%d, Signal=%d, Network=%s, Provider=%s",
		data.MonthlyTxBytes, data.MonthlyRxBytes, data.SignalBar, data.NetworkType, data.Provider)
	return data, nil
}

// setCommonHeaders 设置通用请求头
func (c *Client) setCommonHeaders(req *http.Request, baseURL string) {
	req.Header.Set("Accept", AcceptHeader)
	req.Header.Set("Accept-Language", AcceptLanguageHeader)
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("X-Requested-With", XRequestedWith)
	req.Header.Set("Referer", fmt.Sprintf("%s%s", baseURL, IndexHTMLPath))
}

// encryptPasswordWithLD 使用双重 SHA256 加密密码: sha256(sha256(password) + LD)
func encryptPasswordWithLD(password, ld string) string {
	// 第一次 SHA256
	hash1 := sha256.Sum256([]byte(password))
	hex1 := strings.ToUpper(fmt.Sprintf("%x", hash1))

	// 第二次 SHA256: sha256(hex1 + ld)
	hash2 := sha256.Sum256([]byte(hex1 + ld))
	return strings.ToUpper(fmt.Sprintf("%x", hash2))
}

// translateNetworkType 将 ZTE API 返回的原始 network_type 值转成人可读的网络制式
func translateNetworkType(networkType string) string {
	t := strings.ToUpper(networkType)
	switch {
	case strings.Contains(t, "5G"), strings.Contains(t, "NR"), strings.Contains(t, "SA"), t == "20", t == "26":
		return "5G"
	case strings.Contains(t, "4G"), strings.Contains(t, "LTE"), t == "13":
		return "4G"
	default:
		return networkType
	}
}

// parseQuotaBytes 解析月流量配额上限(字节)，与 f50-cli.sh quota_bytes 逻辑对齐。
// 当 switch=1 且 unit="data" 时，size 形如 "470_1024"，表示 470 * 1024 MiB。
func parseQuotaBytes(switchVal, sizeVal, unitVal interface{}) int64 {
	sw := parseStringFromJSON(switchVal)
	unit := parseStringFromJSON(unitVal)
	size := parseStringFromJSON(sizeVal)
	if sw == "1" && unit == "data" {
		idx := strings.Index(size, "_")
		if idx > 0 {
			a, err1 := strconv.ParseInt(size[:idx], 10, 64)
			b, err2 := strconv.ParseInt(size[idx+1:], 10, 64)
			if err1 == nil && err2 == nil && a > 0 && b > 0 {
				return a * b * 1024 * 1024
			}
		}
	}
	return 0
}

// parseIntFromJSON 从 JSON 值中解析整数 (可能是 string 或 float64)
func parseIntFromJSON(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case string:
		if val == "" {
			return 0
		}
		n, _ := strconv.ParseInt(val, 10, 64)
		return n
	default:
		return 0
	}
}

// parseStringFromJSON 从 JSON 值中解析字符串
func parseStringFromJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
