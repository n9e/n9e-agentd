// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package utils

import (
	"encoding/json"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery/common/types"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"k8s.io/klog/v2"
)

const (
	openmetricsCheckName  = "openmetrics"
	openmetricsInitConfig = "{}"
)

// buildInstances generates check config instances based on the Prometheus config and the object annotations
// The second returned value is true if more than one instance is found
func buildInstances(pc *types.PrometheusCheck, annotations map[string]string, namespacedName string) ([]integration.Data, bool) {
	instances := []integration.Data{}
	for k, v := range pc.AD.KubeAnnotations.Incl {
		if annotations[k] == v {
			klog.V(5).Infof("'%s' matched the annotation '%s=%s' to schedule an openmetrics check", namespacedName, k, v)
			for _, instance := range pc.Instances {
				instanceValues := *instance
				if instanceValues.URL == "" {
					instanceValues.URL = types.BuildURL(annotations)
				}
				instanceJSON, err := json.Marshal(instanceValues)
				if err != nil {
					klog.Warningf("Error processing prometheus configuration: %v", err)
					continue
				}
				instances = append(instances, instanceJSON)
			}
			return instances, len(instances) > 0
		}
	}

	return instances, false
}
