// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package integration

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/containers"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/tmplvar"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

// Data contains YAML code
type Data []byte

// RawMap is the generic type to hold YAML configurations
type RawMap map[interface{}]interface{}

// JSONMap is the generic type to hold JSON configurations
type JSONMap map[string]interface{}

// CreationTime represents the moment when the service was launched compare to the agent start.
type CreationTime int

const (
	// Before indicates the service was launched before the agent start
	Before CreationTime = iota
	// After indicates the service was launched after the agent start
	After
)

// Config is a generic container for configuration files
// When a new field is added to this struct, please evaluate whether it should be computed in the config Digest
// and update the field's documentation and the Digest method accordingly
type Config struct {
	Name                    string       `yaml:"check_name" json:"check_name,omitempty"`                               // the name of the check (include in digest: true)
	Instances               []Data       `yaml:"instances" json:"instances,omitempty"`                                 // the list of instances in Yaml (include in digest: true)
	InitConfig              Data         `yaml:"init_config" json:"init_config,omitempty"`                             // the initConfig in Yaml (include in digest: true)
	MetricConfig            Data         `yaml:"metric_config" json:"metric_config,omitempty"`                         // the metric config in Yaml (jmx check only) (include in digest: false)
	LogsConfig              Data         `yaml:"logs" json:"logs,omitempty"`                                           // the logs config in Yaml (logs-agent only) (include in digest: true)
	ADIdentifiers           []string     `yaml:"ad_identifiers" json:"ad_identifiers,omitempty"`                       // the list of AutoDiscovery identifiers (optional) (include in digest: true)
	Provider                string       `yaml:"provider" json:"provider,omitempty"`                                   // the provider that issued the config (include in digest: false)
	Entity                  string       `yaml:"-" json:"-"`                                                           // the entity ID (optional) (include in digest: true)
	TaggerEntity            string       `yaml:"-" json:"-"`                                                           // the tagger entity ID (optional) (include in digest: false)
	ClusterCheck            bool         `yaml:"cluster_check" json:"cluster_check,omitempty"`                         // cluster-check configuration flag (include in digest: false)
	NodeName                string       `yaml:"node_name" json:"node_name,omitempty"`                                 // node name in case of an endpoint check backed by a pod (include in digest: true)
	CreationTime            CreationTime `yaml:"-" json:"-"`                                                           // creation time of service (include in digest: false)
	Source                  string       `yaml:"source" json:"source,omitempty"`                                       // the source of the configuration (include in digest: false)
	IgnoreAutodiscoveryTags bool         `yaml:"ignore_autodiscovery_tags" json:"ignore_autodiscovery_tags,omitempty"` // used to ignore tags coming from autodiscovery (include in digest: true)
	MetricsExcluded         bool         `yaml:"-" json:"-"`                                                           // whether metrics collection is disabled (set by container listeners only) (include in digest: false)
	LogsExcluded            bool         `yaml:"-" json:"-"`                                                           // whether logs collection is disabled (set by container listeners only) (include in digest: false)
}

// CommonInstanceConfig holds the reserved fields for the yaml instance data
type CommonInstanceConfig struct {
	MinCollectionInterval int      `yaml:"min_collection_interval" json:"min_collection_interval,omitempty"` // This changes the collection interval of the check - default: 15
	EmptyDefaultHostname  bool     `yaml:"empty_default_hostname" json:"empty_default_hostname,omitempty"`   // This forces the check to send metrics with no hostname. This is useful for cluster-level checks.
	Tags                  []string `yaml:"tags" json:"tags,omitempty"`                                       // A list of tags to attach to every metric and service check emitted by this instance, <key_1>:<value_1>
	Service               string   `yaml:"service" json:"service,omitempty"`                                 // Attach the tag `service:<SERVICE>` to every metric, event, and service check emitted by this integration.
	Name                  string   `yaml:"name" json:"name,omitempty"`                                       //
	Namespace             string   `yaml:"namespace" json:"namespace,omitempty"`                             //
}

// CommonGlobalConfig holds the reserved fields for the yaml init_config data
type CommonGlobalConfig struct {
	Service string `yaml:"service"`
}

// Equal determines whether the passed config is the same
func (c *Config) Equal(cfg *Config) bool {
	if cfg == nil {
		return false
	}

	return c.Digest() == cfg.Digest()
}

// String YAML representation of the config
func (c *Config) String() string {
	rawConfig := make(map[interface{}]interface{})
	var initConfig interface{}
	var instances []interface{}
	var logsConfig interface{}

	rawConfig["checkName"] = c.Name

	yaml.Unmarshal(c.InitConfig, &initConfig) //nolint:errcheck
	rawConfig["initConfig"] = initConfig

	for _, i := range c.Instances {
		var instance interface{}
		yaml.Unmarshal(i, &instance) //nolint:errcheck
		instances = append(instances, instance)
	}
	rawConfig["instances"] = instances

	yaml.Unmarshal(c.LogsConfig, &logsConfig) //nolint:errcheck
	rawConfig["logsConfig"] = logsConfig

	buffer, err := yaml.Marshal(&rawConfig)
	if err != nil {
		klog.Error(err)
	}

	return string(buffer)
}

// IsTemplate returns if the config has AD identifiers
func (c *Config) IsTemplate() bool {
	return len(c.ADIdentifiers) > 0
}

// IsCheckConfig returns true if the config is a node-agent check configuration,
func (c *Config) IsCheckConfig() bool {
	return c.ClusterCheck == false && len(c.Instances) > 0
}

// IsLogConfig returns true if config contains a logs config.
func (c *Config) IsLogConfig() bool {
	return c.LogsConfig != nil
}

// HasFilter returns true if metrics or logs collection must be disabled for this config.
// no containers.GlobalFilter case here because we don't create services that are globally excluded in AD
func (c *Config) HasFilter(filter containers.FilterType) bool {
	switch filter {
	case containers.MetricsFilter:
		return c.MetricsExcluded
	case containers.LogsFilter:
		return c.LogsExcluded
	}
	return false
}

// AddMetrics adds metrics to a check configuration
func (c *Config) AddMetrics(metrics Data) error {
	var rawInitConfig RawMap
	err := yaml.Unmarshal(c.InitConfig, &rawInitConfig)
	if err != nil {
		return err
	}

	var rawMetricsConfig []interface{}
	err = yaml.Unmarshal(metrics, &rawMetricsConfig)
	if err != nil {
		return err
	}

	// Grab any metrics currently in init_config
	var conf []interface{}
	currMetrics := make(map[string]bool)
	if _, ok := rawInitConfig["conf"]; ok {
		if currentMetrics, ok := rawInitConfig["conf"].([]interface{}); ok {
			for _, metric := range currentMetrics {
				conf = append(conf, metric)

				if metricS, e := yaml.Marshal(metric); e == nil {
					currMetrics[string(metricS)] = true
				}
			}
		}
	}

	// Add new metrics, skip duplicates
	for _, metric := range rawMetricsConfig {
		if metricS, e := yaml.Marshal(metric); e == nil {
			if !currMetrics[string(metricS)] {
				conf = append(conf, metric)
			}
		}
	}

	// JMX fetch expects the metrics to be a part of init_config, under "conf"
	rawInitConfig["conf"] = conf
	initConfig, err := yaml.Marshal(rawInitConfig)
	if err != nil {
		return err
	}

	c.InitConfig = initConfig
	return nil
}

// GetTemplateVariablesForInstance returns a slice of raw template variables
// it found in a config instance template.
func (c *Config) GetTemplateVariablesForInstance(i int) []tmplvar.TemplateVar {
	if len(c.Instances) < i {
		return nil
	}
	return tmplvar.Parse(c.Instances[i])
}

// GetNameForInstance returns the name from an instance if specified, fallback on namespace
func (c *Data) GetNameForInstance() string {
	commonOptions := CommonInstanceConfig{}
	err := yaml.Unmarshal(*c, &commonOptions)
	if err != nil {
		klog.Errorf("invalid instance section: %s", err)
		return ""
	}

	if commonOptions.Name != "" {
		return commonOptions.Name
	}

	// Fallback on `namespace` if we don't find `name`, can be empty
	return commonOptions.Namespace
}

// MergeAdditionalTags merges additional tags to possible existing config tags
func (c *Data) MergeAdditionalTags(tags []string) error {
	rawConfig := RawMap{}
	err := yaml.Unmarshal(*c, &rawConfig)
	if err != nil {
		return err
	}
	rTags, _ := rawConfig["tags"].([]interface{})
	// convert raw tags to string
	cTags := make([]string, len(rTags))
	for i, t := range rTags {
		cTags[i] = fmt.Sprint(t)
	}
	tagList := append(cTags, tags...)
	if len(tagList) == 0 {
		return nil
	}
	// use set keys to remove duplicate
	tagSet := make(map[string]struct{})
	for _, t := range tagList {
		tagSet[t] = struct{}{}
	}
	// override config tags
	rawConfig["tags"] = []string{}
	for k := range tagSet {
		rawConfig["tags"] = append(rawConfig["tags"].([]string), k)
	}
	// modify original config
	out, err := yaml.Marshal(&rawConfig)
	if err != nil {
		return err
	}
	*c = Data(out)

	return nil
}

// SetField allows to set an arbitrary field to a given value,
// overriding the existing value if present
func (c *Data) SetField(key string, value interface{}) error {
	rawConfig := RawMap{}
	err := yaml.Unmarshal(*c, &rawConfig)
	if err != nil {
		return err
	}

	rawConfig[key] = value
	out, err := yaml.Marshal(&rawConfig)
	if err != nil {
		return err
	}
	*c = Data(out)

	return nil
}

// Digest returns an hash value representing the data stored in this configuration.
// The ClusterCheck field is intentionally left out to keep a stable digest
// between the cluster-agent and the node-agents
func (c *Config) Digest() string {
	h := fnv.New64()
	h.Write([]byte(c.Name)) //nolint:errcheck
	for _, i := range c.Instances {
		inst := RawMap{}
		err := yaml.Unmarshal(i, &inst)
		if err != nil {
			klog.V(5).Infof("Error while calculating config digest for %s, skipping: %v", c.Name, err)
			continue
		}
		if val, found := inst["tags"]; found {
			// sort the list of tags so the digest stays stable for
			// identical configs with the same tags but with different order
			tagsInterface, ok := val.([]interface{})
			if !ok {
				klog.V(5).Infof("Error while calculating config digest for %s, skipping: cannot read tags from config", c.Name)
				continue
			}
			tags := make([]string, len(tagsInterface))
			for i, tag := range tagsInterface {
				tags[i] = fmt.Sprint(tag)
			}
			sort.Strings(tags)
			inst["tags"] = tags
		}
		out, err := yaml.Marshal(&inst)
		if err != nil {
			klog.V(5).Infof("Error while calculating config digest for %s, skipping: %v", c.Name, err)
			continue
		}
		h.Write(out) //nolint:errcheck
	}
	h.Write([]byte(c.InitConfig)) //nolint:errcheck
	for _, i := range c.ADIdentifiers {
		h.Write([]byte(i)) //nolint:errcheck
	}
	h.Write([]byte(c.NodeName))                                    //nolint:errcheck
	h.Write([]byte(c.LogsConfig))                                  //nolint:errcheck
	h.Write([]byte(c.Entity))                                      //nolint:errcheck
	h.Write([]byte(strconv.FormatBool(c.IgnoreAutodiscoveryTags))) //nolint:errcheck

	return strconv.FormatUint(h.Sum64(), 16)
}