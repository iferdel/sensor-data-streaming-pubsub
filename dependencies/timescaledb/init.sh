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

# Create roles, databases and set privileges
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

    -- 3.7 Grant monitor role to iot_monitoring.
    GRANT pg_monitor TO iot_monitoring;

    --------------------------------------------------------------------------------------------
    -- 4. Create Tables and Constraints.
    --------------------------------------------------------------------------------------------

    -- 4.1 Connect to the 'iot' database.
    \connect iot

    -- 4.2 Set role to 'iot_app'.
    SET ROLE iot_app;

    -- 4.3 Create tables.

    CREATE TABLE sensor (
	id SERIAL PRIMARY KEY, 
	serial_number VARCHAR(50) UNIQUE NOT NULL
	);

    CREATE TABLE sensor_measurement (
	time TIMESTAMPTZ NOT NULL,
	sensor_id INTEGER,
	measurement DOUBLE PRECISION,
	UNIQUE (time, sensor_id),
	CONSTRAINT fk_sensor 
	    FOREIGN KEY (sensor_id) 
		REFERENCES sensor(id) 
		    ON DELETE CASCADE
	);
    SELECT create_hypertable('sensor_measurement', by_range('time'));

    -- 4.4 Reset role back to initial user
    RESET ROLE;

EOSQL
#
# #--------------------------------------------------------------------------------
# # EXTENSIONS
# #--------------------------------------------------------------------------------
# psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
#     --------------------------------------------------------------------------------------------
#     -- 1. Monitoring extensions
#     --------------------------------------------------------------------------------------------
#
#     -- * Connect to the monitoring database
#     \c monitoring
#
#     -- * Activate the pg_stat_statements extension
#     CREATE EXTENSION pg_stat_statements;
#
#     -- ** auto_explain is loaded/enabled by default if used in shared_preload_libraries.
# EOSQL
