# IoT System Monitoring Guide

## TL;DR

**Dashboard location:** http://localhost:3000 → "IoT System Monitoring"

**Key metrics to watch:**
| Metric | Healthy | Warning | Critical |
|--------|---------|---------|----------|
| Cache Hit Ratio | > 99% | 95-99% | < 95% |
| Dead Tuples % | < 5% | 5-10% | > 10% |
| Sensor Data Lag | < 5 sec | 5-30 sec | > 30 sec |
| Idle in Transaction | 0 | 1-2 | > 2 |

**Quick fixes:**
- Sensor not sending data? Check MQTT client ID uniqueness
- High dead tuples? Tune autovacuum or check for blocking transactions
- Low cache hit? Increase `shared_buffers` or add indexes

---

## Dashboard Architecture

```
iot database                              monitoring database
├── sensor_measurement (hypertable)       ├── statements_history (query stats)
├── sensor (registry)                     ├── table_stats (vacuum/tuples)
└── timescaledb_information.*             ├── database_stats (cache, connections)
                                          └── hypertables_size_history
         │                                           │
         └──────── collected every 15 seconds ───────┘
                              │
                              ▼
                      Grafana Dashboard
```

---

## Quick Reference: All Panels

| Panel | Question It Answers | Data Source |
|-------|---------------------|-------------|
| Cache Hit Ratio | Is PostgreSQL using memory efficiently? | monitoring |
| Connections by State | How many connections, doing what? | monitoring |
| Dead Tuples % | Is autovacuum keeping up? | monitoring |
| Insert Rate | Are sensors publishing at expected rates? | iot |
| Total Measurements | Is retention policy working? | iot |
| Chunk Size Distribution | How is data distributed? | iot |
| Chunk Count | Are chunks being created/dropped? | iot |
| Compression Status | Are old chunks compressed? | iot |
| Sensor Data Lag | Are sensors actively publishing? | iot |
| Autovacuum Activity | When did vacuum last run? | monitoring |
| Dead Tuples by Chunk | Which chunks need attention? | monitoring |
| Slowest Queries | What's consuming database time? | monitoring |

---

## Troubleshooting Decision Trees

### System Slow?

```
Is the system slow?
├── Check "Cache Hit Ratio"
│   └── < 95%? → Increase shared_buffers or add indexes
├── Check "Slowest Queries"
│   └── High mean_exec? → Run EXPLAIN ANALYZE and optimize
├── Check "Dead Tuples %"
│   └── > 10%? → Tune autovacuum or check for blocking transactions
└── Check "Disk I/O Activity"
    └── High reads? → Cache miss issue, review queries
```

### Sensor Data Missing?

```
Sensors not showing data?
├── Check "Sensor Data Lag"
│   └── High lag? → Check sensor containers
├── Check "Insert Rate per Sensor"
│   └── Rate = 0? → MQTT or ingester issue (see mqtt-client-id-issue.md)
└── Check "Connections by State"
    └── No active? → Application connectivity issue
```

### Storage Growing?

```
Is storage growing unexpectedly?
├── Check "Hypertable Size Over Time"
│   └── Growing unbounded? → Retention policy not working
├── Check "Chunk Count"
│   └── Continuously increasing? → Chunks not being dropped
└── Check "Compression Status"
    └── Old chunks uncompressed? → Compression policy issue
```

---

## Metric Details

### System Health

#### Cache Hit Ratio

**What:** Percentage of data reads served from RAM vs disk.

**Why it matters:** Cache hits are 100-1000x faster than disk reads.

| Ratio | Status | Action |
|-------|--------|--------|
| 99-100% | Excellent | None |
| 95-99% | Good | Monitor |
| 90-95% | Fair | Review query patterns |
| < 90% | Poor | Increase shared_buffers, add indexes |

**Investigate:**
```sql
SHOW shared_buffers;  -- Check current value

-- See what's in the buffer cache
SELECT c.relname, pg_size_pretty(count(*) * 8192) as buffered
FROM pg_buffercache b
JOIN pg_class c ON b.relfilenode = c.relfilenode
WHERE b.reldatabase IN (0, (SELECT oid FROM pg_database WHERE datname = current_database()))
GROUP BY c.relname ORDER BY 2 DESC LIMIT 10;
```

#### Database Connections

**What:** Count of connections grouped by state (active, idle, idle in transaction).

**Why it matters:** Each connection consumes ~10MB. "Idle in transaction" blocks autovacuum.

**Expected for this system:**
- Active: 1-10 (ingesters + API + Grafana)
- Idle: Depends on connection pooling
- Idle in transaction: Should be 0

**Find problematic connections:**
```sql
SELECT pid, usename, application_name, state, query,
       NOW() - state_change AS duration
FROM pg_stat_activity
WHERE state = 'idle in transaction'
  AND NOW() - state_change > interval '1 minute';
```

#### Dead Tuples Percentage

**What:** Ratio of obsolete rows to live rows (PostgreSQL MVCC creates dead tuples on UPDATE/DELETE).

**Why it matters:** High dead tuples = table bloat = slower queries.

| Percentage | Status | Action |
|------------|--------|--------|
| 0-5% | Excellent | None |
| 5-10% | Good | Monitor |
| 10-20% | Warning | Tune autovacuum |
| > 20% | Critical | Check for blocking queries, manual VACUUM |

**Check autovacuum status:**
```sql
SELECT relname, n_live_tup, n_dead_tup,
       ROUND(n_dead_tup::numeric / NULLIF(n_live_tup, 0) * 100, 2) AS dead_pct,
       last_autovacuum
FROM pg_stat_user_tables WHERE schemaname = 'public' ORDER BY n_dead_tup DESC;
```

**Tune for high-churn tables:**
```sql
ALTER TABLE sensor_measurement SET (
    autovacuum_vacuum_scale_factor = 0.01,  -- Vacuum at 1% dead (vs 20% default)
    autovacuum_vacuum_threshold = 1000
);
```

---

### Write Performance

#### Insert Rate per Sensor

**What:** Measurements written per second for each sensor.

**Why it matters:** Validates sensors are publishing at configured frequencies.

**Expected:** Should match sensor sample frequencies (default: 100 Hz per sensor).

**If rate = 0:**
1. Check sensor container: `docker logs iot-sensor-simulation-0 --tail 50`
2. Check for MQTT client ID conflict: `docker logs iot-rabbitmq | grep "duplicate"`
3. Check ingester: `docker logs iot-sensor-measurements-ingester-0 --tail 50`

#### Total Measurements Count

**What:** Cumulative row count per sensor in `sensor_measurement`.

**Why it matters:** Should stabilize after retention window (30 min). Growing unbounded = retention broken.

**Expected steady state:** sample_rate × 1,800 seconds per sensor.

**Check retention policy:**
```sql
SELECT * FROM timescaledb_information.jobs WHERE proc_name = 'policy_retention';

-- Manually trigger if needed
CALL run_job((SELECT job_id FROM timescaledb_information.jobs
              WHERE proc_name = 'policy_retention' LIMIT 1));
```

---

### TimescaleDB Storage

#### Chunk Count & Compression

**Configuration:**
- Chunk interval: 5 minutes
- Hash partitions: 4 (by sensor_id)
- Compression: After 15 minutes
- Retention: 30 minutes

**Expected chunks:** ~24-32 (6 time ranges × 4 hash partitions, plus boundary conditions).

**Sawtooth pattern is normal:** +4 chunks every 5 min, -4 when retention expires.

**Check compression status:**
```sql
SELECT chunk_name, range_start, range_end, is_compressed,
       EXTRACT(EPOCH FROM (NOW() - range_end))/60 AS age_minutes
FROM timescaledb_information.chunks
WHERE hypertable_name = 'sensor_measurement'
ORDER BY range_start DESC LIMIT 20;
```

Chunks > 15 min old should be compressed. If not:
```sql
-- Check compression job
SELECT * FROM timescaledb_information.jobs WHERE proc_name LIKE '%compression%';

-- Force compression
SELECT compress_chunk(i, if_not_compressed => true)
FROM show_chunks('sensor_measurement') i
WHERE range_end < NOW() - INTERVAL '15 minutes';
```

---

### Sensor Health

#### Data Lag

**What:** Seconds since last measurement recorded per sensor.

| Lag | Status | Likely Cause |
|-----|--------|--------------|
| 0-5 sec | Healthy | Normal |
| 5-15 sec | Warning | Backpressure |
| 15-30 sec | Critical | Connection issue |
| > 60 sec | Offline | Sensor down |

**Data flow and potential delays:**
```
Sensor → Batch (1 sec) → MQTT → RabbitMQ (1-100ms) → Ingester → DB (1-10ms)
Expected total lag: 1-3 seconds
```

**If one sensor has high lag (others normal):**
```bash
docker ps --filter "name=iot-sensor-simulation"
docker logs iot-sensor-simulation-1 --tail 50
docker logs iot-rabbitmq | grep "duplicate"  # Client ID issue
```

**If all sensors have high lag:**
```bash
docker logs iot-sensor-measurements-ingester-0 --tail 50
docker exec iot-rabbitmq rabbitmq-streams list_stream_consumer_groups
```

---

### Query Performance

#### Active Queries

**What:** Currently running queries from `pg_stat_activity`.

**Warning signs:**
| Duration | Status |
|----------|--------|
| > 10 sec | Slow query, may need optimization |
| > 30 sec | Very slow or blocking |
| idle in transaction > 1 min | Application bug |

**Terminate a stuck query:**
```sql
SELECT pg_cancel_backend(<pid>);     -- Graceful
SELECT pg_terminate_backend(<pid>);  -- Force
```

**Find blocking queries:**
```sql
SELECT blocked_activity.query AS blocked_statement,
       blocking_activity.query AS blocking_statement
FROM pg_locks blocked_locks
JOIN pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
  AND blocking_locks.pid != blocked_locks.pid
JOIN pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;
```

---

## Alerting Recommendations

**Critical (immediate action):**
- Sensor data lag > 60 seconds
- Cache hit ratio < 85%
- Dead tuple percentage > 20%
- Connections > 80% of max_connections

**Warning (investigate):**
- Sensor data lag > 30 seconds
- Cache hit ratio < 95%
- Dead tuple percentage > 10%
- Chunk count growing continuously

---

## Performance Tuning Checklist

When investigating performance issues:

1. Cache hit ratio > 99%?
2. Dead tuples < 5%?
3. Autovacuum running?
4. No idle in transaction?
5. Compression working?
6. Retention policy active?
7. Sensor data lag < 5 sec?
8. Chunk count stable?

---

## Appendix: I/O Interpretation

**Interpreting shared_blks_written vs exec_writes (pg_stat_kcache):**

| Scenario | Meaning |
|----------|---------|
| High shared_blks_written, low exec_writes | PostgreSQL dirtied many pages, but they weren't flushed to disk yet (coalesced later) |
| Low shared_blks_written, high exec_writes | Disk writes came from WAL, background writer, or other backends |
| Both high | Query workload is flushing dirty pages aggressively and hitting disk quickly |
| Both low | System mostly idle/reads only, or WAL/buffer management is effective |

---

## References

- Dashboard JSON: `dependencies/grafana/provisioning/dashboards/iot-system-monitoring.json`
- [PostgreSQL Monitoring](https://www.postgresql.org/docs/current/monitoring.html)
- [TimescaleDB Performance](https://docs.timescale.com/timescaledb/latest/how-to-guides/performance/)
- [Autovacuum Tuning](https://www.postgresql.org/docs/current/routine-vacuuming.html#AUTOVACUUM)
