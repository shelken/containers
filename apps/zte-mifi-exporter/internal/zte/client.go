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

// TrafficData 流量数据
type TrafficData struct {
	MonthlyTxBytes int64
	MonthlyRxBytes int64
	DateMonth      int
	Success        bool
}

// Client ZTE 设备客户端
type Client struct {
	config      Config
	cacheMu     sync.RWMutex
	cacheData   *TrafficData
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

// GetTrafficData 获取流量数据（带缓存）
func (c *Client) GetTrafficData() *TrafficData {
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

	data := c.fetchTrafficData()
	c.cacheData = data
	c.cacheTime = time.Now()
	return data
}

// fetchTrafficData 从 ZTE 设备获取流量数据
func (c *Client) fetchTrafficData() *TrafficData {
	// 检查是否需要登录
	if time.Now().After(c.loginExpiry) {
		if err := c.login(); err != nil {
			log.Printf("Login failed: %v", err)
			return &TrafficData{Success: false}
		}
		// 登录成功，设置过期时间 (5分钟)
		c.loginExpiry = time.Now().Add(5 * time.Minute)
	}

	// 获取流量数据
	data, err := c.getMonthlyTraffic()
	if err != nil {
		log.Printf("Failed to get traffic data: %v", err)
		// 可能是会话过期，清除登录状态
		c.loginExpiry = time.Time{}
		return &TrafficData{Success: false}
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

// getMonthlyTraffic 获取月度流量数据
func (c *Client) getMonthlyTraffic() (*TrafficData, error) {
	baseURL := fmt.Sprintf("http://%s", c.config.Host)

	// GET 请求不需要 AD 参数，添加时间戳防止缓存
	timestamp := time.Now().UnixMilli()
	reqURL := fmt.Sprintf("%s/goform/goform_get_cmd_process?isTest=false&cmd=monthly_tx_bytes,monthly_rx_bytes,date_month&multi_data=1&_=%d", baseURL, timestamp)

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

	data := &TrafficData{Success: true}

	// 解析数据 (API 返回数字类型，不是字符串)
	data.MonthlyTxBytes = parseIntFromJSON(result["monthly_tx_bytes"])
	data.MonthlyRxBytes = parseIntFromJSON(result["monthly_rx_bytes"])
	data.DateMonth = int(parseIntFromJSON(result["date_month"]))

	log.Printf("Traffic data: TX=%d, RX=%d, Month=%d", data.MonthlyTxBytes, data.MonthlyRxBytes, data.DateMonth)
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
