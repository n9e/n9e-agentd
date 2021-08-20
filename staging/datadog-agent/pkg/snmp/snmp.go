// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2020-present Datadog, Inc.

package snmp

import (
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/config/snmp"
)

// NewListenerConfig parses configuration and returns a built ListenerConfig
func NewListenerConfig() (snmp.ListenerConfig, error) {
	return config.C.SnmpListener, nil
}
