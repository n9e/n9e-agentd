// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build kubelet

package utils

import (
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery/common/types"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/providers/names"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/kubernetes/kubelet"
	"k8s.io/klog/v2"
)

// ConfigsForPod returns the openmetrics configurations for a given pod if it matches the AD configuration
func ConfigsForPod(pc *types.PrometheusCheck, pod *kubelet.Pod) []integration.Config {
	var configs []integration.Config
	namespacedName := fmt.Sprintf("%s/%s", pod.Metadata.Namespace, pod.Metadata.Name)
	if pc.IsExcluded(pod.Metadata.Annotations, namespacedName) {
		return configs
	}

	instances, found := buildInstances(pc, pod.Metadata.Annotations, namespacedName)
	if found {
		for _, container := range pod.Status.GetAllContainers() {
			if !pc.AD.MatchContainer(container.Name) {
				klog.V(5).Infof("Container '%s' doesn't match the AD configuration 'kubernetes_container_names', ignoring it", container.Name)
				continue
			}
			configs = append(configs, integration.Config{
				Name:          openmetricsCheckName,
				InitConfig:    integration.Data(openmetricsInitConfig),
				Instances:     instances,
				Provider:      names.PrometheusPods,
				Source:        "prometheus_pods:" + container.ID,
				ADIdentifiers: []string{container.ID},
			})
		}
		return configs
	}

	// TODO: Support AD matching based on label selectors

	return configs
}
