package script

import (
	"encoding/json"
	"strings"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"
)

// from github.com/didi/nightingale/src/common/dataobj
type MetricValue struct {
	Metric      string  `json:"metric"`
	Endpoint    string  `json:"endpoint"`
	Value       float64 `json:"value"`
	CounterType string  `json:"counterType"` // GAUGE | COUNTER | SUBTRACT | DERIVE
	Tags        string  `json:"tags"`        // a=1,b=2,c=3
}

func send(sender aggregator.Sender, data []byte) error {
	var series []*MetricValue
	if err := json.Unmarshal(data, &series); err != nil {
		return err
	}

	for _, serie := range series {
		switch strings.ToLower(serie.CounterType) {
		case "gauge":
			sender.Gauge(serie.Metric, serie.Value, serie.Endpoint, makeTags(serie))
		case "counter": // same as rrdtool / open-falcon
			sender.Rate(serie.Metric, serie.Value, serie.Endpoint, makeTags(serie))
		case "monotonic_count":
			sender.MonotonicCount(serie.Metric, serie.Value, serie.Endpoint, makeTags(serie))
		default:
			sender.Gauge(serie.Metric, serie.Value, serie.Endpoint, makeTags(serie))
		}
	}
	return nil
}

func makeTags(m *MetricValue) []string {
	if m.Tags == "" {
		return nil
	}

	tags := strings.Split(m.Tags, ",")
	if strings.IndexByte(tags[0], '=') > 0 && strings.IndexByte(tags[0], ':') < 0 {
		for i, tag := range tags {
			tags[i] = strings.Replace(tag, "=", ":", 1)
		}
	}

	return tags
}
