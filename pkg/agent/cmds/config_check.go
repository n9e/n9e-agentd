package cmds

import (
	"bytes"
	"fmt"

	"github.com/DataDog/datadog-agent/pkg/flare"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newConfigCheckCmd(env *agent.EnvSettings) *cobra.Command {
	var withDebug bool

	cmd := &cobra.Command{
		Use:   "configcheck",
		Short: "Print all configurations loaded & resolved of a running agent",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := config.SetupLogger(loggerName, config.GetEnvDefault("N9E_LOG_LEVEL", "off"), "", "", false, true, false)
			if err != nil {
				fmt.Printf("Cannot setup logger, exiting: %v\n", err)
				return err
			}

			var b bytes.Buffer
			color.Output = &b
			err = flare.GetConfigCheck(color.Output, withDebug)
			if err != nil {
				return fmt.Errorf("unable to get config: %v", err)
			}

			scrubbed, err := log.CredentialsCleanerBytes(b.Bytes())
			if err != nil {
				return fmt.Errorf("unable to scrub sensitive data configcheck output: %v", err)
			}

			fmt.Println(string(scrubbed))
			return nil
		},
	}

	cmd.Flags().BoolVar(&withDebug, "debug", false, "print additional debug info")

	return cmd
}
