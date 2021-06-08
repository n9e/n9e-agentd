package api

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

type CollectRule struct {
	Name      string `json:"name"`       // -> Config.instances[$]
	Type      string `json:"type"`       // -> Config.Name
	Data      string `json:"data"`       // -> Config.Instances[$]
	Interval  int    `json:"step"`       // -> Config.Instances[$].MinCollectionInterval
	Tags      string `json:"appendTags"` // -> Config.Instances[$].Tags   a:b,b:c
	CreatedAt int64  `json:"createAt"`   // deprecated
	UpdatedAt int64  `json:"updateAt"`   // deprecated
	Creator   string `json:"createBy"`   // deprecated
	Updater   string `json:"updateBy"`   // deprecated
}

type CollectRulesSummary struct {
	LatestUpdatedAt int64 `json:"latestUpdatedAt"`
	Total           int   `json:"total"`
}

// from pkg/autodiscovery/providers/file.go: configFormat
// format of collectRule.Data
type ConfigFormat struct {
	ADIdentifiers           []string      `json:"adIdentifiers"`
	ClusterCheck            bool          `json:"clusterCheck"`
	InitConfig              interface{}   `json:"initConfig"`
	MetricConfig            interface{}   `json:"jmxMetrics"`
	LogsConfig              interface{}   `json:"logs"`
	Instances               []interface{} `json:"instances"`
	IgnoreAutodiscoveryTags bool          `json:"ignoreAutodiscoveryTags"` // Use to ignore tags coming from autodiscovery
}

// Converts YAML to JSON then uses JSON to unmarshal into ConfigFormat
func ParseConfigFormatYaml(data []byte) (*ConfigFormat, error) {
	var cf ConfigFormat
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal %s", err)
	}

	return &cf, nil
}

func ParseConfigFormatJson(data []byte) (*ConfigFormat, error) {
	var cf ConfigFormat
	err := json.Unmarshal(data, &cf)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal %s", err)
	}

	return &cf, nil
}
