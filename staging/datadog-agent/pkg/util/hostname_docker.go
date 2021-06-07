// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build linux windows darwin
// I don't think windows and darwin can actually be docker hosts
// but keeping it this way for build consistency (for now)

package util

import (
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/hostname"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/hostname/validate"
	"k8s.io/klog/v2"
)

func getContainerHostname() (bool, string) {
	var name string

	// Cluster-agent logic: Kube apiserver
	if getKubeHostname, found := hostname.ProviderCatalog["kube_apiserver"]; found {
		klog.V(5).Info("GetHostname trying Kubernetes trough API server...")
		name, err := getKubeHostname()
		if err == nil && validate.ValidHostname(name) == nil {
			return true, name
		}
	}

	if config.IsContainerized() == false {
		return false, name
	}

	// Node-agent logic: docker or kubelet

	// Docker
	klog.V(5).Info("GetHostname trying Docker API...")
	if getDockerHostname, found := hostname.ProviderCatalog["docker"]; found {
		name, err := getDockerHostname()
		if err == nil && validate.ValidHostname(name) == nil {
			return true, name
		}
	}

	if config.IsKubernetes() == false {
		return false, name
	}
	// Kubelet
	if getKubeletHostname, found := hostname.ProviderCatalog["kubelet"]; found {
		klog.V(5).Info("GetHostname trying Kubernetes trough kubelet API...")
		name, err := getKubeletHostname()
		if err == nil && validate.ValidHostname(name) == nil {
			return true, name
		}
	}
	return false, name
}
