FROM timescale/timescaledb-ha:pg17.4-ts2.19.3

COPY ./.psqlrc ./.psqlrc

COPY ./init.sh /docker-entrypoint-initdb.d/init.sh

COPY ./iot_database_app_setup.sql /scripts/iot_database_app_setup.sql
COPY ./iot_database_hypertables_size_history_setup.sql /scripts/iot_database_hypertables_size_history_setup.sql

COPY ./monitoring_database_stat_statements_history_setup.sql /scripts/monitoring_database_stat_statements_history_setup.sql
COPY ./monitoring_database_fdw_setup.sql /scripts/monitoring_database_fdw_setup.sql

COPY ./replication.sql /scripts/replication.sql
