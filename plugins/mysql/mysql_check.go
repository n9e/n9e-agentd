package mysql

import (
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/n9e/n9e-agentd/plugins/mysql/db"
	"k8s.io/klog/v2"
)

func (c *Check) init() error {
	tlsConfig, err := c.config.TLS.TLSConfig()
	if err != nil {
		return fmt.Errorf("registering TLS config: %s", err)
	}

	if tlsConfig != nil {
		if err := mysql.RegisterTLSConfig(string(c.ID()), tlsConfig); err != nil {
			return err
		}
	}

	c.queryManager = db.NewQueryManager(db.QueryManagerOptions{})
	c.innodbStats = NewInnoDBMetrics(c.db)
	c.statementMetrics = NewMySQLStatementMetrics()
	c.statementSamples = NewMySQLStatementSamples()

	return nil
}

func (c *Check) cancel() {
	if tlsConfig, _ := c.config.TLS.TLSConfig(); tlsConfig != nil {
		mysql.DeregisterTLSConfig(string(c.ID()))
	}
	c.statementSamples.Cancel()
}

func (c *Check) check() error {

	// version collection
	if err := c.getVersion(); err != nil {
		klog.Error("get mysql version %s", err)
	} else {
		c.sendMetadata()
	}

	// Metric collection
	c.collectMetrics()
	c.collectSystemMetrics()
	if c.config.DeepDatabaseMonitoring {
		c.collectStatementMetrics()
		c.statementSamples.runSampler()
	}

	// keeping track of these:
	c.putQcacheStats()

	//Custom queries
	c.queryManager.Execute()
	return nil
}

func (c *Check) sendMetadata() error {
	klog.V(5).Infof("version %s %s", c.version.version, c.version.build)
	klog.V(5).Infof("flavor %s", c.version.flavor)
	return nil
}
func (c *Check) collectMetrics() error {
	return nil
}
func (c *Check) collectSystemMetrics() error {
	return nil
}
func (c *Check) collectStatementMetrics() error {
	return nil
}
func (c *Check) runSampler() error {
	return nil
}
func (c *Check) putQcacheStats() error {
	return nil
}
func (c *Check) queryManagerExecute() error {
	return nil
}
func (c *Check) statementSampleCancel() error {
	return nil
}

var STATEMENT_METRICS = map[string]string{
	"count":              "mysql.queries.count",
	"errors":             "mysql.queries.errors",
	"time":               "mysql.queries.time",
	"select_scan":        "mysql.queries.select_scan",
	"select_full_join":   "mysql.queries.select_full_join",
	"no_index_used":      "mysql.queries.no_index_used",
	"no_good_index_used": "mysql.queries.no_good_index_used",
	"lock_time":          "mysql.queries.lock_time",
	"rows_affected":      "mysql.queries.rows_affected",
	"rows_sent":          "mysql.queries.rows_sent",
	"rows_examined":      "mysql.queries.rows_examined",
}

type StatementMetrics struct {
	*Check
}

func (p *StatementMetrics) collectPerStatementMetrics() {
	//monotonicRows, err := p.querySummaryPerStatement()
	//if err != nil {
	//	klog.Errorf("query summary per statement err %s", err)
	//	return
	//}
	//monotonicRow = p.mergeDuplicateRows(monotonicRows, key=keyfunc)
	//rows := p.computeDerivativeRows(monotonicRows, )

	//p.sender.Count("dd.mysql.queries.query_rows_raw", 0, nil)
	//p.sender.Count("dd.mysql.queries.query_rows_raw", float64(len(rows)), "", nil)
}

func (p *StatementMetrics) querySummaryPerStatement() ([]map[string]interface{}, error) {
	return p.queryRows(`
	SELECT ifnull(SCHEMA_NAME, 'NONE') as 'schema',
		DIGEST as 'digest',
		DIGEST_TEXT as 'query',
		COUNT_STAR as 'count',
		SUM_TIMER_WAIT / 1000 as 'time',
		SUM_LOCK_TIME / 1000 as 'lock_time',
		SUM_ERRORS as 'errors',
		SUM_ROWS_AFFECTED as 'rows_affected',
		SUM_ROWS_SENT as 'rows_sent',
		SUM_ROWS_EXAMINED as 'rows_examined',
		SUM_SELECT_SCAN as 'select_scan',
		SUM_SELECT_FULL_JOIN as 'select_full_join',
		SUM_NO_INDEX_USED as 'no_index_used',
		SUM_NO_GOOD_INDEX_USED as 'no_good_index_used'
	FROM performance_schema.events_statements_summary_by_digest
	WHERE 'digest_text' NOT LIKE 'EXPLAIN %'
	ORDER BY 'count_star' DESC
	LIMIT 10000`, func() []interface{} {
		return []interface{}{
			new(string), new(string), new(string), new(float64), new(float64),
			new(float64), new(float64), new(float64), new(float64), new(float64),
			new(float64), new(float64), new(float64), new(float64),
		}
	})
}

func (p *StatementMetrics) mergeDuplicateRows() {
}
func (p *StatementMetrics) computeDerivativeRows() {
}

func (c *Check) queryRow(sql string, factory func() []interface{}) (map[string]interface{}, error) {
	return db.QueryRow(c.db, sql, factory)
}

func (c *Check) queryRows(sql string, factory func() []interface{}) ([]map[string]interface{}, error) {
	return db.QueryRows(c.db, sql, factory)
}
