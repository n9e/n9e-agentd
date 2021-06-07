// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build linux windows darwin

package v5

import (
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metadata/common"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metadata/gohai"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metadata/host"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metadata/resources"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util"
)

// GetPayload returns the complete metadata payload as seen in Agent v5
func GetPayload(hostnameData util.HostnameData) *Payload {
	cp := common.GetPayload(hostnameData.Hostname)
	hp := host.GetPayload(hostnameData)
	rp := resources.GetPayload(hostnameData.Hostname)

	p := &Payload{
		CommonPayload: CommonPayload{*cp},
		HostPayload:   HostPayload{*hp},
	}

	if rp != nil {
		p.ResourcesPayload = ResourcesPayload{*rp}
	}

	if config.C.EnableGohai {
		p.GohaiPayload = GohaiPayload{MarshalledGohaiPayload{*gohai.GetPayload()}}
	}

	return p
}
