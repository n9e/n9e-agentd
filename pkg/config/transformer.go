package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/n9e/n9e-agentd/pkg/util"
	"k8s.io/klog/v2"
)

func init() {
	defaultTransformer = NewTransformer()
	defaultTransformer.SetMetric(transformMetricMap)
}

var (
	defaultTransformer *Transformer
)

type Transformer struct {
	metrics map[string]string
}

func NewTransformer() *Transformer { return &Transformer{metrics: make(map[string]string)} }

func (t *Transformer) SetMetric(config map[string]string) {
	for k, v := range config {
		t.metrics[util.SanitizeMetric(k)] = v
	}
}

func (t *Transformer) Metric(name string) string {
	name = util.SanitizeMetric(name)
	if v, ok := t.metrics[name]; ok {
		return v
	}

	return name
}

func (t *Transformer) SetMetricFromFile(file string) error {
	if file == "" {
		return nil
	}

	klog.V(6).Infof("load transformer metric from %s", file)

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	config := map[string]string{}
	if err := json.Unmarshal(b, &config); err != nil {
		return nil
	}

	t.SetMetric(config)

	return nil
}

func TransformMetric(metric string) string             { return defaultTransformer.Metric(metric) }
func TransformMapTags(tags []string) map[string]string { return util.SanitizeMapTags(tags) }
func TransformTags(tags []string) []string             { return util.SanitizeTags(tags) }

var (
	transformMetricMap = map[string]string{
		"demo": "demo777",
	}
)
