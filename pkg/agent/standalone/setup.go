// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package standalone

import (
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/config"
)

// SetupCLI sets up the shared utilities for a standalone CLI command:
// - config, with defaults to avoid conflicting with an agent process running in parallel
// - logger
// and returns the log level resolved from cliLogLevel and defaultLogLevel
func SetupCLI(env *agent.EnvSettings, loggerName config.LoggerName, cliLogFile string, cliLogLevel string) (string, *config.Warnings, error) {
	var resolvedLogLevel string

	if cliLogLevel != "" {
		// Honour the deprecated --log-level argument
		//overrides := make(map[string]interface{})
		//overrides["log_level"] = cliLogLevel
		config.C.LogLevel = cliLogLevel
		//config.AddOverrides(overrides)
	}

	resolvedLogLevel = env.Agent.LogLevel

	//overrides := make(map[string]interface{})
	//overrides["cmd_port"] = 0
	//config.AddOverrides(overrides)
	//env.Apiserver.Port = 0 // let the OS assign an available port for the HTTP server

	//warnings, err := common.SetupConfigWithWarnings(confFilePath, configName)
	//if err != nil {
	//	return resolvedLogLevel, warnings, fmt.Errorf("unable to set up global agent configuration: %v", err)
	//}

	err := config.SetupLogger(loggerName, resolvedLogLevel, cliLogFile, "", false, true, false)
	if err != nil {
		return resolvedLogLevel, nil, fmt.Errorf("unable to set up logger: %v", err)
	}

	return resolvedLogLevel, nil, nil
}
