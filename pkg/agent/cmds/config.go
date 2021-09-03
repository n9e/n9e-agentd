package cmds

import (
	"bytes"
	"fmt"

	"github.com/DataDog/datadog-agent/pkg/flare"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/fatih/color"
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/config"

	"github.com/spf13/cobra"
)

// Config returns the main cobra config command.
func newConfigCmd(env *agent.EnvSettings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Print the runtime configuration of a running agent",
	}

	cmd.AddCommand(
		newConfigAgentCmd(env),
		newConfigCheckCmd(env),
		newConfigJmxCmd(env),
		newConfigZshCmd(env),
		newConfigBashCmd(env),
	)

	return cmd
}

func newConfigAgentCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "agent",
		Short: "Print the runtime configuration of a running agent",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := agent.NewSettingsClient(env).FullConfig()
			if err != nil {
				return err
			}
			env.Out.Write([]byte(config))
			return nil
		},
	}
}

func newConfigCheckCmd(env *agent.EnvSettings) *cobra.Command {
	var withDebug bool
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Print all configurations loaded & resolved of a running agent",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := config.SetupLogger(loggerName, config.GetEnvDefault("N9E_LOG_LEVEL", "off"), "", "", false, true, false)
			if err != nil {
				fmt.Printf("Cannot setup logger, exiting: %v\n", err)
				return err
			}

			var query string
			if len(args) > 0 {
				query = args[0]
			}

			var b bytes.Buffer
			color.Output = &b
			err = flare.GetConfigCheck(color.Output, withDebug, query)
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

func newConfigJmxCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "jmx",
		Short: "Print the jmx configuration of a running agent",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return env.ApiCall("GET", "/api/v1/config/jmx", nil, nil, env.Out)
		},
	}
}
func newConfigZshCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "set zsh completion config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configZshCompletion(env)
		},
	}
}

const zshConfigUsage = `# Zsh completion file has been generated(~/.n9e/zsh/_n9e)

# So, to initialize the compsys add the following code into your ~/.zshrc file
cat >> ~/.zshrc <<'EOF'
# add custom completion scripts
fpath=(~/.n9e/zsh $fpath)

# compsys initialization
autoload -U compinit
compinit

# show completion menu when number of 
zstyle ':completion:*' menu select=2
EOF
`

func configZshCompletion(env *agent.EnvSettings) error {
	if err := env.TopCmd.GenZshCompletion(env.Out); err != nil {
		return err
	}

	fmt.Fprintf(env.Out, zshConfigUsage)

	return nil
}

func newConfigBashCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "set bash completion config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configBashCompletion(env)
		},
	}
}

const bashConfigUsage = `# Bash completion file has been generated(~/.n9e/n9e.bash)

# So, to initialize the compsys add the following code into your ~/.bashrc file
cat >> ~/.bashrc <<'EOF'
if [ -f ~/.n9e/n9e.bash ]; then
	source ~/.n9e/n9e.bash
fi
EOF
`

func configBashCompletion(env *agent.EnvSettings) error {
	if err := env.TopCmd.GenBashCompletion(env.Out); err != nil {
		return err
	}

	fmt.Fprintf(env.Out, bashConfigUsage)
	return nil
}
