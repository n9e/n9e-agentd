package db

import (
	"database/sql"
	"fmt"

	"k8s.io/klog/v2"
)

type Query struct {
	query_data *CustomQuery
	name       string
	query      string
	columns    []*columnItem
	extras     []*extraItem
	tags       []string
}

type columnItem struct {
	column_name string
	column_type string
	transformer transformHandle
}

type sourceItem struct {
	Type  string
	Index int
}

type extraItem struct {
	extra_name  string
	transformer transformHandle
}

func NewQuery(q *CustomQuery) *Query {
	return &Query{query_data: q}
}

func (p *Query) compile(column_transformers, extra_transformers mapinterface) error {
	// Check for previous compilation
	if p.name != "" {
		return nil
	}

	query_name := p.query_data.Name
	if query_name == "" {
		return fmt.Errorf("query field `name` is required")
	}

	query := p.query_data.Query
	if query == "" {
		return fmt.Errorf("field `query` for %s is required", query_name)
	}

	columns := p.query_data.Columns
	if len(columns) == 0 {
		return fmt.Errorf("field `columns` for %s is required", query_name)
	}

	tags := p.query_data.Tags

	// Keep track of all defined names
	sources := map[string]*sourceItem{}

	column_data := []*columnItem{}
	for i, column := range columns {
		if len(column) == 0 {
			continue
		}

		column_name, ok := column["name"].(string)
		if !ok {
			return fmt.Errorf("field `name` for column[%d]  of %s is required", i, query_name)
		}

		if s, ok := sources[column_name]; ok {
			return fmt.Errorf("the name %s of %s was alread defined in %s [%d]",
				column_name, query_name, s.Type, s.Index)
		}

		sources[column_name] = &sourceItem{Type: "column", Index: i}

		column_type, ok := column["type"].(string)
		if !ok {
			return fmt.Errorf("field `type` for column %s of %s is required", column_name, query_name)
		}

		if column_type == "source" {
			column_data = append(column_data, &columnItem{
				column_name: column_name,
			})
			continue
		}

		column_transformer, err := _transformer(column_transformers[column_type])
		if err != nil {
			return fmt.Errorf("unknown type `%s` for column %s of %s err %s",
				column_type, column_name, p.name, err)
		}

		modifiers := make(mapinterface)
		for k, v := range column {
			if k != "name" && k != "type" {
				modifiers[k] = v
			}
		}

		transformer, err := column_transformer(column_transformers, column_name, modifiers)
		if err != nil {
			klog.ErrorS(err, "column_transformer", "column_type", column_type, "column_name", column_name, "query_name", query_name)
			return err
		}

		if t, err := _transformer(transformer); err != nil {
			return fmt.Errorf("unknown type `%s` for column %s of %s err %s",
				column_type, column_name, p.name, err)
		} else {
			if column_type == "tag" || column_type == "tag_list" {
				column_data = append(column_data, &columnItem{
					column_name: column_name,
					column_type: column_type,
					transformer: t,
				})
			} else {
				column_data = append(column_data, &columnItem{
					column_name: column_name,
					transformer: t,
				})
			}
		}
	}

	submission_transformers := make(mapinterface)
	for k, v := range column_transformers {
		if k == "tag" || k == "tag_list" {
			continue
		}
		submission_transformers[k] = v
	}

	extra_data := []*extraItem{}
	for i, extra := range p.query_data.Extras {
		extra_name, ok := extra["name"].(string)
		if !ok {
			return fmt.Errorf("field `name` for extra #%d of %s is required", i, query_name)
		}
		if s, ok := sources[extra_name]; ok {

			return fmt.Errorf("the name %s of %s was already defined in %s #%d",
				extra_name, query_name, s.Type, s.Index)
		}

		sources[extra_name] = &sourceItem{Type: "extra", Index: i}

		extra_type, ok := extra["type"].(string)
		if !ok {
			if extra["expression"] != nil {
				extra_type = "expression"
			} else {
				return fmt.Errorf("field `type` for extra %s of %s is required", extra_name, query_name)
			}
		} else if extra_transformers[extra_type] == nil && submission_transformers[extra_type] == nil {
			return fmt.Errorf("unknown type `%s` for extra %s of %s", extra_type, extra_name, query_name)
		}

		transformer_factory, err := _transformer(extra_transformers[extra_type])
		if err != nil {
			transformer_factory, err = _transformer(submission_transformers[extra_type])
		}
		if err != nil {
			return fmt.Errorf("unable get extra transformer[%s] %s", extra_type, err)
		}

		modifiers := make(mapinterface)
		extra_source, ok := extra["source"].(string)
		if submission_transformers[extra_type] != nil {
			if extra_source == "" {
				return fmt.Errorf("field `source` for extra %s of %s is required", extra_name, query_name)
			}
			for k, v := range extra {
				if k != "name" && k != "type" && k != "source" {
					modifiers[k] = v
				}
			}
		} else {
			for k, v := range extra {
				if k != "name" && k != "type" {
					modifiers[k] = v
				}
			}
			modifiers["sources"] = sources
		}

		transformer_, err := transformer_factory(submission_transformers, extra_name, modifiers)

		if err != nil {
			return fmt.Errorf("error compiling type `%s` for extra %s of %s: %s", extra_type, extra_name, query_name, err)
		}

		if t, err := _transformer(transformer_); err != nil {
			return err
		} else {
			if submission_transformers[extra_type] != nil {
				if t, err = create_extra_transformer(t, extra_source); err != nil {
					return err
				}
			}

			extra_data = append(extra_data, &extraItem{
				extra_name:  extra_name,
				transformer: t,
			})
		}
	}

	p.name = query_name
	p.query = query
	p.columns = column_data
	p.extras = extra_data
	p.tags = tags
	p.query_data = nil

	return nil
}

func QueryMapRow(db *sql.DB, ql string, values ...interface{}) (map[string]interface{}, error) {
	var fn func() []interface{}
	if len(values) > 0 {
		fn = func() []interface{} { return values }
	}
	if rows, err := queryMapRows(db, ql, false, fn); err != nil {
		return nil, err
	} else if len(rows) > 0 {
		return rows[0], nil
	}
	return nil, nil
}

func QueryMapRows(db *sql.DB, ql string, factory func() []interface{}) ([]map[string]interface{}, error) {
	return queryMapRows(db, ql, true, factory)
}

func queryMapRows(db *sql.DB, ql string, all bool, factory func() []interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Query(ql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []map[string]interface{}{}

	// first
	var cols []string
	for rows.Next() {
		if cols == nil {
			if cols, err = rows.Columns(); err != nil {
				return nil, err
			}
		}

		var values []interface{}
		if factory != nil {
			if values = factory(); len(values) != len(cols) {
				return nil, fmt.Errorf("columns %d != factory return %d", len(cols), len(values))
			}
		} else {
			values = make([]interface{}, len(cols))
			// fill the array with sql.Rawbytes
			for i := range values {
				values[i] = &sql.RawBytes{}
			}
		}

		if err = rows.Scan(values...); err != nil {
			return nil, err
		}

		m := map[string]interface{}{}
		for i := 0; i < len(cols); i++ {
			m[cols[i]] = values[i]
		}
		ret = append(ret, m)

		if !all {
			return ret, nil
		}
	}

	return ret, nil
}

func QueryRow(db *sql.DB, ql string, values ...interface{}) ([]interface{}, error) {
	var fn func() []interface{}
	if len(values) > 0 {
		fn = func() []interface{} { return values }
	}
	if rows, err := queryRows(db, ql, false, fn); err != nil {
		return nil, err
	} else if len(rows) > 0 {
		return rows[0], nil
	}
	return nil, nil
}

func QueryRows(db *sql.DB, ql string, factory func() []interface{}) ([][]interface{}, error) {
	return queryRows(db, ql, true, factory)
}

func queryRows(db *sql.DB, ql string, all bool, factory func() []interface{}) ([][]interface{}, error) {
	rows, err := db.Query(ql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := [][]interface{}{}
	var cols []string

	// first
	for rows.Next() {
		if cols == nil {
			if cols, err = rows.Columns(); err != nil {
				return nil, err
			}
		}

		var values []interface{}
		if factory != nil {
			if values = factory(); len(values) != len(cols) {
				return nil, fmt.Errorf("columns %d != factory return %d", len(cols), len(values))
			}
		} else {
			values = make([]interface{}, len(cols))
			// fill the array with sql.Rawbytes
			for i := range values {
				values[i] = &sql.RawBytes{}
			}
		}

		if err = rows.Scan(values...); err != nil {
			return nil, err
		}

		ret = append(ret, values)

		if !all {
			return ret, nil
		}
	}

	return ret, nil
}
