# Documentation Index

## Writing Guidelines

When creating new documentation, follow these patterns:

### Structure (BLIND principle)
1. **Bottom line** - TL;DR at the top: what, why, how in 3-5 lines
2. **Impact** - Why should the reader care?
3. **Next steps** - Clear actions they can take
4. **Details** - Supporting information (progressive disclosure)

### Naming Convention
| Type | Prefix | Example |
|------|--------|---------|
| Guides/How-tos | `guide-` or descriptive | `monitoring-guide.md` |
| Known issues | `issue-` | `issue-mqtt-client-id.md` |
| Architecture decisions | `adr-` | `adr-001-stream-protocol.md` |
| Runbooks | `runbook-` | `runbook-database-recovery.md` |

### Formatting Checklist
- [ ] TL;DR section at the top
- [ ] Tables over prose for reference data
- [ ] Code blocks with copy-paste commands
- [ ] Decision trees for troubleshooting
- [ ] Actionable headings ("Fix X" not "About X")
- [ ] Links to related docs
- [ ] Read time estimate in index

### Anti-patterns
- Walls of text without structure
- Buried conclusions
- Missing context (assume reader has 5 minutes)
- Outdated values (keep in sync with compose.yaml)

---

## What's Here

| Document | Purpose | Read Time |
|----------|---------|-----------|
| [monitoring-guide.md](monitoring-guide.md) | Grafana dashboard metrics reference and troubleshooting | 15 min |
| [performance-testing.md](performance-testing.md) | Load testing and video demo guide | 5 min |
| [issue-mqtt-client-id.md](issue-mqtt-client-id.md) | Known issue: MQTT duplicate client IDs | 3 min |

## Quick Start

**To monitor the system:**
1. Open Grafana: http://localhost:3000
2. Navigate to "IoT System Monitoring" dashboard
3. See [monitoring-guide.md](monitoring-guide.md) for metric explanations

**To run load tests:**
1. Use `./scripts/scale-sensors.sh 10 5000` to generate 10 sensors at 5kHz
2. Run `docker compose up -d`
3. See [performance-testing.md](performance-testing.md) for details

## System Architecture (30-second overview)

```
Sensors (MQTT) --> RabbitMQ Streams --> Ingesters (3x) --> TimescaleDB
                                                              |
                                                              v
                                        Grafana <-- Prometheus/Monitoring DB
```

**Default configuration:**
- 2 sensors: AAD-1123 and BBB-3423 (both 100 Hz)
- 3 ingester replicas (Single Active Consumer pattern)
- 30-minute data retention with 5-minute chunks
- Compression after 15 minutes
