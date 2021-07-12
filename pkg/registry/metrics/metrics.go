package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/util"
)

type MetricDesc struct {
	Name   string
	Metric string
	Desc   string
	Extra  []string
}

type MetricGroup struct {
	Name       string
	Extra      []string
	Metrics    metricDescs
	metricsMap map[string]*MetricDesc
}

type metricDescs []*MetricDesc

func (p metricDescs) Len() int           { return len(p) }
func (p metricDescs) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p metricDescs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type MetricGroups []*MetricGroup

func (p MetricGroups) Len() int           { return len(p) }
func (p MetricGroups) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p MetricGroups) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

var (
	once        sync.Once
	groupsMap   = map[string]*MetricGroup{}
	metricGrops MetricGroups
)

func GetMetricGroup(group string, extra ...string) *MetricGroup {
	if v, ok := groupsMap[group]; ok {
		return v
	}

	groupsMap[group] = &MetricGroup{
		Name:       group,
		Extra:      extra,
		metricsMap: map[string]*MetricDesc{},
	}
	return groupsMap[group]
}

// name, [type], [unitType]
func (p MetricGroup) Register(name string, extra ...string) {
	p.metricsMap[name] = &MetricDesc{Name: name, Extra: extra}
}

func initMetrics() {
	for _, group := range groupsMap {
		for _, v := range group.metricsMap {
			v.Metric = util.SanitizeMetric(v.Name)
			v.Desc = i18n.Sprintf(v.Name)
			group.Metrics = append(group.Metrics, v)
		}
		sort.Sort(group.Metrics)
		metricGrops = append(metricGrops, group)
	}
	sort.Sort(metricGrops)
}

func Handler() http.Handler {
	once.Do(initMetrics)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if q := r.FormValue("output"); q == "json" {
			util.WriteRawJSON(http.StatusOK, metricGrops, w)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		for _, g := range metricGrops {
			fmt.Fprintf(w, fmt.Sprintf("%s %s\r\n", g.Name, strings.Join(g.Extra, ",")))
			for _, m := range g.Metrics {
				fmt.Fprintf(w, fmt.Sprintf("  %s:%s:%s\r\n",
					m.Metric, m.Desc, strings.Join(m.Extra, ",")))
			}
		}
	})
}
