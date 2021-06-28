package port

import (
	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/registry/metrics"
)

var langStrings = map[string]map[string]string{
	"zh": map[string]string{
		"proc.port.listen": "进程监听端口",
	},
}

func init() {
	metrics.Register("proc.port.listen")
	i18n.SetLangStrings(langStrings)
}
