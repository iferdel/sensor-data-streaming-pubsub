FROM timescale/timescaledb-ha:pg17

COPY ./init.sh /docker-entrypoint-initdb.d/init.sh
COPY ./monitoring_database_statements_history_setup.sql /scripts/monitoring_database_statements_history_setup.sql
COPY ./iot_database_hypertable_size_history_setup.sql /scripts/iot_database_hypertable_size_history_setup.sql
