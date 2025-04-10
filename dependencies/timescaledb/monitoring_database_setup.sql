/*
 * Create a dedicated schema to hold the info
 */
CREATE SCHEMA IF NOT EXISTS statements_history;

/*
 * The snapshots table holds the cluster-wide values
 * each time an overall snapshot is taken. There is
 * no database or user information stored. This allows
 * you to create cluster dashboards for very fast, high-level
 * information on the trending state of the cluster.
 */
CREATE TABLE IF NOT EXISTS statements_history.snapshots (
    created timestamp with time zone NOT NULL,
    calls bigint NOT NULL,
    total_plan_time double precision NOT NULL,
    total_exec_time double precision NOT NULL,
    rows bigint NOT NULL,
    shared_blks_hit bigint NOT NULL,
    shared_blks_read bigint NOT NULL,
    shared_blks_dirtied bigint NOT NULL,
    shared_blks_written bigint NOT NULL,
    local_blks_hit bigint NOT NULL,
    local_blks_read bigint NOT NULL,
    local_blks_dirtied bigint NOT NULL,
    local_blks_written bigint NOT NULL,
    temp_blks_read bigint NOT NULL,
    temp_blks_written bigint NOT NULL,
    wal_records bigint NOT NULL,
    wal_fpi bigint NOT NULL,
    wal_bytes numeric NOT NULL,
		stats_reset timestamp with time zone NOT NULL,
    PRIMARY KEY (created)
);

COMMENT ON TABLE statements_history.snapshots IS
$$This table contains a full aggregate of the pg_stat_statements view
'https://www.postgresql.org/docs/current/pgstatstatements.html#PGSTATSTATEMENTS'
at the time of the snapshot. This allows for very fast queries that require
a very high level overview. It also contains other 
'https://www.postgresql.org/docs/current/monitoring-stats.html'$$;

/*
 * To reduce the storage requirement of saving query statistics
 * at a consistent interval, we store the query text in a separate
 * table and join it as necessary. The queryid is the identifier
 * for each query across tables.
 */
CREATE TABLE IF NOT EXISTS statements_history.queries (
    queryid bigint NOT NULL,
    rolname text,
    datname text,
    query text,
    PRIMARY KEY (queryid, datname, rolname)
);

COMMENT ON TABLE statements_history.queries IS
$$This table contains all query text, this allows us to not repeatably store the query text$$;


/*
 * Finally, we store the individual statistics for each queryid
 * each time we take a snapshot. This allows you to dig into a
 * specific interval of time and see the snapshot-by-snapshot view
 * of query performance and resource usage
 */
CREATE TABLE IF NOT EXISTS statements_history.statements (
    created timestamp with time zone NOT NULL,
    queryid bigint NOT NULL,
    plans bigint NOT NULL,
    total_plan_time double precision NOT NULL,
    calls bigint NOT NULL,
    total_exec_time double precision NOT NULL,
    rows bigint NOT NULL,
    shared_blks_hit bigint NOT NULL,
    shared_blks_read bigint NOT NULL,
    shared_blks_dirtied bigint NOT NULL,
    shared_blks_written bigint NOT NULL,
    local_blks_hit bigint NOT NULL,
    local_blks_read bigint NOT NULL,
    local_blks_dirtied bigint NOT NULL,
    local_blks_written bigint NOT NULL,
    temp_blks_read bigint NOT NULL,
    temp_blks_written bigint NOT NULL,
    wal_records bigint NOT NULL,
    wal_fpi bigint NOT NULL,
    wal_bytes numeric NOT NULL,
    rolname text NOT NULL,
    datname text NOT NULL,
    PRIMARY KEY (created, queryid, rolname, datname),
    FOREIGN KEY (queryid, datname, rolname) REFERENCES statements_history.queries (queryid, datname, rolname) ON DELETE CASCADE
);


/*
 * These next statements create each fo these tables as
 * TimescaleDB hypertables to unlock automatic table partitioning
 * and other features like columnar compression and data retention.
 */

SELECT * FROM create_hypertable(
    'statements_history.statements',
    'created',
    create_default_indexes => false,
    chunk_time_interval => interval '1 week',
    migrate_data => true
);

/*
* Enable hypertable compression on the statements
* hypertable. This will automatically compress chunks
* that are more than one week old. Adjust as appropriate.
*/
ALTER TABLE statements_history.statements SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'datname,rolname,queryid',
    timescaledb.compress_orderby = 'created'
);

SELECT add_compression_policy(
    'statements_history.statements',
    compress_after => interval '1 week',
    if_not_exists => true
);

/*
 * We need to fill the tables with data on a timed basis. This
 * can be done with TimescaleDB User-Defined Actions or other
 * tools like pg_cron.
 * 
 * This example procedure is specifically written for TimescaleDB
 * User Defined Actions. The inner SQL can be adapted for other
 * job scheduling sessions.
 */
CREATE OR REPLACE PROCEDURE statements_history.create_snapshot(
    job_id int,
    config jsonb
)
LANGUAGE plpgsql AS
$function$
DECLARE
    snapshot_time timestamp with time zone := now();
BEGIN
	/*
	 * This first CTE queries pg_stat_statements and joins
	 * to the roles and database table for more detail that
	 * we will store later.
	 */
    WITH statements AS (
        SELECT
            *
        FROM
            pg_stat_statements(true)
        JOIN
            pg_roles ON (userid=pg_roles.oid)
        JOIN
            pg_database ON (dbid=pg_database.oid)
    ), 
    /*
     * We then get the individual queries out of the result
     * and store the text and queryid separately to avoid
     * storing the query text often.
     */
    queries AS (
        INSERT INTO
            statements_history.queries (queryid, query, datname, rolname)
        SELECT
            queryid, query, datname, rolname
        FROM
            statements
        ON CONFLICT
            DO NOTHING
        RETURNING
            queryid
    ), 
    /*
     * This query SUMs all data from all queries and databases
     * to get high-level cluster statistics each time the snapshot
     * is taken.
     */
    snapshot AS (
        INSERT INTO
            statements_history.snapshots
        SELECT
            now(),
            sum(calls),
            sum(total_plan_time) AS total_plan_time,
            sum(total_exec_time) AS total_exec_time,
            sum(rows) AS rows,
            sum(shared_blks_hit) AS shared_blks_hit,
            sum(shared_blks_read) AS shared_blks_read,
            sum(shared_blks_dirtied) AS shared_blks_dirtied,
            sum(shared_blks_written) AS shared_blks_written,
            sum(local_blks_hit) AS local_blks_hit,
            sum(local_blks_read) AS local_blks_read,
            sum(local_blks_dirtied) AS local_blks_dirtied,
            sum(local_blks_written) AS local_blks_written,
            sum(temp_blks_read) AS temp_blks_read,
            sum(temp_blks_written) AS temp_blks_written,
            sum(wal_records) AS wal_records,
            sum(wal_fpi) AS wal_fpi,
            sum(wal_bytes) AS wal_bytes,
						pg_postmaster_start_time()
        FROM
            statements
    )
    /*
     * And finally, we store the individual pg_stat_statement 
     * aggregated results for each query, for each snapshot time.
     */
    INSERT INTO
        statements_history.statements
    SELECT
        snapshot_time,
        queryid,
        plans,
        total_plan_time,
        calls,
        total_exec_time,
        rows,
        shared_blks_hit,
        shared_blks_read,
        shared_blks_dirtied,
        shared_blks_written,
        local_blks_hit,
        local_blks_read,
        local_blks_dirtied,
        local_blks_written,
        temp_blks_read,
        temp_blks_written,
        wal_records,
        wal_fpi,
        wal_bytes,
        rolname,
        datname
    FROM
        statements
		ON CONFLICT DO NOTHING;

END;
$function$;

/*
* Check that the stored procedure works as expected
*/
CALL statements_history.create_snapshot(null, null);

EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM statements_history.statements;

/*
 * Add the recurring UDA that will create the snapshot.
 * 
 * As configured below, a snapshot will be taken every minute. This
 * should be adjusted for your use case, query load, server resources,
 * etc.
 * 
 * This job runs every minute. If you want to store data less often
 * adjust the interval in the statement below.
 */
SELECT add_job(
    'statements_history.create_snapshot',
    interval '1 minutes'
)
WHERE NOT EXISTS (
    SELECT
    FROM
        timescaledb_information.jobs
    WHERE
        proc_name='create_snapshot'
        AND proc_schema='statements_history'
);

/*
* Check that the job was created and is running
*/ 
SELECT * FROM timescaledb_information.jobs;
