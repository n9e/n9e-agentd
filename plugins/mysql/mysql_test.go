package mysql

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/n9e/n9e-agentd/pkg/util/testsender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testTableName = "test.metrics"
)

func initdb(dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	if _, err = db.Exec("CREATE TABLE " + testTableName + " (name varchar(31), used int, total int);"); err != nil {
		return err
	}

	// | name | used | total |
	// | --   | --   | --    |
	// | foo  | 1    | 2     |
	// | baz  | 3    | 4     |
	// | bar  | 5    | 6     |
	if _, err = db.Exec("INSERT INTO "+testTableName+
		" VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?)",
		"foo", 2, 3,
		"baz", 3, 4,
		"bar", 5, 6); err != nil {
		return err
	}
	return nil
}

func teardowndb(dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	if _, err = db.Exec("DROP TABLE IF EXISTS " + testTableName); err != nil {
		return err
	}

	return nil
}

func TestMysql(t *testing.T) {
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		t.Fatalf("unable to get TEST_DSN from env")
	}

	if err := initdb(dsn); err != nil {
		t.Fatalf("db %s", err)
	}
	defer teardowndb(dsn)

	check := new(Check)
	err := check.Configure([]byte(`
dsn: `+dsn+`
options:
  extraInnodbMetrics: false
  extraStatusMetrics: false
  schemaSizeMetrics: false
  extraPerformanceMetrics: false
customQueries:
  - query: SELECT name, used, total FROM `+testTableName+` where name != 'baz'
    columns:
      - name: name
        type: tag
      - name: disk_used
        type: gauge
      - name: disk_total
        type: gauge
    tags:
      - collector:test
    extras:
      - name: free
        expression: disk_total - disk_used
      - name: disk_free
        type: gauge
        source: free
`), nil, "test")
	assert.Nil(t, err)

	sender := testsender.NewTestSender(check.ID(), t)

	sender.On("Rate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("Gauge", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("Count", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("MonotonicCount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("Histogram", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("Historate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("ServiceCheck", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("Event", mock.Anything).Return()
	sender.On("Commit").Return()

	//assert.Nil(t, check.Run())
	//time.Sleep(time.Second)

	assert.Nil(t, check.Run())

	sender.AssertCalled(t, "Gauge", "disk_used", float64(2.0), "", []string{"collector:test", "name:foo"})
	sender.AssertCalled(t, "Gauge", "disk_total", float64(3.0), "", []string{"collector:test", "name:foo"})
	sender.AssertCalled(t, "Gauge", "disk_free", float64(1.0), "", []string{"collector:test", "name:foo"})
	sender.AssertCalled(t, "Gauge", "disk_used", float64(5.0), "", []string{"collector:test", "name:bar"})
	sender.AssertCalled(t, "Gauge", "disk_total", float64(6.0), "", []string{"collector:test", "name:bar"})
	sender.AssertCalled(t, "Gauge", "disk_free", float64(1.0), "", []string{"collector:test", "name:bar"})

}
