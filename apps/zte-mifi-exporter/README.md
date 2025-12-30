# ZTE MiFi Exporter

ZTE MiFi Exporter 是一个 Prometheus exporter，用于从 ZTE 便携式 WiFi 设备（如 ZTE F50）收集指标数据。

## 简介

该项目允许您从 ZTE 便携式 WiFi 设备（如 MiFi 热点）收集网络使用情况和其他指标，并将其暴露给 Prometheus 进行监控和告警。该工具通过模拟设备 Web 界面的 API 调用来获取数据，使用安全的双重 SHA256 加密认证方式。

## 特性

- 从 ZTE 便携式 WiFi 设备收集全面的网络使用情况指标
- 提供标准的 Prometheus 指标格式
- 支持健康检查端点
- 轻量级 Go 应用，资源占用少
- 安全的双重 SHA256 加密认证
- 支持多种 ZTE 设备（已测试 ZTE F50）
- 包含数据缓存机制（30秒TTL），减少对设备的请求压力
- 智能登录管理：仅在获取数据失败时重新登录，而非固定时间间隔
- 支持多种网络类型和详细的信号指标

## 指标

该 exporter 提供以下指标：

### 设备信息
- `zte_mifi_info` - 设备信息（固件版本等）

### 流量统计
- `zte_mifi_monthly_tx_bytes_total` - 月度传输字节数
- `zte_mifi_monthly_rx_bytes_total` - 月度接收字节数
- `zte_mifi_monthly_bytes_total` - 月度总字节数（传输 + 接收）
- `zte_mifi_monthly_time_seconds` - 月度在线时长（秒）

### 实时速率
- `zte_mifi_realtime_tx_bytes_per_second` - 实时上传速率（字节/秒）
- `zte_mifi_realtime_rx_bytes_per_second` - 实时下载速率（字节/秒）

### 信号状态
- `zte_mifi_signal_bar` - 信号格数（0-5）
- `zte_mifi_rsrp_5g_dbm` - 5G RSRP 信号强度（dBm）
- `zte_mifi_rssi` - RSSI 信号强度
- `zte_mifi_ppp_connected` - PPP 连接状态（1=连接，0=断开）

### 5G NR 详细信号
- `zte_mifi_nr_rsrp_dbm` - NR RSRP 信号强度（dBm）
- `zte_mifi_nr_rsrq_db` - NR RSRQ（dB）
- `zte_mifi_nr_snr_db` - NR SNR（信噪比，dB）

### WiFi 状态
- `zte_mifi_wifi_clients` - 连接的 WiFi 客户端数量
- `zte_mifi_wifi_enabled` - WiFi 启用状态（1=开，0=关）

### 元数据
- `zte_mifi_scrape_success` - 抓取是否成功

## 环境变量

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ZTE_HOST` | 是 | - | ZTE 设备 IP 地址 |
| `ZTE_PASSWORD` | 是 | - | ZTE 设备管理员密码 |
| `LISTEN_ADDR` | 否 | `:9586` | Exporter 监听地址 |

## 安装和使用

### Docker 部署

```bash
docker run -d \
  -e ZTE_HOST=192.168.10.1 \
  -e ZTE_PASSWORD=your_password \
  -p 9586:9586 \
  ghcr.io/shelken/zte-mifi-exporter:0.4.0
```

### Docker Compose

```yaml
version: '3.8'
services:
  zte-mifi-exporter:
    image: ghcr.io/shelken/zte-mifi-exporter:0.4.0
    container_name: zte-mifi-exporter
    environment:
      - ZTE_HOST=192.168.10.1
      - ZTE_PASSWORD=your_password
    ports:
      - "9586:9586"
    restart: unless-stopped
```

### 直接运行

如果您有 Go 环境，也可以直接运行：

```bash
git clone https://github.com/shelken/zte-mifi-exporter.git
cd zte-mifi-exporter
go build -o zte-mifi-exporter .
ZTE_HOST=192.168.10.1 ZTE_PASSWORD=your_password ./zte-mifi-exporter
```

## 端点

- `/metrics` - Prometheus 指标
- `/health` - 健康检查

## 配置 Prometheus

在 Prometheus 配置文件中添加以下 job：

```yaml
scrape_configs:
  - job_name: 'zte-mifi'
    static_configs:
      - targets: ['localhost:9586']
    scrape_interval: 60s  # 建议不要过于频繁，以减少对设备的压力
```

## 支持的设备

- ZTE F50（已测试）
- 具有类似 Web API 的其他 ZTE MiFi 设备（未测试）

## 构建

要构建 Docker 镜像，请确保您已安装 Docker 和 docker-bake：

```bash
# 构建本地镜像
docker buildx bake image-local

# 构建多架构镜像
docker buildx bake image-all
```

## 开发

本项目使用 Go 1.23 编写。要进行开发：

1. 克隆项目
2. 安装 Go 1.23+
3. 运行 `go mod download` 下载依赖
4. 修改代码
5. 运行 `go run main.go` 测试

## 许可证

本项目采用 MIT 许可证 - 请参阅 [LICENSE](LICENSE) 文件了解详情。

## 故障排除

- 确保 ZTE_HOST 指向正确的设备 IP 地址
- 确保 ZTE_PASSWORD 与设备管理员密码匹配
- 检查防火墙设置，确保可以访问 ZTE 设备
- 如果指标无法获取，请检查设备的 Web API 是否可用
- 注意：Exporter 使用缓存机制（30秒TTL），数据不会实时更新
- 登录逻辑：Exporter 仅在获取数据失败时重新登录，而不是基于固定时间间隔
