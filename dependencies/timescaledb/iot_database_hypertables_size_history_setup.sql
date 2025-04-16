CREATE SCHEMA IF NOT EXISTS hypertables_size_history;

CREATE table hypertables_size_history.detailed (
	created TIMESTAMP with time zone NOT NULL,
	hypertable_name TEXT NOT NULL,
	table_bytes BIGINT NOT NULL,
	index_bytes BIGINT NOT NULL,
	toast_bytes BIGINT NOT NULL,
	total_size BIGINT NOT NULL,
	node_name TEXT,
	PRIMARY KEY (created, hypertable_name)
);

SELECT * FROM create_hypertable(
    'hypertables_size_history.detailed',
    'created',
    create_default_indexes => false,
    chunk_time_interval => interval '1 day',
    migrate_data => true
);

ALTER TABLE hypertables_size_history.detailed SET (
  timescaledb.compress,
  timescaledb.compress_orderby = 'created',
	timescaledb.compress_segmentby = 'hypertable_name'
);

SELECT add_compression_policy(
    'hypertables_size_history.detailed',
    compress_after => interval '1 hour',
    if_not_exists => true
);

CREATE OR REPLACE PROCEDURE hypertables_size_history.add_hypertables_size_record(
    job_id int,
    config jsonb
)
LANGUAGE plpgsql AS
$function$
DECLARE
    record_time timestamp with time zone := now();
BEGIN
    WITH detailed AS (
        SELECT
					hypertable_name,
					s.table_bytes,
					s.index_bytes,
					s.toast_bytes,
					s.total_bytes,
					s.node_name
        FROM
        	timescaledb_information.hypertables
				CROSS JOIN LATERAL
					hypertable_detailed_size(format('%I.%I', hypertable_schema, hypertable_name)) as s
				WHERE
					hypertable_name != 'detailed'
		)
    INSERT INTO
        hypertables_size_history.detailed
    SELECT
        record_time,
				hypertable_name,
				table_bytes,
				index_bytes,
				toast_bytes,
				total_bytes,
				node_name
    FROM
        detailed;

END;
$function$;

/*
* Check that the stored procedure works as expected
*/
CALL hypertables_size_history.add_hypertables_size_record(null, null);

EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM hypertables_size_history.detailed;

SELECT add_job(
    'hypertables_size_history.add_hypertables_size_record',
    interval '15 seconds'
)
WHERE NOT EXISTS (
    SELECT
    FROM
        timescaledb_information.jobs
    WHERE
        proc_name='add_hypertables_size_record'
        AND proc_schema='hypertables_size_history'
);

/*
* Check that the job was created and is running
*/ 
SELECT * FROM timescaledb_information.jobs;
