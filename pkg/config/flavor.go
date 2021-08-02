// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package config

import "github.com/n9e/n9e-agentd/pkg/config/flavor"

var agentFlavor = flavor.DefaultAgent

// SetFlavor sets the Agent flavor
func SetFlavor(f string) {
	agentFlavor = f

	if agentFlavor == flavor.IotAgent {
		C.m.Lock()
		C.IotHost = true
		C.m.Unlock()
	}
}

// GetFlavor gets the running Agent flavor
// it MUST NOT be called before the main package is initialized;
// e.g. in init functions or to initialize package constants or variables.
func GetFlavor() string {
	return agentFlavor
}
