package mysql

import (
	"database/sql"

	libdb "github.com/n9e/n9e-agentd/plugins/mysql/db"
)

func queryRow(db *sql.DB, ql string, factory func() []interface{}) (map[string]interface{}, error) {
	return libdb.QueryRow(db, ql, factory)
}

func queryRows(db *sql.DB, ql string, factory func() []interface{}) ([]map[string]interface{}, error) {
	return libdb.QueryRows(db, ql, factory)
}
