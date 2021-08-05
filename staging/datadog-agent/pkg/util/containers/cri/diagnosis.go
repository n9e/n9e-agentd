// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build cri

package cri

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/diagnose/diagnosis"
	"k8s.io/klog/v2"
)

func init() {
	diagnosis.Register("CRI availability", diagnose)
}

// diagnose the CRI socket connectivity
func diagnose() error {
	_, err := GetUtil()
	if err != nil {
		klog.Error(err)
	}
	return err
}