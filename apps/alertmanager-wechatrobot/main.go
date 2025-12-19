package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

	buffer.WriteString(fmt.Sprintf("### Alertmanager %s\n", statusStr))
	for _, alert := range notification.Alerts {
		severity := alert.Labels["severity"]
		severityColor := "comment"
		if severity == "critical" || severity == "error" {
			severityColor = "warning"
		}

		details := alert.Annotations["message"]
		if details == "" {
			details = alert.Annotations["summary"]
		}
		if details == "" {
			details = alert.Annotations["description"]
		}
		if details == "" {
			details = "无详情"
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
		buffer.WriteString(fmt.Sprintf("> **%s**: %s\n", timeLabel, timeStr))
		buffer.WriteString("\n")
	}
	return buffer.String()
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
