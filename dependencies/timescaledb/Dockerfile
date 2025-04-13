FROM timescale/timescaledb-ha:pg17

COPY ./init.sh /docker-entrypoint-initdb.d/init.sh
COPY ./monitoring_database_statements_history_setup.sql /scripts/monitoring_database_statements_history_setup.sql
COPY ./iot_database_hypertables_size_history_setup.sql /scripts/iot_database_hypertables_size_history_setup.sql
COPY ./monitoring_database_fdw_setup.sql /scripts/monitoring_database_fdw_setup.sql

