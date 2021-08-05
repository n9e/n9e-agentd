// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build kubelet

package hostname

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/hostname/kubelet"
)

func init() {
	RegisterHostnameProvider("kubelet", kubelet.HostnameProvider)
}