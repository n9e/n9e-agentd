// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package config

const (
	defaultConfdPath            = "/etc/n9e-agentd/conf.d"
	defaultAdditionalChecksPath = "/etc/n9e-agentd/checks.d"
	defaultSystemProbeAddress   = "/opt/n9e-agentd/run/sysprobe.sock"
	defaultGuiPort              = 5002
	defaultSyslogURI            = "unixgram:///var/run/syslog"
)
