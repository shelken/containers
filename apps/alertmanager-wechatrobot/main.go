package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

type Notification struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
}

type WeChatMsg struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	webhookURL := os.Getenv("WECHAT_WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("WECHAT_WEBHOOK_URL environment variable is required")
	}

	timezone := os.Getenv("TIMEZONE")
	if timezone != "" {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			log.Printf("Error loading location %s: %v, using default UTC", timezone, err)
		} else {
			time.Local = loc
		}
	}

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}

		var notification Notification
		if err := json.Unmarshal(body, &notification); err != nil {
			http.Error(w, "Error unmarshaling request body", http.StatusBadRequest)
			return
		}

		content := formatMarkdown(notification)
		if err := sendToWeChat(webhookURL, content); err != nil {
			log.Printf("Error sending to WeChat: %v", err)
			http.Error(w, "Error sending to WeChat", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func formatMarkdown(notification Notification) string {
	var buffer bytes.Buffer
	statusStr := "<font color=\"warning\">告警</font>"
	if notification.Status == "resolved" {
		statusStr = "<font color=\"info\">恢复</font>"
	}

	// 改进标题：包含告警名称和级别
	if len(notification.Alerts) > 0 {
		alert := notification.Alerts[0]
		alertName := alert.Labels["alertname"]
		severity := alert.Labels["severity"]
		
		// 如果有多个告警，显示数量
		if len(notification.Alerts) > 1 {
			buffer.WriteString(fmt.Sprintf("### Alertmanager %s · %s · %s (%d个)\n", statusStr, alertName, severity, len(notification.Alerts)))
		} else {
			buffer.WriteString(fmt.Sprintf("### Alertmanager %s · %s · %s\n", statusStr, alertName, severity))
		}
	} else {
		buffer.WriteString(fmt.Sprintf("### Alertmanager %s\n", statusStr))
	}
	for _, alert := range notification.Alerts {
		severity := alert.Labels["severity"]
		severityColor := "comment"
		if severity == "critical" || severity == "error" {
			severityColor = "warning"
		}

		// 只显示 summary 字段（一句话摘要）
		details := alert.Annotations["summary"]
		if details == "" {
			details = "无摘要"
		}

		// 收集并排序 Labels 以提供上下文，排除无意义字段
		var keys []string
		excludeKeys := map[string]bool{
			"alertname":          true,
			"severity":           true,
			"oci-digest":         true, // 排除哈希值
			"revision":           true, // 排除版本哈希
			"reportingcontroller": true, // 排除控制器名称
		}
		
		for k := range alert.Labels {
			if !excludeKeys[k] {
				// 排除看起来像哈希的值（以 sha256: 开头或长度超过40的十六进制）
				value := alert.Labels[k]
				if strings.HasPrefix(value, "sha256:") || 
				   (len(value) > 40 && isHexString(value)) ||
				   strings.Contains(value, "@sha") {
					continue
				}
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)

		var contextStr bytes.Buffer
		for i, k := range keys {
			if i > 0 {
				contextStr.WriteString("\n")
			}
			contextStr.WriteString(fmt.Sprintf("%s=%s", k, alert.Labels[k]))
		}

		t := alert.StartsAt
		timeLabel := "开始时间"
		if alert.Status == "resolved" {
			t = alert.EndsAt
			timeLabel = "恢复时间"
		}

		timeStr := t.Local().Format("2006-01-02 15:04:05")

		buffer.WriteString(fmt.Sprintf("> **告警名称**: <font color=\"info\">%s</font>\n", alert.Labels["alertname"]))
		buffer.WriteString(fmt.Sprintf("> **告警级别**: <font color=\"%s\">%s</font>\n", severityColor, severity))
		buffer.WriteString(fmt.Sprintf("> **告警详情**: %s\n", details))
		buffer.WriteString(fmt.Sprintf("> **告警上下文**: %s\n", contextStr.String()))
		buffer.WriteString(fmt.Sprintf("> **%s**: %s\n", timeLabel, timeStr))
		buffer.WriteString("\n")
	}
	return buffer.String()
}

// isHexString 检查字符串是否只包含十六进制字符
func isHexString(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func sendToWeChat(webhookURL, content string) error {
	msg := WeChatMsg{
		MsgType: "markdown",
	}
	msg.Markdown.Content = content

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(msgBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("wechat api returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
