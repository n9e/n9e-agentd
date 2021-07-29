// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package settings

import (
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/log"
)

// logLevelRuntimeSetting wraps operations to change log level at runtime.
type logLevelRuntimeSetting string

func (l logLevelRuntimeSetting) Description() string {
	return "Set/get the log level, valid values are: trace, debug, info, warn, error, critical and off"
}

func (l logLevelRuntimeSetting) Hidden() bool {
	return false
}

func (l logLevelRuntimeSetting) Name() string {
	return string(l)
}

func (l logLevelRuntimeSetting) Get() (interface{}, error) {
	level, err := log.GetLogLevel()
	if err != nil {
		return "", err
	}
	return level.String(), nil
}

func (l logLevelRuntimeSetting) Set(v interface{}) error {
	logLevel := v.(string)
	// TODO
	//err := config.ChangeLogLevel(logLevel)
	//if err != nil {
	//	return err
	//}
	config.C.LogLevel = logLevel
	return nil
}
