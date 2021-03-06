// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package common

import (
	"strings"

	"github.com/DataDog/datadog-agent/pkg/autodiscovery"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/providers"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/scheduler"
	"github.com/n9e/n9e-agentd/pkg/config"
	configad "github.com/n9e/n9e-agentd/pkg/config/autodiscovery"
	"k8s.io/klog/v2"
)

// This is due to an AD limitation that does not allow several listeners to work in parallel
// if they can provide for the same objects.
// When this is solved, we can remove this check and simplify code below
var (
	incompatibleListeners = map[string]map[string]struct{}{
		"kubelet": {"docker": struct{}{}},
		"docker":  {"kubelet": struct{}{}},
	}
)

// TODO
func setupAutoDiscovery(confSearchPaths []string, metaScheduler *scheduler.MetaScheduler) *autodiscovery.AutoConfig {
	ad := autodiscovery.NewAutoConfig(metaScheduler)
	ad.AddConfigProvider(providers.NewFileConfigProvider(confSearchPaths), false, 0)

	cf := config.C

	// Autodiscovery cannot easily use config.RegisterOverrideFunc() due to Unmarshalling
	extraConfigProviders, extraConfigListeners := configad.DiscoverComponentsFromConfig()

	var extraEnvProviders []config.ConfigurationProviders
	var extraEnvListeners []config.Listeners
	if cf.Autoconfig.Enabled && !cf.IsCLCRunner() {
		extraEnvProviders, extraEnvListeners = configad.DiscoverComponentsFromEnv()
	}

	configProviders := cf.ConfigProviders
	if cf.EnableN9eProvider && len(cf.Endpoints) > 0 {
		configProviders = append(configProviders, config.ConfigurationProviders{
			Name:        "http",
			Polling:     true,
			TemplateURL: strings.Join(cf.Endpoints, ","),
			Token:       cf.ApiKey,
		})
	}

	// Register additional configuration providers
	var uniqueConfigProviders map[string]config.ConfigurationProviders
	uniqueConfigProviders = make(map[string]config.ConfigurationProviders, len(configProviders)+len(extraEnvProviders)+len(configProviders))
	for _, provider := range configProviders {
		uniqueConfigProviders[provider.Name] = provider
	}

	// Add extra config providers
	for _, name := range cf.ExtraConfigProviders {
		if _, found := uniqueConfigProviders[name]; !found {
			uniqueConfigProviders[name] = config.ConfigurationProviders{Name: name, Polling: true}
		} else {
			klog.Infof("Duplicate AD provider from extraConfigProviders discarded as already present in configProviders: %s", name)
		}
	}

	for _, provider := range extraConfigProviders {
		if _, found := uniqueConfigProviders[provider.Name]; !found {
			uniqueConfigProviders[provider.Name] = provider
		}
	}

	for _, provider := range extraEnvProviders {
		if _, found := uniqueConfigProviders[provider.Name]; !found {
			uniqueConfigProviders[provider.Name] = provider
		}
	}

	// Adding all found providers
	for _, cp := range uniqueConfigProviders {
		factory, found := providers.ProviderCatalog[cp.Name]
		if found {
			configProvider, err := factory(cp)
			if err != nil {
				klog.Errorf("Error while adding config provider %v: %v", cp.Name, err)
				continue
			}

			pollInterval := providers.GetPollInterval(cp)
			if cp.Polling {
				klog.Infof("Registering %s config provider polled every %s", cp.Name, pollInterval.String())
			} else {
				klog.Infof("Registering %s config provider", cp.Name)
			}
			ad.AddConfigProvider(configProvider, cp.Polling, pollInterval)
		} else {
			klog.Errorf("Unable to find this provider in the catalog: %v", cp.Name)
		}
	}

	listeners := cf.Listeners
	// Add extra listeners
	for _, name := range cf.ExtraListeners {
		listeners = append(listeners, config.Listeners{Name: name})
	}

	for _, listener := range extraConfigListeners {
		alreadyPresent := false
		for _, existingListener := range listeners {
			if listener.Name == existingListener.Name {
				alreadyPresent = true
				break
			}
		}

		if !alreadyPresent {
			listeners = append(listeners, listener)
		}
	}

	// For extraEnvListeners, we need to check incompatibleListeners to avoid generation of duplicate checks
	for _, listener := range extraEnvListeners {
		skipListener := false
		incomp := incompatibleListeners[listener.Name]

		for _, existingListener := range listeners {
			if listener.Name == existingListener.Name {
				skipListener = true
				break
			}

			if _, found := incomp[existingListener.Name]; found {
				klog.V(5).Infof("Discarding discovered listener: %s as incompatible with listener from config: %s", listener.Name, existingListener.Name)
				break
			}
		}

		if !skipListener {
			listeners = append(listeners, listener)
		}
	}

	ad.AddListeners(listeners)

	return ad
}

// StartAutoConfig starts auto discovery
func StartAutoConfig() {
	AC.LoadAndRun()
}
