// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package common

import (
	"path/filepath"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/scheduler"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector"
	lsched "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/scheduler"
	lstatus "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/status"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/tagger"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/tagger/collectors"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/tagger/local"
)

// LoadComponents configures several common Agent components:
// tagger, collector, scheduler and autodiscovery
func LoadComponents(cf *config.Config) {
	// start tagging system
	tagger.SetDefaultTagger(local.NewTagger(collectors.DefaultCatalog))
	tagger.Init()

	// create the Collector instance and start all the components
	// NOTICE: this will also setup the Python environment, if available
	Coll = collector.NewCollector(GetPythonPaths()...)

	// creating the meta scheduler
	metaScheduler := scheduler.NewMetaScheduler()

	// registering the check scheduler
	metaScheduler.Register("check", collector.InitCheckScheduler(Coll))

	// registering the logs scheduler
	if lstatus.Get().IsRunning {
		metaScheduler.Register("logs", lsched.GetScheduler())
	}

	// setup autodiscovery
	confSearchPaths := []string{
		cf.ConfdPath,
		filepath.Join(cf.FileUsedDir, "conf.d"),
		"",
	}

	AC = setupAutoDiscovery(cf, confSearchPaths, metaScheduler)
}
