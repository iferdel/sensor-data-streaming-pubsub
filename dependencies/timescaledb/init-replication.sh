#!/bin/bash
set -e

export PGPASSWORD=iot_replication

if [ ! -s "$PGDATA/PG_VERSION" ]; then
  echo "Running first‑time pg_basebackup..."
  pg_basebackup -h iot-timescaledb-primary \
               -U iot_replication \
               -D "$PGDATA" -Fp -Xs -P -R -C -S replica_1_slot
fi

exec postgres -c config_file=/etc/postgresql.conf
