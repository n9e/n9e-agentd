package db

import (
	"fmt"
	"reflect"

	"github.com/n9e/n9e-agentd/pkg/aggregator"
	"k8s.io/klog/v2"
)

// - query: SELECT foo, COUNT(*) FROM table.events GROUP BY foo
//   columns:
//   - name: foo
//     type: tag
//   - name: event.total
//     type: gauge
//   tags:
//   - test:mysql
type CustomQuery struct {
	Name         string                   `json:"-"`
	MetricPrefix string                   `json:"metric_prefix"`
	Query        string                   `json:"query"`
	Columns      []map[string]interface{} `json:"columns"`
	Extras       []map[string]interface{} `json:"extras"`
	Tags         []string                 `json:"tags"`
}

//type GlobalCustomQuerie struct {
//	Query   string                   `json:"query"`
//	Columns []map[string]interface{} `json:"columns"`
//	Tags    []string                 `json:"tags"`
//}

type AgentCheck interface {
	CustomQueries() []CustomQuery
	GlobalCustomQueries() []CustomQuery
	UseGlobalCustomQueries() string
	Sender() aggregator.Sender
	Executor(query string) ([][]interface{}, error)
}

type QueryManager struct {
	check   AgentCheck
	queries []*Query
	tags    []string // ?
}

//   - **check** (_AgentCheck_) - an instance of a Check
//     a sequence representing either the full result set or an iterator over the result set
//   - **queries** (_List[Query]_) - a list of `Query` instances
func NewQueryManager(c AgentCheck, queries []*CustomQuery, tags []string) (*QueryManager, error) {
	p := &QueryManager{
		check: c,
		tags:  tags,
	}

	for _, q := range queries {
		p.queries = append(p.queries, NewQuery(q))
	}

	custom_queries := c.CustomQueries()
	use_global_custom_queries := c.UseGlobalCustomQueries()

	// Handle overrides
	if use_global_custom_queries == "extend" || (len(custom_queries) == 0 && len(c.GlobalCustomQueries()) > 0 && use_global_custom_queries == "true") {
		custom_queries = append(custom_queries, c.GlobalCustomQueries()...)
	}

	// Deduplicate
	for i, q := range custom_queries {
		query := NewQuery(&q)
		query.query_data.Name = fmt.Sprintf("custom query #%d", i)
		p.queries = append(p.queries, query)
	}

	return p, nil
}

// This method compiles every `Query` object.
func (p *QueryManager) Compile_queries() error {
	// This method compiles every `Query` object.
	column_transformers := make(mapinterface)
	for k, v := range COLUMN_TRANSFORMERS {
		column_transformers[k] = v
	}

	sender := reflect.ValueOf(p.check.Sender())

	for submission_method, transformer_name := range SUBMISSION_METHODS {
		// Save each method in the initializer -> callable format
		column_transformers[transformer_name] = create_submission_transformer(sender.MethodByName(submission_method).Interface())
	}

	for _, query := range p.queries {
		err := query.compile(column_transformers, EXTRA_TRANSFORMERS)
		if err != nil {
			return fmt.Errorf("query.compile err %s", err)
		}
	}

	return nil

}
func (p *QueryManager) Execute(extra_tags []string) {
	// This method is what you call every check run."""
	global_tags := append(p.tags, extra_tags...)

	for _, query := range p.queries {
		query_name := query.name
		query_columns := query.columns
		query_extras := query.extras
		num_columns := len(query_columns)
		query_tags := query.tags

		rows, err := p.execute_query(query.query)
		if err != nil {
			klog.Errorf("Error querying %s: %s", query_name, err)
			continue
		}

		for _, row := range rows {
			if num_columns != len(row) {
				klog.Errorf("Query %s expected %d column, got %d", query_name, num_columns, len(row))
			}

			var tags []string
			tags = append(tags, global_tags...)
			tags = append(tags, query_tags...)

			sources := make(mapinterface)
			submission_queue := []submission_item{}

			for i := 0; i < num_columns; i++ {
				column_name := query_columns[i].column_name
				column_type := query_columns[i].column_type
				value := row[i]

				// Columns can be ignored via configuration
				if column_name == "" {
					continue
				}

				sources[column_name] = value

				// The transformer can be None for `source` types. Those such columns do not submit
				// anything but are collected into the row values for other columns to reference.

				transformer := query_columns[i].transformer
				if transformer == nil {
					continue
				}
				if column_type == "tag" {
					ret, err := transformer(nil, value, nil)
					if err != nil {
						klog.ErrorS(err, "transformer",
							"column_name", column_name,
							"column_type", column_type)
						continue
					}
					tags = append(tags, ret.(string))

				} else if column_type == "tag_list" {
					ret, err := transformer(nil, value, nil)
					if err != nil {
						klog.ErrorS(err, "transformer",
							"column_name", column_name,
							"column_type", column_type)
						continue
					}
					tags = append(tags, ret.([]string)...)

				} else {
					submission_queue = append(submission_queue,
						submission_item{transformer, value, column_name, column_type})
				}
			}

			for _, v := range submission_queue {
				_, err := v.transformer(sources, v.value, mapinterface{"tags": tags})
				if err != nil {
					klog.ErrorS(err, "submission_transformer",
						"column_name", v.column_name,
						"column_type", v.column_type)
				}
			}

			for _, v := range query_extras {
				result, err := v.transformer(sources, mapinterface{"tags": tags}, nil)
				if err != nil {
					klog.ErrorS(err, "extras_transformer",
						"extra_name", v.extra_name)
					continue
				}

				if result != nil {
					sources[v.extra_name] = result
				}
			}
		}
	}
}

func (p *QueryManager) execute_query(query string) ([][]interface{}, error) {
	// Called by `execute`, this triggers query execution to check for errors immediately in a way that is compatible
	// with any library. If there are no errors, this is guaranteed to return an iterator over the result set.
	return p.check.Executor(query)
}
