// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package common

import (
	"os"
	"path/filepath"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/winutil"
	"k8s.io/klog/v2"

	// Init packages
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/containers/providers/windows"

	"github.com/cihub/seelog"
	"golang.org/x/sys/windows/registry"
)

var (
	// PyChecksPath holds the path to the python checks from integrations-core shipped with the agent
	PyChecksPath = filepath.Join(_here, "..", "checks.d")
	distPath     string
	// ViewsPath holds the path to the folder containing the GUI support files
	viewsPath   string
	enabledVals = map[string]bool{"yes": true, "true": true, "1": true,
		"no": false, "false": false, "0": false}
	subServices = map[string]string{"logs_enabled": "logs_enabled",
		"apm_enabled":     "apm_config.enabled",
		"process_enabled": "process_config.enabled"}
)

var (
	// DefaultConfPath points to the folder containing datadog.yaml
	DefaultConfPath = "c:\\programdata\\n9e"
	// DefaultLogFile points to the log file that will be used if not configured
	DefaultLogFile = "c:\\programdata\\n9e\\logs\\agent.log"
	// DefaultDCALogFile points to the log file that will be used if not configured
	DefaultDCALogFile = "c:\\programdata\\n9e\\logs\\cluster-agent.log"
	//DefaultJmxLogFile points to the jmx fetch log file that will be used if not configured
	DefaultJmxLogFile = "c:\\programdata\\n9e\\logs\\jmxfetch.log"
	// DefaultCheckFlareDirectory a flare friendly location for checks to be written
	DefaultCheckFlareDirectory = "c:\\programdata\\n9e\\logs\\checks\\"
	// DefaultJMXFlareDirectory a flare friendly location for jmx command logs to be written
	DefaultJMXFlareDirectory = "c:\\programdata\\n9e\\logs\\jmxinfo\\"
)

func init() {
	pd, err := winutil.GetProgramDataDir()
	if err == nil {
		DefaultConfPath = pd
		DefaultLogFile = filepath.Join(pd, "logs", "agent.log")
		DefaultDCALogFile = filepath.Join(pd, "logs", "cluster-agent.log")
	} else {
		winutil.LogEventViewer(config.ServiceName, 0x8000000F, DefaultConfPath)
	}
}

// EnableLoggingToFile -- set up logging to file
//func EnableLoggingToFile() {
//	seeConfig := `
//<seelog>
//	<outputs>
//		<rollingfile type="size" filename="c:\\ProgramData\\n9e\\Logs\\agent.log" maxsize="1000000" maxrolls="2" />
//	</outputs>
//</seelog>`
//	logger, _ := seelog.LoggerFromConfigAsBytes([]byte(seeConfig))
//	log.ReplaceLogger(logger)
//}

func getInstallPath() string {
	// fetch the installation path from the registry
	installpath := filepath.Join(_here, "..")
	var s string
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\n9e\n9e agentd`, registry.QUERY_VALUE)
	if err != nil {
		klog.Warningf("Failed to open registry key: %s", err)
	} else {
		defer k.Close()
		s, _, err = k.GetStringValue("InstallPath")
		if err != nil {
			klog.Warningf("Installpath not found in registry: %s", err)
		}
	}
	// if unable to figure out the install path from the registry,
	// just compute it relative to the executable.
	if s == "" {
		s = installpath
	}
	return s
}

// GetDistPath returns the fully qualified path to the 'dist' directory
func GetDistPath() string {
	if len(distPath) == 0 {
		var s string
		if s = getInstallPath(); s == "" {
			return ""
		}
		distPath = filepath.Join(s, `bin/agent/dist`)
	}
	return distPath
}

// GetViewsPath returns the fully qualified path to the GUI's 'views' directory
func GetViewsPath() string {
	if len(viewsPath) == 0 {
		var s string
		if s = getInstallPath(); s == "" {
			return ""
		}
		viewsPath = filepath.Join(s, "bin", "agent", "dist", "views")
		klog.V(5).Infof("ViewsPath is now %s", viewsPath)
	}
	return viewsPath
}

// CheckAndUpgradeConfig checks to see if there's an old n9e.conf, and if
// n9e.yaml is either missing or incomplete (no API key).  If so, upgrade it
func CheckAndUpgradeConfig() error {
	return nil
}
