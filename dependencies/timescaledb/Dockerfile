FROM timescale/timescaledb-ha:pg17

COPY ./init.sh /docker-entrypoint-initdb.d/init.sh
COPY ./monitoring_database_setup.sql /scripts/monitoring_database_setup.sql
