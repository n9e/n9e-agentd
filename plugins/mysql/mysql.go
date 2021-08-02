package mysql

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/pkg/util/db"
	"github.com/DataDog/datadog-agent/pkg/aggregator"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	core "github.com/DataDog/datadog-agent/pkg/collector/corechecks"
	"github.com/DataDog/datadog-agent/pkg/metrics"
	"github.com/shirou/gopsutil/process"
	"k8s.io/klog/v2"
)

const (
	checkName                  = "mysql"
	SERVICE_CHECK_NAME         = "mysql.can_connect"
	SLAVE_SERVICE_CHECK_NAME   = "mysql.replication.slave_running"
	REPLICA_SERVICE_CHECK_NAME = "mysql.replication.replica_running"
	DEFAULT_MAX_CUSTOM_QUERIES = 20
)

type qcacheStats struct {
	hits       int64
	inserts    int64
	not_cached int64
}

// Check doesn't need additional fields
type Check struct {
	initDone bool
	core.CheckBase
	sender aggregator.Sender

	qcache_stats map[string]*qcacheStats

	_qcache_hits       int64
	_qcache_inserts    int64
	_qcache_not_cached int64
	service_check_tags []string

	config *Config
	db     *sql.DB

	qcacheStats        interface{}
	version            MySQLVersion
	_query_manager     *db.QueryManager
	innodb_stats       *InnoDBMetrics
	_statement_metrics *MySQLStatementMetrics
	_statement_samples *MySQLStatementSamples
}

// Run executes the check
func (c *Check) Run() (err error) {
	if c.sender, err = aggregator.GetSender(c.ID()); err != nil {
		return err
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s", e)
		}
	}()

	if err := c.init(); err != nil {
		return err
	}

	if err := c.check(); err != nil {
		klog.V(5).Infof("error %s", err)
		return err
	}

	c.sender.Commit()
	return nil
}

func (c *Check) Cancel() {
	defer c.CheckBase.Cancel()
	c.cancel()
}

// Configure the Prom check
func (c *Check) Configure(rawInstance integration.Data, rawInitConfig integration.Data, source string) error {
	// Must be called before c.CommonConfigure
	c.BuildID(rawInstance, rawInitConfig)

	err := c.CommonConfigure(rawInstance, source)
	if err != nil {
		return fmt.Errorf("common configure failed: %s", err)
	}

	config, err := buildConfig(rawInstance, rawInitConfig)
	if err != nil {
		return fmt.Errorf("build config failed: %s", err)
	}

	c.config = config

	return nil
}

func promFactory() check.Check {
	return &Check{
		CheckBase: core.NewCheckBase(checkName),
	}
}

func init() {
	core.RegisterCheck(checkName, promFactory)
}

// ############### mysql

func (c *Check) init() error {
	if c.initDone {
		return nil
	}

	tlsConfig, err := c.config.TLS.TLSConfig()
	if err != nil {
		return fmt.Errorf("registering TLS config: %s", err)
	}

	if tlsConfig != nil {
		if err := mysql.RegisterTLSConfig(string(c.ID()), tlsConfig); err != nil {
			return err
		}
	}

	c.qcache_stats = make(map[string]*qcacheStats)
	if c._query_manager, err = db.NewQueryManager(c, nil, nil); err != nil {
		return err
	}
	c.innodb_stats = NewInnoDBMetrics(c)
	c._statement_metrics = NewMySQLStatementMetrics(c)
	c._statement_samples = NewMySQLStatementSamples(c)

	// check_initializations
	if err := c._query_manager.Compile_queries(); err != nil {
		return err
	}

	c.initDone = true

	return nil
}

func (c *Check) CustomQueries() []db.CustomQuery                { return c.config.CustomQueries }
func (c *Check) GlobalCustomQueries() []db.CustomQuery          { return c.config.GlobalCustomQueries }
func (c *Check) UseGlobalCustomQueries() string                 { return c.config.UseGlobalCustomQueries }
func (c *Check) Sender() aggregator.Sender                      { return c.sender }
func (c *Check) Executor(query string) ([][]interface{}, error) { return c.queryRows(query) }

func (c *Check) execute_query_raw(query string) {
	//queryRowArray()
}

func (c *Check) check() error {
	c._set_qcache_stats()

	if err := c._connect(); err != nil {
		return err
	}
	defer c.db.Close()

	// version collection
	if err := c.getVersion(); err != nil {
		klog.Errorf("get mysql version %s", err)
	} else {
		c.sendMetadata()
	}

	// Metric collection
	c._collect_metrics()
	c._collect_system_metrics()
	if c.config.DeepDatabaseMonitoring {
		c._collect_statement_metrics()
		c._statement_samples.run_sampler(c.service_check_tags)
	}

	// keeping track of these:
	c._put_qcache_stats()

	//Custom queries
	c._query_manager.Execute(nil)
	return nil
}

func (c *Check) cancel() {
	if tlsConfig, _ := c.config.TLS.TLSConfig(); tlsConfig != nil {
		mysql.DeregisterTLSConfig(string(c.ID()))
	}
	c._statement_samples.cancel()
}

func (c *Check) _set_qcache_stats() {
	if st, ok := c.qcache_stats[c.config.Dsn]; ok {
		c._qcache_hits = st.hits
		c._qcache_inserts = st.inserts
		c._qcache_not_cached = st.not_cached
	}
}

func (c *Check) _put_qcache_stats() error {
	c.qcache_stats[c.config.Dsn] = &qcacheStats{
		hits:       c._qcache_hits,
		inserts:    c._qcache_inserts,
		not_cached: c._qcache_not_cached,
	}

	return nil
}

func (c *Check) sendMetadata() error {
	klog.V(5).Infof("version %s %s", c.version.version, c.version.build)
	klog.V(5).Infof("flavor %s", c.version.flavor)
	return nil
}

func (c *Check) _connect() (err error) {
	service_check_tags := []string{"server:" + c.config.server, "port:" + c.config.port}

	if c.db, err = sql.Open("mysql", c.config.Dsn); err != nil {
		c.sender.ServiceCheck(SERVICE_CHECK_NAME, metrics.ServiceCheckCritical, "", service_check_tags, "")
		return err
	}
	klog.V(5).Infof("Connected to MySQL")

	c.service_check_tags = service_check_tags
	c.sender.ServiceCheck(SERVICE_CHECK_NAME, metrics.ServiceCheckOK, "", service_check_tags, "")

	return
}

func (c *Check) _collect_metrics() error {
	// Get aggregate of all VARS we want to collect
	metrics := make(MetricItems)
	metrics.update(STATUS_VARS)

	// collect results from db
	results := c._get_stats_from_status()
	results.update(c._get_stats_from_variables())

	if !c.config.Options.DisableInnodbMetrics && c._is_innodb_engine_enabled() {
		values, err := c.innodb_stats.get_stats_from_innodb_status()
		if err != nil {
			klog.Warningf("getStatsFromInnodbStatus err %s", err)
		} else {
			results.update(values)
		}
		err = c.innodb_stats.process_innodb_stats(results, &c.config.Options, metrics)
		if err != nil {
			klog.Warningf("process_innodb_stats err %s", err)
		}
	}

	// Binary log statistics
	if c._get_variable_enabled(results, "log_bin") {
		results["Binlog_space_usage_bytes"] = c._get_binary_log_stats()
	}

	// Compute key cache utilization metric
	key_blocks_unused := Float(results["Key_blocks_unused"])
	key_cache_block_size := Float(results["key_cache_block_size"])
	key_buffer_size := Float(results["key_buffer_size"])
	results["Key_buffer_size"] = key_buffer_size

	// can be null if the unit is missing in the user config (4 instead of 4G for eg.)
	if key_buffer_size != 0 {
		key_cache_utilization := 1 - ((key_blocks_unused * key_cache_block_size) / key_buffer_size)
		results["Key_cache_utilization"] = key_cache_utilization
	}

	results["Key_buffer_bytes_used"] = Float(results["Key_blocks_used"]) * key_cache_block_size
	results["Key_buffer_bytes_unflushed"] = Float(results["Key_blocks_not_flushed"]) * key_cache_block_size

	metrics.update(VARIABLES_VARS)
	metrics.update(INNODB_VARS)
	metrics.update(BINLOG_VARS)

	if c.config.Options.ExtraStatusMetrics {
		klog.V(6).Info("Collecting Extra Status Metrics")
		metrics.update(OPTIONAL_STATUS_VARS)

		if c.version.versionCompatible(5, 6, 6) {
			metrics.update(OPTIONAL_STATUS_VARS_5_6_6)
		}
	}

	if c.config.Options.GaleraCluster {
		// already in result-set after 'SHOW STATUS' just add vars to collect
		klog.V(6).Info("Collecting Galera Metrics.")
		metrics.update(GALERA_VARS)
	}

	performance_schema_enabled := c._get_variable_enabled(results, "performance_schema")
	above_560 := c.version.versionCompatible(5, 6, 0)
	if c.config.Options.ExtraPerformanceMetrics && above_560 && performance_schema_enabled {

		// report avg query response time per schema to Datadog
		results["perf_digest_95th_percentile_avg_us"] = c._get_query_exec_time_95th_us()
		results["query_run_time_avg"] = c._query_exec_time_per_schema()
		metrics.update(PERFORMANCE_VARS)
	}

	if c.config.Options.SchemaSizeMetrics {
		// report avg query response time per schema to Datadog
		results["information_schema_size"] = c._query_size_per_schema()
		metrics.update(SCHEMA_VARS)
	}

	if c.config.Options.Replication {
		replication_metrics := c._collect_replication_metrics(results, above_560)
		metrics.update(replication_metrics)
		c._check_replication_status(results)
	}

	// "synthetic" metrics
	metrics.update(SYNTHETIC_VARS)
	c._compute_synthetic_results(results)

	// remove uncomputed metrics
	for k, _ := range SYNTHETIC_VARS {
		if _, ok := results[k]; !ok {
			metrics.pop(k)
		}
	}

	// add duped metrics - reporting some as both rate and gauge
	dupes := map[string]string{
		"Table_locks_waited":    "Table_locks_waited_rate",
		"Table_locks_immediate": "Table_locks_immediate_rate",
	}
	for src, dst := range dupes {
		if v, ok := results[src]; ok {
			results[dst] = v
		}
	}

	c._submit_metrics(metrics, results)

	return nil
}

func (c *Check) _collect_replication_metrics(results mapinterface, above_560 bool) MetricItems {
	// Get replica stats
	results.update(c._get_replica_stats())
	results.update(c._get_replica_status(above_560))
	return REPLICA_VARS
}

func (c *Check) _check_replication_status(results mapinterface) {
	// get replica running form global status page
	replica_running_status := metrics.ServiceCheckUnknown
	// Replica_IO_Running: Whether the I/O thread for reading the source's binary log is running.
	// You want this to be Yes unless you have not yet started replication or have explicitly stopped it.
	replica_io_running, replica_io_running_ok := results.collectMap("Slave_IO_Running")
	if !replica_io_running_ok {
		replica_io_running, replica_io_running_ok = results.collectMap("Replica_IO_Running")
	}

	replica_sql_running, replica_sql_running_ok := results.collectMap("Slave_SQL_Running")
	if !replica_sql_running_ok {
		replica_sql_running, replica_sql_running_ok = results.collectMap("Replica_SQL_Running")
	}

	binlog_running := results.collectBool("Binlog_enabled")

	// replicas will only be collected if user has PROCESS privileges.

	replicas, replicas_ok := results.collectFloat("Slaves_connected")
	if !replicas_ok {
		replicas, replicas_ok = results.collectFloat("Replicas_connected")
	}

	if !(!replica_io_running_ok && !replica_sql_running_ok) {
		if len(replica_io_running) == 0 && len(replica_sql_running) == 0 {
			klog.V(6).Infof("Replica_IO_Running and Replica_SQL_Running are not ok")
			replica_running_status = metrics.ServiceCheckCritical
		} else if len(replica_io_running) == 0 || len(replica_sql_running) == 0 {
			klog.V(6).Infof("Either Replica_IO_Running or Replica_SQL_Running are not ok")
			replica_running_status = metrics.ServiceCheckWarning

		}
	}

	if replica_running_status == metrics.ServiceCheckUnknown {
		if c._is_source_host(replicas, results) { // master
			if replicas > 0 && binlog_running {
				klog.V(6).Infof("Host is master, there are replicas and binlog is running")
				replica_running_status = metrics.ServiceCheckOK
			} else {
				replica_running_status = metrics.ServiceCheckWarning
			}

		} else { // replica ( or standalone)
			if !(!replica_io_running_ok && !replica_sql_running_ok) {
				if len(replica_io_running) > 0 && len(replica_sql_running) > 0 {
					klog.V(6).Infof("Replica_IO_Running and Replica_SQL_Running are ok")
					replica_running_status = metrics.ServiceCheckOK
				}
			}

		}
	}

	// deprecated in favor of service_check("mysql.replication.slave_running")
	{
		v := 0
		if replica_running_status == metrics.ServiceCheckOK {
			v = 1
		}
		c.sender.Gauge(SLAVE_SERVICE_CHECK_NAME, float64(v), "", nil)
	}
	// deprecated in favor of service_check("mysql.replication.replica_running")
	c.sender.ServiceCheck(SLAVE_SERVICE_CHECK_NAME, replica_running_status, "", nil, "")
	c.sender.ServiceCheck(REPLICA_SERVICE_CHECK_NAME, replica_running_status, "", nil, "")
}

func (c *Check) _collect_statement_metrics() error {
	metrics := c._statement_metrics.collect_per_statement_metrics()
	for _, metric := range metrics {
		c.sender.Count(metric.name, metric.value, "", metric.tags)
	}
	return nil
}

func (c *Check) _is_source_host(replicas float64, results mapinterface) bool {
	// type: (float, Dict[str, Any]) -> bool
	// master uuid only collected in replicas
	source_host, ok := results.collectString("Master_Host")
	if !ok {
		source_host, _ = results.collectString("Source_Host")
	}
	if replicas > 0 || source_host == "" {
		return true
	}
	return false
}

func (c *Check) _submit_metrics(variables MetricItems, db_results mapinterface) {
	for variable, metric := range variables {
		if len(metric.Items) > 0 {
			for _, m := range metric.Items {
				c.__submit_metric(m.Metric, m.Type, variable, db_results)
			}
		} else {
			c.__submit_metric(metric.Metric, metric.Type, variable, db_results)
		}
	}
}

func (c *Check) __submit_metric(metric_name string, metric_type metrics.MetricType, variable string, db_results mapinterface) {
	v, ok := db_results[variable]
	if !ok {
		return
	}

	var values []interface{}
	var tags []string

	if m, ok := v.(map[string]interface{}); ok {
		for k2, v2 := range m {
			tags = append(tags, k2)
			values = append(values, v2)
		}
	} else {
		tags = append(tags, "")
		values = append(values, v)
	}

	for i, tag := range tags {
		var metric_tags []string
		if len(tag) > 0 {
			metric_tags = append(metric_tags, tag)
		}
		value := values[i]

		if value == nil {
			continue
		}

		switch metric_type {
		case metrics.RateType:
			c.sender.Rate(metric_name, Float(value), "", metric_tags)
		case metrics.GaugeType:
			c.sender.Gauge(metric_name, Float(value), "", metric_tags)
		case metrics.CountType:
			c.sender.Count(metric_name, Float(value), "", metric_tags)
		case metrics.MonotonicCountType:
			c.sender.MonotonicCount(metric_name, Float(value), "", metric_tags)
		}
	}
}

func (c *Check) _collect_system_metrics() error {
	pid := -1
	cfg := c.config
	if strings.Contains(cfg.server, "localhost") ||
		strings.Contains(cfg.server, "127.0.0.1") ||
		strings.Contains(cfg.server, "0.0.0.0") ||
		strings.Contains(cfg.port, "0") ||
		strings.Contains(cfg.port, "unix_socket") {
		pid = c._get_server_pid()
	}

	if pid > -1 {
		klog.V(6).Infof("System metrics for mysql w/ pid: %d", pid)
		// At last, get mysql cpu data out of psutil or procfs
		// github.com/shirou/gopsutil/
		proc, err := process.NewProcess(int32(pid))
		if err != nil {
			return err
		}

		stat, err := proc.Times()
		if err != nil {
			return err
		}

		c.sender.Rate("mysql.performance.user_time", stat.User, "", nil)
		// should really be system_time
		c.sender.Rate("mysql.performance.kernel_time", stat.System, "", nil)
		c.sender.Rate("mysql.performance.cpu_time", stat.User+stat.System, "", nil)
	}
	return nil
}

func (c *Check) _get_pid_file_variable() string {
	//  """
	//  Get the `pid_file` variable
	//  """
	var pid_file string
	_, err := c.queryRow("SHOW VARIABLES LIKE 'pid_file'", new(string), &pid_file)
	if err != nil {
		klog.Warning("Error while fetching pid_file variable of MySQL.")
	}

	return pid_file
}

func (c *Check) _get_server_pid() int {
	pid := -1

	// Try to get pid from pid file, it can fail for permission reason
	pid_file := c._get_pid_file_variable()
	if pid_file != "" {
		klog.V(6).Infof("pid file: %s", pid_file)
		b, err := ioutil.ReadFile(pid_file)
		if err != nil {
			klog.Warning("read pidfile %s err %s", pid_file, err)
			return -1
		}
		pid, err = strconv.Atoi(strings.TrimSpace(string(b)))
		if err != nil {
			klog.Warningf("read pid file %v err %v", strings.TrimSpace(string(b)), err)
		}
	}

	// If pid has not been found, read it from ps
	// TODO
	//if pid == -1 {
	//    for proc in psutil.process_iter():
	//        try:
	//            if proc.name() == PROC_NAME:
	//                pid = proc.pid
	//        except (psutil.AccessDenied, psutil.ZombieProcess, psutil.NoSuchProcess):
	//            continue
	//        except Exception:
	//            self.log.exception("Error while fetching mysql pid from psutil")
	//}

	return pid
}

func (c *Check) _get_stats_from_status() mapinterface {
	results, err := c.queryKv("SHOW /*!50002 GLOBAL */ STATUS;")
	if err != nil {
		klog.Errorf("_get_stats_from_status err %s", err)
	}

	return results
}

func (c *Check) _get_stats_from_variables() map[string]interface{} {
	results, err := c.queryKv("SHOW GLOBAL VARIABLES")
	if err != nil {
		klog.Errorf("_get_stats_from_variables err %s", err)
	}

	return results
}

func (c *Check) _get_binary_log_stats() float64 {
	rows, err := c.queryKv("SHOW BINARY LOGS;")
	if err != nil {
		klog.Errorf("show binary logs err %s", err)
		return 0
	}

	binary_log_space := float64(0)
	for _, v := range rows {
		binary_log_space += Float(v)
	}
	return binary_log_space
}

func (c *Check) _is_innodb_engine_enabled() bool {
	rows, err := c.queryRows(SQL_INNODB_ENGINES)
	if err != nil {
		klog.Warningf("Possibly innodb stats unavailable - error querying engines table: %s", err)
		return false
	}

	return len(rows) > 0
}

func (c *Check) _get_replica_stats() mapinterface {
	ret := make(mapinterface)
	is_mariadb := c.version.flavor == "MariaDB"
	replication_channel := c.config.Options.ReplicationChannel

	if is_mariadb && replication_channel != "" {
		_, err := c.db.Exec("SET @@default_master_connection = '?';", replication_channel)
		if err != nil {
			klog.Warningf("get replica stats err %s", err)
			return nil
		}
	}

	results, err := c.queryMapRows(showReplicaStatusQuery(c.version, is_mariadb, replication_channel))
	if err != nil {
		klog.Warningf("Privileges error getting replication status (must grant REPLICATION CLIENT): %s", err)
		return nil
	}

	klog.V(6).Infof("Getting replication status: %s", results)

	for _, result := range results {
		// MySQL <5.7 does not have Channel_Name.
		// For MySQL >=5.7 'Channel_Name' is set to an empty string by default
		channel := "default"
		if replication_channel != "" {
			channel = replication_channel
		}
		if result["Channel_Name"] != "" {
			channel = String(result["Channel_Name"])
		}
		for key, value := range result {
			ret[key] = map[string]interface{}{"channel:" + channel: value}
		}
	}

	if row, _ := c.queryRow("SHOW MASTER STATUS;"); row != nil {
		ret["Binlog_enabled"] = true
	}

	return ret
}

// Retrieve the replicas statuses using:
// 1. The `performance_schema.threads` table. Non-blocking, requires version > 5.6.0
// 2. The `information_schema.processlist` table. Blocking
func (c *Check) _get_replica_status(above_560 bool) mapinterface {
	ret := make(mapinterface)

	var rows [][]interface{}
	var err error
	if above_560 && c.config.Options.ReplicationNonBlockingStatus {
		rows, err = c.queryRows(SQL_WORKER_THREADS)
	} else {
		rows, err = c.queryRows(SQL_PROCESS_LIST)
	}
	if err != nil {
		klog.Warningf("Privileges error accessing the process tables (must grant PROCESS): %s", err)
	}
	ret["Replicas_connected"] = float64(len(rows))
	return ret
}

func (c *Check) _get_variable_enabled(results map[string]interface{}, k string) bool {
	return strings.TrimSpace(strings.ToLower(String(results[k]))) == "on"
}

func (c *Check) _get_query_exec_time_95th_us() interface{} {
	// Fetches the 95th percentile query execution time and returns the value
	// in microseconds
	var query_exec_time_95th_per float64
	_, err := c.queryRow(SQL_95TH_PERCENTILE, new(float64), &query_exec_time_95th_per)
	if err != nil {
		klog.Warningf("Failed to fetch records from the perf schema 'events_statements_summary_by_digest' table.")
		return nil
	}
	return &query_exec_time_95th_per
}
func (c *Check) _query_exec_time_per_schema() interface{} {
	// Fetches the avg query execution time per schema and returns the
	// value in microseconds
	rows, err := c.queryRows(SQL_AVG_QUERY_RUN_TIME)
	if err != nil {
		klog.Warningf("Avg exec time performance metrics unavailable at this time %s", err)
		return nil
	}

	schema_query_avg_run_time := map[string]interface{}{}
	for _, row := range rows {
		// set the tag as the dictionary key
		schema_query_avg_run_time["schema:"+*row[0].(*string)] = row[1]
	}

	return schema_query_avg_run_time
}

func (c *Check) _query_size_per_schema() interface{} {
	rows, err := c.queryRows(SQL_QUERY_SCHEMA_SIZE)
	if len(rows) < 1 {
		klog.Warning("Failed to fetch records from the information schema 'tables' table.")
		return nil
	}
	if err != nil {
		klog.Warning("Avg exec time performance metrics unavailable at this time: %s", err)
		return nil
	}
	schema_size := map[string]interface{}{}
	for _, row := range rows {
		schema_size["schema:"+String(row[0])] = row[1]
	}

	return schema_size
}

func (c *Check) _compute_synthetic_results(results mapinterface) {
	var hits, inserts, notCached int64
	var ok bool

	if hits, ok = results.collectInt("Qcache_hits"); !ok {
		return
	}
	if inserts, ok = results.collectInt("Qcache_inserts"); !ok {
		return
	}
	if notCached, ok = results.collectInt("Qcache_not_cached"); !ok {
		return
	}

	if hits == 0 {
		results["Qcache_utilization"] = float64(0)
	} else {
		results["Qcache_utilization"] = float64(hits) / float64(inserts+notCached+hits) * 100
	}

	if c._qcache_hits > 0 || c._qcache_inserts > 0 || c._qcache_not_cached > 0 {
	}

	if hits-c._qcache_hits == 0 {
		results["Qcache_instant_utilization"] = float64(0)
	} else {
		top := float64(hits - c._qcache_hits)
		bottom := float64((inserts - c._qcache_inserts) + (notCached - c._qcache_not_cached) + (hits - c._qcache_hits))
		results["Qcache_instant_utilization"] = (top / bottom) * 100
	}

	// update all three, or none - for consistent samples.
	c._qcache_hits = hits
	c._qcache_inserts = inserts
	c._qcache_not_cached = notCached
}
