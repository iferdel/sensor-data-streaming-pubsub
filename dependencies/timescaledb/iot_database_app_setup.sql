CREATE TABLE account (
	id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
	username VARCHAR(50) UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	created_at TIMESTAMPTZ DEFAULT NOW()
  );

CREATE TABLE api_key (
	id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
	account_id UUID NOT NULL, 
	api_key_hash TEXT NOT NULL,
	created_at TIMESTAMPTZ DEFAULT NOW(),
	expires_at TIMESTAMPTZ DEFAULT NULL,
	revoked BOOL DEFAULT FALSE,
	CONSTRAINT fk_account
	  FOREIGN KEY (account_id)
			REFERENCES account(id)
		    ON DELETE CASCADE
  );
  COMMENT ON TABLE api_key IS 'one user may have more than one api key depending on the machine that is being logged into the system';
  CREATE INDEX idx_api_key_hash ON api_key (api_key_hash);

CREATE TABLE target (
	id SERIAL PRIMARY KEY,
	name VARCHAR(50) NOT NULL
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
	measurement DOUBLE PRECISION NOT NULL,
	CONSTRAINT fk_sensor 
	  FOREIGN KEY (sensor_id) 
			REFERENCES sensor(id) 
		    ON DELETE CASCADE
	);
  COMMENT ON COLUMN sensor_measurement.measurement IS 'double precision is best for this kind of data since we dont need exact-like precision covered by NUMERIC, as rounding errors can be tolerated';
    
SELECT * FROM create_hypertable(
	'sensor_measurement', 
	'time', 
	chunk_time_interval=>'5 minutes'::interval
	);
SELECT add_dimension('sensor_measurement', by_hash('sensor_id', 4)); -- add sensor_id as partitioning column
SELECT add_retention_policy('sensor_measurement', drop_after => INTERVAL '30 minutes');
CREATE UNIQUE INDEX idx_sensorid_time ON sensor_measurement(sensor_id, time);
ALTER TABLE sensor_measurement SET (
  timescaledb.compress,
  timescaledb.compress_orderby = 'time DESC',
  timescaledb.compress_segmentby = 'sensor_id'
);
SELECT add_compression_policy('sensor_measurement', INTERVAL '15 minutes');
SELECT enable_chunk_skipping('sensor_measurement', 'sensor_id');

CREATE TABLE target_location(
	time TIMESTAMPTZ NOT NULL,
	target_id INTEGER NOT NULL,
	location GEOGRAPHY(POINT, 4326),
	CONSTRAINT fk_target
		FOREIGN KEY (target_id)
			REFERENCES target(id)
		    ON DELETE CASCADE
	);
  COMMENT ON TABLE target_location IS 'multiple sensors may be on the same target, that is why target_location table makes more sense than sensor_location. Another reason is that the sample frequency from sensor_measurement is scoped to the domain of that variable';
  SELECT create_hypertable('target_location', by_range('time'));
  CREATE INDEX ON target_location (target_id, time DESC);
