// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package settings

import (
	"fmt"

	"github.com/n9e/n9e-agentd/cmd/agentd/common"
	"github.com/n9e/n9e-agentd/pkg/config"
)

// DsdStatsRuntimeSetting wraps operations to change log level at runtime.
type DsdStatsRuntimeSetting string

func (s DsdStatsRuntimeSetting) Description() string {
	return "Enable/disable the dogstatsd debug stats. Possible values: true, false"
}

func (s DsdStatsRuntimeSetting) Hidden() bool {
	return false
}

func (s DsdStatsRuntimeSetting) Name() string {
	return string(s)
}

func (s DsdStatsRuntimeSetting) Get() (interface{}, error) {
	return config.Get(string(s)), nil
	//return atomic.LoadUint64(&common.DSD.Debug.Enabled) == 1, nil
}

func (s DsdStatsRuntimeSetting) Set(v interface{}) error {
	if !config.C.Statsd.Enabled {
		return fmt.Errorf("statsd has been disabled")
	}

	var newValue bool
	var err error

	if newValue, err = GetBool(v); err != nil {
		return fmt.Errorf("dsdStatsRuntimeSetting: %v", err)
	}

	if newValue {
		common.DSD.EnableMetricsStats()
	} else {
		common.DSD.DisableMetricsStats()
	}

	config.Set(string(s), newValue)
	return nil
}
