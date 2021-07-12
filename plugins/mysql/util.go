package mysql

import (
	"database/sql"
	"fmt"
	"hash/crc64"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/n9e/n9e-agentd/plugins/mysql/db"
)

var (
	hashTable = crc64.MakeTable(crc64.ISO)
)

func (c *Check) queryMapRow(sql string, values ...interface{}) (map[string]interface{}, error) {
	return db.QueryMapRow(c.db, sql, values...)
}

func (c *Check) queryMapRows(sql string, factory ...func() []interface{}) ([]map[string]interface{}, error) {
	if len(factory) > 0 {
		return db.QueryMapRows(c.db, sql, factory[0])
	}
	return db.QueryMapRows(c.db, sql, nil)

}

func (c *Check) queryRow(sql string, values ...interface{}) ([]interface{}, error) {
	return db.QueryRow(c.db, sql, values...)
}

func (c *Check) queryRows(sql string, factory ...func() []interface{}) ([][]interface{}, error) {
	if len(factory) > 0 {
		return db.QueryRows(c.db, sql, factory[0])
	}
	return db.QueryRows(c.db, sql, nil)
}

func (c *Check) queryKv(ql string) (mapinterface, error) {
	results := make(mapinterface)
	rows, err := c.queryRows(ql, func() []interface{} { return []interface{}{new(string), new(sql.RawBytes)} })
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		results[strings.ToLower(*(row[0]).(*string))] = row[1]
	}

	return results, nil
}

func Int(a interface{}) int64 {
	switch v := a.(type) {
	case int64:
		return v
	case *sql.RawBytes:
		i, _ := strconv.ParseInt(string(*v), 10, 0)
		return i
	case string:
		i, _ := strconv.ParseInt(v, 10, 0)
		return i
	case float64:
		return int64(v)
	case nil:
		return 0
	default:
		panic(fmt.Sprintf("unsupported type %s", reflect.TypeOf(a)))
	}
}

func Float(a interface{}) float64 {
	switch v := a.(type) {
	case float64:
		return v
	case *sql.RawBytes:
		i, _ := strconv.ParseFloat(string(*v), 0)
		return i
	case string:
		i, _ := strconv.ParseFloat(v, 0)
		return i
	case nil:
		return 0
	default:
		panic(fmt.Sprintf("unsupported type %s", reflect.TypeOf(a)))
	}
}

func String(a interface{}) string {
	switch v := a.(type) {
	case string:
		return v
	case *string:
		return *v
	case *sql.RawBytes:
		return string(*v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(v, 10)
	case nil:
		return ""
	default:
		panic(fmt.Sprintf("unsupported type %v", reflect.TypeOf(a)))
	}
}

func _are_values_numeric(in []string) bool {
	for _, v := range in {
		for _, c := range []byte(v) {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

func collect_scalar(key string, v map[string]interface{}) float64 {
	return Float(v[key])
}

func hash64(s string) uint64 {
	c := crc64.New(hashTable)
	io.WriteString(c, s)
	return c.Sum64()
}
