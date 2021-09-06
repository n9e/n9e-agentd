package processor

import (
	"io/ioutil"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/yubo/golib/util/yaml"
	"k8s.io/klog/v2"
)

type Processor struct {
	enabled bool
	config  string
	*seriesProcessor
}

type options struct {
	Metrics map[string]*MetricPolicy `json:"metrics"`
}

func NewProcessor() (*Processor, error) {
	p := &Processor{}
	if !config.C.EnableN9eProvider {
		klog.V(1).Infof("payload processor disabled")
		return p, nil
	}

	p.enabled = true

	opts := &options{
		Metrics: make(map[string]*MetricPolicy),
	}
	if file := config.C.PayloadProcessorConfig; file != "" {
		klog.V(1).Infof("load payload processor config from %s", file)
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, opts); err != nil {
			return nil, err
		}
	}

	var err error
	p.seriesProcessor, err = newSeriesProcessor(opts)
	if err != nil {
		return nil, err
	}

	return p, nil
}
