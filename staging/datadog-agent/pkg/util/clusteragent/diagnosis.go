// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package clusteragent

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/diagnose/diagnosis"
	"k8s.io/klog/v2"
)

func init() {
	diagnosis.Register("Cluster Agent availability", diagnose)
}

func diagnose() error {
	_, err := GetClusterAgentClient()
	if err != nil {
		klog.Error(err)
	}
	return err
}
