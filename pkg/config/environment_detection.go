// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package config

import (
	"regexp"
	"strings"
	"sync"

	"github.com/DataDog/datadog-agent/pkg/util/log"
)

const (
	autoconfEnvironmentVariable = "AUTOCONFIG_FROM_ENVIRONMENT"
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

var (
	knownFeatures    = make(FeatureMap)
	detectedFeatures FeatureMap
	featureLock      sync.RWMutex
)

// GetDetectedFeatures returns all detected features (detection only performed once)
func GetDetectedFeatures() FeatureMap {
	featureLock.RLock()
	defer featureLock.RUnlock()

	if detectedFeatures == nil {
		// If this function is called while feature detection has not run
		// it means Confifguration has not been loaded, which is an unexpected flow in our code
		// It's not useful to do lazy detection as it would also mean Configuration has not been loaded
		panic("Trying to access features before detection has run")
	}

	return detectedFeatures
}

// IsFeaturePresent returns if a particular feature is activated
func IsFeaturePresent(feature Feature) bool {
	featureLock.RLock()
	defer featureLock.RUnlock()

	if detectedFeatures == nil {
		// If this function is called while feature detection has not run
		// it means Confifguration has not been loaded, which is an unexpected flow in our code
		// It's not useful to do lazy detection as it would also mean Configuration has not been loaded
		panic("Trying to access features before detection has run")
	}

	_, found := detectedFeatures[feature]
	return found
}

// IsAutoconfigEnabled returns if autoconfig from environment is activated or not
func IsAutoconfigEnabled() bool {
	return C.AutoconfigFromEnvironment
}

// DetectFeatures runs the feature detection.
// We guarantee that Datadog configuration is entirely loaded (env + YAML)
// before this function is called
func DetectFeatures() {
	featureLock.Lock()
	defer featureLock.Unlock()

	newFeatures := make(FeatureMap)
	if IsAutoconfigEnabled() {
		detectContainerFeatures(newFeatures)
		excludedFeatures := C.AutoconfigExcludeFeatures
		excludeFeatures(newFeatures, excludedFeatures)

		includedFeatures := C.AutoconfigIncludeFeatures
		for _, f := range includedFeatures {
			f = strings.ToLower(f)
			if _, found := knownFeatures[Feature(f)]; found {
				newFeatures[Feature(f)] = struct{}{}
			} else {
				log.Warnf("Unknown feature in autoconfig_include_features: %s - discarding", f)
			}
		}

		log.Infof("Features detected from environment: %v", newFeatures)
	} else {
		log.Warnf("Deactivating Autoconfig will disable most components. It's recommended to use autoconfig_exclude_features and autoconfig_include_features to activate/deactivate features selectively")
	}
	detectedFeatures = newFeatures
}

func excludeFeatures(detectedFeatures FeatureMap, excludedFeatures []string) {
	rFilters := make([]*regexp.Regexp, 0, len(excludedFeatures))

	for _, filter := range excludedFeatures {
		filter = strings.ToLower(strings.TrimPrefix(filter, "name:"))
		r, err := regexp.Compile(filter)
		if err != nil {
			log.Warnf("Unbale to parse exclude feature filter: '%s'", filter)
			continue
		}

		rFilters = append(rFilters, r)
	}

	for feature := range detectedFeatures {
		for _, r := range rFilters {
			if r.MatchString(string(feature)) {
				delete(detectedFeatures, feature)
				break
			}
		}
	}
}

func registerFeature(f Feature) {
	knownFeatures[f] = struct{}{}
}
