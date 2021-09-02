// +build jmx

package cmds

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/n9e/n9e-agentd/cmd/agent/common"
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/agent/standalone"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/spf13/cobra"
)

type jmxCmd struct {
	env               *agent.EnvSettings
	CliSelectedChecks []string
	JmxLogLevel       string
	SaveFlare         bool
}

func newJmxCmd(env *agent.EnvSettings) *cobra.Command {
	jc := &jmxCmd{
		CliSelectedChecks: []string{},
		env:               env,
	}
	cmd := &cobra.Command{
		Use:   "jmx",
		Short: "Run troubleshooting commands on JMXFetch integrations",
	}

	cmd.AddCommand(jc.newJmxListCmd())
	cmd.AddCommand(jc.newJmxCollectCmd())

	cmd.PersistentFlags().StringVarP(&jc.JmxLogLevel, "log-level", "l", "", "set the log level (default 'debug') (deprecated, use the env var DD_LOG_LEVEL instead)")
	return cmd
}

func (p *jmxCmd) newJmxListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List attributes matched by JMXFetch.",
	}

	cmd.AddCommand(
		p.newJmxListEverythingCmd(),
		p.newJmxListMatchingCmd(),
		p.newJmxListLimitedCmd(),
		p.newJmxListCollectedCmd(),
		p.newJmxListNotMatchingCmd(),
		p.newJmxListWithMetricsCmd(),
		p.newJmxListWithRateMetricsCmd(),
	)

	cmd.PersistentFlags().StringSliceVar(&p.CliSelectedChecks, "checks", []string{}, "JMX checks (ex: jmx,tomcat)")
	cmd.PersistentFlags().BoolVarP(&p.SaveFlare, "flare", "", false, "save jmx list results to the log dir so it may be reported in a flare")

	return cmd
}

func (p *jmxCmd) newJmxListEverythingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "everything",
		Short: "List every attributes available that has a type supported by JMXFetch.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runJmxCommandConsole("list_everything")
		},
	}
}

func (p *jmxCmd) newJmxListMatchingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "matching",
		Short: "List attributes that match at least one of your instances configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runJmxCommandConsole("list_matching_attributes")
		},
	}
}
func (p *jmxCmd) newJmxListLimitedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "limited",
		Short: "List attributes that do match one of your instances configuration but that are not being collected because it would exceed the number of metrics that can be collected.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runJmxCommandConsole("list_limited_attributes")
		},
	}
}
func (p *jmxCmd) newJmxListCollectedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "collected",
		Short: "List attributes that will actually be collected by your current instances configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runJmxCommandConsole("list_collected_attributes")
		},
	}
}
func (p *jmxCmd) newJmxListNotMatchingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "not-matching",
		Short: "List attributes that don’t match any of your instances configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runJmxCommandConsole("list_not_matching_attributes")
		},
	}
}
func (p *jmxCmd) newJmxListWithMetricsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "with-metrics",
		Short: "List attributes and metrics data that match at least one of your instances configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runJmxCommandConsole("list_with_metrics")
		},
	}
}
func (p *jmxCmd) newJmxListWithRateMetricsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "with-rate-metrics",
		Short: "List attributes and metrics data that match at least one of your instances configuration, including rates and counters.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runJmxCommandConsole("list_with_rate_metrics")
		},
	}
}

func (p *jmxCmd) newJmxCollectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect",
		Short: "Start the collection of metrics based on your current configuration and display them in the console.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runJmxCommandConsole("collect")
		},
	}

	cmd.Flags().StringSliceVar(&p.CliSelectedChecks, "checks", []string{}, "JMX checks (ex: jmx,tomcat)")
	cmd.Flags().BoolVarP(&p.SaveFlare, "flare", "", false, "save jmx list results to the log dir so it may be reported in a flare")
	return cmd
}

// runJmxCommandConsole sets up the common utils necessary for JMX, and executes the command
// with the Console reporter
func (p *jmxCmd) runJmxCommandConsole(command string) error {
	logFile := ""
	if p.SaveFlare {
		// Windows cannot accept ":" in file names
		filenameSafeTimeStamp := strings.ReplaceAll(time.Now().UTC().Format(time.RFC3339), ":", "-")
		logFile = filepath.Join(config.C.JmxFlareDir, "jmx_"+command+"_"+filenameSafeTimeStamp+".log")
		p.JmxLogLevel = "debug"
	}

	logLevel, _, err := standalone.SetupCLI(p.env, loggerName, logFile, p.JmxLogLevel)
	if err != nil {
		fmt.Printf("Cannot initialize command: %v\n", err)
		return err
	}

	err = config.SetupJMXLogger(jmxLoggerName, logLevel, logFile, "", false, true, false)
	if err != nil {
		return fmt.Errorf("Unable to set up JMX logger: %v", err)
	}

	common.LoadComponents()

	err = standalone.ExecJMXCommandConsole(p.env, command, p.CliSelectedChecks, logLevel)

	if runtime.GOOS == "windows" {
		standalone.PrintWindowsUserWarning("jmx")
	}

	return err
}

func init() {
	agent.RegisterCmd([]agent.CmdOps{
		{CmdFactory: newJmxCmd, GroupNum: agent.CMD_G_GENERIC},
	}...)
}
