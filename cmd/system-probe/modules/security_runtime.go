// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.
// +build linux

package modules

import (
	"github.com/n9e/n9e-agentd/cmd/system-probe/api"
	"github.com/n9e/n9e-agentd/cmd/system-probe/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/ebpf"
	sconfig "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/security/config"
	secmodule "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/security/module"
	"k8s.io/klog/v2"
	"github.com/pkg/errors"
)

const (
	// DefaultRuntimePoliciesDir is the default policies directory used by the runtime security module
	DefaultRuntimePoliciesDir = "/etc/datadog-agent/runtime-security.d"
)

// SecurityRuntime - Security runtime Factory
var SecurityRuntime = api.Factory{
	Name: config.SecurityRuntimeModule,
	Fn: func(agentConfig *config.Config) (api.Module, error) {
		config, err := sconfig.NewConfig(agentConfig)
		if err != nil {
			return nil, errors.Wrap(err, "invalid security runtime module configuration")
		}

		module, err := secmodule.NewModule(config)
		if err == ebpf.ErrNotImplemented {
			klog.Info("Datadog runtime security agent is only supported on Linux")
			return nil, api.ErrNotEnabled
		}
		return module, err
	},
}
