// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2020-present Datadog, Inc.

package snmp

import (
	"github.com/n9e/n9e-agentd/pkg/config"
	. "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/snmp/types"
)

const (
	defaultPort    = 161
	defaultTimeout = 5
	defaultRetries = 3
)

// NewListenerConfig parses configuration and returns a built ListenerConfig
func NewListenerConfig() (ListenerConfig, error) {
	snmpConfig := config.C.SnmpListener

	// Set the default values, we can't otherwise on an array
	for i := range snmpConfig.Configs {
		// We need to modify the struct in place
		config := &snmpConfig.Configs[i]
		if config.Port == 0 {
			config.Port = defaultPort
		}
		if config.Timeout == 0 {
			config.Timeout = defaultTimeout
		}
		if config.Retries == 0 {
			config.Retries = defaultRetries
		}
	}
	return snmpConfig, nil
}
