// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package config

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/flavor"
	"k8s.io/klog/v2"
)

// DiscoverComponentsFromConfig returns a list of AD Providers and Listeners based on the agent configuration
func (cf Config) DiscoverComponentsFromConfig() ([]ConfigurationProviders, []Listeners) {
	detectedProviders := []ConfigurationProviders{}
	detectedListeners := []Listeners{}

	// Auto-add Prometheus config provider based on `prometheus_scrape.enabled`
	if cf.PrometheusScrape.Enabled {
		var prometheusProvider ConfigurationProviders
		if flavor.GetFlavor() == flavor.ClusterAgent {
			prometheusProvider = ConfigurationProviders{Name: "prometheus_services", Polling: true}
		} else {
			prometheusProvider = ConfigurationProviders{Name: "prometheus_pods", Polling: true}
		}
		klog.Infof("Prometheus scraping is enabled: Adding the Prometheus config provider '%s'", prometheusProvider.Name)
		detectedProviders = append(detectedProviders, prometheusProvider)
	}

	return detectedProviders, detectedListeners
}

// DiscoverComponentsFromEnv returns a list of AD Providers and Listeners based on environment characteristics
func (cf *Config) DiscoverComponentsFromEnv() ([]ConfigurationProviders, []Listeners) {
	detectedProviders := []ConfigurationProviders{}
	detectedListeners := []Listeners{}

	// When using automatic discovery of providers/listeners
	// We automatically activate the environment listener
	detectedListeners = append(detectedListeners, Listeners{Name: "environment"})

	if cf.IsFeaturePresent(Docker) {
		detectedProviders = append(detectedProviders, ConfigurationProviders{Name: "docker", Polling: true, PollInterval: "1s"})
		if !cf.IsFeaturePresent(Kubernetes) {
			detectedListeners = append(detectedListeners, Listeners{Name: "docker"})
			klog.Info("Adding Docker listener from environment")
		}
		klog.Info("Adding Docker provider from environment")
	}

	if cf.IsFeaturePresent(Kubernetes) {
		detectedProviders = append(detectedProviders, ConfigurationProviders{Name: "kubelet", Polling: true})
		detectedListeners = append(detectedListeners, Listeners{Name: "kubelet"})
		klog.Info("Adding Kubelet autodiscovery provider and listener from environment")
	}

	if cf.IsFeaturePresent(ECSFargate) {
		detectedProviders = append(detectedProviders, ConfigurationProviders{Name: "ecs", Polling: true})
		detectedListeners = append(detectedListeners, Listeners{Name: "ecs"})
		klog.Info("Adding ECS Fargate autodiscovery provider and listener from environment")
	}

	return detectedProviders, detectedListeners
}
