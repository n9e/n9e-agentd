package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/util"
	"golang.org/x/text/message"
)

type MetricDesc struct {
	Name   string   // raw name
	Metric string   // transformed . -> _
	Desc   string   // desc should be transformed by i18n
	Extra  []string // extra info, e.g. gauge,Shown as connection
}

type MetricGroup struct {
	Name         string      // group name
	Extra        []string    // no used
	MetricsDescs MetricDescs //
	metricsMap   map[string]*MetricDesc
}

type MetricDescs []*MetricDesc

func (p MetricDescs) Len() int           { return len(p) }
func (p MetricDescs) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p MetricDescs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type MetricGroups []*MetricGroup

func (p MetricGroups) Len() int           { return len(p) }
func (p MetricGroups) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p MetricGroups) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p MetricGroups) Output(lang *message.Printer) MetricGroups {
	m := make(MetricGroups, len(p))
	for k, v := range metricGroups {
		metricDescs := make(MetricDescs, len(v.MetricsDescs))
		for k1, v1 := range v.MetricsDescs {
			metricDescs[k1] = &MetricDesc{
				Name:   v1.Name,
				Metric: v1.Metric,
				Desc:   lang.Sprintf(v1.Name),
				Extra:  v1.Extra,
			}
		}
		m[k] = &MetricGroup{
			Name:         v.Name,
			Extra:        v.Extra,
			MetricsDescs: metricDescs,
		}
	}
	return m
}

var (
	once         sync.Once
	groupsMap    = map[string]*MetricGroup{}
	metricGroups MetricGroups
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
			v.Desc = v.Name
			group.MetricsDescs = append(group.MetricsDescs, v)
		}
		sort.Sort(group.MetricsDescs)
		metricGroups = append(metricGroups, group)
	}
	sort.Sort(metricGroups)
}

// curl -Ss 'http://localhost:8070/docs/metrics?output=json&lang=zh'
// curl -Ss 'http://localhost:8070/docs/metrics?output=json&lang=en'
func Handler() http.Handler {
	once.Do(initMetrics)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		output := r.FormValue("output")
		lang := i18n.NewPrinter(r.FormValue("lang"), "zh", "en")
		data := metricGroups.Output(lang)

		if output == "json" {
			util.WriteRawJSON(http.StatusOK, data, w)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		for _, g := range data {
			fmt.Fprintf(w, fmt.Sprintf("%s %s\r\n", g.Name, strings.Join(g.Extra, ",")))
			for _, m := range g.MetricsDescs {
				fmt.Fprintf(w, fmt.Sprintf("  %s:%s:%s\r\n",
					m.Metric, m.Desc, strings.Join(m.Extra, ",")))
			}
		}
	})
}
