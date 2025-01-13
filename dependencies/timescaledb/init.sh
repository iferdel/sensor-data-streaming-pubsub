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
# CREATE TABLES/SCHEMA
#--------------------------------------------------------------------------------
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    --------------------------------------------------------------------------------------------
    -- 1. Create Tables and Constraints.
    --------------------------------------------------------------------------------------------

    -- 1.1 Connect to the 'iot' database.
    \connect iot

    -- 1.2 Set role to 'iot_app'.
    SET ROLE iot_app;

    -- 1.3 Create tables.

    CREATE TABLE target (
	id SERIAL PRIMARY KEY,
	name VARCHAR(50)
    );

    CREATE TABLE sensor (
	id SERIAL PRIMARY KEY, 
	serial_number VARCHAR(20) UNIQUE NOT NULL,
	sample_frequency DOUBLE PRECISION CHECK(sample_frequency > 0.0),
	target_id INTEGER,
	CONSTRAINT fk_target
	    FOREIGN KEY (target_id)
		REFERENCES target(id)
		    ON DELETE CASCADE
	);
	COMMENT ON COLUMN sensor.target_id IS 'one target may have more than 1 sensor, but one sensor is only associated with one target. The sensor id gets populated through new sensors entering the system, but the target_id do not autopopulate on sensors turned on, but afterwards though command-line tooling which would make the association from an user to which target is used for which sensor. This is like a step afterwards since after booting and register serial number, the sensor enters into a --wait-- mode until someone enters the target of the sensor and then this triggers the starting point for start the measurements';

    CREATE TABLE sensor_measurement ( 
	time TIMESTAMPTZ NOT NULL,
	sensor_id INTEGER NOT NULL,
	measurement DOUBLE PRECISION,
	UNIQUE (time, sensor_id),
	CONSTRAINT fk_sensor 
	    FOREIGN KEY (sensor_id) 
		REFERENCES sensor(id) 
		    ON DELETE CASCADE
	);
	COMMENT ON COLUMN sensor_measurement.measurement IS 'double precision is best for this kind of data since we dont need exact-like precision covered by NUMERIC, as rounding errors can be tolerated';
    SELECT create_hypertable('sensor_measurement', by_range('time'));

    CREATE TABLE target_location(
	time TIMESTAMPTZ NOT NULL,
	target_id INTEGER NOT NULL,
	location GEOGRAPHY(POINT, 4326),
	CONSTRAINT fk_target
	    FOREIGN KEY (target_id)
		REFERENCES target(id)
		    ON DELETE CASCADE
	);
	COMMENT ON TABLE target_location IS 'multiple sensors may be on the same target, that is why target_location table makes more sense than sensor_location';
    SELECT create_hypertable('target_location', by_range('time'));
    CREATE INDEX ON target_location (target_id, time DESC);

    -- 1.4 Reset role back to initial user
    RESET ROLE;

EOSQL

