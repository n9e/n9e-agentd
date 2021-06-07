package api

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

// format of collectRule.Data
type CollectRuleConfig struct {
	ADIdentifiers           []string      `json:"adIdentifiers" yaml:"adIdentifiers"`
	ClusterCheck            bool          `json:"clusterCheck" yaml:"clusterCheck"`
	InitConfig              interface{}   `yaml:"initConfig"`
	MetricConfig            interface{}   `json:"jmxMetrics" yaml:"jmxMetrics"`
	LogsConfig              interface{}   `json:"logs" yaml:"logs"`
	Instances               []interface{} `json:"instances" yaml:"instances"`
	DockerImages            []string      `json:"dockerImages" yaml:"dockerImages"`                       // Only imported for deprecation warning
	IgnoreAutodiscoveryTags bool          `json:"ignoreAutodiscoveryTags" yaml:"ignoreAutodiscoveryTags"` // Use to ignore tags coming from autodiscovery
}
