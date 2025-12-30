package main

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"
)

const (
	defaultListenAddr = ":9586"
	userAgent         = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"
)

var (
	zteHost     string
	ztePassword string
	listenAddr  string
	httpClient  *http.Client
)

func init() {
	zteHost = os.Getenv("ZTE_HOST")
	ztePassword = os.Getenv("ZTE_PASSWORD")
	listenAddr = os.Getenv("LISTEN_ADDR")

	if zteHost == "" {
		log.Fatal("ZTE_HOST environment variable is required")
	}
	if ztePassword == "" {
		log.Fatal("ZTE_PASSWORD environment variable is required")
	}
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}

	// 初始化带 cookie jar 的 HTTP 客户端
	jar, _ := cookiejar.New(nil)
	httpClient = &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}
}

func main() {
	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/health", healthHandler)

	log.Printf("Starting zte-mifi-exporter on %s", listenAddr)
	log.Printf("Monitoring ZTE device at %s", zteHost)

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
