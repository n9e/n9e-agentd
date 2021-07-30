// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package uptime

import (
	"github.com/n9e/n9e-agentd/pkg/aggregator"
	"github.com/n9e/n9e-agentd/pkg/collector/check"
	core "github.com/n9e/n9e-agentd/pkg/collector/corechecks"
	"k8s.io/klog/v2"
)

const checkName = "uptime"

// Check doesn't need additional fields
type Check struct {
	core.CheckBase
}

// Run executes the check
func (c *Check) Run() error {
	sender, err := aggregator.GetSender(c.ID())
	if err != nil {
		return err
	}

	t, err := uptime()
	if err != nil {
		klog.Errorf("uptime.Check: could not retrieve uptime: %s", err)
		return err
	}

	sender.Gauge("system.uptime", float64(t), "", nil)
	sender.Commit()

	return nil
}

func uptimeFactory() check.Check {
	return &Check{
		CheckBase: core.NewCheckBase(checkName),
	}
}

func init() {
	core.RegisterCheck(checkName, uptimeFactory)
}
