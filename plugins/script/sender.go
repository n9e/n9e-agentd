package script

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/n9e/n9e-agentd/pkg/aggregator"
	"k8s.io/klog/v2"
)

// from github.com/didi/nightingale/src/common/dataobj
type MetricValue struct {
	Metric string            `json:"metric"` //
	Value  interface{}       `json:"value"`  //
	Type   string            `json:"type"`   // GAUGE | COUNTER | SUBTRACT | DERIVE
	Tags   map[string]string `json:"tags"`   //
}

func send(sender aggregator.Sender, data []byte) error {
	var series []*MetricValue
	if err := json.Unmarshal(data, &series); err != nil {
		return err
	}

	for _, serie := range series {
		switch strings.ToLower(serie.Type) {
		case "gauge":
			sender.Gauge(serie.Metric, Float(serie.Value), "", makeTags(serie))
		case "rate", "counter": // same as rrdtool / open-falcon
			sender.Rate(serie.Metric, Float(serie.Value), "", makeTags(serie))
		case "monotonic_count", "subtract", "increase":
			sender.MonotonicCount(serie.Metric, Float(serie.Value), "", makeTags(serie))
		default:
			sender.Gauge(serie.Metric, Float(serie.Value), "", makeTags(serie))
		}
	}
	return nil
}

func makeTags(m *MetricValue) []string {
	if len(m.Tags) == 0 {
		return nil
	}

	tags := make([]string, 0, len(m.Tags))
	for k, v := range m.Tags {
		tags = append(tags, k+":"+v)
	}

	return tags
}

func Float(a interface{}) float64 {
	switch v := a.(type) {
	case float64:
		return v
	case string:
		i, _ := strconv.ParseFloat(v, 0)
		return i
	default:
		fmt.Printf("unsupported metric value type %s\n", reflect.TypeOf(a))
		klog.V(5).Infof("unsupported metric value type %s", reflect.TypeOf(a))
		return 0
	}
}
