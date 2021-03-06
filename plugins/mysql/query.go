package mysql

import "fmt"

const (
	SQL_95TH_PERCENTILE = "SELECT `avg_us`, `ro` as `percentile` FROM " +
		" (SELECT `avg_us`, @rownum := @rownum + 1 as `ro` FROM " +
		"     (SELECT ROUND(avg_timer_wait / 1000000) as `avg_us` " +
		"         FROM performance_schema.events_statements_summary_by_digest " +
		"         ORDER BY `avg_us` ASC) p, " +
		"     (SELECT @rownum := 0) r) q " +
		" WHERE q.`ro` > ROUND(.95*@rownum) " +
		" ORDER BY `percentile` ASC " +
		" LIMIT 1"

	SQL_QUERY_SCHEMA_SIZE = `
SELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb
FROM     information_schema.tables
GROUP BY table_schema`

	SQL_AVG_QUERY_RUN_TIME = `
SELECT schema_name, ROUND((SUM(sum_timer_wait) / SUM(count_star)) / 1000000) AS avg_us
FROM performance_schema.events_statements_summary_by_digest
WHERE schema_name IS NOT NULL
GROUP BY schema_name`

	SQL_WORKER_THREADS = "SELECT THREAD_ID, NAME FROM performance_schema.threads WHERE NAME LIKE '%worker'"

	SQL_PROCESS_LIST = "SELECT * FROM INFORMATION_SCHEMA.PROCESSLIST WHERE COMMAND LIKE '%Binlog dump%'"

	SQL_INNODB_ENGINES = `
SELECT engine
FROM information_schema.ENGINES
WHERE engine='InnoDB' and support != 'no' and support != 'disabled'`
)

func showReplicaStatusQuery(version MySQLVersion, is_mariadb bool, channel string) string {
	var base_query string
	if version.versionCompatible(10, 5, 1) || (!is_mariadb && version.versionCompatible(8, 0, 22)) {
		base_query = "SHOW REPLICA STATUS"
	} else {
		base_query = "SHOW SLAVE STATUS"
	}
	if channel != "" && !is_mariadb {
		return fmt.Sprintf("%s FOR CHANNEL '%s';", base_query, channel)
	} else {
		return fmt.Sprintf("%s;", base_query)
	}
}
