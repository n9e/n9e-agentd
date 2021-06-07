// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package disk

import (
	"regexp"

	yaml "gopkg.in/yaml.v2"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/check"
	core "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks"
	"k8s.io/klog/v2"
)

const (
	// SectorSize is exported in github.com/shirou/gopsutil/disk (but not working!)
	SectorSize       = 512
	kB               = (1 << 10)
	iostatsCheckName = "io"
)

// Configure the IOstats check
func (c *IOCheck) commonConfigure(data integration.Data, initConfig integration.Data, source string) error {
	if err := c.CommonConfigure(data, source); err != nil {
		return err
	}

	conf := make(map[interface{}]interface{})

	err := yaml.Unmarshal([]byte(initConfig), &conf)
	if err != nil {
		return err
	}

	blacklistRe, ok := conf["device_exclude_re"]
	if !ok {
		blacklistRe, ok = conf["device_blacklist_re"]
		if ok {
			klog.Warning("'device_blacklist_re' has been deprecated, use 'device_exclude_re' instead")
		}
	}
	if ok && blacklistRe != "" {
		if regex, ok := blacklistRe.(string); ok {
			c.blacklist, err = regexp.Compile(regex)
		}
	}

	if lowercaseDeviceTagOption, ok := conf["lowercase_device_tag"]; ok {
		if lowercaseDeviceTag, ok := lowercaseDeviceTagOption.(bool); ok {
			c.lowercaseDeviceTag = lowercaseDeviceTag
		} else {
			klog.Warning("Can't cast value of 'lowercase_device_tag' option to boolean: ", lowercaseDeviceTagOption)
		}
	}

	return err
}

func init() {
	core.RegisterCheck(iostatsCheckName, ioFactory)
}

func ioFactory() check.Check {
	return &IOCheck{
		CheckBase: core.NewCheckBase(iostatsCheckName),
	}
}
