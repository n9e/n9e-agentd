package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"sync"

	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/util"
)

type metricDesc struct {
	Name   string
	Metric string
	Desc   string
	Extra  []string
}

type metricsDesc []*metricDesc

func (p metricsDesc) Len() int           { return len(p) }
func (p metricsDesc) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p metricsDesc) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

var (
	metricsMap = map[string]*metricDesc{}
	metrics    metricsDesc
	metricOnce sync.Once
)

// name, [type], [unitType]
func Register(name string, extra ...string) {
	metricsMap[name] = &metricDesc{Name: name, Extra: extra}
}

func initMetrics() {
	for _, v := range metricsMap {
		v.Metric = util.SanitizeMetric(v.Name)
		v.Desc = i18n.Sprintf(v.Name)
		metrics = append(metrics, v)
	}
	sort.Sort(metrics)
}

func Handler() http.Handler {
	metricOnce.Do(initMetrics)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		for _, m := range metrics {
			fmt.Fprintf(w, fmt.Sprintf("%s:%s\r\n", m.Metric, m.Desc))
		}
	})
}
