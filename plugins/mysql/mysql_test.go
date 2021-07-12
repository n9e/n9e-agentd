package mysql

import (
	"fmt"
	"os"
	"testing"

	"github.com/n9e/n9e-agentd/pkg/util/testsender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMysql(t *testing.T) {
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		t.Fatalf("unable to get TEST_DSN from env")
	}

	check := new(Check)
	err := check.Configure([]byte(fmt.Sprintf(`
dsn: %s 
options:
  extraInnodbMetrics: true
  extraStatusMetrics: true
  schemaSizeMetrics: true
  extraPerformanceMetrics: true
custom_queries:
  - query: SELECT foo, COUNT(*) FROM test.events GROUP BY foo
    columns:
    - name: foo
      type: tag
    - name: event.total
      type: gauge
    tags:
    - test:mysql
`, dsn)), nil, "test")
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
}
