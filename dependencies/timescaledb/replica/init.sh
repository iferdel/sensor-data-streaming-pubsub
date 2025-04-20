#!/bin/bash

set -e

if [ ! -s "/home/postgres/pgdata/data/PG_VERSION" ]; then
  echo "Running pg_basebackup to initialize replica..."

  until pg_basebackup \
    --pgdata=/home/postgres/pgdata/data \
    -R \
    --slot=replica_1_slot \
    --host=iot-timescaledb-primary \
    --port=5432 \
    --username="$PGUSER"
  do
    echo "Waiting for primary to connect..."
    sleep 1s
  done

  chmod 0700 /home/postgres/pgdata/data
else
  echo "Replica already initialized. Skipping basebackup."
fi

exec postgres
