package mysql

import (
	"github.com/DataDog/datadog-agent/pkg/metrics"
)

const (
	GAUGE     = "gauge"
	RATE      = "rate"
	COUNT     = "count"
	MONOTONIC = "monotonic_count"
	PROC_NAME = "mysqld"
)

// Vars found in "SHOW STATUS;"
var STATUS_VARS = MetricItems{
	// Command Metrics
	"slow_queries":       {"mysql.performance.slow_queries", metrics.RateType, nil},
	"questions":          {"mysql.performance.questions", metrics.RateType, nil},
	"queries":            {"mysql.performance.queries", metrics.RateType, nil},
	"com_select":         {"mysql.performance.com_select", metrics.RateType, nil},
	"com_insert":         {"mysql.performance.com_insert", metrics.RateType, nil},
	"com_update":         {"mysql.performance.com_update", metrics.RateType, nil},
	"com_delete":         {"mysql.performance.com_delete", metrics.RateType, nil},
	"com_replace":        {"mysql.performance.com_replace", metrics.RateType, nil},
	"com_load":           {"mysql.performance.com_load", metrics.RateType, nil},
	"com_insert_select":  {"mysql.performance.com_insert_select", metrics.RateType, nil},
	"com_update_multi":   {"mysql.performance.com_update_multi", metrics.RateType, nil},
	"com_delete_multi":   {"mysql.performance.com_delete_multi", metrics.RateType, nil},
	"com_replace_select": {"mysql.performance.com_replace_select", metrics.RateType, nil},
	// Connection Metrics
	"connections":          {"mysql.net.connections", metrics.RateType, nil},
	"max_used_connections": {"mysql.net.max_connections", metrics.GaugeType, nil},
	"aborted_clients":      {"mysql.net.aborted_clients", metrics.RateType, nil},
	"aborted_connects":     {"mysql.net.aborted_connects", metrics.RateType, nil},
	// Table Cache Metrics
	"open_files":  {"mysql.performance.open_files", metrics.GaugeType, nil},
	"open_tables": {"mysql.performance.open_tables", metrics.GaugeType, nil},
	// Network Metrics
	"bytes_sent":     {"mysql.performance.bytes_sent", metrics.RateType, nil},
	"bytes_received": {"mysql.performance.bytes_received", metrics.RateType, nil},
	// Query Cache Metrics
	"qcache_hits":          {"mysql.performance.qcache_hits", metrics.RateType, nil},
	"qcache_inserts":       {"mysql.performance.qcache_inserts", metrics.RateType, nil},
	"qcache_lowmem_prunes": {"mysql.performance.qcache_lowmem_prunes", metrics.RateType, nil},
	// Table Lock Metrics
	"table_locks_waited":      {"mysql.performance.table_locks_waited", metrics.GaugeType, nil},
	"table_locks_waited_rate": {"mysql.performance.table_locks_waited.rate", metrics.RateType, nil},
	// Temporary Table Metrics
	"created_tmp_tables":      {"mysql.performance.created_tmp_tables", metrics.RateType, nil},
	"created_tmp_disk_tables": {"mysql.performance.created_tmp_disk_tables", metrics.RateType, nil},
	"created_tmp_files":       {"mysql.performance.created_tmp_files", metrics.RateType, nil},
	// Thread Metrics
	"threads_connected": {"mysql.performance.threads_connected", metrics.GaugeType, nil},
	"threads_running":   {"mysql.performance.threads_running", metrics.GaugeType, nil},
	// MyISAM Metrics
	"key_buffer_bytes_unflushed": {"mysql.myisam.key_buffer_bytes_unflushed", metrics.GaugeType, nil},
	"key_buffer_bytes_used":      {"mysql.myisam.key_buffer_bytes_used", metrics.GaugeType, nil},
	"key_read_requests":          {"mysql.myisam.key_read_requests", metrics.RateType, nil},
	"key_reads":                  {"mysql.myisam.key_reads", metrics.RateType, nil},
	"key_write_requests":         {"mysql.myisam.key_write_requests", metrics.RateType, nil},
	"key_writes":                 {"mysql.myisam.key_writes", metrics.RateType, nil},
}

// Possibly from SHOW GLOBAL VARIABLES
var VARIABLES_VARS = MetricItems{
	"key_buffer_size":       {"mysql.myisam.key_buffer_size", metrics.GaugeType, nil},
	"key_cache_utilization": {"mysql.performance.key_cache_utilization", metrics.GaugeType, nil},
	"max_connections":       {"mysql.net.max_connections_available", metrics.GaugeType, nil},
	"query_cache_size":      {"mysql.performance.qcache_size", metrics.GaugeType, nil},
	"table_open_cache":      {"mysql.performance.table_open_cache", metrics.GaugeType, nil},
	"thread_cache_size":     {"mysql.performance.thread_cache_size", metrics.GaugeType, nil},
}

var INNODB_VARS = MetricItems{
	// InnoDB metrics
	"innodb_data_reads":                    {"mysql.innodb.data_reads", metrics.RateType, nil},
	"innodb_data_writes":                   {"mysql.innodb.data_writes", metrics.RateType, nil},
	"innodb_os_log_fsyncs":                 {"mysql.innodb.os_log_fsyncs", metrics.RateType, nil},
	"innodb_mutex_spin_waits":              {"mysql.innodb.mutex_spin_waits", metrics.RateType, nil},
	"innodb_mutex_spin_rounds":             {"mysql.innodb.mutex_spin_rounds", metrics.RateType, nil},
	"innodb_mutex_os_waits":                {"mysql.innodb.mutex_os_waits", metrics.RateType, nil},
	"innodb_row_lock_waits":                {"mysql.innodb.row_lock_waits", metrics.RateType, nil},
	"innodb_row_lock_time":                 {"mysql.innodb.row_lock_time", metrics.RateType, nil},
	"innodb_row_lock_current_waits":        {"mysql.innodb.row_lock_current_waits", metrics.GaugeType, nil},
	"innodb_current_row_locks":             {"mysql.innodb.current_row_locks", metrics.GaugeType, nil},
	"innodb_buffer_pool_bytes_dirty":       {"mysql.innodb.buffer_pool_dirty", metrics.GaugeType, nil},
	"innodb_buffer_pool_bytes_free":        {"mysql.innodb.buffer_pool_free", metrics.GaugeType, nil},
	"innodb_buffer_pool_bytes_used":        {"mysql.innodb.buffer_pool_used", metrics.GaugeType, nil},
	"innodb_buffer_pool_bytes_total":       {"mysql.innodb.buffer_pool_total", metrics.GaugeType, nil},
	"innodb_buffer_pool_read_requests":     {"mysql.innodb.buffer_pool_read_requests", metrics.RateType, nil},
	"innodb_buffer_pool_reads":             {"mysql.innodb.buffer_pool_reads", metrics.RateType, nil},
	"innodb_buffer_pool_pages_utilization": {"mysql.innodb.buffer_pool_utilization", metrics.GaugeType, nil},
}

// Calculated from "SHOW MASTER LOGS;"
var BINLOG_VARS = MetricItems{
	"Binlog_space_usage_bytes": {"mysql.binlog.disk_use", metrics.GaugeType, nil},
}

// Additional Vars found in "SHOW STATUS;"
// Will collect if [FLAG NAME] is True
var OPTIONAL_STATUS_VARS = MetricItems{
	"binlog_cache_disk_use":      {"mysql.binlog.cache_disk_use", metrics.GaugeType, nil},
	"binlog_cache_use":           {"mysql.binlog.cache_use", metrics.GaugeType, nil},
	"handler_commit":             {"mysql.performance.handler_commit", metrics.RateType, nil},
	"handler_delete":             {"mysql.performance.handler_delete", metrics.RateType, nil},
	"handler_prepare":            {"mysql.performance.handler_prepare", metrics.RateType, nil},
	"handler_read_first":         {"mysql.performance.handler_read_first", metrics.RateType, nil},
	"handler_read_key":           {"mysql.performance.handler_read_key", metrics.RateType, nil},
	"handler_read_next":          {"mysql.performance.handler_read_next", metrics.RateType, nil},
	"handler_read_prev":          {"mysql.performance.handler_read_prev", metrics.RateType, nil},
	"handler_read_rnd":           {"mysql.performance.handler_read_rnd", metrics.RateType, nil},
	"handler_read_rnd_next":      {"mysql.performance.handler_read_rnd_next", metrics.RateType, nil},
	"handler_rollback":           {"mysql.performance.handler_rollback", metrics.RateType, nil},
	"handler_update":             {"mysql.performance.handler_update", metrics.RateType, nil},
	"handler_write":              {"mysql.performance.handler_write", metrics.RateType, nil},
	"opened_tables":              {"mysql.performance.opened_tables", metrics.RateType, nil},
	"qcache_total_blocks":        {"mysql.performance.qcache_total_blocks", metrics.GaugeType, nil},
	"qcache_free_blocks":         {"mysql.performance.qcache_free_blocks", metrics.GaugeType, nil},
	"qcache_free_memory":         {"mysql.performance.qcache_free_memory", metrics.GaugeType, nil},
	"qcache_not_cached":          {"mysql.performance.qcache_not_cached", metrics.RateType, nil},
	"qcache_queries_in_cache":    {"mysql.performance.qcache_queries_in_cache", metrics.GaugeType, nil},
	"select_full_join":           {"mysql.performance.select_full_join", metrics.RateType, nil},
	"select_full_range_join":     {"mysql.performance.select_full_range_join", metrics.RateType, nil},
	"select_range":               {"mysql.performance.select_range", metrics.RateType, nil},
	"select_range_check":         {"mysql.performance.select_range_check", metrics.RateType, nil},
	"select_scan":                {"mysql.performance.select_scan", metrics.RateType, nil},
	"sort_merge_passes":          {"mysql.performance.sort_merge_passes", metrics.RateType, nil},
	"sort_range":                 {"mysql.performance.sort_range", metrics.RateType, nil},
	"sort_rows":                  {"mysql.performance.sort_rows", metrics.RateType, nil},
	"sort_scan":                  {"mysql.performance.sort_scan", metrics.RateType, nil},
	"table_locks_immediate":      {"mysql.performance.table_locks_immediate", metrics.GaugeType, nil},
	"table_locks_immediate_rate": {"mysql.performance.table_locks_immediate.rate", metrics.RateType, nil},
	"threads_cached":             {"mysql.performance.threads_cached", metrics.GaugeType, nil},
	"threads_created":            {"mysql.performance.threads_created", metrics.MonotonicCountType, nil},
}

// Status Vars added in Mysql 5.6.6
var OPTIONAL_STATUS_VARS_5_6_6 = MetricItems{
	"Table_open_cache_hits":   {"mysql.performance.table_cache_hits", metrics.RateType, nil},
	"Table_open_cache_misses": {"mysql.performance.table_cache_misses", metrics.RateType, nil},
}

// Will collect if [extra_innodb_metrics] is True
var OPTIONAL_INNODB_VARS = MetricItems{
	"innodb_active_transactions":            {"mysql.innodb.active_transactions", metrics.GaugeType, nil},
	"innodb_buffer_pool_bytes_data":         {"mysql.innodb.buffer_pool_data", metrics.GaugeType, nil},
	"innodb_buffer_pool_pages_data":         {"mysql.innodb.buffer_pool_pages_data", metrics.GaugeType, nil},
	"innodb_buffer_pool_pages_dirty":        {"mysql.innodb.buffer_pool_pages_dirty", metrics.GaugeType, nil},
	"innodb_buffer_pool_pages_flushed":      {"mysql.innodb.buffer_pool_pages_flushed", metrics.RateType, nil},
	"innodb_buffer_pool_pages_free":         {"mysql.innodb.buffer_pool_pages_free", metrics.GaugeType, nil},
	"innodb_buffer_pool_pages_total":        {"mysql.innodb.buffer_pool_pages_total", metrics.GaugeType, nil},
	"innodb_buffer_pool_read_ahead":         {"mysql.innodb.buffer_pool_read_ahead", metrics.RateType, nil},
	"innodb_buffer_pool_read_ahead_evicted": {"mysql.innodb.buffer_pool_read_ahead_evicted", metrics.RateType, nil},
	"innodb_buffer_pool_read_ahead_rnd":     {"mysql.innodb.buffer_pool_read_ahead_rnd", metrics.GaugeType, nil},
	"innodb_buffer_pool_wait_free":          {"mysql.innodb.buffer_pool_wait_free", metrics.MonotonicCountType, nil},
	"innodb_buffer_pool_write_requests":     {"mysql.innodb.buffer_pool_write_requests", metrics.RateType, nil},
	"innodb_checkpoint_age":                 {"mysql.innodb.checkpoint_age", metrics.GaugeType, nil},
	"innodb_current_transactions":           {"mysql.innodb.current_transactions", metrics.GaugeType, nil},
	"innodb_data_fsyncs":                    {"mysql.innodb.data_fsyncs", metrics.RateType, nil},
	"innodb_data_pending_fsyncs":            {"mysql.innodb.data_pending_fsyncs", metrics.GaugeType, nil},
	"innodb_data_pending_reads":             {"mysql.innodb.data_pending_reads", metrics.GaugeType, nil},
	"innodb_data_pending_writes":            {"mysql.innodb.data_pending_writes", metrics.GaugeType, nil},
	"innodb_data_read":                      {"mysql.innodb.data_read", metrics.RateType, nil},
	"innodb_data_written":                   {"mysql.innodb.data_written", metrics.RateType, nil},
	"innodb_dblwr_pages_written":            {"mysql.innodb.dblwr_pages_written", metrics.RateType, nil},
	"innodb_dblwr_writes":                   {"mysql.innodb.dblwr_writes", metrics.RateType, nil},
	"innodb_hash_index_cells_total":         {"mysql.innodb.hash_index_cells_total", metrics.GaugeType, nil},
	"innodb_hash_index_cells_used":          {"mysql.innodb.hash_index_cells_used", metrics.GaugeType, nil},
	"innodb_history_list_length":            {"mysql.innodb.history_list_length", metrics.GaugeType, nil},
	"innodb_ibuf_free_list":                 {"mysql.innodb.ibuf_free_list", metrics.GaugeType, nil},
	"innodb_ibuf_merged":                    {"mysql.innodb.ibuf_merged", metrics.RateType, nil},
	"innodb_ibuf_merged_delete_marks":       {"mysql.innodb.ibuf_merged_delete_marks", metrics.RateType, nil},
	"innodb_ibuf_merged_deletes":            {"mysql.innodb.ibuf_merged_deletes", metrics.RateType, nil},
	"innodb_ibuf_merged_inserts":            {"mysql.innodb.ibuf_merged_inserts", metrics.RateType, nil},
	"innodb_ibuf_merges":                    {"mysql.innodb.ibuf_merges", metrics.RateType, nil},
	"innodb_ibuf_segment_size":              {"mysql.innodb.ibuf_segment_size", metrics.GaugeType, nil},
	"innodb_ibuf_size":                      {"mysql.innodb.ibuf_size", metrics.GaugeType, nil},
	"innodb_lock_structs":                   {"mysql.innodb.lock_structs", metrics.RateType, nil},
	"innodb_locked_tables":                  {"mysql.innodb.locked_tables", metrics.GaugeType, nil},
	"innodb_locked_transactions":            {"mysql.innodb.locked_transactions", metrics.GaugeType, nil},
	"innodb_log_waits":                      {"mysql.innodb.log_waits", metrics.RateType, nil},
	"innodb_log_write_requests":             {"mysql.innodb.log_write_requests", metrics.RateType, nil},
	"innodb_log_writes":                     {"mysql.innodb.log_writes", metrics.RateType, nil},
	"innodb_lsn_current":                    {"mysql.innodb.lsn_current", metrics.RateType, nil},
	"innodb_lsn_flushed":                    {"mysql.innodb.lsn_flushed", metrics.RateType, nil},
	"innodb_lsn_last_checkpoint":            {"mysql.innodb.lsn_last_checkpoint", metrics.RateType, nil},
	"innodb_mem_adaptive_hash":              {"mysql.innodb.mem_adaptive_hash", metrics.GaugeType, nil},
	"innodb_mem_additional_pool":            {"mysql.innodb.mem_additional_pool", metrics.GaugeType, nil},
	"innodb_mem_dictionary":                 {"mysql.innodb.mem_dictionary", metrics.GaugeType, nil},
	"innodb_mem_file_system":                {"mysql.innodb.mem_file_system", metrics.GaugeType, nil},
	"innodb_mem_lock_system":                {"mysql.innodb.mem_lock_system", metrics.GaugeType, nil},
	"innodb_mem_page_hash":                  {"mysql.innodb.mem_page_hash", metrics.GaugeType, nil},
	"innodb_mem_recovery_system":            {"mysql.innodb.mem_recovery_system", metrics.GaugeType, nil},
	"innodb_mem_thread_hash":                {"mysql.innodb.mem_thread_hash", metrics.GaugeType, nil},
	"innodb_mem_total":                      {"mysql.innodb.mem_total", metrics.GaugeType, nil},
	"innodb_os_file_fsyncs":                 {"mysql.innodb.os_file_fsyncs", metrics.RateType, nil},
	"innodb_os_file_reads":                  {"mysql.innodb.os_file_reads", metrics.RateType, nil},
	"innodb_os_file_writes":                 {"mysql.innodb.os_file_writes", metrics.RateType, nil},
	"innodb_os_log_pending_fsyncs":          {"mysql.innodb.os_log_pending_fsyncs", metrics.GaugeType, nil},
	"innodb_os_log_pending_writes":          {"mysql.innodb.os_log_pending_writes", metrics.GaugeType, nil},
	"innodb_os_log_written":                 {"mysql.innodb.os_log_written", metrics.RateType, nil},
	"innodb_pages_created":                  {"mysql.innodb.pages_created", metrics.RateType, nil},
	"innodb_pages_read":                     {"mysql.innodb.pages_read", metrics.RateType, nil},
	"innodb_pages_written":                  {"mysql.innodb.pages_written", metrics.RateType, nil},
	"innodb_pending_aio_log_ios":            {"mysql.innodb.pending_aio_log_ios", metrics.GaugeType, nil},
	"innodb_pending_aio_sync_ios":           {"mysql.innodb.pending_aio_sync_ios", metrics.GaugeType, nil},
	"innodb_pending_buffer_pool_flushes":    {"mysql.innodb.pending_buffer_pool_flushes", metrics.GaugeType, nil},
	"innodb_pending_checkpoint_writes":      {"mysql.innodb.pending_checkpoint_writes", metrics.GaugeType, nil},
	"innodb_pending_ibuf_aio_reads":         {"mysql.innodb.pending_ibuf_aio_reads", metrics.GaugeType, nil},
	"innodb_pending_log_flushes":            {"mysql.innodb.pending_log_flushes", metrics.GaugeType, nil},
	"innodb_pending_log_writes":             {"mysql.innodb.pending_log_writes", metrics.GaugeType, nil},
	"innodb_pending_normal_aio_reads":       {"mysql.innodb.pending_normal_aio_reads", metrics.GaugeType, nil},
	"innodb_pending_normal_aio_writes":      {"mysql.innodb.pending_normal_aio_writes", metrics.GaugeType, nil},
	"innodb_queries_inside":                 {"mysql.innodb.queries_inside", metrics.GaugeType, nil},
	"innodb_queries_queued":                 {"mysql.innodb.queries_queued", metrics.GaugeType, nil},
	"innodb_read_views":                     {"mysql.innodb.read_views", metrics.GaugeType, nil},
	"innodb_rows_deleted":                   {"mysql.innodb.rows_deleted", metrics.RateType, nil},
	"innodb_rows_inserted":                  {"mysql.innodb.rows_inserted", metrics.RateType, nil},
	"innodb_rows_read":                      {"mysql.innodb.rows_read", metrics.RateType, nil},
	"innodb_rows_updated":                   {"mysql.innodb.rows_updated", metrics.RateType, nil},
	"innodb_s_lock_os_waits":                {"mysql.innodb.s_lock_os_waits", metrics.RateType, nil},
	"innodb_s_lock_spin_rounds":             {"mysql.innodb.s_lock_spin_rounds", metrics.RateType, nil},
	"innodb_s_lock_spin_waits":              {"mysql.innodb.s_lock_spin_waits", metrics.RateType, nil},
	"innodb_semaphore_wait_time":            {"mysql.innodb.semaphore_wait_time", metrics.GaugeType, nil},
	"innodb_semaphore_waits":                {"mysql.innodb.semaphore_waits", metrics.GaugeType, nil},
	"innodb_tables_in_use":                  {"mysql.innodb.tables_in_use", metrics.GaugeType, nil},
	"innodb_x_lock_os_waits":                {"mysql.innodb.x_lock_os_waits", metrics.RateType, nil},
	"innodb_x_lock_spin_rounds":             {"mysql.innodb.x_lock_spin_rounds", metrics.RateType, nil},
	"innodb_x_lock_spin_waits":              {"mysql.innodb.x_lock_spin_waits", metrics.RateType, nil},
}

var GALERA_VARS = MetricItems{
	"wsrep_cluster_size":           {"mysql.galera.wsrep_cluster_size", metrics.GaugeType, nil},
	"wsrep_local_recv_queue_avg":   {"mysql.galera.wsrep_local_recv_queue_avg", metrics.GaugeType, nil},
	"wsrep_flow_control_paused":    {"mysql.galera.wsrep_flow_control_paused", metrics.GaugeType, nil},
	"wsrep_flow_control_paused_ns": {"mysql.galera.wsrep_flow_control_paused_ns", metrics.MonotonicCountType, nil},
	"wsrep_flow_control_recv":      {"mysql.galera.wsrep_flow_control_recv", metrics.MonotonicCountType, nil},
	"wsrep_flow_control_sent":      {"mysql.galera.wsrep_flow_control_sent", metrics.MonotonicCountType, nil},
	"wsrep_cert_deps_distance":     {"mysql.galera.wsrep_cert_deps_distance", metrics.GaugeType, nil},
	"wsrep_local_send_queue_avg":   {"mysql.galera.wsrep_local_send_queue_avg", metrics.GaugeType, nil},
}

var PERFORMANCE_VARS = MetricItems{
	"query_run_time_avg":                 {"mysql.performance.query_run_time.avg", metrics.GaugeType, nil},
	"perf_digest_95th_percentile_avg_us": {"mysql.performance.digest_95th_percentile.avg_us", metrics.GaugeType, nil},
}

var SCHEMA_VARS = MetricItems{"information_schema_size": {"mysql.info.schema.size", metrics.GaugeType, nil}}

var REPLICA_VARS = MetricItems{
	"seconds_Behind_Master": {"mysql.replication.seconds_behind_master", metrics.GaugeType, nil},
	"replicas_connected": {Items: []MetricItem{
		{"mysql.replication.slaves_connected", metrics.GaugeType, nil},
		{"mysql.replication.replicas_connected", metrics.GaugeType, nil},
	}},
}

var SYNTHETIC_VARS = MetricItems{
	"qcache_utilization":         {"mysql.performance.qcache.utilization", metrics.GaugeType, nil},
	"qcache_instant_utilization": {"mysql.performance.qcache.utilization.instant", metrics.GaugeType, nil},
}

var BUILDS = []string{"log", "standard", "debug", "valgrind", "embedded"}
