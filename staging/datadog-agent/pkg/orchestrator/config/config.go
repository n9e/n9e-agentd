// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package config

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/DataDog/datadog-agent/pkg/forwarder"
	"github.com/DataDog/datadog-agent/pkg/orchestrator/redact"
	apicfg "github.com/DataDog/datadog-agent/pkg/process/util/api/config"
	coreutil "github.com/DataDog/datadog-agent/pkg/util"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/clustername"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/config/flavor"
)

const (
	orchestratorNS  = "orchestrator_explorer"
	processNS       = "process_config"
	defaultEndpoint = "https://orchestrator.datadoghq.com"
	maxMessageBatch = 100
)

// OrchestratorConfig is the global config for the Orchestrator related packages. This information
// is sourced from config files and the environment variables.
type OrchestratorConfig struct {
	OrchestrationCollectionEnabled bool
	KubeClusterName                string
	IsScrubbingEnabled             bool
	Scrubber                       *redact.DataScrubber
	OrchestratorEndpoints          []apicfg.Endpoint
	MaxPerMessage                  int
	PodQueueBytes                  int // The total number of bytes that can be enqueued for delivery to the orchestrator endpoint
	ExtraTags                      []string
}

// NewDefaultOrchestratorConfig returns an NewDefaultOrchestratorConfig using a configuration file. It can be nil
// if there is no file available. In this case we'll configure only via environment.
func NewDefaultOrchestratorConfig() *OrchestratorConfig {
	orchestratorEndpoint, err := url.Parse(defaultEndpoint)
	if err != nil {
		// This is a hardcoded URL so parsing it should not fail
		panic(err)
	}
	oc := OrchestratorConfig{
		Scrubber:              redact.NewDefaultDataScrubber(),
		MaxPerMessage:         100,
		OrchestratorEndpoints: []apicfg.Endpoint{{Endpoint: orchestratorEndpoint}},
		PodQueueBytes:         15 * 1000 * 1000,
	}
	return &oc
}

func key(pieces ...string) string {
	return strings.Join(pieces, ".")
}

// Load load orchestrator-specific configuration
// at this point secrets should already be resolved by the core/process/cluster agent
func (oc *OrchestratorConfig) Load() error {
	URL, err := extractOrchestratorDDUrl()
	if err != nil {
		return err
	}
	oc.OrchestratorEndpoints[0].Endpoint = URL
	cf := config.C.OrchestratorExplorer

	if key := config.C.ApiKey; len(key) > 0 {
		oc.OrchestratorEndpoints[0].APIKey = key
	}

	if err := extractOrchestratorAdditionalEndpoints(URL, &oc.OrchestratorEndpoints); err != nil {
		return err
	}

	// A custom word list to enhance the default one used by the DataScrubber
	if v := cf.CustomSensitiveWords; len(v) > 0 {
		oc.Scrubber.AddCustomSensitiveWords(v)
	}

	// The maximum number of pods, nodes, replicaSets, deployments and services per message. Note: Only change if the defaults are causing issues.
	if v := cf.MaxPerMessage; v > 0 {
		if v <= maxMessageBatch {
			oc.MaxPerMessage = v
		} else if v > 0 {
			log.Warn("Overriding the configured item count per message limit because it exceeds maximum")
		}
	}

	if v := cf.PodQueueBytes; v > 0 {
		oc.PodQueueBytes = v
	}

	// Orchestrator Explorer
	if cf.Enabled {
		oc.OrchestrationCollectionEnabled = true
		// Set clustername
		hostname, _ := coreutil.GetHostname(context.TODO())
		if clusterName := clustername.GetClusterName(context.TODO(), hostname); clusterName != "" {
			oc.KubeClusterName = clusterName
		}
	}
	oc.IsScrubbingEnabled = cf.ContainerScrubbingEnabled
	oc.ExtraTags = cf.ExtraTags

	return nil
}

func extractOrchestratorAdditionalEndpoints(URL *url.URL, orchestratorEndpoints *[]apicfg.Endpoint) error {
	if v := config.C.OrchestratorExplorer.AdditionalEndpoints; len(v) > 0 {
		if err := extractEndpoints(URL, v, orchestratorEndpoints); err != nil {
			return err
		}
	} else if v := config.C.ProcessConfig.OrchestratorAdditionalEndpoints; len(v) > 0 {
		if err := extractEndpoints(URL, v, orchestratorEndpoints); err != nil {
			return err
		}
	}
	return nil
}

func extractEndpoints(URL *url.URL, in map[string][]string, endpoints *[]apicfg.Endpoint) error {
	for endpointURL, apiKeys := range in {
		u, err := URL.Parse(endpointURL)
		if err != nil {
			return fmt.Errorf("invalid additional endpoint url '%s': %s", endpointURL, err)
		}
		for _, k := range apiKeys {
			*endpoints = append(*endpoints, apicfg.Endpoint{
				APIKey:   config.SanitizeAPIKey(k),
				Endpoint: u,
			})
		}
	}
	return nil
}

// extractOrchestratorDDUrl contains backward compatible config parsing code.
func extractOrchestratorDDUrl() (*url.URL, error) {
	u := config.C.OrchestratorExplorer.Url
	if u == "" {
		u = config.C.ProcessConfig.Url
	}
	URL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("error parsing orchestrator_dd_url: %s", err)
	}
	return URL, nil
}

// NewOrchestratorForwarder returns an orchestratorForwarder
// if the feature is activated on the cluster-agent/cluster-check runner, nil otherwise
func NewOrchestratorForwarder() *forwarder.DefaultForwarder {
	if !config.C.OrchestratorExplorer.Enabled {
		return nil
	}
	if config.GetFlavor() == flavor.DefaultAgent && !config.IsCLCRunner() {
		return nil
	}
	orchestratorCfg := NewDefaultOrchestratorConfig()
	if err := orchestratorCfg.Load(); err != nil {
		log.Errorf("Error loading the orchestrator config: %s", err)
	}
	keysPerDomain := apicfg.KeysPerDomains(orchestratorCfg.OrchestratorEndpoints)
	orchestratorForwarderOpts := forwarder.NewOptions(keysPerDomain)
	orchestratorForwarderOpts.DisableAPIKeyChecking = true

	return forwarder.NewDefaultForwarder(orchestratorForwarderOpts)
}
