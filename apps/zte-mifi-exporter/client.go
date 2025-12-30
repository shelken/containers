package main

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

// TrafficData 流量数据
type TrafficData struct {
	MonthlyTxBytes int64
	MonthlyRxBytes int64
	DateMonth      int
	Success        bool
}

// ZTEResponse ZTE API 响应
type ZTEResponse map[string]interface{}

var (
	cacheMu     sync.RWMutex
	cacheData   *TrafficData
	cacheTime   time.Time
	cacheTTL    = 30 * time.Second
	loginExpiry time.Time
)

// GetTrafficData 获取流量数据（带缓存）
func GetTrafficData() *TrafficData {
	cacheMu.RLock()
	if cacheData != nil && time.Since(cacheTime) < cacheTTL {
		data := cacheData
		cacheMu.RUnlock()
		return data
	}
	cacheMu.RUnlock()

	// 需要刷新缓存
	cacheMu.Lock()
	defer cacheMu.Unlock()

	// 双重检查
	if cacheData != nil && time.Since(cacheTime) < cacheTTL {
		return cacheData
	}

	data := fetchTrafficData()
	cacheData = data
	cacheTime = time.Now()
	return data
}

// fetchTrafficData 从 ZTE 设备获取流量数据
func fetchTrafficData() *TrafficData {
	// 检查是否需要登录
	if time.Now().After(loginExpiry) {
		if err := login(); err != nil {
			log.Printf("Login failed: %v", err)
			return &TrafficData{Success: false}
		}
		// 登录成功，设置过期时间 (5分钟)
		loginExpiry = time.Now().Add(5 * time.Minute)
	}

	// 获取流量数据
	data, err := getMonthlyTraffic()
	if err != nil {
		log.Printf("Failed to get traffic data: %v", err)
		// 可能是会话过期，清除登录状态
		loginExpiry = time.Time{}
		return &TrafficData{Success: false}
	}

	return data
}

// login 登录 ZTE 设备
func login() error {
	baseURL := fmt.Sprintf("http://%s", zteHost)

	// 计算加密密码 (SHA256 大写 hex)
	encryptedPassword := encryptPassword(ztePassword)

	// 发送登录请求 (LOGIN 不需要 AD 参数)
	loginURL := fmt.Sprintf("%s/goform/goform_set_cmd_process", baseURL)
	formData := url.Values{}
	formData.Set("isTest", "false")
	formData.Set("goformId", "LOGIN")
	formData.Set("password", encryptedPassword)

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	setCommonHeaders(req, baseURL)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Origin", baseURL)

	resp, err := httpClient.Do(req)
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

	// 检查登录结果
	if result["result"] == "0" || result["result"] == 0 {
		return fmt.Errorf("login failed: invalid password")
	}

	log.Printf("Login successful to %s", zteHost)
	return nil
}

// getMonthlyTraffic 获取月度流量数据
func getMonthlyTraffic() (*TrafficData, error) {
	baseURL := fmt.Sprintf("http://%s", zteHost)

	// GET 请求不需要 AD 参数，添加时间戳防止缓存
	timestamp := time.Now().UnixMilli()
	reqURL := fmt.Sprintf("%s/goform/goform_get_cmd_process?isTest=false&cmd=monthly_tx_bytes,monthly_rx_bytes,date_month&multi_data=1&_=%d", baseURL, timestamp)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	setCommonHeaders(req, baseURL)

	resp, err := httpClient.Do(req)
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

// encryptPassword 使用 SHA256 加密密码
func encryptPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return strings.ToUpper(fmt.Sprintf("%x", hash))
}

// setCommonHeaders 设置通用请求头
func setCommonHeaders(req *http.Request, baseURL string) {
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", fmt.Sprintf("%s/index.html", baseURL))
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
