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

		// 如果有多个告警，显示数量，标题保持原样（仅使用 alertname）
		if len(notification.Alerts) > 1 {
			fmt.Fprintf(&buffer, "### Alertmanager %s · %s · %s (%d个)\n", statusStr, alertName, severity, len(notification.Alerts))
		} else {
			// 如果仅有一个告警，尝试提取 name 或 pod 拼接到标题
			displayName := getAlertDisplayName(alert, true)
			fmt.Fprintf(&buffer, "### Alertmanager %s · %s · %s\n", statusStr, displayName, severity)
		}
	} else {
		fmt.Fprintf(&buffer, "### Alertmanager %s\n", statusStr)
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
			"alertname":           true,
			"severity":            true,
			"oci-digest":          true, // 排除哈希值
			"revision":            true, // 排除版本哈希
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
			fmt.Fprintf(&contextStr, "%s=%s", k, alert.Labels[k])
		}

		t := alert.StartsAt
		timeLabel := "开始时间"
		if alert.Status == "resolved" {
			t = alert.EndsAt
			timeLabel = "恢复时间"
		}

		timeStr := t.Local().Format("2006-01-02 15:04:05")

		displayName := getAlertDisplayName(alert, false)
		fmt.Fprintf(&buffer, "> **告警名称**: <font color=\"info\">%s</font>\n", displayName)
		fmt.Fprintf(&buffer, "> **告警级别**: <font color=\"%s\">%s</font>\n", severityColor, severity)
		fmt.Fprintf(&buffer, "> **告警详情**: %s\n", details)
		fmt.Fprintf(&buffer, "> **告警上下文**: %s\n", contextStr.String())
		fmt.Fprintf(&buffer, "> **%s**: %s\n", timeLabel, timeStr)
		buffer.WriteString("\n")
	}
	return buffer.String()
}

// getAlertDisplayName 根据 label 生成告警显示名称
// 逻辑：
// 1. 如果有 name label，返回 alertname - name (当includeAlertName=true) 或 name (当includeAlertName=false)
// 2. 如果没有 name 但有 pod label，处理 pod 名称（按 - 分割剔除倒数后两个），返回 alertname - processed_pod 或 processed_pod
// 3. 否则只返回 alertname
func getAlertDisplayName(alert Alert, includeAlertName bool) string {
	alertName := alert.Labels["alertname"]
	var resourceName string

	if name, ok := alert.Labels["name"]; ok && name != "" {
		resourceName = name
	} else if pod, ok := alert.Labels["pod"]; ok && pod != "" {
		parts := strings.Split(pod, "-")
		// 如果 pod 名称分割后超过 2 部分，剔除倒数后两个（通常是 replica set hash 和 pod hash）
		if len(parts) > 2 {
			resourceName = strings.Join(parts[:len(parts)-2], "-")
		} else {
			resourceName = pod
		}
	}

	if resourceName != "" {
		if includeAlertName {
			return fmt.Sprintf("%s - %s", alertName, resourceName)
		}
		return resourceName
	}

	return alertName
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
