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
	data := h.client.GetTrafficData()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// 月度发送字节数
	fmt.Fprintf(w, "# HELP zte_mifi_monthly_tx_bytes_total Monthly transmitted bytes\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_monthly_tx_bytes_total gauge\n")
	fmt.Fprintf(w, "zte_mifi_monthly_tx_bytes_total{host=\"%s\"} %d\n", h.host, data.MonthlyTxBytes)

	// 月度接收字节数
	fmt.Fprintf(w, "# HELP zte_mifi_monthly_rx_bytes_total Monthly received bytes\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_monthly_rx_bytes_total gauge\n")
	fmt.Fprintf(w, "zte_mifi_monthly_rx_bytes_total{host=\"%s\"} %d\n", h.host, data.MonthlyRxBytes)

	// 月度总字节数
	fmt.Fprintf(w, "# HELP zte_mifi_monthly_bytes_total Monthly total bytes (tx + rx)\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_monthly_bytes_total gauge\n")
	fmt.Fprintf(w, "zte_mifi_monthly_bytes_total{host=\"%s\"} %d\n", h.host, data.MonthlyTxBytes+data.MonthlyRxBytes)

	// 抓取成功状态
	fmt.Fprintf(w, "# HELP zte_mifi_scrape_success Whether the scrape was successful\n")
	fmt.Fprintf(w, "# TYPE zte_mifi_scrape_success gauge\n")
	if data.Success {
		fmt.Fprintf(w, "zte_mifi_scrape_success{host=\"%s\"} 1\n", h.host)
	} else {
		fmt.Fprintf(w, "zte_mifi_scrape_success{host=\"%s\"} 0\n", h.host)
	}
}
