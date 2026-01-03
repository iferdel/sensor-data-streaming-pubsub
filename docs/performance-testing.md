# Performance Testing Guide

## TL;DR

**Run a load test:**
```bash
./scripts/scale-sensors.sh 10 5000  # 10 sensors at 5kHz = 50k Hz total
docker compose up -d
```

**Monitor:** http://localhost:3000 → "IoT Performance Monitoring - Real-time"

**Cleanup:**
```bash
docker compose down
rm compose.override.yml
docker compose up -d
```

---

## Default Configuration

| Sensor | Frequency | Measurements/sec |
|--------|-----------|------------------|
| AAD-1123 | 100 Hz | 100 |
| BBB-3423 | 100 Hz | 100 |
| **Total** | | **200** |

**Infrastructure:**
- 3 ingester replicas (Single Active Consumer pattern)
- TimescaleDB with 5-min chunks, 4-way hash partitioning
- RabbitMQ Streams with MQTT protocol
- 30-minute data retention

---

## Load Testing Scenarios

### Quick Start

```bash
# Generate override file with load test sensors
./scripts/scale-sensors.sh 10 1000

# Start everything
docker compose up -d

# Open monitoring
open http://localhost:3000
```

### Progressive Load Tests

| Scenario | Command | Expected Throughput |
|----------|---------|---------------------|
| Moderate | `./scripts/scale-sensors.sh 5 1000` | ~5k Hz |
| High | `./scripts/scale-sensors.sh 10 10000` | ~100k Hz |
| Extreme | `./scripts/scale-sensors.sh 20 10000` | ~200k Hz |

---

## Key Metrics

### Throughput (Hz)
- **Location:** Large gauge at top of performance dashboard
- **Expected:** Should match total sensor sampling frequency
- **Good value:** Stable, matching configured frequencies

### Processing Latency
- **p50:** < 50ms
- **p95:** < 100ms
- **p99:** < 200ms

### Success Rate
- **Expected:** > 99.9%
- **Investigate if:** < 95%

---

## Monitoring Endpoints

| Service | URL |
|---------|-----|
| Grafana | http://localhost:3000 |
| Prometheus | http://localhost:9090 |
| RabbitMQ Management | http://localhost:15672 |
| Ingester 0 Metrics | http://localhost:2112/metrics |
| Ingester 1 Metrics | http://localhost:2113/metrics |
| Ingester 2 Metrics | http://localhost:2114/metrics |

---

## Troubleshooting

### Low Throughput
1. Check sensor containers: `docker compose ps`
2. Verify MQTT client IDs unique (see [issue-mqtt-client-id.md](issue-mqtt-client-id.md))
3. Check RabbitMQ stream consumer status

### High Latency
1. Check database connection pool saturation
2. Verify TimescaleDB CPU/memory usage
3. Review autovacuum activity

### Metrics Not Showing
1. Verify Prometheus scraping: http://localhost:9090/targets
2. Check ingester containers expose port 2112
3. Restart Grafana: `docker compose restart grafana`

---

## Video Demo Guide

### Setup (Before Recording)

```bash
docker compose down -v  # Clean state
docker compose up -d
# Wait 30 seconds to stabilize
```

**Grafana setup:**
1. Open http://localhost:3000
2. Navigate to "IoT Performance Monitoring - Real-time"
3. Set time range: Last 5 minutes
4. Set refresh: 1 second
5. Full-screen mode (F11)

### Recording Script

| Scene | Duration | What to Show |
|-------|----------|--------------|
| Intro | 15s | Architecture overview |
| Baseline | 30s | Default 2 sensors, ~200 Hz throughput |
| Scale up | 60s | Run scale script, watch throughput climb |
| Peak | 45s | Show ~100k+ Hz, stable latency |
| Database | 30s | Switch to System Monitoring dashboard |
| Conclusion | 15s | Final stats overlay |

### Stats Overlay Template

```
System Performance Summary
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Peak Throughput: XXX,XXX Hz
Total Sensors: XX
Average Latency: XX ms
Success Rate: 99.XX%

Stack: Go 1.23 | RabbitMQ Streams | TimescaleDB
```

---

## Related Files

- Scale script: `scripts/scale-sensors.sh`
- Metrics code: `internal/metrics/metrics.go`
- Handler code: `cmd/sensor-measurements-ingester/handlers.go`
- Dashboard: `dependencies/grafana/provisioning/dashboards/`
