// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build containerd

package containerd

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/diagnose/diagnosis"
)

func init() {
	diagnosis.Register("Containerd availability", diagnose)
}

// diagnose the Containerd socket connectivity
func diagnose() error {
	_, err := GetContainerdUtil()
	return err
}
