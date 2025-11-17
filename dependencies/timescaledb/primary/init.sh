#!/bin/bash

#------------------------------------------------------------------------------
# USER MANAGEMENT
#------------------------------------------------------------------------------
set -e

# Read passwords from Docker secrets
iot_password="$IOT_PASSWORD"
iot_replication_password="$IOT_REPLICATION_PASSWORD"
iot_monitoring_password="$IOT_MONITORING_PASSWORD"
iot_app_password="$IOT_APP_PASSWORD"
iot_readonly_password="$IOT_READONLY_PASSWORD"

#--------------------------------------------------------------------------------
# CREATE ROLES, DATABASES AND SET PRIVILEGES
#--------------------------------------------------------------------------------
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    --------------------------------------------------------------------------------------------
    -- 1. Roles
    --------------------------------------------------------------------------------------------

    -- 1.1 Create an administrative role iot to avoid using 'postgres'.
    CREATE ROLE iot WITH
        LOGIN
        PASSWORD '${iot_password}'
        CREATEDB
        CREATEROLE
        INHERIT;

    -- 1.2. Create iot_replication role for replication
    CREATE ROLE iot_replication WITH 
        LOGIN 
        PASSWORD '${iot_replication_password}' 
        CONNECTION LIMIT 5 
        REPLICATION;

    -- 1.3. Create iot_monitoring role for monitoring tasks such as with pg_stat_statements
    CREATE ROLE iot_monitoring WITH
        LOGIN
        PASSWORD '${iot_monitoring_password}'
        NOSUPERUSER
        INHERIT;

    -- 1.4. Create the application role 'iot_app' for the application ifself.
    CREATE ROLE iot_app WITH
        LOGIN
        PASSWORD '${iot_app_password}'
        NOSUPERUSER
        INHERIT;

    -- 1.5. Create the read-only role for reporting.
    CREATE ROLE iot_readonly WITH
        LOGIN
        PASSWORD '${iot_readonly_password}'
        NOSUPERUSER
        INHERIT;

    --------------------------------------------------------------------------------------------
    -- 2. Databases
    --------------------------------------------------------------------------------------------

    /*
     * -- Monitoring Database --
     * 
     * pg_stat_statements logs queries for the entire
     * PostgreSQL cluster. To avoid the monitoring queries also
     * showing up as part of the overall monitoring under normal
     * operations, it is advisable to create a separate monitoring
     * database within the cluster and filter that 'dbid' out in the 
     * monitoring queries. At the end of the day, what is sayid is that 
     * pg_stat_statements queries, if not filtered, would appear in the 
     * monitoring, and as such be only noise.
     * 
    */
    -- 2.1 Create a monitoring database
    CREATE DATABASE monitoring OWNER iot_monitoring;

    -- 2.2. Create the 'iot' database owned by 'iot_app'.
    CREATE DATABASE iot OWNER iot_app;

    --------------------------------------------------------------------------------------------
    -- 3. Implement permissions.
    --------------------------------------------------------------------------------------------

    -- 3.1 Connect to the 'iot' database.
    \connect iot

    -- 3.2 Set role to 'iot_app'.
    SET ROLE iot_app;

    -- 3.3 Revoke default privileges from 'public'.
    REVOKE ALL ON SCHEMA public FROM public;

    -- 3.4 Reset role back to initial user
    RESET ROLE;

    -- 3.5 Grant predefined roles to iot as superuser-like role.
    GRANT pg_read_all_data, pg_write_all_data, pg_maintain, pg_signal_backend, pg_monitor, pg_use_reserved_connections TO iot;

    -- 3.6. Grant read_all_data predefined role to iot_readonly.
    GRANT pg_read_all_data, pg_read_all_stats TO iot_readonly;

    \connect monitoring

    -- 3.7 Grant monitor role to iot_monitoring.
    GRANT pg_monitor TO iot_monitoring;

EOSQL

#--------------------------------------------------------------------------------
# EXTENSIONS
#--------------------------------------------------------------------------------
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    --------------------------------------------------------------------------------------------
    -- 1. Monitoring extensions
    --------------------------------------------------------------------------------------------

    -- * Connect to the monitoring database
    \c monitoring

    -- * Activate the pg_stat_statements extension
    CREATE EXTENSION pg_stat_statements;

    -- * Activate the pg_stat_kcache extension
    CREATE EXTENSION pg_stat_kcache;

    -- * Activate the postgres_fdw extension so we can query stats by sensor when possible
    CREATE EXTENSION postgres_fdw;
    CREATE SERVER iot_server 
	FOREIGN DATA WRAPPER postgres_fdw
	OPTIONS (host 'localhost', port '5432', dbname 'iot');
    CREATE USER MAPPING for iot_monitoring
	SERVER iot_server
	OPTIONS (user 'iot_readonly');
    ALTER USER MAPPING FOR iot_monitoring SERVER iot_server
	OPTIONS (ADD password_required 'false');
    GRANT USAGE ON FOREIGN SERVER iot_server TO iot_monitoring;

    --------------------------------------------------------------------------------------------
    -- 2. Geospatial extensions
    --------------------------------------------------------------------------------------------

    -- * Connect to the iot database
    \c iot

    -- * Activate the postgis extension (installed but disabled by default in timescaledb docker image)
    CREATE EXTENSION postgis;

    -- ** auto_explain is loaded/enabled by default if used in shared_preload_libraries.

EOSQL


#--------------------------------------------------------------------------------
# CREATE APPLICATION TABLES iot DATABASE
#--------------------------------------------------------------------------------
psql -v on_error_stop=1 --username iot_app -d iot -f /scripts/iot_database_app_setup.sql

#--------------------------------------------------------------------------------
# CREATE HYPERTABLE_SIZE'S TABLE IN IOT DATABASE FOR MONITORING PURPOSES
#--------------------------------------------------------------------------------
psql -v on_error_top=1 --username iot_app -d iot -f /scripts/iot_database_hypertables_size_history_setup.sql

#--------------------------------------------------------------------------------
# ONE-TIME JOB THAT USES FOREIGN DATA WRAPPER TO CONNECT TO IOT DB AN GATHER HYPERTABLE_SIZE TABLE
#--------------------------------------------------------------------------------
psql -v on_error_top=1 --username iot_monitoring -d monitoring -f /scripts/monitoring_database_fdw_setup.sql

#--------------------------------------------------------------------------------
# CREATE PG_STAT_STATEMENTS + PG_STAT_KCACHE MONITORING SCHEMA
#--------------------------------------------------------------------------------
psql -v on_error_stop=1 --username iot_monitoring -d monitoring -f /scripts/monitoring_database_stat_statements_history_setup.sql

#--------------------------------------------------------------------------------
# REPLICATION SETUP
#--------------------------------------------------------------------------------
psql -v on_error_stop=1 --username iot_replication -d iot -f /scripts/replication.sql

# Use environment variable with fallback
REPLICATION_SUBNET="${REPLICATION_SUBNET:-172.16.0.0/12}"

if [ -f "$PGDATA/pg_hba.conf" ]; then
  echo "host replication iot_replication $REPLICATION_SUBNET md5" >> "$PGDATA/pg_hba.conf"
  echo "Configured replication for subnet: $REPLICATION_SUBNET"
else
  echo "pg_hba.conf not found, skipping update"
fi
