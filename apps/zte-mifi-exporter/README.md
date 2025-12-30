# ZTE MiFi Exporter

Prometheus exporter for ZTE portable WiFi devices (e.g., ZTE F50).

## Metrics

| Metric | Description |
|--------|-------------|
| `zte_mifi_monthly_tx_bytes_total` | Monthly transmitted bytes |
| `zte_mifi_monthly_rx_bytes_total` | Monthly received bytes |
| `zte_mifi_monthly_bytes_total` | Monthly total bytes (tx + rx) |
| `zte_mifi_scrape_success` | Whether the scrape was successful |

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ZTE_HOST` | Yes | - | ZTE device IP address |
| `ZTE_PASSWORD` | Yes | - | ZTE device admin password |
| `LISTEN_ADDR` | No | `:9586` | Exporter listen address |

## Usage

```bash
docker run -d \
  -e ZTE_HOST=192.168.10.1 \
  -e ZTE_PASSWORD=your_password \
  -p 9586:9586 \
  ghcr.io/shelken/zte-mifi-exporter:0.1.0
```

## Endpoints

- `/metrics` - Prometheus metrics
- `/health` - Health check

## Supported Devices

- ZTE F50 (tested)
- Other ZTE MiFi devices with similar Web API (untested)
