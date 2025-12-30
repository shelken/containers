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
	MonthlyTxBytes int64
	MonthlyRxBytes int64
	MonthlyTime    int64 // 月在线时长(秒)
	DateMonth      int

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
		cacheTTL: 30 * time.Second,
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
	// 检查是否需要登录
	if time.Now().After(c.loginExpiry) {
		if err := c.login(); err != nil {
			log.Printf("Login failed: %v", err)
			return &DeviceData{Success: false}
		}
		// 登录成功，设置过期时间 (5分钟)
		c.loginExpiry = time.Now().Add(5 * time.Minute)
	}

	// 获取设备数据
	data, err := c.getDeviceStatus()
	if err != nil {
		log.Printf("Failed to get device data: %v", err)
		// 可能是会话过期，清除登录状态
		c.loginExpiry = time.Time{}
		return &DeviceData{Success: false}
	}

	return data
}

// getLD 获取登录所需的 LD token
func (c *Client) getLD() (string, error) {
	baseURL := fmt.Sprintf("http://%s", c.config.Host)
	timestamp := time.Now().UnixMilli()
	reqURL := fmt.Sprintf("%s/goform/goform_get_cmd_process?isTest=false&cmd=LD&_=%d", baseURL, timestamp)

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

	ld, ok := result["LD"].(string)
	if !ok || ld == "" {
		return "", fmt.Errorf("LD not found in response")
	}

	return ld, nil
}

// login 登录 ZTE 设备
func (c *Client) login() error {
	baseURL := fmt.Sprintf("http://%s", c.config.Host)

	// 1. 获取 LD token
	ld, err := c.getLD()
	if err != nil {
		return fmt.Errorf("failed to get LD: %w", err)
	}

	// 2. 计算加密密码: sha256(sha256(password) + LD)
	encryptedPassword := encryptPasswordWithLD(c.config.Password, ld)

	// 3. 发送登录请求
	loginURL := fmt.Sprintf("%s/goform/goform_set_cmd_process", baseURL)
	formData := url.Values{}
	formData.Set("isTest", "false")
	formData.Set("goformId", "LOGIN")
	formData.Set("user", "admin")
	formData.Set("password", encryptedPassword)

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	c.setCommonHeaders(req, baseURL)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
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
	resultVal := fmt.Sprintf("%v", result["result"])
	if resultVal != "0" {
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
		"data_volume_alert_percent,data_volume_limit_size,realtime_rx_thrpt,realtime_tx_thrpt," +
		"realtime_time,monthly_tx_bytes,monthly_rx_bytes,monthly_time,network_type,network_provider," +
		"ppp_status,signalbar,wifi_onoff_state,date_month"

	timestamp := time.Now().UnixMilli()
	reqURL := fmt.Sprintf("%s/goform/goform_get_cmd_process?isTest=false&cmd=%s&multi_data=1&_=%d", baseURL, cmd, timestamp)

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
	if errMsg, ok := result["Error"].(string); ok && errMsg != "" {
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

	// 实时速率
	data.RealtimeTxThrpt = parseIntFromJSON(result["realtime_tx_thrpt"])
	data.RealtimeRxThrpt = parseIntFromJSON(result["realtime_rx_thrpt"])

	// 信号状态
	data.SignalBar = int(parseIntFromJSON(result["signalbar"]))
	data.NetworkType = parseStringFromJSON(result["network_type"])
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

	log.Printf("Device data: TX=%d, RX=%d, Signal=%d, Network=%s, Provider=%s",
		data.MonthlyTxBytes, data.MonthlyRxBytes, data.SignalBar, data.NetworkType, data.Provider)
	return data, nil
}

// setCommonHeaders 设置通用请求头
func (c *Client) setCommonHeaders(req *http.Request, baseURL string) {
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", fmt.Sprintf("%s/index.html", baseURL))
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
