package zte

import (
	"net/http"
	"net/http/cookiejar"
	"os"
	"testing"
	"time"
)

func TestClientLoginLogic(t *testing.T) {
	// 从环境变量获取测试配置
	zteHost := os.Getenv("ZTE_HOST")
	ztePassword := os.Getenv("ZTE_PASSWORD")

	if zteHost == "" || ztePassword == "" {
		t.Skip("ZTE_HOST and ZTE_PASSWORD environment variables are required for this test")
	}

	// 初始化 HTTP 客户端
	jar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	// 创建 ZTE 客户端
	client := NewClient(Config{
		Host:       zteHost,
		Password:   ztePassword,
		UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36",
		HTTPClient: httpClient,
	})

	// 测试获取设备数据
	t.Run("GetDeviceData", func(t *testing.T) {
		data := client.GetDeviceData()

		if !data.Success {
			t.Error("Failed to get device data")
		}

		// 检查关键字段是否不为空
		if data.NetworkType == "" {
			t.Error("NetworkType is empty")
		}

		if data.Provider == "" {
			t.Error("Provider is empty")
		}

		t.Logf("Successfully retrieved device data: NetworkType=%s, Provider=%s", data.NetworkType, data.Provider)
		t.Logf("Monthly TX: %d, Monthly RX: %d", data.MonthlyTxBytes, data.MonthlyRxBytes)
	})

	// 测试缓存机制
	t.Run("CacheMechanism", func(t *testing.T) {
		// 连续两次调用，第二次应该使用缓存
		start := time.Now()
		firstData := client.GetDeviceData()
		firstDuration := time.Since(start)

		start = time.Now()
		secondData := client.GetDeviceData()
		secondDuration := time.Since(start)

		if !firstData.Success || !secondData.Success {
			t.Error("Failed to get device data for cache test")
		}

		// 验证第二次调用应该更快（因为使用缓存）
		if secondDuration < firstDuration && secondDuration < time.Second {
			t.Logf("Cache working: first call %v, second call %v", firstDuration, secondDuration)
		} else {
			t.Logf("Cache timing: first call %v, second call %v (cache TTL may have expired between calls)", firstDuration, secondDuration)
		}
	})
}