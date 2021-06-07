package config

import (
	"strings"

	"k8s.io/klog/v2"
)

// Feature represents a feature of current environment
type Feature string

// FeatureMap represents all detected features
type FeatureMap map[Feature]struct{}

func (fm FeatureMap) String() string {
	features := make([]string, 0, len(fm))
	for f := range fm {
		features = append(features, string(f))
	}

	return strings.Join(features, ",")
}

// GetDetectedFeatures returns all detected features (detection only performed once)
func (cf *Config) GetDetectedFeatures() FeatureMap {
	return cf.Autoconfig.features
}

// IsFeaturePresent returns if a particular feature is activated
func (cf *Config) IsFeaturePresent(feature Feature) bool {
	_, found := cf.Autoconfig.features[feature]
	return found
}

// IsAutoconfigEnabled returns if autoconfig from environment is activated or not
// We cannot rely on Datadog config as this function may be called before configuration is read
func (cf Config) IsAutoconfigEnabled() bool {
	return cf.Autoconfig.Enabled
}

// We guarantee that Datadog configuration is entirely loaded (env + YAML)
// before this function is called
func (cf *Config) detectFeatures() {
	newFeatures := make(FeatureMap)
	if cf.IsAutoconfigEnabled() {
		cf.detectContainerFeatures(newFeatures)
		for _, ef := range cf.Autoconfig.ExcludeFeatures {
			delete(newFeatures, Feature(ef))
		}

		klog.Infof("Features detected from environment: %v", newFeatures)
	}
	cf.Autoconfig.features = newFeatures
}
