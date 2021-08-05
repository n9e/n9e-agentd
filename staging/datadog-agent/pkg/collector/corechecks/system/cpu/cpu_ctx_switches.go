// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.
// +build !linux

package cpu

import "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"

func (c *Check) collectCtxSwitches(sender aggregator.Sender) error {
	// On non-linux systems, do nothing
	return nil
}