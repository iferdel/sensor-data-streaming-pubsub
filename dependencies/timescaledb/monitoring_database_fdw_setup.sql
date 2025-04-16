CREATE SCHEMA IF NOT EXISTS hypertables_size_history;

CREATE OR REPLACE PROCEDURE hypertables_size_history.import_fdw_detailed(
	job_id int,
	config jsonb
)
LANGUAGE plpgsql AS $function$
BEGIN
	RAISE NOTICE 'Executing action % with config %', job_id, config;
	IMPORT FOREIGN SCHEMA hypertables_size_history
  	  LIMIT TO (detailed)
    	FROM SERVER iot_server
    	INTO hypertables_size_history;
	PERFORM delete_job(job_id);
END;
$function$;

SELECT add_job(
    'hypertables_size_history.import_fdw_detailed',
  	interval '30 seconds'
)
WHERE NOT EXISTS (
    SELECT
    FROM
        timescaledb_information.jobs
    WHERE
        proc_name = 'import_fdw_detailed'
        AND proc_schema = 'hypertables_size_history'
);

SELECT * FROM timescaledb_information.jobs;
