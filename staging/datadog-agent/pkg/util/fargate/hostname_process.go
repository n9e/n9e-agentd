// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build fargateprocess

package fargate

import (
	"errors"
	"fmt"

	ecsmeta "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/ecs/metadata"
	"k8s.io/klog/v2"
)

// GetFargateHost returns the hostname to be used
// by the process Agent based on the Fargate orchestrator
// - ECS: fargate_task:<TaskARN>
// - EKS: value of kubernetes_kubelet_nodename
func GetFargateHost() (string, error) {
	return getFargateHost(GetOrchestrator(), getECSHost, getEKSHost)
}

// getFargateHost is separated from GetFargateHost for testing purpose
func getFargateHost(orchestrator OrchestratorName, ecsFunc, eksFunc func() (string, error)) (string, error) {
	// Fargate should have no concept of host names
	// we set the hostname depending on the orchestrator
	switch orchestrator {
	case ECS:
		return ecsFunc()
	case EKS:
		return eksFunc()
	}
	return "", errors.New("unknown Fargate orchestrator")
}

func getECSHost() (string, error) {
	client, err := ecsmeta.V2()
	if err != nil {
		klog.V(5).Infof("error while initializing ECS metadata V2 client: %s", err)
		return "", err
	}

	// Use the task ARN as hostname
	taskMeta, err := client.GetTask()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("fargate_task:%s", taskMeta.TaskARN), nil
}

func getEKSHost() (string, error) {
	// Use the node name as hostname
	return GetEKSFargateNodename()
}