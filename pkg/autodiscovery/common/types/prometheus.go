// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package types

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/klog/v2"
)

const (
	// Default openmetrics check configuration values
	openmetricsURLPrefix   = "http://%%host%%:"
	openmetricsDefaultPort = "%%port%%"
	openmetricsDefaultPath = "/metrics"
	openmetricsDefaultNS   = ""

	// PrometheusScrapeAnnotation standard Prometheus scrape annotation key
	PrometheusScrapeAnnotation = "prometheus.io/scrape"
	// PrometheusPathAnnotation standard Prometheus path annotation key
	PrometheusPathAnnotation = "prometheus.io/path"
	// PrometheusPortAnnotation standard Prometheus port annotation key
	PrometheusPortAnnotation = "prometheus.io/port"
)

var (
	// PrometheusStandardAnnotations contains the standard Prometheus AD annotations
	PrometheusStandardAnnotations = []string{
		PrometheusScrapeAnnotation,
		PrometheusPathAnnotation,
		PrometheusPortAnnotation,
	}
	openmetricsDefaultMetrics = []string{"*"}
)

// PrometheusCheck represents the openmetrics check instances and the corresponding autodiscovery rules
type PrometheusCheck struct {
	Instances []*OpenmetricsInstance `mapstructure:"configurations" yaml:"configurations,omitempty" json:"configurations"`
	AD        *ADConfig              `mapstructure:"autodiscovery" yaml:"autodiscovery,omitempty" json:"autodiscovery"`
}

// OpenmetricsInstance contains the openmetrics check instance fields
type OpenmetricsInstance struct {
	URL                           string                      `mapstructure:"prometheusUrl" yaml:"prometheusUrl,omitempty" json:"prometheusUrl,omitempty"`
	Namespace                     string                      `mapstructure:"namespace" yaml:"namespace,omitempty" json:"namespace"`
	Metrics                       []string                    `mapstructure:"metrics" yaml:"metrics,omitempty" json:"metrics,omitempty"`
	Prefix                        string                      `mapstructure:"prometheusMetricsPrefix" yaml:"prometheusMetricsPrefix,omitempty" json:"prometheusMetricsPrefix,omitempty"`
	HealthCheck                   bool                        `mapstructure:"healthServiceCheck" yaml:"healthServiceCheck,omitempty" json:"healthServiceCheck,omitempty"`
	LabelToHostname               bool                        `mapstructure:"labelToHostname" yaml:"labelToHostname,omitempty" json:"labelToHostname,omitempty"`
	LabelJoins                    map[string]LabelJoinsConfig `mapstructure:"labelJoins" yaml:"labelJoins,omitempty" json:"labelJoins,omitempty"`
	LabelsMapper                  map[string]string           `mapstructure:"labelsMapper" yaml:"labelsMapper,omitempty" json:"labelsMapper,omitempty"`
	TypeOverride                  map[string]string           `mapstructure:"typeOverrides" yaml:"typeOverrides,omitempty" json:"typeOverrides,omitempty"`
	HistogramBuckets              bool                        `mapstructure:"sendHistogramsBuckets" yaml:"sendHistogramsBuckets,omitempty" json:"sendHistogramsBuckets,omitempty"`
	DistributionBuckets           bool                        `mapstructure:"sendDistributionBuckets" yaml:"sendDistributionBuckets,omitempty" json:"sendDistributionBuckets,omitempty"`
	MonotonicCounter              bool                        `mapstructure:"sendMonotonicCounter" yaml:"sendMonotonicCounter,omitempty" json:"sendMonotonicCounter,omitempty"`
	DistributionCountsAsMonotonic bool                        `mapstructure:"sendDistributionCountsAsMonotonic" yaml:"sendDistributionCountsAsMonotonic,omitempty" json:"sendDistributionCountsAsMonotonic,omitempty"`
	DistributionSumsAsMonotonic   bool                        `mapstructure:"sendDistributionSumsAsMonotonic" yaml:"sendDistributionSumsAsMonotonic,omitempty" json:"sendDistributionSumsAsMonotonic,omitempty"`
	ExcludeLabels                 []string                    `mapstructure:"excludeLabels" yaml:"excludeLabels,omitempty" json:"excludeLabels,omitempty"`
	BearerTokenAuth               bool                        `mapstructure:"bearerTokenAuth" yaml:"bearerTokenAuth,omitempty" json:"bearerTokenAuth,omitempty"`
	BearerTokenPath               string                      `mapstructure:"bearerTokenPath" yaml:"bearerTokenPath,omitempty" json:"bearerTokenPath,omitempty"`
	IgnoreMetrics                 []string                    `mapstructure:"ignoreMetrics" yaml:"ignoreMetrics,omitempty" json:"ignoreMetrics,omitempty"`
	Proxy                         map[string]string           `mapstructure:"proxy" yaml:"proxy,omitempty" json:"proxy,omitempty"`
	SkipProxy                     bool                        `mapstructure:"skipProxy" yaml:"skipProxy,omitempty" json:"skipProxy,omitempty"`
	Username                      string                      `mapstructure:"username" yaml:"username,omitempty" json:"username,omitempty"`
	Password                      string                      `mapstructure:"password" yaml:"password,omitempty" json:"password,omitempty"`
	TLSVerify                     bool                        `mapstructure:"tlsVerify" yaml:"tlsVerify,omitempty" json:"tlsVerify,omitempty"`
	TLSHostHeader                 bool                        `mapstructure:"tlsUseHostHeader" yaml:"tlsUseHostHeader,omitempty" json:"tlsUseHostHeader,omitempty"`
	TLSIgnoreWarn                 bool                        `mapstructure:"tlsIgnoreWarning" yaml:"tlsIgnoreWarning,omitempty" json:"tlsIgnoreWarning,omitempty"`
	TLSCert                       string                      `mapstructure:"tlsCert" yaml:"tlsCert,omitempty" json:"tlsCert,omitempty"`
	TLSPrivateKey                 string                      `mapstructure:"tlsPrivateKey" yaml:"tlsPrivateKey,omitempty" json:"tlsPrivateKey,omitempty"`
	TLSCACert                     string                      `mapstructure:"tlsCaCert" yaml:"tlsCaCert,omitempty" json:"tlsCaCert,omitempty"`
	Headers                       map[string]string           `mapstructure:"headers" yaml:"headers,omitempty" json:"headers,omitempty"`
	ExtraHeaders                  map[string]string           `mapstructure:"extraHeaders" yaml:"extraHeaders,omitempty" json:"extraHeaders,omitempty"`
	Timeout                       int                         `mapstructure:"timeout" yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Tags                          []string                    `mapstructure:"tags" yaml:"tags,omitempty" json:"tags,omitempty"`
	Service                       string                      `mapstructure:"service" yaml:"service,omitempty" json:"service,omitempty"`
	MinCollectInterval            int                         `mapstructure:"minCollectionInterval" yaml:"minCollectionInterval,omitempty" json:"minCollectionInterval,omitempty"`
	EmptyDefaultHost              bool                        `mapstructure:"emptyDefaultHostname" yaml:"emptyDefaultHostname,omitempty" json:"emptyDefaultHostname,omitempty"`
}

// LabelJoinsConfig contains the label join configuration fields
type LabelJoinsConfig struct {
	LabelsToMatch []string `mapstructure:"labelsToMatch" yaml:"labelsToMatch,omitempty" json:"labelsToMatch"`
	LabelsToGet   []string `mapstructure:"labelsToGet" yaml:"labelsToGet,omitempty" json:"labelsToGet"`
}

// ADConfig contains the autodiscovery configuration data for a PrometheusCheck
type ADConfig struct {
	KubeAnnotations    *InclExcl      `mapstructure:"kubernetesAnnotations,omitempty" yaml:"kubernetesAnnotations,omitempty" json:"kubernetesAnnotations,omitempty"`
	KubeContainerNames []string       `mapstructure:"kubernetesContainerNames,omitempty" yaml:"kubernetesContainerNames,omitempty" json:"kubernetesContainerNames,omitempty"`
	ContainersRe       *regexp.Regexp `mapstructure:",omitempty" yaml:",omitempty"`
}

// InclExcl contains the include/exclude data structure
type InclExcl struct {
	Incl map[string]string `mapstructure:"include" yaml:"include,omitempty" json:"include,omitempty"`
	Excl map[string]string `mapstructure:"exclude" yaml:"exclude,omitempty" json:"exclude,omitempty"`
}

// Init prepares the PrometheusCheck structure and defaults its values
// init must be called only once
func (pc *PrometheusCheck) Init() error {
	pc.initInstances()
	return pc.initAD()
}

// initInstances defaults the Instances field in PrometheusCheck
func (pc *PrometheusCheck) initInstances() {
	if len(pc.Instances) == 0 {
		// Put a default config
		pc.Instances = append(pc.Instances, &OpenmetricsInstance{
			Metrics:   openmetricsDefaultMetrics,
			Namespace: openmetricsDefaultNS,
		})
		return
	}

	for _, instance := range pc.Instances {
		// Default the required config values if not set
		if len(instance.Metrics) == 0 {
			instance.Metrics = openmetricsDefaultMetrics
		}
	}
}

// initAD defaults the AD field in PrometheusCheck
// It also prepares the regex to match the containers by name
func (pc *PrometheusCheck) initAD() error {
	if pc.AD == nil {
		pc.AD = &ADConfig{}
	}

	pc.AD.defaultAD()
	return pc.AD.setContainersRegex()
}

// IsExcluded returns whether is the annotations match an AD exclusion rule
func (pc *PrometheusCheck) IsExcluded(annotations map[string]string, namespacedName string) bool {
	for k, v := range pc.AD.KubeAnnotations.Excl {
		if annotations[k] == v {
			klog.V(5).Infof("'%s' matched the exclusion annotation '%s=%s' ignoring it", namespacedName, k, v)
			return true
		}
	}
	return false
}

// GetIncludeAnnotations returns the AD include annotations
func (ad *ADConfig) GetIncludeAnnotations() map[string]string {
	annotations := map[string]string{}
	if ad.KubeAnnotations != nil && ad.KubeAnnotations.Incl != nil {
		return ad.KubeAnnotations.Incl
	}
	return annotations
}

// GetExcludeAnnotations returns the AD exclude annotations
func (ad *ADConfig) GetExcludeAnnotations() map[string]string {
	annotations := map[string]string{}
	if ad.KubeAnnotations != nil && ad.KubeAnnotations.Excl != nil {
		return ad.KubeAnnotations.Excl
	}
	return annotations
}

// defaultAD defaults the values of the autodiscovery structure
func (ad *ADConfig) defaultAD() {
	if ad.KubeContainerNames == nil {
		ad.KubeContainerNames = []string{}
	}

	if ad.KubeAnnotations == nil {
		ad.KubeAnnotations = &InclExcl{
			Excl: map[string]string{PrometheusScrapeAnnotation: "false"},
			Incl: map[string]string{PrometheusScrapeAnnotation: "true"},
		}
		return
	}

	if ad.KubeAnnotations.Excl == nil {
		ad.KubeAnnotations.Excl = map[string]string{PrometheusScrapeAnnotation: "false"}
	}

	if ad.KubeAnnotations.Incl == nil {
		ad.KubeAnnotations.Incl = map[string]string{PrometheusScrapeAnnotation: "true"}
	}
}

// setContainersRegex precompiles the regex to match the container names for autodiscovery
// returns an error if the container names cannot be converted to a valid regex
func (ad *ADConfig) setContainersRegex() error {
	ad.ContainersRe = nil
	if len(ad.KubeContainerNames) == 0 {
		return nil
	}

	regexString := strings.Join(ad.KubeContainerNames, "|")
	re, err := regexp.Compile(regexString)
	if err != nil {
		return fmt.Errorf("Invalid container names - regex: '%s': %v", regexString, err)
	}

	ad.ContainersRe = re
	return nil
}

// MatchContainer returns whether a container name matches the 'kubernetes_container_names' configuration
func (ad *ADConfig) MatchContainer(name string) bool {
	if ad.ContainersRe == nil {
		return true
	}
	return ad.ContainersRe.MatchString(name)
}

// DefaultPrometheusCheck has the default openmetrics check values
// To be used when the checks configuration is empty
var DefaultPrometheusCheck = &PrometheusCheck{
	Instances: []*OpenmetricsInstance{
		{
			Metrics:   openmetricsDefaultMetrics,
			Namespace: openmetricsDefaultNS,
		},
	},
	AD: &ADConfig{
		KubeAnnotations: &InclExcl{
			Excl: map[string]string{PrometheusScrapeAnnotation: "false"},
			Incl: map[string]string{PrometheusScrapeAnnotation: "true"},
		},
		KubeContainerNames: []string{},
	},
}

// BuildURL returns the 'prometheus_url' based on the default values
// and the prometheus path and port annotations
func BuildURL(annotations map[string]string) string {
	port := openmetricsDefaultPort
	if portFromAnnotation, found := annotations[PrometheusPortAnnotation]; found {
		port = portFromAnnotation
	}

	path := openmetricsDefaultPath
	if pathFromAnnotation, found := annotations[PrometheusPathAnnotation]; found {
		path = pathFromAnnotation
	}

	return openmetricsURLPrefix + port + path
}

// PrometheusAnnotations abstracts a map of prometheus annotations
type PrometheusAnnotations map[string]string

// IsMatchingAnnotations returns whether annotations matches the AD include rules for Prometheus
func (a PrometheusAnnotations) IsMatchingAnnotations(svcAnnotations map[string]string) bool {
	for k, v := range a {
		if svcAnnotations[k] == v {
			return true
		}
	}
	return false
}

// AnnotationsDiffer returns whether the Prometheus AD include annotations have changed
func (a PrometheusAnnotations) AnnotationsDiffer(first, second map[string]string) bool {
	for k := range a {
		if first[k] != second[k] {
			return true
		}
	}
	return false
}
