// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build linux freebsd netbsd openbsd solaris dragonfly aix
// +build !android

package config

const (
	defaultConfdPath            = "/opt/n9e/agentd/conf.d"
	defaultAdditionalChecksPath = "/opt/n9e/agentd/checks.d"
	defaultRunPath              = "/opt/n9e/agentd/run"
	defaultSystemProbeAddress   = "/opt/n9e/agentd/run/sysprobe.sock"
	defaultSyslogURI            = "unixgram:///dev/log"
	defaultGuiPort              = -1
	// defaultSecurityAgentLogFile points to the log file that will be used by the security-agent if not configured
	//defaultSecurityAgentLogFile = "/var/log/n9e/security-agent.log"
	// defaultSystemProbeAddress is the default unix socket path to be used for connecting to the system probe
	//defaultSystemProbeLogFilePath = "/var/log/n9e/system-probe.log"
)

// called by init in config.go, to ensure any os-specific config is done
// in time
func osinit() {
}
