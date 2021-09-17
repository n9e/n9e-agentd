// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package common provides a set of common symbols needed by different packages,
// to avoid circular dependencies.
package common

import (
	"encoding/json"
	"net/http"

	"github.com/DataDog/datadog-agent/pkg/autodiscovery"
	"github.com/DataDog/datadog-agent/pkg/collector"
	"github.com/DataDog/datadog-agent/pkg/dogstatsd"
	"github.com/DataDog/datadog-agent/pkg/forwarder"
	"github.com/DataDog/datadog-agent/pkg/metadata"
	"github.com/DataDog/datadog-agent/pkg/util/executable"
	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/version"
)

var (
	// AC is the global object orchestrating checks' loading and running
	AC *autodiscovery.AutoConfig

	// Coll is the global collector instance
	Coll *collector.Collector

	// DSD is the global dogstatsd instance
	DSD *dogstatsd.Server

	// MetadataScheduler is responsible to orchestrate metadata collection
	MetadataScheduler *metadata.Scheduler

	// Forwarder is the global forwarder instance
	Forwarder forwarder.Forwarder

	// utility variables
	_here, _ = executable.Folder()

	Client api.Client
)

// GetPythonPaths returns the paths (in order of precedence) from where the agent
// should load python modules and checks
func GetPythonPaths() []string {
	// wheels install in default site - already in sys.path; takes precedence over any additional location
	return []string{
		config.C.PyChecksPath,
		config.C.AdditionalChecksd, // custom checks, least precedent check location
	}
}

// GetVersion returns the version of the agent in a http response json
func GetVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	av, _ := version.Agent()
	j, _ := json.Marshal(av)
	w.Write(j)
}
