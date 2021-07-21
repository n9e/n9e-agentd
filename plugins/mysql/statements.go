package mysql

import (
	"fmt"
	"strings"

	"github.com/n9e/n9e-agentd/pkg/util/db"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/obfuscate"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/pb"
	"k8s.io/klog/v2"
)

var (
	STATEMENT_METRICS = map[string]string{
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

	// These limits define the top K and bottom K unique query rows for each metric. For each check run the
	// max metrics sent will be sum of all numbers below (in practice, much less due to overlap in rows).
	DEFAULT_STATEMENT_METRICS_LIMITS = map[string][2]int{
		"count":              {400, 0},
		"errors":             {100, 0},
		"time":               {400, 0},
		"select_scan":        {50, 0},
		"select_full_join":   {50, 0},
		"no_index_used":      {50, 0},
		"no_good_index_used": {50, 0},
		"lock_time":          {50, 0},
		"rows_affected":      {100, 0},
		"rows_sent":          {100, 0},
		"rows_examined":      {100, 0},
		// Synthetic column limits
		"avg_time":        {400, 0},
		"rows_sent_ratio": {0, 50},
	}
)

func generate_synthetic_rows(rows []map[string]interface{}) []map[string]interface{} {
	// type: (List[PyMysqlRow]) -> List[PyMysqlRow]
	// Given a list of rows, generate a new list of rows with "synthetic" column values derived from
	// the existing row values.
	synthetic_rows := []map[string]interface{}{}
	for _, row := range rows {
		if count := Float(row["count"]); count > 0 {
			row["avg_time"] = Float(row["time"]) / count
		} else {
			row["avg_time"] = float64(0)
		}
		if examined := Float(row["rows_examined"]); examined > 0 {
			row["rows_sent_ratio"] = Float(row["rows_sent"]) / examined
		} else {
			row["rows_sent_ratio"] = float64(0)
		}

		synthetic_rows = append(synthetic_rows, row)
	}

	return synthetic_rows
}

func NewMySQLStatementMetrics(c *Check) *MySQLStatementMetrics {
	return &MySQLStatementMetrics{
		Check:      c,
		state:      db.NewMySQLStatementMetrics(),
		obfuscator: obfuscate.NewObfuscator(nil),
	}
}

type MySQLStatementMetrics struct {
	*Check
	state      *db.StatementMetrics
	obfuscator *obfuscate.Obfuscator
}

type statementMetric struct {
	name  string
	value float64
	tags  []string
}

func (p *MySQLStatementMetrics) collect_per_statement_metrics() []statementMetric {
	metrics := []statementMetric{}

	keyfunc := func(row map[string]interface{}) string {
		return (string(row["schema"].([]byte)) + string(row["digest"].([]byte)))
	}

	monotonic_rows := p._query_summary_per_statement()
	monotonic_rows = p._merge_duplicate_rows(monotonic_rows, keyfunc)
	rows := p.state.ComputeDerivativeRows(monotonic_rows, STATEMENT_METRICS, keyfunc)
	metrics = append(metrics, statementMetric{
		name:  "mysql.queries.query_rows_raw",
		value: float64(len(rows)),
	})

	rows = generate_synthetic_rows(rows)

	rows = db.ApplyRowLimits(
		rows,
		p.config.StatementMetricsLimits,
		"count",
		true,
		keyfunc,
	)
	metrics = append(metrics, statementMetric{
		name:  "mysql.queries.query_rows_limited",
		value: float64(len(rows)),
	})

	for _, row := range rows {
		var tags []string
		tags = append(tags, "digest:"+String(row["digest"]))

		if schema := String(row["schema"]); len(schema) > 0 {
			tags = append(tags, "schema:"+schema)
		}

		span := &pb.Span{
			Type:     "sql",
			Resource: String(row["query"]),
		}
		p.obfuscator.Obfuscate(span)
		obfuscated_statement := span.Resource
		tags = append(tags, fmt.Sprintf("query_signature:%x", hash64(obfuscated_statement)))
		tags = append(tags, "query_signature:"+strings.TrimSpace(obfuscated_statement))

		for col, name := range STATEMENT_METRICS {
			metrics = append(metrics, statementMetric{
				name:  name,
				value: Float(row[col]),
				tags:  tags,
			})
		}
	}

	return metrics
}

func (p *MySQLStatementMetrics) _merge_duplicate_rows(rows []map[string]interface{}, key func(map[string]interface{}) string) []map[string]interface{} {
	//# type: (List[PyMysqlRow], RowKeyFunction) -> List[PyMysqlRow]
	//Merges the metrics from duplicate rows because the (schema, digest) identifier may not be
	//unique, see: https://bugs.mysql.com/bug.php?id=79533
	merged := make(map[string]map[string]interface{})

	for i, row := range rows {
		k := key(row)
		if _, ok := merged[k]; ok {
			for k2, _ := range STATEMENT_METRICS {
				merged[k][k2] = Int(merged[k][k2]) + Int(row[k2])
			}
		} else {
			merged[k] = rows[i]
		}
		/*
		   if k in merged:
		       for m in STATEMENT_METRICS:
		           merged[k][m] += row[m]
		   else:
		       merged[k] = copy.copy(row)
		*/
	}

	out := make([]map[string]interface{}, 0, len(merged))
	for _, v := range merged {
		out = append(out, v)
	}

	return out
}

func (p *MySQLStatementMetrics) _query_summary_per_statement() []map[string]interface{} {
	// # type: (pymysql.connections.Connection) -> List[PyMysqlRow]
	// """
	// Collects per-statement metrics from performance schema. Because the statement sums are
	// cumulative, the results of the previous run are stored and subtracted from the current
	// values to get the counts for the elapsed period. This is similar to monotonic_count, but
	// several fields must be further processed from the delta values.
	// """

	sql_statement_summary := "SELECT `schema_name` as `schema`, " +
		"`digest` as `digest`, " +
		"`digest_text` as `query`, " +
		"`count_star` as `count`, " +
		"`sum_timer_wait` / 1000 as `time`, " +
		"`sum_lock_time` / 1000 as `lock_time`, " +
		"`sum_errors` as `errors`, " +
		"`sum_rows_affected` as `rows_affected`, " +
		"`sum_rows_sent` as `rows_sent`, " +
		"`sum_rows_examined` as `rows_examined`, " +
		"`sum_select_scan` as `select_scan`, " +
		"`sum_select_full_join` as `select_full_join`, " +
		"`sum_no_index_used` as `no_index_used`, " +
		"`sum_no_good_index_used` as `no_good_index_used` " +
		"FROM performance_schema.events_statements_summary_by_digest " +
		"WHERE `digest_text` NOT LIKE 'EXPLAIN %' " +
		"ORDER BY `count_star` DESC " +
		"LIMIT 10000"

	rows, err := p.queryMapRows(sql_statement_summary)
	if err != nil {
		klog.Warningf("sql statement summary error %s", err)
	}
	return rows
}
