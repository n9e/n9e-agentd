package port

import (
	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/registry/metrics"
)

var langStrings = map[string]map[string]string{
	"zh": map[string]string{
		"proc.port.listen": "进程监听端口",
	},
	"en": map[string]string{
		"proc.port.listen": "Process listening port",
	},
}

func registerMetric() {
	m := metrics.GetMetricGroup("process")
	m.Register("proc.port.listen")
}

func init() {
	registerMetric()
	i18n.SetLangStrings(langStrings)
}
