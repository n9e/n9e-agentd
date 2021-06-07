// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build !kubeapiserver

package status

import (
	"k8s.io/klog/v2"
)

func getLeaderElectionDetails() map[string]string {
	klog.Info("Not implemented")
	return nil
}

func getDCAStatus() map[string]string {
	klog.Info("Not implemented")
	return nil
}
