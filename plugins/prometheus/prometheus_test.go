package prometheus

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/n9e/n9e-agentd/pkg/aggregator/mocksender"
)

const sampleTextFormat = `
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0.00010425500000000001
go_gc_duration_seconds{quantile="0.25"} 0.000139108
go_gc_duration_seconds{quantile="0.5"} 0.00015749400000000002
go_gc_duration_seconds{quantile="0.75"} 0.000331463
go_gc_duration_seconds{quantile="1"} 0.000667154
go_gc_duration_seconds_sum 0.0018183950000000002
go_gc_duration_seconds_count 7
# HELP test_histogram A histogram for test
# TYPE test_histogram histogram
test_histogram_bucket{le="0.1", start="positive"} 1
test_histogram_bucket{le=".2", start="positive"} 2
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 15
# HELP test_metric An untyped metric with a timestamp
# TYPE test_metric untyped
test_metric{label="value"} 1.0
`

func TestCollect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, sampleTextFormat)
	}))
	defer ts.Close()

	check := new(Check)
	err := check.Configure([]byte(fmt.Sprintf(`
prometheusUrl: %s
namespace: test
`, ts.URL)), nil, "test")
	assert.Nil(t, err)

	sender := mocksender.NewMockSender(check.ID())

	sender.On("Gauge", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("Rate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("MonotonicCount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("Commit").Return()

	err = check.Run()
	assert.Nil(t, err)

	sender.AssertCalled(t, "Gauge", "go_gc_duration_seconds_quantile", float64(0.00010425500000000001), "", []string{"quantile:0"})
	sender.AssertCalled(t, "Gauge", "go_gc_duration_seconds_quantile", float64(0.000139108), "", []string{"quantile:0.25"})
	sender.AssertCalled(t, "Gauge", "go_gc_duration_seconds_quantile", float64(0.00015749400000000002), "", []string{"quantile:0.5"})
	sender.AssertCalled(t, "Gauge", "go_gc_duration_seconds_quantile", float64(0.000331463), "", []string{"quantile:0.75"})
	sender.AssertCalled(t, "Gauge", "go_gc_duration_seconds_quantile", float64(0.000331463), "", []string{"quantile:0.75"})
	sender.AssertCalled(t, "Gauge", "go_gc_duration_seconds_quantile", float64(0.000667154), "", []string{"quantile:1"})
	sender.AssertCalled(t, "MonotonicCount", "go_gc_duration_seconds_sum", float64(0.0018183950000000002), "", []string{})
	sender.AssertCalled(t, "MonotonicCount", "go_gc_duration_seconds_count", float64(7), "", []string{})

	sender.AssertCalled(t, "Gauge", "test_histogram_bucket", float64(1), "", []string{"start:positive", "le:0.1"})
	sender.AssertCalled(t, "Gauge", "test_histogram_bucket", float64(2), "", []string{"start:positive", "le:0.2"})

	sender.AssertCalled(t, "Gauge", "test_metric", float64(1), "", []string{"label:value"})
	//sender.AssertCalled(t, "Commit")
}
