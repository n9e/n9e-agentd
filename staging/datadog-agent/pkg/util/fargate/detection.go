// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package fargate

import (
	"errors"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/ecs"
)

// IsFargateInstance returns whether the Agent is running in Fargate.
func IsFargateInstance() bool {
	return ecs.IsFargateInstance() || IsEKSFargateInstance()
}

// GetOrchestrator returns whether the Agent is running on ECS or EKS.
func GetOrchestrator() OrchestratorName {
	if IsEKSFargateInstance() {
		return EKS
	}
	if ecs.IsFargateInstance() {
		return ECS
	}
	return Unknown
}

// IsEKSFargateInstance returns whether the Agent is running in EKS Fargate.
func IsEKSFargateInstance() bool {
	return config.C.EKSFargate
}

// GetEKSFargateNodename returns the node name in EKS Fargate
func GetEKSFargateNodename() (string, error) {
	if nodename := config.C.KubernetesKubeletNodename; nodename != "" {
		return nodename, nil
	}
	return "", errors.New("kubeletNodename is not defined")
}
