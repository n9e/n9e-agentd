package cmds

import (
	"fmt"

	"github.com/DataDog/datadog-agent/pkg/diagnose"
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/config"
	"k8s.io/klog/v2"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newDiagnoseCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "diagnose",
		Short: "Execute some connectivity diagnosis on your system",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doDiagnose(env)
		},
	}
}

func doDiagnose(env *agent.EnvSettings) error {
	klog.InfoS("diagnose", "loggerName", loggerName,
		"logLeve", config.C.LogLevel,
		"LogFile", config.C.LogFile,
		"syslogURI", config.GetSyslogURI(),
		"syslogrfc", config.C.SyslogRfc,
		"to cononsole", config.C.LogToConsole,
		"logformatJson", config.C.LogFormatJson,
	)
	err := config.SetupLogger(
		loggerName,
		config.C.LogLevel,
		config.C.LogFile,
		config.GetSyslogURI(),
		config.C.SyslogRfc,
		config.C.LogToConsole,
		config.C.LogFormatJson,
	)
	if err != nil {
		return fmt.Errorf("Error while setting up logging, exiting: %v", err)
	}

	return diagnose.RunAll(color.Output)
}
