// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package util

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metadata/inventories"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/alibaba"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/azure"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/ec2"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/ecs"
	ecscommon "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/ecs/common"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/gce"
	"k8s.io/klog/v2"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/tencent"
)

type cloudProviderDetector struct {
	name     string
	callback func() bool
}

type cloudProviderNTPDetector struct {
	name     string
	callback func() []string
}

// DetectCloudProvider detects the cloud provider where the agent is running in order:
// * AWS ECS/Fargate
// * AWS EC2
// * GCE
// * Azure
// * Alibaba
// * Tencent
func DetectCloudProvider() {
	detectors := []cloudProviderDetector{
		{name: ecscommon.CloudProviderName, callback: ecs.IsRunningOn},
		{name: ec2.CloudProviderName, callback: ec2.IsRunningOn},
		{name: gce.CloudProviderName, callback: gce.IsRunningOn},
		{name: azure.CloudProviderName, callback: azure.IsRunningOn},
		{name: alibaba.CloudProviderName, callback: alibaba.IsRunningOn},
		{name: tencent.CloudProviderName, callback: tencent.IsRunningOn},
	}

	for _, cloudDetector := range detectors {
		if cloudDetector.callback() {
			inventories.SetAgentMetadata(inventories.CloudProviderMetatadaName, cloudDetector.name)
			klog.Infof("Cloud provider %s detected", cloudDetector.name)
			return
		}
	}
	klog.Info("No cloud provider detected")
}

// GetCloudProviderNTPHosts detects the cloud provider where the agent is running in order and returns its NTP host name.
func GetCloudProviderNTPHosts() []string {
	detectors := []cloudProviderNTPDetector{
		{name: ecscommon.CloudProviderName, callback: ecs.GetNTPHosts},
		{name: ec2.CloudProviderName, callback: ec2.GetNTPHosts},
		{name: gce.CloudProviderName, callback: gce.GetNTPHosts},
		{name: azure.CloudProviderName, callback: azure.GetNTPHosts},
		{name: alibaba.CloudProviderName, callback: alibaba.GetNTPHosts},
		{name: tencent.CloudProviderName, callback: tencent.GetNTPHosts},
	}

	for _, cloudNTPDetector := range detectors {
		if cloudNTPServers := cloudNTPDetector.callback(); cloudNTPServers != nil {
			klog.Infof("Using NTP servers from %s cloud provider: %+q", cloudNTPDetector.name, cloudNTPServers)
			return cloudNTPServers
		}
	}

	return nil
}
