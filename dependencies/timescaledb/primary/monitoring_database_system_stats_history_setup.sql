/*
 * System Statistics History Setup
 *
 * This file captures pg_stat_user_tables, pg_stat_database, and pg_stat_activity
 * snapshots for historical analysis. These stats are stored in the monitoring
 * database and replicated to the replica, allowing dashboards to query the
 * replica instead of hitting the primary.
 *
 * Complements: monitoring_database_stat_statements_history_setup.sql
 */

/*
 * Table statistics history - captures dead tuples, autovacuum activity
 */
CREATE TABLE IF NOT EXISTS statements_history.table_stats (
    created timestamp with time zone NOT NULL,
    schemaname text NOT NULL,
    relname text NOT NULL,
    n_live_tup bigint NOT NULL,
    n_dead_tup bigint NOT NULL,
    dead_tup_ratio numeric,
    last_vacuum timestamp with time zone,
    last_autovacuum timestamp with time zone,
    last_analyze timestamp with time zone,
    last_autoanalyze timestamp with time zone,
    vacuum_count bigint NOT NULL,
    autovacuum_count bigint NOT NULL,
    analyze_count bigint NOT NULL,
    autoanalyze_count bigint NOT NULL,
    n_tup_ins bigint NOT NULL,
    n_tup_upd bigint NOT NULL,
    n_tup_del bigint NOT NULL,
    n_tup_hot_upd bigint NOT NULL,
    seq_scan bigint NOT NULL,
    seq_tup_read bigint NOT NULL,
    idx_scan bigint,
    idx_tup_fetch bigint,
    PRIMARY KEY (created, schemaname, relname)
);

COMMENT ON TABLE statements_history.table_stats IS
$$Snapshots of pg_stat_user_tables for tracking dead tuples, autovacuum activity,
and table access patterns over time. Captured every 15 seconds.$$;

SELECT * FROM create_hypertable(
    'statements_history.table_stats',
    'created',
    create_default_indexes => false,
    chunk_time_interval => interval '1 day',
    if_not_exists => true,
    migrate_data => true
);

ALTER TABLE statements_history.table_stats SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'schemaname,relname',
    timescaledb.compress_orderby = 'created'
);

SELECT add_compression_policy(
    'statements_history.table_stats',
    compress_after => interval '1 hour',
    if_not_exists => true
);

SELECT add_retention_policy(
    'statements_history.table_stats',
    drop_after => interval '7 days',
    if_not_exists => true
);

/*
 * Database statistics history - captures cache hit ratio and database-level stats
 */
CREATE TABLE IF NOT EXISTS statements_history.database_stats (
    created timestamp with time zone NOT NULL,
    datname text NOT NULL,
    numbackends integer NOT NULL,
    xact_commit bigint NOT NULL,
    xact_rollback bigint NOT NULL,
    blks_read bigint NOT NULL,
    blks_hit bigint NOT NULL,
    cache_hit_ratio numeric,
    tup_returned bigint NOT NULL,
    tup_fetched bigint NOT NULL,
    tup_inserted bigint NOT NULL,
    tup_updated bigint NOT NULL,
    tup_deleted bigint NOT NULL,
    conflicts bigint NOT NULL,
    temp_files bigint NOT NULL,
    temp_bytes bigint NOT NULL,
    deadlocks bigint NOT NULL,
    blk_read_time double precision NOT NULL,
    blk_write_time double precision NOT NULL,
    stats_reset timestamp with time zone,
    PRIMARY KEY (created, datname)
);

COMMENT ON TABLE statements_history.database_stats IS
$$Snapshots of pg_stat_database for tracking cache hit ratio, transaction counts,
and database-level performance metrics over time. Captured every 15 seconds.$$;

SELECT * FROM create_hypertable(
    'statements_history.database_stats',
    'created',
    create_default_indexes => false,
    chunk_time_interval => interval '1 day',
    if_not_exists => true,
    migrate_data => true
);

ALTER TABLE statements_history.database_stats SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'datname',
    timescaledb.compress_orderby = 'created'
);

SELECT add_compression_policy(
    'statements_history.database_stats',
    compress_after => interval '1 hour',
    if_not_exists => true
);

SELECT add_retention_policy(
    'statements_history.database_stats',
    drop_after => interval '7 days',
    if_not_exists => true
);

/*
 * Connection statistics history - captures active connections by state
 */
CREATE TABLE IF NOT EXISTS statements_history.connection_stats (
    created timestamp with time zone NOT NULL,
    datname text NOT NULL,
    state text NOT NULL,
    connection_count integer NOT NULL,
    PRIMARY KEY (created, datname, state)
);

COMMENT ON TABLE statements_history.connection_stats IS
$$Snapshots of pg_stat_activity aggregated by connection state for tracking
connection patterns over time. Captured every 15 seconds.$$;

SELECT * FROM create_hypertable(
    'statements_history.connection_stats',
    'created',
    create_default_indexes => false,
    chunk_time_interval => interval '1 day',
    if_not_exists => true,
    migrate_data => true
);

ALTER TABLE statements_history.connection_stats SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'datname,state',
    timescaledb.compress_orderby = 'created'
);

SELECT add_compression_policy(
    'statements_history.connection_stats',
    compress_after => interval '1 hour',
    if_not_exists => true
);

SELECT add_retention_policy(
    'statements_history.connection_stats',
    drop_after => interval '7 days',
    if_not_exists => true
);

/*
 * Stored procedure to capture all system stats in one snapshot
 */
CREATE OR REPLACE PROCEDURE statements_history.create_system_stats_snapshot(
    job_id int,
    config jsonb
)
LANGUAGE plpgsql AS
$function$
DECLARE
    snapshot_time timestamp with time zone := now();
BEGIN
    /*
     * Capture table statistics from pg_stat_user_tables
     * Includes both public schema tables and TimescaleDB internal chunks
     */
    INSERT INTO statements_history.table_stats
    SELECT
        snapshot_time,
        schemaname,
        relname,
        n_live_tup,
        n_dead_tup,
        CASE
            WHEN n_live_tup > 0 THEN ROUND(n_dead_tup::numeric / n_live_tup * 100, 2)
            ELSE 0
        END AS dead_tup_ratio,
        last_vacuum,
        last_autovacuum,
        last_analyze,
        last_autoanalyze,
        vacuum_count,
        autovacuum_count,
        analyze_count,
        autoanalyze_count,
        n_tup_ins,
        n_tup_upd,
        n_tup_del,
        n_tup_hot_upd,
        seq_scan,
        seq_tup_read,
        idx_scan,
        idx_tup_fetch
    FROM dblink(
        'host=iot-timescaledb-primary port=5432 dbname=iot user=iot_monitoring password=iot_monitoring',
        'SELECT schemaname, relname, n_live_tup, n_dead_tup,
                last_vacuum, last_autovacuum, last_analyze, last_autoanalyze,
                vacuum_count, autovacuum_count, analyze_count, autoanalyze_count,
                n_tup_ins, n_tup_upd, n_tup_del, n_tup_hot_upd,
                seq_scan, seq_tup_read, idx_scan, idx_tup_fetch
         FROM pg_stat_user_tables
         WHERE schemaname IN (''public'', ''_timescaledb_internal'')'
    ) AS t(
        schemaname text, relname text, n_live_tup bigint, n_dead_tup bigint,
        last_vacuum timestamp with time zone, last_autovacuum timestamp with time zone,
        last_analyze timestamp with time zone, last_autoanalyze timestamp with time zone,
        vacuum_count bigint, autovacuum_count bigint, analyze_count bigint, autoanalyze_count bigint,
        n_tup_ins bigint, n_tup_upd bigint, n_tup_del bigint, n_tup_hot_upd bigint,
        seq_scan bigint, seq_tup_read bigint, idx_scan bigint, idx_tup_fetch bigint
    )
    ON CONFLICT DO NOTHING;

    /*
     * Capture database statistics from pg_stat_database
     * Calculates cache hit ratio inline
     */
    INSERT INTO statements_history.database_stats
    SELECT
        snapshot_time,
        datname,
        numbackends,
        xact_commit,
        xact_rollback,
        blks_read,
        blks_hit,
        CASE
            WHEN blks_hit + blks_read > 0
            THEN ROUND(100.0 * blks_hit / (blks_hit + blks_read), 2)
            ELSE 100
        END AS cache_hit_ratio,
        tup_returned,
        tup_fetched,
        tup_inserted,
        tup_updated,
        tup_deleted,
        conflicts,
        temp_files,
        temp_bytes,
        deadlocks,
        blk_read_time,
        blk_write_time,
        stats_reset
    FROM dblink(
        'host=iot-timescaledb-primary port=5432 dbname=iot user=iot_monitoring password=iot_monitoring',
        'SELECT datname, numbackends, xact_commit, xact_rollback,
                blks_read, blks_hit, tup_returned, tup_fetched,
                tup_inserted, tup_updated, tup_deleted, conflicts,
                temp_files, temp_bytes, deadlocks, blk_read_time, blk_write_time, stats_reset
         FROM pg_stat_database
         WHERE datname = ''iot'''
    ) AS t(
        datname text, numbackends integer, xact_commit bigint, xact_rollback bigint,
        blks_read bigint, blks_hit bigint, tup_returned bigint, tup_fetched bigint,
        tup_inserted bigint, tup_updated bigint, tup_deleted bigint, conflicts bigint,
        temp_files bigint, temp_bytes bigint, deadlocks bigint,
        blk_read_time double precision, blk_write_time double precision,
        stats_reset timestamp with time zone
    )
    ON CONFLICT DO NOTHING;

    /*
     * Capture connection statistics from pg_stat_activity
     * Aggregated by state to track connection patterns
     */
    INSERT INTO statements_history.connection_stats
    SELECT
        snapshot_time,
        datname,
        COALESCE(state, 'unknown') AS state,
        count
    FROM dblink(
        'host=iot-timescaledb-primary port=5432 dbname=iot user=iot_monitoring password=iot_monitoring',
        'SELECT datname, COALESCE(state, ''unknown'') AS state, COUNT(*)::integer AS count
         FROM pg_stat_activity
         WHERE datname = ''iot''
         GROUP BY datname, state'
    ) AS t(datname text, state text, count integer)
    ON CONFLICT DO NOTHING;

END;
$function$;

/*
 * Test the stored procedure
 */
CALL statements_history.create_system_stats_snapshot(null, null);

/*
 * Schedule the job to run every 15 seconds
 */
SELECT add_job(
    'statements_history.create_system_stats_snapshot',
    interval '15 seconds'
)
WHERE NOT EXISTS (
    SELECT
    FROM timescaledb_information.jobs
    WHERE proc_name = 'create_system_stats_snapshot'
      AND proc_schema = 'statements_history'
);

/*
 * Verify job was created
 */
SELECT * FROM timescaledb_information.jobs
WHERE proc_schema = 'statements_history';
