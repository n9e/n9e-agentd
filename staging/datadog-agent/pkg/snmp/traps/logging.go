// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2020-present Datadog, Inc.

package traps

import (
	"k8s.io/klog/v2"
	"github.com/soniah/gosnmp"
)

// trapLogger is a GoSNMP logger interface implementation.
type trapLogger struct {
	gosnmp.Logger
}

// NOTE: GoSNMP logs show the full content of decoded trap packets. Logging as DEBUG would be too noisy.
func (logger *trapLogger) Print(v ...interface{}) {
	klog.V(6).Info(v...)
}
func (logger *trapLogger) Printf(format string, v ...interface{}) {
	klog.V(6).Infof(format, v...)
}
