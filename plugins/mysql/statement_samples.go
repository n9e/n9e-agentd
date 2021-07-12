package mysql

import (
	"strings"

	"k8s.io/klog/v2"
)

type frozenset []interface{}

func (p frozenset) has(s interface{}) bool {
	for _, v := range p {
		if v == s {
			return true
		}
	}
	return false
}

var (
	VALID_EXPLAIN_STATEMENTS = frozenset{"select", "table", "delete", "insert", "replace", "update"}

	// unless a specific table is configured, we try all of the events_statements tables in descending order of
	// preference
	EVENTS_STATEMENTS_PREFERRED_TABLES = []string{
		"events_statements_history_long",
		"events_statements_current",
		// events_statements_history is the lowest in preference because it keeps the history only as long as the thread
		// exists, which means if an application uses only short-lived connections that execute a single query then we
		// won"t be able to catch any samples of it. By querying events_statements_current we at least guarantee we"ll
		// be able to catch queries from short-lived connections.
		"events_statements_history",
	}

	// default sampling settings for events_statements_* tables
	// rate limit is in samples/second
	// {table -> rate-limit}
	DEFAULT_EVENTS_STATEMENTS_COLLECTIONS_PER_SECOND = map[string]float64{
		"events_statements_history_long": 0.1,
		"events_statements_history":      0.1,
		"events_statements_current":      1,
	}
	DEFAULT_EVENTS_STATEMENTS_COLLECTIONS_PER_SECOND_KEYS = []string{
		"events_statements_history_long",
		"events_statements_history",
		"events_statements_current",
	}

	// columns from events_statements_summary tables which correspond to attributes common to all databases and are
	// therefore stored under other standard keys
	EVENTS_STATEMENTS_SAMPLE_EXCLUDE_KEYS = []string{
		// gets obfuscated
		"sql_text",
		// stored as "instance"
		"current_schema",
		// used for signature
		"digest_text",
		"timer_end_time_s",
		"max_timer_wait_ns",
		"timer_start",
		// included as network.client.ip
		"processlist_host",
	}

	CREATE_TEMP_TABLE = `CREATE TEMPORARY TABLE {temp_table} SELECT
        current_schema,
        sql_text,
        digest,
        digest_text,
        timer_start,
        timer_end,
        timer_wait,
        lock_time,
        rows_affected,
        rows_sent,
        rows_examined,
        select_full_join,
        select_full_range_join,
        select_range,
        select_range_check,
        select_scan,
        sort_merge_passes,
        sort_range,
        sort_rows,
        sort_scan,
        no_index_used,
        no_good_index_used,
        event_name,
        thread_id
     FROM {statements_table}
        WHERE sql_text IS NOT NULL
        AND event_name like 'statement/%%'
        AND digest_text is NOT NULL
        AND digest_text NOT LIKE 'EXPLAIN %%'
        AND timer_start > %s
    LIMIT %s`

	// neither window functions nor this variable-based window function emulation can be used directly on performance_schema
	// tables due to some underlying issue regarding how the performance_schema storage engine works (for some reason
	// many of the rows end up making it past the WHERE clause when they should have been filtered out)
	SUB_SELECT_EVENTS_NUMBERED = `(SELECT
        *,
        @row_num := IF(@current_digest = digest, @row_num + 1, 1) AS row_num,
        @current_digest := digest
    FROM {statements_table}
    ORDER BY digest, timer_wait)`

	SUB_SELECT_EVENTS_WINDOW = `(SELECT
        *,
        row_number() over (partition by digest order by timer_wait desc) as row_num
    FROM {statements_table})`

	STARTUP_TIME_SUBQUERY = `(SELECT UNIX_TIMESTAMP()-VARIABLE_VALUE
    FROM {global_status_table}
    WHERE VARIABLE_NAME='UPTIME')`

	EVENTS_STATEMENTS_QUERY = `SELECT
        current_schema,
        sql_text,
        digest,
        digest_text,
        timer_start,
        @startup_time_s+timer_end*1e-12 as timer_end_time_s,
        timer_wait / 1000 AS timer_wait_ns,
        lock_time / 1000 AS lock_time_ns,
        rows_affected,
        rows_sent,
        rows_examined,
        select_full_join,
        select_full_range_join,
        select_range,
        select_range_check,
        select_scan,
        sort_merge_passes,
        sort_range,
        sort_rows,
        sort_scan,
        no_index_used,
        no_good_index_used,
        processlist_user,
        processlist_host,
        processlist_db
    FROM {statements_numbered} as E
    LEFT JOIN performance_schema.threads as T
        ON E.thread_id = T.thread_id
    WHERE sql_text IS NOT NULL
        AND timer_start > %s
        AND row_num = 1
    ORDER BY timer_wait DESC
    LIMIT %s`

	ENABLED_STATEMENTS_CONSUMERS_QUERY = `SELECT name
    FROM performance_schema.setup_consumers
    WHERE enabled = 'YES'
    AND name LIKE 'events_statements_%'`

	PYMYSQL_NON_RETRYABLE_ERRORS = frozenset{
		1044, // access denied on database
		1046, // no permission on statement
		1049, // unknown database
		1305, // procedure does not exist
		1370, // no execute on procedure
	}
)

// TODO: use sender.Event() instead of statement samples client
func _new_statement_samples_client() interface{} {
	return nil
}

func NewMySQLStatementSamples(c *Check) *MySQLStatementSamples {
	cfg := c.config.StatementSamples
	p := &MySQLStatementSamples{
		Check:                               c,
		_version_processed:                  false,
		_service:                            "mysql",
		_db_hostname:                        c.config.server,
		_enabled:                            cfg.Enabled,
		_run_sync:                           cfg.RunSync,
		_collections_per_second:             cfg.CollectionsPerSecond,
		_events_statements_row_limit:        cfg.EventsStatementsRowLimit,
		_explain_procedure:                  cfg.ExplainProcedure,
		_fully_qualified_explain_procedure:  cfg.FullyQualifiedExplainProcedure,
		_events_statements_temp_table:       cfg.EventsStatementsTempTableName,
		_events_statements_enable_procedure: cfg.EventsStatementsEnableProcedure,
		_preferred_events_statements_tables: EVENTS_STATEMENTS_PREFERRED_TABLES,
	}

	if events_statements_table := cfg.EventsStatementsTable; events_statements_table != "" {
		if _, ok := DEFAULT_EVENTS_STATEMENTS_COLLECTIONS_PER_SECOND[events_statements_table]; ok {
			klog.V(6).Infof("Configured preferred events_statements_table: %s", events_statements_table)
			p._preferred_events_statements_tables = []string{events_statements_table}
		} else {
			klog.Warningf("Invalid events_statements_table: %s. Must be one of %s. Falling back to trying all tables.", events_statements_table, strings.Join(DEFAULT_EVENTS_STATEMENTS_COLLECTIONS_PER_SECOND_KEYS, ", "))
		}
	}

	p._explain_strategies = map[string]interface{}{
		"PROCEDURE":    p._run_explain_procedure,
		"FQ_PROCEDURE": p._run_fully_qualified_explain_procedure,
		"STATEMENT":    p._run_explain,
	}
	p._preferred_explain_strategies = []string{"PROCEDURE", "FQ_PROCEDURE", "STATEMENT"}
	p._init_caches()
	p._statement_samples_client = _new_statement_samples_client()

	return p
}

type MySQLStatementSamples struct {
	*Check
	_version_processed                  bool
	_connection_args                    bool
	_checkpoint                         interface{}
	_last_check_run                     interface{}
	_tags                               interface{}
	_tags_str                           interface{}
	_service                            string
	_collection_loop_future             interface{}
	_cancel_event                       interface{}
	_rate_limiter                       interface{}
	_db_hostname                        string
	_enabled                            bool
	_run_sync                           bool
	_collections_per_second             int
	_events_statements_row_limit        int
	_explain_procedure                  string
	_fully_qualified_explain_procedure  string
	_events_statements_temp_table       string
	_events_statements_enable_procedure string
	_preferred_events_statements_tables []string
	_has_window_functions               interface{}
	_explain_strategies                 interface{}
	_preferred_explain_strategies       interface{}
	_statement_samples_client           interface{}
}

// TODO
func (p *MySQLStatementSamples) run_sampler(tags []string) {
	klog.Warningf("unsupported statement samples")
}
func (p *MySQLStatementSamples) _init_caches()                              {}
func (p *MySQLStatementSamples) cancel()                                    {}
func (p *MySQLStatementSamples) _get_db_connection()                        {}
func (p *MySQLStatementSamples) _close_db_conn()                            {}
func (p *MySQLStatementSamples) collection_loop()                           {}
func (p *MySQLStatementSamples) _cursor_run()                               {}
func (p *MySQLStatementSamples) _get_new_events_statements()                {}
func (p *MySQLStatementSamples) _filter_valid_statement_rows()              {}
func (p *MySQLStatementSamples) _collect_plan_for_statement()               {}
func (p *MySQLStatementSamples) _collect_plans_for_statements()             {}
func (p *MySQLStatementSamples) _get_enabled_performance_schema_consumers() {}
func (p *MySQLStatementSamples) _enable_events_statements_consumers()       {}
func (p *MySQLStatementSamples) _get_sample_collection_strategy()           {}
func (p *MySQLStatementSamples) _collect_statement_samples()                {}
func (p *MySQLStatementSamples) _explain_statement()                        {}
func (p *MySQLStatementSamples) _run_explain()                              {}
func (p *MySQLStatementSamples) _run_explain_procedure()                    {}
func (p *MySQLStatementSamples) _run_fully_qualified_explain_procedure()    {}
func (p *MySQLStatementSamples) _can_explain()                              {}
func (p *MySQLStatementSamples) _parse_execution_plan_cost()                {}
