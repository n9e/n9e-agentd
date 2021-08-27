package api

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

type SettingInput struct {
	Setting string `param:"path"`
	Value   string `param:"query"`
}

type StatsdReplayInput struct {
	ReplayFile string `param:"query" flag:"file,d" description:"Input file with TCP traffic to replay."`
	TaggerFile string `param:"query" flag:"tagger" description:"Input file with TCP traffic to replay."`
}

type StatsdCaptureTriggerInput struct {
	Duration   int  `json:"duration" flag:"duration,d" default:"60" description:"capture duration (second)"`
	Compressed bool `json:"compressed" flag:"compressed,z" default:"true" description:"Should capture be zstd compressed."`
}

func (p *StatsdCaptureTriggerInput) Validate() error {
	if p.Duration <= 0 {
		p.Duration = 60
	}

	return nil
}

type CollectRule struct {
	ID   int64  `json:"id"`   // option rule.ID
	Name string `json:"name"` // option Config.instances[$] ruleID
	Type string `json:"type"` // required Config.Name checkName
	Data string `json:"data"` // required Config.Instances[$]

	Interval  int    `json:"-"` // -> Config.Instances[$].MinCollectionInterval
	Tags      string `json:"-"` // -> Config.Instances[$].Tags   a:b,b:c
	CreatedAt int64  `json:"-"` // deprecated
	UpdatedAt int64  `json:"-"` // deprecated
	Creator   string `json:"-"` // deprecated
	Updater   string `json:"-"` // deprecated
}

type CollectRuleWrap struct {
	Data []CollectRule `json:"dat"`
	Err  string        `json:"err"`
}

type CollectRulesSummary struct {
	LatestUpdatedAt int64 `json:"latest_updated_at"`
	Total           int   `json:"total"`
}

type CollectRulesSummaryWrap struct {
	Data CollectRulesSummary `json:"dat"`
	Err  string              `json:"err"`
}

// from pkg/autodiscovery/providers/file.go: configFormat
// format of collectRule.Data
type ConfigFormat struct {
	ADIdentifiers           []string      `json:"ad_identifiers"`
	ClusterCheck            bool          `json:"cluster_check"`
	InitConfig              interface{}   `json:"init_config"`
	MetricConfig            interface{}   `json:"jmx_metrics"`
	LogsConfig              interface{}   `json:"logs"`
	Instances               []interface{} `json:"instances"`
	IgnoreAutodiscoveryTags bool          `json:"ignore_autodiscovery_tags"` // Use to ignore tags coming from autodiscovery
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
	MinCollectionInterval int      `json:"min_collection_interval"` // This changes the collection interval of the check - default: 15
	EmptyDefaultHostname  bool     `json:"empty_default_hostname"`  // This forces the check to send metrics with no hostname. This is useful for cluster-level checks.
	Tags                  []string `json:"tags"`                    // A list of tags to attach to every metric and service check emitted by this instance, <key_1>:<value_1>
	Service               string   `json:"service"`                 // Attach the tag `service:<SERVICE>` to every metric, event, and service check emitted by this integration.
	Name                  string   `json:"name"`                    //
	Namespace             string   `json:"namespace"`               //
}

type ScriptCollectFormat struct {
	InitConfig struct {
		Root    string `json:"root"`
		Env     string `json:"env"`
		Timeout int    `json:"timeout"`
	} `json:"init_config"`
	Instances []struct {
		CommonInstanceConfig
		FilePath string `json:"file_path"`
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
	} `json:"init_config"`
	Instances []struct {
		CommonInstanceConfig
		Protocol string `json:"protocol" description:"udp or tcp"`
		Port     int    `json:"port"`
	} `json:"instances"`
}

type LogCollectFormat struct {
	Instances []struct {
		CommonInstanceConfig
		MetricName  string            `json:"metric_name"`  //
		FilePath    string            `json:"file_path"`    //
		Pattern     string            `json:"pattern"`      //
		TagsPattern map[string]string `json:"tags_pattern"` //
		Func        string            `json:"func"`         // count(c), histogram(h)
	} `json:"instances"`
}

type ProcCollectFormat struct {
	Instances []struct {
		CommonInstanceConfig
		Target        string `json:"target"`
		CollectMethod string `json:"collect_method" description:"name or cmdline"`
		Name          string `json:"name"`    // no used
		Comment       string `json:"comment"` // no used
	} `json:"instances"`
}

// Status represents the current status of registered components
// it is built and returned by GetStatus()
type HealthStatus struct {
	Healthy   []string
	Unhealthy []string
}
