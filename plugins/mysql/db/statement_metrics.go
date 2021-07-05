package db

import (
	"sort"

	"k8s.io/klog/v2"
)

// This class supports normalized statement-level metrics, which are collected from the database's
// statistics tables, ex:
//
//     - Postgres: pg_stat_statements
//     - MySQL: performance_schema.events_statements_summary_by_digest
//     - Oracle: V$SQLAREA
//     - SQL Server: sys.dm_exec_query_stats
//     - DB2: mon_db_summary
//
// These tables are monotonically increasing, so the metrics are computed from the difference
// in values between check runs.
type StatementMetrics struct {
	previousStatements map[string]map[string]interface{}
}

// Compute the first derivative of column-based metrics for a given set of rows. This function
// takes the difference of the previous check run's values and the current check run's values
// to produce a new set of rows whose values represent the total counts in the time elapsed
// between check runs.
//
// This differs from `AgentCheck.monotonic_count` in that state for the entire row is kept,
// regardless of whether or not the tags used to uniquely identify the row are submitted as
// metric tags. There is also custom logic around stats resets to discard all rows when a
// negative value is found, rather than just the single metric of that row/column.
//
// This function resets the statement cache so it should only be called once per check run.
//
// - **rows** (_List[dict]_) - rows from current check run
// - **metrics** (_List[str]_) - the metrics to compute for each row
// - **key** (_callable_) - function for an ID which uniquely identifies a row across runs
func (p *StatementMetrics) computeDerivativeRows(rows []map[string]interface{}, metrics map[string]string, key func(map[string]interface{}) string) []map[string]interface{} {
	newCache := map[string]map[string]interface{}{}
	result := []map[string]interface{}{}

	for _, row := range rows {
		rowKey := key(row)
		if _, ok := newCache[rowKey]; ok {
			klog.V(5).Infof("Collision in cached query metrics. Dropping existing row, row_key=%s new=%+v dropped=%+v", rowKey, row, newCache[rowKey])
		}

		newCache[rowKey] = row

		prev, ok := p.previousStatements[rowKey]
		if !ok {
			continue
		}

		var sum float64
		diffedRow := map[string]interface{}{}
		for k, v := range row { // prev: 0
			dv := itof64(v)

			if _, ok := metrics[k]; ok {
				dv = dv - itof64(prev[k])
			}

			if dv < 0 {
				goto NEXT_ROW
			}

			sum += dv
			diffedRow[k] = dv
		}

		// No changes to the query; no metric needed
		if sum == 0 {
			continue
		}

		result = append(result, diffedRow)
	NEXT_ROW:
	}

	p.previousStatements = newCache

	return result

}

/*
   Given a list of query rows, apply limits ensuring that the top K and bottom K of each metric (columns)
   are present. To increase the overlap of rows across metics with the same values (such as 0), the tiebreaker metric
   is used as a second sort dimension.

   The reason for this custom limit function on metrics is to guarantee that metric `top()` functions show the true
   top and true bottom K, even if some limits are applied to drop less interesting queries that fall in the middle.

   Longer Explanation of the Algorithm
   -----------------------------------

   Simply taking the top K and bottom K of all metrics is insufficient. For instance, for K=2 you might have rows
   with values:

       | query               | count      | time        | errors      |
       | --------------------|------------|-------------|-------------|
       | select * from dogs  | 1 (bottom) | 10 (top)    |  1 (top)    |
       | delete from dogs    | 2 (bottom) |  8 (top)    |  0 (top)    |
       | commit              | 3          |  7          |  0 (bottom) |
       | rollback            | 4          |  3          |  0 (bottom) |
       | select now()        | 5 (top)    |  2 (bottom) |  0          |
       | begin               | 6 (top)    |  2 (bottom) |  0          |

   If you only take the top 2 and bottom 2 values of each column and submit those metrics, then each query is
   missing a lot of metrics:

       | query               | count      | time        | errors      |
       | --------------------|------------|-------------|-------------|
       | select * from dogs  | 1          | 10          |  1          |
       | delete from dogs    | 2          |  8          |  0          |
       | commit              |            |             |  0          |
       | rollback            |            |             |  0          |
       | select now()        | 5          |  2          |             |
       | begin               | 6          |  2          |             |

   This is fine for showing only one metric, but if the user copies the query tag to find our more information,
   that query should have all of the metrics because it is an "interesting" query.

   To solve that, you can submit all metrics for all rows with at least on metric submitted, but then the worst-case
   for total cardinality is:

       (top K + bottom K) * metric count

   Note that this only applies to one check run and a completely different set of "tied" metrics can be submitted on
   the next check run. Since a large number of rows will have value '0', a tiebreaker is used to bias the selected
   rows to rows already picked in the top K / bottom K for the tiebreaker.


       | query               | count      | time        | errors      |
       | --------------------|------------|-------------|-------------|
       | select * from dogs  | 1          | 10          |  1          |
       | delete from dogs    | 2          |  8          |  0          |
       | commit              |            |             |             |
       | rollback            |            |             |             |
       | select now()        | 5          |  2          |  0          | <-- biased toward top K count
       | begin               | 6          |  2          |  0          | <-- biased toward top K count

   The queries `commit` and `rollback` were not interesting to keep; they were only selected because they have error
   counts 0 (but so do the other queries). So we use the `count` as a tiebreaker to instead choose queries which are
   interesting because they have higher execution counts.

   - **rows** (_List[dict]_) - rows with columns as metrics
   - **metric_limits** (_Dict[str,Tuple[int,int]]_) - dict of the top k and bottom k limits for each metric
           ex:
           >>> metric_limits = {
           >>>     'count': (200, 50),
           >>>     'time': (200, 100),
           >>>     'lock_time': (50, 50),
           >>>     ...
           >>>     'rows_sent': (100, 0),
           >>> }

           The first item in each tuple guarantees the top K rows will be chosen for this metric. The second item
           guarantees the bottom K rows will also be chosen. Both of these numbers are configurable because you
           may want to keep the top 100 slowest queries, but are only interested in the top 10 fastest queries.
           That configuration would look like:

           >>> metric_limits = {
           >>>     'time': (100, 10),  # Top 100, bottom 10
           >>>     ...
           >>> }

   - **tiebreaker_metric** (_str_) - metric used to resolve ties, intended to increase row overlap in different metrics
   - **tiebreaker_reverse** (_bool_) - whether the tiebreaker metric should be in reverse order (descending)
   - **key** (_callable_) - function for an ID which uniquely identifies a row
*/
func ApplyRowLimits(
	rows []map[string]interface{},
	metricLimits map[string][2]int,
	tiebreakerMetric string,
	tiebreakerReverse bool,
	key func(map[string]interface{}) string) []map[string]interface{} {

	if len(rows) == 0 {
		return rows
	}

	sorter := orderedBy(rows, tiebreakerMetric, tiebreakerReverse)

	sample := rows[0]
	limited := map[string]map[string]interface{}{}
	for metric, limit := range metricLimits {
		if _, ok := sample[metric]; !ok {
			continue
		}

		sorter.Sort(metric)

		// top
		for _, row := range rows[len(rows)-limit[0]:] {
			limited[key(row)] = row
		}

		// bottom
		for _, row := range rows[:limit[1]] {
			limited[key(row)] = row
		}
	}

	ret := make([]map[string]interface{}, 0, len(limited))
	for _, v := range limited {
		ret = append(ret, v)
	}
	return ret
}

func orderedBy(rows []map[string]interface{}, key2 string, reverse bool) *rowLimitsSorter {
	return &rowLimitsSorter{
		key2:       key2,
		key2Revert: reverse,
		data:       rows,
	}
}

type rowLimitsSorter struct {
	key        string
	key2       string
	key2Revert bool
	data       []map[string]interface{}
}

func (p *rowLimitsSorter) Swap(i, j int) {
	p.data[i], p.data[j] = p.data[j], p.data[i]
}
func (p *rowLimitsSorter) Len() int {
	return len(p.data)
}

func (p *rowLimitsSorter) Less(i, j int) bool {
	i1, j1 := itof64(p.data[i][p.key]), itof64(p.data[j][p.key])
	if i1 == j1 {
		i2, j2 := itof64(p.data[i][p.key2]), itof64(p.data[j][p.key2])
		if p.key2Revert {
			return i2 > j2
		}
		return i2 < j2
	}
	return i1 < j1
}

func (p *rowLimitsSorter) Sort(metric string) {
	p.key = metric
	sort.Sort(p)
}

func itof64(i interface{}) float64 {
	if i == nil {
		return 0
	}
	return *i.(*float64)
}
