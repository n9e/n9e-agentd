package db

import "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metrics"

type transformHandle func(transformers mapinterface, name, modifiers interface{}) (interface{}, error)

type metricHandle func(string, float64, string, []string)
type serviceCheckHandle func(metric string, status metrics.ServiceCheckStatus, hostname string, tags []string, message string)

type sourceTransform struct {
	source      string
	transformer transformHandle
}

type submission_item struct {
	transformer transformHandle
	value       interface{}
	column_name string
	column_type string
}

type mapstring map[string]string

func (p mapstring) update(in mapstring) {
	for k, v := range in {
		p[k] = v
	}
}

func (p mapstring) tags() (tags []string) {
	for k, v := range p {
		tags = append(tags, k+":"+v)
	}
	return
}

type mapinterface map[string]interface{}

func (p mapinterface) update(in mapinterface) {
	for k, v := range in {
		p[k] = v
	}
}

func (p mapinterface) tags() (tags []string) {
	if v, ok := p["tags"].([]string); ok {
		return v
	}

	return nil
}

func (p mapinterface) pop(k string, def ...interface{}) interface{} {
	if v, ok := p[k]; ok {
		delete(p, k)
		return v
	}

	if len(def) > 0 {
		return def[0]
	}

	return nil
}
