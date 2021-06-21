package api

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

type CollectRule struct {
	ID   int64  `jso:id`      // option rule.ID
	Name string `json:"name"` // option Config.instances[$] ruleID
	Type string `json:"type"` // required Config.Name checkName
	Data string `json:"data"` // required Config.Instances[$]

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

// CommonInstanceConfig holds the reserved fields for the yaml instance data
type CommonInstanceConfig struct {
	MinCollectionInterval int      `json:"minCollectionInterval"` // This changes the collection interval of the check - default: 15
	EmptyDefaultHostname  bool     `json:"emptyDefaultHostname"`  // This forces the check to send metrics with no hostname. This is useful for cluster-level checks.
	Tags                  []string `json:"tags"`                  // A list of tags to attach to every metric and service check emitted by this instance, <key_1>:<value_1>
	Service               string   `json:"service"`               // Attach the tag `service:<SERVICE>` to every metric, event, and service check emitted by this integration.
	Name                  string   `json:"name"`                  //
	Namespace             string   `json:"namespace"`             //
}

type ScriptCollectFormat struct {
	InitConfig struct {
		Root    string `json:"root"`
		Env     string `json:"env"`
		Timeout int    `json:"timeout"`
	} `json:"initConfig"`
	Instances []struct {
		CommonInstanceConfig
		FilePath string `json:"filePath"`
		Root     string `json:"root"`
		Params   string `json:"params"`
		Env      string `json:"env"`
		Stdin    string `json:"stdin"`
		Timeout  int    `json:"timeout"`
	} `json:"instances"`
}

type PortCollectFormat struct {
	InitConfig struct {
		Timeout int `json:"timeout"`
	} `json:"initConfig"`
	Instances []struct {
		CommonInstanceConfig
		Protocol string `json:"protocol" description:"udp or tcp"`
		Port     int    `json:"port"`
	} `json:"instances"`
}

type LogCollectFormat struct {
	Instances []struct {
		CommonInstanceConfig
		MetricName  string            `json:"metricName"`  //
		FilePath    string            `json:"filePath"`    //
		Pattern     string            `json:"pattern"`     //
		TagsPattern map[string]string `json:"tagsPattern"` //
		Func        string            `json:"func"`        // count(c), histogram(h)
	} `json:"instances"`
}

type ProcCollectFormat struct {
	Instances []struct {
		CommonInstanceConfig
		Target        string `json:"target"`
		CollectMethod string `json:"collectMethod" description:"name or cmdline"`
		Name          string `json:"name"`    // no used
		Comment       string `json:"comment"` // no used
	} `json:"instances"`
}
