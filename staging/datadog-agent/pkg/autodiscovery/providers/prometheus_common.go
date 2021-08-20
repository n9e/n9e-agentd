// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package providers

import (
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/common/types"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/n9e/n9e-agentd/pkg/config"
)

// getPrometheusConfigs reads and initializes the openmetrics checks from the configuration
// It defines a default openmetrics instances with default AD if the checks configuration is empty
func getPrometheusConfigs() ([]*types.PrometheusCheck, error) {
	checks := config.C.PrometheusScrape.Checks

	if len(checks) == 0 {
		log.Info("The 'prometheus_scrape.checks' configuration is empty, a default openmetrics check configuration will be used")
		return []*types.PrometheusCheck{types.DefaultPrometheusCheck}, nil
	}

	validChecks := []*types.PrometheusCheck{}
	for i, check := range checks {
		if err := check.Init(); err != nil {
			log.Errorf("Ignoring check configuration (# %d): %v", i+1, err)
			continue
		}
		validChecks = append(validChecks, check)
	}

	return validChecks, nil
}
