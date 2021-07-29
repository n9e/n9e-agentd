/// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package settings

import (
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/config"
)

// profilingRuntimeSetting wraps operations to change log level at runtime
type profilingRuntimeSetting string

func (l profilingRuntimeSetting) Description() string {
	return "Enable/disable profiling on the agent, valid values are: true, false"
}

func (l profilingRuntimeSetting) Hidden() bool {
	return true
}

func (l profilingRuntimeSetting) Name() string {
	return string(l)
}

func (l profilingRuntimeSetting) Get() (interface{}, error) {
	return config.C.ProfilingEnabled, nil
}

func (l profilingRuntimeSetting) Set(v interface{}) error {
	var profile bool
	var err error

	profile, err = getBool(v)

	if err != nil {
		return fmt.Errorf("Unsupported type for profile runtime setting: %v", err)
	}

	//if profile {
	//	// populate site
	//	s := config.DefaultSite
	//	if config.Datadog.IsSet("site") {
	//		s = config.Datadog.GetString("site")
	//	}

	//	// allow full url override for development use
	//	site := fmt.Sprintf(profiling.ProfileURLTemplate, s)
	//	if config.Datadog.IsSet("profiling.profile_dd_url") {
	//		site = config.Datadog.GetString("profiling.profile_dd_url")
	//	}

	//	err := profiling.Start(
	//		config.SanitizeAPIKey(config.C.ApiKey),
	//		site,
	//		config.Datadog.GetString("env"),
	//		profiling.ProfileCoreService,
	//		fmt.Sprintf("version:%v", options.Version),
	//	)
	//	if err == nil {
	//		config.Datadog.Set("profiling.enabled", true)
	//	}
	//} else {
	//	profiling.Stop()
	//	config.C.ProfilingEnabled = false
	//}
	config.C.ProfilingEnabled = profile

	return nil
}
