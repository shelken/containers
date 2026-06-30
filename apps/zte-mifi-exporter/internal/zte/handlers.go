package zte

import (
	"fmt"
	"net/http"
)

// Handlers HTTP 处理器
type Handlers struct {
	client *Client
	host   string
}

// NewHandlers 创建处理器
func NewHandlers(client *Client, host string) *Handlers {
	return &Handlers{
		client: client,
		host:   host,
	}
}

// HealthHandler 健康检查端点
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// MetricsHandler Prometheus 指标端点
func (h *Handlers) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	data := h.client.GetDeviceData()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// 通用 labels
	labels := fmt.Sprintf(`host="%s",network_type="%s",provider="%s"`, h.host, data.NetworkType, data.Provider)

	// === 设备信息 ===
	fmt.Fprintf(w, "# HELP zte_mifi_info Device information\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_info gauge\n")
	fmt.Fprintf(w, "zte_mifi_info{host=\"%s\",firmware=\"%s\"} 1\n", h.host, data.FirmwareVersion)

	// === 流量统计 ===
	// 月度发送字节数
	fmt.Fprintf(w, "# HELP zte_mifi_monthly_tx_bytes_total Monthly transmitted bytes\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_monthly_tx_bytes_total gauge\n")
	fmt.Fprintf(w, "zte_mifi_monthly_tx_bytes_total{%s} %d\n", labels, data.MonthlyTxBytes)

	// 月度接收字节数
	fmt.Fprintf(w, "# HELP zte_mifi_monthly_rx_bytes_total Monthly received bytes\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_monthly_rx_bytes_total gauge\n")
	fmt.Fprintf(w, "zte_mifi_monthly_rx_bytes_total{%s} %d\n", labels, data.MonthlyRxBytes)

	// 月度总字节数
	fmt.Fprintf(w, "# HELP zte_mifi_monthly_bytes_total Monthly total bytes (tx + rx)\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_monthly_bytes_total gauge\n")
	fmt.Fprintf(w, "zte_mifi_monthly_bytes_total{%s} %d\n", labels, data.MonthlyTxBytes+data.MonthlyRxBytes)

	// 月流量配额上限
	fmt.Fprintf(w, "# HELP zte_mifi_monthly_quota_bytes Monthly traffic quota limit in bytes (0 if device has no limit set)\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_monthly_quota_bytes gauge\n")
	fmt.Fprintf(w, "zte_mifi_monthly_quota_bytes{%s} %d\n", labels, data.MonthlyQuotaBytes)

	// 月流量剩余
	fmt.Fprintf(w, "# HELP zte_mifi_monthly_remaining_bytes Monthly remaining bytes before quota exhausted\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_monthly_remaining_bytes gauge\n")
	remaining := data.MonthlyQuotaBytes - (data.MonthlyTxBytes + data.MonthlyRxBytes)
	if remaining < 0 {
		remaining = 0
	}
	fmt.Fprintf(w, "zte_mifi_monthly_remaining_bytes{%s} %d\n", labels, remaining)

	// 月度在线时长
	fmt.Fprintf(w, "# HELP zte_mifi_monthly_time_seconds Monthly online time in seconds\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_monthly_time_seconds gauge\n")
	fmt.Fprintf(w, "zte_mifi_monthly_time_seconds{%s} %d\n", labels, data.MonthlyTime)

	// === 实时速率 ===
	// 实时上传速率
	fmt.Fprintf(w, "# HELP zte_mifi_realtime_tx_bytes_per_second Realtime upload speed in bytes/s\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_realtime_tx_bytes_per_second gauge\n")
	fmt.Fprintf(w, "zte_mifi_realtime_tx_bytes_per_second{%s} %d\n", labels, data.RealtimeTxThrpt)

	// 实时下载速率
	fmt.Fprintf(w, "# HELP zte_mifi_realtime_rx_bytes_per_second Realtime download speed in bytes/s\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_realtime_rx_bytes_per_second gauge\n")
	fmt.Fprintf(w, "zte_mifi_realtime_rx_bytes_per_second{%s} %d\n", labels, data.RealtimeRxThrpt)

	// === 信号状态 ===
	// 信号格数
	fmt.Fprintf(w, "# HELP zte_mifi_signal_bar Signal bar level (0-5)\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_signal_bar gauge\n")
	fmt.Fprintf(w, "zte_mifi_signal_bar{%s} %d\n", labels, data.SignalBar)

	// 5G RSRP (Z5g_rsrp)
	fmt.Fprintf(w, "# HELP zte_mifi_rsrp_5g_dbm 5G RSRP signal strength in dBm\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_rsrp_5g_dbm gauge\n")
	fmt.Fprintf(w, "zte_mifi_rsrp_5g_dbm{%s} %d\n", labels, data.RSRP5G)

	// RSSI
	fmt.Fprintf(w, "# HELP zte_mifi_rssi RSSI signal level\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_rssi gauge\n")
	fmt.Fprintf(w, "zte_mifi_rssi{%s} %d\n", labels, data.RSSI)

	// === 5G NR 详细信号 ===
	// NR RSRP
	fmt.Fprintf(w, "# HELP zte_mifi_nr_rsrp_dbm NR RSRP signal strength in dBm\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_nr_rsrp_dbm gauge\n")
	fmt.Fprintf(w, "zte_mifi_nr_rsrp_dbm{%s,nr_band=\"%d\"} %d\n", labels, data.NrBands, data.NrRSRP)

	// NR RSRQ
	fmt.Fprintf(w, "# HELP zte_mifi_nr_rsrq_db NR RSRQ in dB\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_nr_rsrq_db gauge\n")
	fmt.Fprintf(w, "zte_mifi_nr_rsrq_db{%s,nr_band=\"%d\"} %d\n", labels, data.NrBands, data.NrRSRQ)

	// NR SNR
	fmt.Fprintf(w, "# HELP zte_mifi_nr_snr_db NR SNR (Signal-to-Noise Ratio) in dB\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_nr_snr_db gauge\n")
	fmt.Fprintf(w, "zte_mifi_nr_snr_db{%s,nr_band=\"%d\"} %d\n", labels, data.NrBands, data.NrSNR)

	// PPP 连接状态
	fmt.Fprintf(w, "# HELP zte_mifi_ppp_connected PPP connection status (1=connected, 0=disconnected)\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_ppp_connected gauge\n")
	pppConnected := 0
	if data.PPPStatus == "ipv4_ipv6_connected" || data.PPPStatus == "ipv4_connected" || data.PPPStatus == "ipv6_connected" {
		pppConnected = 1
	}
	fmt.Fprintf(w, "zte_mifi_ppp_connected{%s,ppp_status=\"%s\"} %d\n", labels, data.PPPStatus, pppConnected)

	// === WiFi 状态 ===
	// WiFi 连接设备数
	fmt.Fprintf(w, "# HELP zte_mifi_wifi_clients Number of connected WiFi clients\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_wifi_clients gauge\n")
	fmt.Fprintf(w, "zte_mifi_wifi_clients{%s} %d\n", labels, data.WifiStaNum)

	// WiFi 开关状态
	fmt.Fprintf(w, "# HELP zte_mifi_wifi_enabled WiFi enabled status (1=on, 0=off)\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_wifi_enabled gauge\n")
	wifiEnabled := 0
	if data.WifiOnOff {
		wifiEnabled = 1
	}
	fmt.Fprintf(w, "zte_mifi_wifi_enabled{%s} %d\n", labels, wifiEnabled)

	// === 元数据 ===
	// 抓取成功状态
	fmt.Fprintf(w, "# HELP zte_mifi_scrape_success Whether the scrape was successful\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_scrape_success gauge\n")
	if data.Success {
		fmt.Fprintf(w, "zte_mifi_scrape_success{host=\"%s\"} 1\n", h.host)
	} else {
		fmt.Fprintf(w, "zte_mifi_scrape_success{host=\"%s\"} 0\n", h.host)
	}
}
