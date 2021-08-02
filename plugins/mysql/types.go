package mysql

import (
	"strconv"
	"strings"

	"github.com/DataDog/datadog-agent/pkg/metrics"
	"k8s.io/klog/v2"
)

type MetricItem struct {
	Metric string
	Type   metrics.MetricType
	Items  []MetricItem
}

type MetricItems map[string]MetricItem

func (p MetricItems) update(in MetricItems) {
	for k, v := range in {
		p[k] = v
	}
}
func (p MetricItems) pop(k string, def ...MetricItem) MetricItem {
	if v, ok := p[k]; ok {
		delete(p, k)
		return v
	}

	if len(def) > 0 {
		return def[0]
	}

	return MetricItem{}
}

type mapinterface map[string]interface{}

func (p mapinterface) dump(filters ...string) {
	if klog.V(6).Enabled() {
		for k, v := range p {
			if len(filters) == 0 {
				klog.Infof("> %s:%v", k, v)
				continue
			}

			for _, f := range filters {
				if strings.Contains(k, f) {
					klog.Infof("> %s:%v", k, v)
					break
				}
			}
		}
	}
}

func (p mapinterface) update(in mapinterface) {
	for k, v := range in {
		p[k] = v
	}
}

func (p mapinterface) collectMap(key string) (map[string]interface{}, bool) {
	v, ok := p[key].(map[string]interface{})
	return v, ok
}

func (p mapinterface) collectFloat(key string) (float64, bool) {
	switch t := p[key].(type) {
	case []byte:
		i, err := strconv.ParseFloat(string(t), 0)
		if err != nil {
			return 0, false
		}
		return i, true
	case string:
		i, err := strconv.ParseFloat(t, 0)
		if err != nil {
			return 0, false
		}
		return i, true
	case float64:
		return t, true
	case int64:
		return float64(t), true
	default:
		return 0, false
	}
}
func (p mapinterface) collectInt(key string) (int64, bool) {
	switch t := p[key].(type) {
	case []byte:
		i, err := strconv.ParseInt(string(t), 10, 0)
		if err != nil {
			return 0, false
		}
		return i, true
	case string:
		i, err := strconv.ParseInt(t, 10, 0)
		if err != nil {
			return 0, false
		}
		return i, true
	case float64:
		return int64(t), true
	case int64:
		return t, true
	default:
		return 0, false
	}
}

func (p mapinterface) collectString(key string) (string, bool) {
	switch t := p[key].(type) {
	case []byte:
		return string(t), true
	case string:
		return t, true
	}

	return "", false
}

func (p mapinterface) collectBool(key string) bool {
	return p[key].(bool)
}
