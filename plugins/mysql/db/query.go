package db

import (
	"database/sql"
	"fmt"
)

type Query struct {
	data    *CustomQuery
	name    string
	query   string
	columns []string
	extras  []map[string]string
	tags    []string
}

func NewQuery(q *CustomQuery) *Query {
	return &Query{data: q}
}

func (p *Query) Compile(columnTransformers map[string]interface{}, extraTransformers interface{}) error {
	// Check for previous compilation
	if p.name == "" {
		return nil
	}

	q := p.data

	// Keep track of all defined names
	sources := map[string]int{}
	columnData := []interface{}{}
	for i, column := range q.Columns {
		columnName := column["name"]
		if columnName == "" {
			return fmt.Errorf("field `name` for column[%d]  of %s is required",
				i, q.Name)
		}

		if _, ok := sources[columnName]; ok {
			return fmt.Errorf("the name %s of %s was alread defined in column[%d]",
				columnName, q.Name, sources[columnName])
		}
		sources[columnName] = i

		columnType := column["type"]
		if columnType == "" {
			return fmt.Errorf("field `type` for column[%d]  of %s is required",
				i, q.Name)
		} else if columnType == "source" {
			columnData = append(columnData, columnName)
			continue
		} else if _, ok := columnTransformers[columnType]; !ok {
			return fmt.Errorf("unknown type `%s` for column %s of %s",
				columnType, columnName, q.Name)
		}

		//modifiers := map[string]struct{}{}
		//for k, v := range column {
		//	if k != "name" && k != "type" {
		//		modifiers[k] = struct{}{}
		//	}
		//}
	}

	return nil
}

func QueryRow(db *sql.DB, ql string, factory func() []interface{}) (map[string]interface{}, error) {
	if rows, err := queryRows(db, ql, false, factory); err != nil {
		return nil, err
	} else {
		return rows[0], nil
	}
}

func QueryRows(db *sql.DB, ql string, factory func() []interface{}) ([]map[string]interface{}, error) {
	if rows, err := queryRows(db, ql, true, factory); err != nil {
		return nil, err
	} else {
		return rows, nil
	}
}

func queryRows(db *sql.DB, ql string, all bool, factory func() []interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Query(ql)
	if err != nil {
		return nil, err
	}

	ret := []map[string]interface{}{}

	// first
	var fields []string
	if rows.Next() {
		if fields == nil {
			if fields, err = rows.Columns(); err != nil {
				return nil, err
			}
		}

		values := factory()
		if len(values) != len(fields) {
			return nil, fmt.Errorf("fields %s != factory return %d", len(fields), len(values))
		}

		if err = rows.Scan(values...); err != nil {
			return nil, err
		}

		m := map[string]interface{}{}
		for i := 0; i < len(fields); i++ {
			m[fields[i]] = values[i]
		}
		ret = append(ret, m)

		if !all {
			return ret, nil
		}
	}

	return ret, nil
}
