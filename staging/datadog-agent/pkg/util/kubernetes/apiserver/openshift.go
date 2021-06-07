// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build kubeapiserver

package apiserver

import (
	"context"

	"k8s.io/klog/v2"
)

// DetectOpenShiftAPILevel looks at known endpoints to detect if OpenShift
// APIs are available on this apiserver. OpenShift transitioned from a
// non-standard `/oapi` URL prefix to standard api groups under the `/apis`
// prefix in 3.6. Detecting both, with a preference for the new prefix.
func (c *APIClient) DetectOpenShiftAPILevel() OpenShiftAPILevel {
	err := c.Cl.CoreV1().RESTClient().Get().AbsPath("/apis/quota.openshift.io").Do(context.TODO()).Error()
	if err == nil {
		klog.V(5).Infof("Found %s", OpenShiftAPIGroup)
		return OpenShiftAPIGroup
	}
	klog.V(5).Infof("Cannot access %s: %s", OpenShiftAPIGroup, err)

	err = c.Cl.CoreV1().RESTClient().Get().AbsPath("/oapi").Do(context.TODO()).Error()
	if err == nil {
		klog.V(5).Infof("Found %s", OpenShiftOAPI)
		return OpenShiftOAPI
	}
	klog.V(5).Infof("Cannot access %s: %s", OpenShiftOAPI, err)

	// Fallback to NotOpenShift
	return NotOpenShift
}
