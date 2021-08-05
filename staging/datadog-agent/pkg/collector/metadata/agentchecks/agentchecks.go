// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package agentchecks

import (
	"encoding/json"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/runner"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metadata/common"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metadata/externalhost"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metadata/host"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/status"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util"
	"k8s.io/klog/v2"
)

// GetPayload builds a payload of all the agentchecks metadata
func GetPayload() *Payload {
	agentChecksPayload := ACPayload{}
	hostnameData, _ := util.GetHostnameData()
	hostname := hostnameData.Hostname
	checkStats := runner.GetCheckStats()
	jmxStartupError := status.GetJMXStartupError()

	for _, stats := range checkStats {
		for _, s := range stats {
			var status []interface{}
			if s.LastError != "" {
				status = []interface{}{
					s.CheckName, s.CheckName, s.CheckID, "ERROR", s.LastError, "",
				}
			} else if len(s.LastWarnings) != 0 {
				status = []interface{}{
					s.CheckName, s.CheckName, s.CheckID, "WARNING", s.LastWarnings, "",
				}
			} else {
				status = []interface{}{
					s.CheckName, s.CheckName, s.CheckID, "OK", "", "",
				}
			}
			if status != nil {
				agentChecksPayload.AgentChecks = append(agentChecksPayload.AgentChecks, status)
			}
		}
	}

	loaderErrors := collector.GetLoaderErrors()

	for check, errs := range loaderErrors {
		jsonErrs, err := json.Marshal(errs)
		if err != nil {
			klog.Warningf("Error formatting loader error from check %s: %v", check, err)
		}
		status := []interface{}{
			check, check, "initialization", "ERROR", string(jsonErrs),
		}
		agentChecksPayload.AgentChecks = append(agentChecksPayload.AgentChecks, status)
	}

	configErrors := autodiscovery.GetConfigErrors()

	for check, e := range configErrors {
		status := []interface{}{
			check, check, "initialization", "ERROR", e,
		}
		agentChecksPayload.AgentChecks = append(agentChecksPayload.AgentChecks, status)
	}

	if jmxStartupError.LastError != "" {
		status := []interface{}{
			"jmx", "jmx", "initialization", "ERROR", jmxStartupError.LastError,
		}
		agentChecksPayload.AgentChecks = append(agentChecksPayload.AgentChecks, status)
	}

	// Grab the non agent checks information
	metaPayload := host.GetMeta(hostnameData)
	metaPayload.Hostname = hostname
	cp := common.GetPayload(hostname)
	ehp := externalhost.GetPayload()
	payload := &Payload{
		CommonPayload{*cp},
		MetaPayload{*metaPayload},
		agentChecksPayload,
		ExternalHostPayload{*ehp},
	}

	return payload
}