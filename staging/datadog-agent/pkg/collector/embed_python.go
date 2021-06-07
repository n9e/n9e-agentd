// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build python

package collector

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/python"
	"github.com/n9e/n9e-agentd/pkg/config"
	"k8s.io/klog/v2"
)

func pySetup(paths ...string) (pythonVersion, pythonHome, pythonPath string) {
	if err := python.Initialize(paths...); err != nil {
		klog.Errorf("Could not initialize Python: %s", err)
	}
	return python.PythonVersion, python.PythonHome, python.PythonPath
}

func pyPrepareEnv() error {
	if config.Datadog.IsSet("procfs_path") {
		procfsPath := config.Datadog.GetString("procfs_path")
		return python.SetPythonPsutilProcPath(procfsPath)
	}
	return nil
}

func pyTeardown() {
	python.Destroy()
}
