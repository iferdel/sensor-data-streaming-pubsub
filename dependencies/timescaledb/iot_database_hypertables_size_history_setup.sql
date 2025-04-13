CREATE SCHEMA IF NOT EXISTS hypertables_size_history;

CREATE table hypertables_size_history.snapshots (
	created TIMESTAMP with time zone NOT NULL,
	hypertable_name TEXT NOT NULL,
	size BIGINT NOT NULL,
	stats_reset timestamp with time zone NOT NULL,
	PRIMARY KEY (created, hypertable_name)
);

CREATE table hypertables_size_history.statements (
	created TIMESTAMP with time zone NOT NULL,
	hypertable_name TEXT NOT NULL,
	size BIGINT NOT NULL,
	PRIMARY KEY (created, hypertable_name)
);

SELECT * FROM create_hypertable(
    'hypertables_size_history.statements',
    'created',
    create_default_indexes => false,
    chunk_time_interval => interval '1 day',
    migrate_data => true
);

ALTER TABLE hypertables_size_history.statements SET (
  timescaledb.compress,
  timescaledb.compress_orderby = 'created',
	timescaledb.compress_segmentby = 'hypertable_name'
);

SELECT add_compression_policy(
    'hypertables_size_history.statements',
    compress_after => interval '1 hour',
    if_not_exists => true
);


CREATE OR REPLACE PROCEDURE hypertables_size_history.create_snapshot(
    job_id int,
    config jsonb
)
LANGUAGE plpgsql AS
$function$
DECLARE
    snapshot_time timestamp with time zone := now();
BEGIN
    WITH statements AS (
        SELECT
					hypertable_name,
					hypertable_size(format('%I.%I', hypertable_schema, hypertable_name)) AS size
        FROM
        	timescaledb_information.hypertables
				WHERE
					hypertable_name != 'statements'
		),
    snapshot AS (
        INSERT INTO
            hypertables_size_history.snapshots
        SELECT
            now(),
            hypertable_name,
						size,
						pg_postmaster_start_time()
        FROM
            statements
    )
    /*
     * And finally, we store the individual pg_stat_statement 
     * aggregated results for each query, for each snapshot time.
     */
    INSERT INTO
        hypertables_size_history.statements
    SELECT
        snapshot_time,
				hypertable_name,
				size
    FROM
        statements
		ON CONFLICT DO NOTHING;

END;
$function$;

/*
* Check that the stored procedure works as expected
*/
CALL hypertables_size_history.create_snapshot(null, null);

EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM hypertables_size_history.statements;


SELECT add_job(
    'hypertables_size_history.create_snapshot',
    interval '15 seconds'
)
WHERE NOT EXISTS (
    SELECT
    FROM
        timescaledb_information.jobs
    WHERE
        proc_name='create_snapshot'
        AND proc_schema='hypertables_size_history'
);

/*
* Check that the job was created and is running
*/ 
SELECT * FROM timescaledb_information.jobs;
