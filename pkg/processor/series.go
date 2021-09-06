package processor

import (
	"fmt"
	"strings"

	"github.com/DataDog/datadog-agent/pkg/metrics"
	agentpayload "github.com/n9e/agent-payload/gogen"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/util"
)

type seriesProcessor struct {
	polices      map[string]*MetricPolicy
	ident        string
	alias        string
	additionTags []string
}

type MetricPolicy struct {
	NewName   string `json:"new_name"`
	ValueFunc string `json:"value_func"`

	valueFunc func(float64) float64
}

func newSeriesProcessor(opts *options) (*seriesProcessor, error) {
	metrics := make(map[string]*MetricPolicy)

	for k, v := range opts.Metrics {
		metrics[k] = v
	}

	for k, v := range _metricPolicies {
		metrics[k] = v
	}

	for _, v := range metrics {
		if len(v.ValueFunc) > 0 {
			if fn, ok := valueFuncs[v.ValueFunc]; !ok {
				return nil, fmt.Errorf("metric func %s not found", v.ValueFunc)
			} else {
				v.valueFunc = fn
			}
		}
	}

	return &seriesProcessor{
		polices:      metrics,
		ident:        config.C.Ident,
		alias:        config.C.Alias,
		additionTags: config.C.Tags,
	}, nil
}

func (p *seriesProcessor) ProcessSketch(sl metrics.SketchSeriesList) error {
	return nil
}

func (p *seriesProcessor) Process(series metrics.Series) *agentpayload.N9EMetricsPayload {
	payload := &agentpayload.N9EMetricsPayload{
		Samples:  []*agentpayload.N9EMetricsPayload_Sample{},
		Metadata: &agentpayload.CommonMetadata{},
	}

	for _, serie := range series {
		for _, point := range serie.Points {
			payload.Samples = append(payload.Samples,
				&agentpayload.N9EMetricsPayload_Sample{
					Ident:          p.ident,
					Alias:          p.alias,
					Metric:         p.MetricName(serie.Name),
					Tags:           p.TagsMap(serie.Tags),
					Time:           int64(point.Ts),
					Value:          p.value(serie.Name, point.Value),
					Type:           serie.MType.String(), // extra, no used
					SourceTypeName: serie.SourceTypeName, // extra, nouse
				})
		}
	}

	return payload
}

func (p *seriesProcessor) MetricName(name string) string {
	name = util.SanitizeMetric(name)

	if p := p.polices[name]; p != nil && p.NewName != "" {
		return p.NewName
	}

	// datadog_dogstatsd_* -> stasd_*
	name = strings.TrimPrefix(name, "datadog_dog")

	return name
}

func (p *seriesProcessor) value(name string, value float64) float64 {
	if p := p.polices["name"]; p != nil && p.valueFunc != nil {
		return p.valueFunc(value)
	}
	return value
}

func (p *seriesProcessor) TagsMap(tags []string) map[string]string {
	return util.SanitizeMapTags(append(tags, p.additionTags...))
}

func (p *seriesProcessor) Tags(tags []string) []string {
	return util.SanitizeTags(append(tags, p.additionTags...))
}
