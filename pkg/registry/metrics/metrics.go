package metrics

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"

	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/util"
)

var (
	metricsMap = map[string]struct{}{}
	metrics    []string
	content    []byte
)

func Register(key string) {
	metricsMap[key] = struct{}{}
}

func InitMetrics(fs ...func(string) string) {
	for k, _ := range metricsMap {
		k2 := k
		for _, f := range fs {
			k2 = f(k2)
		}
		metrics = append(metrics, fmt.Sprintf("%s:%s", k2, i18n.Sprintf(k)))
	}
	sort.Strings(metrics)

	buf := bytes.Buffer{}
	for _, v := range metrics {
		buf.WriteString(fmt.Sprintf("%s\r\n", v))
	}
	content = buf.Bytes()
}

func MetricFilter(metric string) string {
	return util.SanitizeMetric(metric)
}

func Handler() http.Handler {
	InitMetrics(MetricFilter)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write(content)
	})
}
