// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.
// +build windows

package filehandles

import (
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/check"
	core "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks"
	"k8s.io/klog/v2"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/winutil/pdhutil"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"
)

const fileHandlesCheckName = "file_handle"

type fhCheck struct {
	core.CheckBase
	counter *pdhutil.PdhMultiInstanceCounterSet
}

// Run executes the check
func (c *fhCheck) Run() error {

	sender, err := aggregator.GetSender(c.ID())
	if err != nil {
		return err
	}
	vals, err := c.counter.GetAllValues()
	if err != nil {
		klog.Warningf("Error getting handle value %v", err)
		return err
	}
	val := vals["_Total"]
	klog.V(5).Infof("Submitting system.fs.file_handles_in_use %v", val)
	sender.Gauge("system.fs.file_handles.in_use", float64(val), "", nil)
	sender.Commit()

	return nil
}

// The check doesn't need configuration
func (c *fhCheck) Configure(data integration.Data, initConfig integration.Data, source string) (err error) {
	if err := c.CommonConfigure(data, source); err != nil {
		return err
	}

	c.counter, err = pdhutil.GetMultiInstanceCounter("Process", "Handle Count", &[]string{"_Total"}, nil)
	return err
}

func fhFactory() check.Check {
	return &fhCheck{
		CheckBase: core.NewCheckBase(fileHandlesCheckName),
	}
}

func init() {
	core.RegisterCheck(fileHandlesCheckName, fhFactory)
}
