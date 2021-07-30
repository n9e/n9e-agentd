package prometheus

import (
	"fmt"
	"math"

	dto "github.com/prometheus/client_model/go"
	"github.com/n9e/n9e-agentd/pkg/aggregator"
)

// https://prometheus.io/docs/concepts/metric_types/#summary
// https://github.com/OpenObservability/OpenMetrics/blob/master/specification/OpenMetrics.md#summary-1
func sendSummary(metricName string, m *dto.Metric, sender aggregator.Sender) {
	tags := makeTags(m)
	data := m.GetSummary()

	quantileMetric := metricName + "_quantile"
	sumMetric := metricName + "_sum"
	countMetric := metricName + "_count"

	for _, v := range data.GetQuantile() {
		sender.Gauge(quantileMetric, v.GetValue(), "", append(tags, fmt.Sprintf("quantile:%v", v.GetQuantile())))
	}
	sender.MonotonicCount(sumMetric, data.GetSampleSum(), "", tags)
	sender.MonotonicCount(countMetric, float64(data.GetSampleCount()), "", tags)
}

// https://prometheus.io/docs/concepts/metric_types/#histogram
// https://github.com/OpenObservability/OpenMetrics/blob/master/specification/OpenMetrics.md#histogram-1
func sendHistogram(metricName string, m *dto.Metric, sender aggregator.Sender) {
	tags := makeTags(m)
	data := m.GetHistogram()

	quantileMetric := metricName + "_bucket"
	sumMetric := metricName + "_sum"
	countMetric := metricName + "_count"

	for _, v := range data.GetBucket() {
		sender.Gauge(quantileMetric, float64(v.GetCumulativeCount()), "",
			append(tags, fmt.Sprintf("le:%v", v.GetUpperBound())))
	}
	sender.MonotonicCount(sumMetric, data.GetSampleSum(), "", tags)
	sender.MonotonicCount(countMetric, float64(data.GetSampleCount()), "", tags)

}

func sendGauge(metricName string, m *dto.Metric, sender aggregator.Sender) {
	if v := m.GetGauge().GetValue(); !math.IsNaN(v) {
		sender.Gauge(metricName, v, "", makeTags(m))
	}
}

func sendCounter(metricName string, m *dto.Metric, sender aggregator.Sender) {
	if v := m.GetCounter().GetValue(); !math.IsNaN(v) {
		sender.MonotonicCount(metricName, v, "", makeTags(m))
	}
}

func sendUntyped(metricName string, m *dto.Metric, sender aggregator.Sender) {
	if v := m.GetUntyped().GetValue(); !math.IsNaN(v) {
		sender.Gauge(metricName, v, "", makeTags(m))
	}
}

// Get labels from metric
func makeTags(m *dto.Metric) []string {
	tags := make([]string, len(m.Label))
	for i, lp := range m.Label {
		tags[i] = lp.GetName() + ":" + lp.GetValue()
	}
	return tags
}
