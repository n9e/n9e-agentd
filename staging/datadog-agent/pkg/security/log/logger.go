// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package log

import "k8s.io/klog/v2"

// DatadogAgentLogger is a wrapper for the agent logger that we use to prevent a dependency on packages that we cannot
// import outside of the agent repository
type DatadogAgentLogger struct{}

// Tracef is used to print a trace level log
func (l DatadogAgentLogger) Tracef(format string, params ...interface{}) {
	klog.V(6).Infof(format, params...)
}

// Debugf is used to print a trace level log
func (l DatadogAgentLogger) Debugf(format string, params ...interface{}) {
	klog.V(5).Infof(format, params...)
}

// Errorf is used to print an error
func (l DatadogAgentLogger) Errorf(format string, params ...interface{}) {
	_ = klog.Errorf(format, params...)
}

// Infof is used to print an error
func (l DatadogAgentLogger) Infof(format string, params ...interface{}) {
	klog.Infof(format, params...)
}
