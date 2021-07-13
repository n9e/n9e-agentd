package demo

import (
	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/registry/metrics"
)

var langStrings = map[string]map[string]string{
	"zh": map[string]string{
		"demo": "演示",
	},
	"en": map[string]string{
		"demo": "demo metric",
	},
}

func registerMetric() {
	m := metrics.GetMetricGroup("demo")
	m.Register("demo", "gauge", "demo metric")
}

func init() {
	registerMetric()
	i18n.SetLangStrings(langStrings)
}
