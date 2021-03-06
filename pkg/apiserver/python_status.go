// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.
//
// Package agent implements the api endpoints for the `/agent` prefix.
// This group of endpoints is meant to provide high-level functionalities
// at the agent level.

// +build python
// +build !windows

package apiserver

import (
	"net/http"

	"github.com/DataDog/datadog-agent/pkg/collector/python"
)

func getPythonStatus(w http.ResponseWriter, r *http.Request) ([]*python.PythonStats, error) {
	return python.GetPythonInterpreterMemoryUsage()
}
