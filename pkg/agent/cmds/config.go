package cmds

import (
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/config/settings"

	"github.com/spf13/cobra"
)

// Config returns the main cobra config command.
func newConfigCmd(env *agent.EnvSettings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Print the runtime configuration of a running agent",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := agent.NewSettingsClient(env).FullConfig()
			if err != nil {
				return err
			}
			env.Out.Write([]byte(config))
			return nil
		},
	}

	cmd.AddCommand(newListConfig(env))
	cmd.AddCommand(newGetConfig(env))
	cmd.AddCommand(newSetConfig(env))
	cmd.AddCommand(newConfigZshCmd(env))
	cmd.AddCommand(newConfigBashCmd(env))

	return cmd
}

// listRuntime returns a cobra command to list the settings that can be changed at runtime.
func newListConfig(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "list-runtime",
		Short: "List settings that can be changed at runtime",
		Long:  ``,
		RunE: func(_ *cobra.Command, _ []string) error {
			return listRuntimeConfigurableValue(env)
		},
	}
}

func listRuntimeConfigurableValue(env *agent.EnvSettings) error {
	configs, err := _listRuntimeConfigurableValue(env)
	if err != nil {
		return err
	}

	fmt.Println("=== Settings that can be changed at runtime ===")
	for setting, details := range configs {
		if !details.Hidden {
			fmt.Printf("%-30s %s\n", setting, details.Description)
		}
	}

	return nil
}

func _listRuntimeConfigurableValue(env *agent.EnvSettings) (map[string]settings.RuntimeSettingResponse, error) {
	output := map[string]settings.RuntimeSettingResponse{}

	if err := env.ApiCall("GET", "/api/v1/config/settings", nil, nil, &output); err != nil {
		return nil, err
	}

	return output, nil
}

// set returns a cobra command to set a config value at runtime.
func newSetConfig(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "set [setting] [value]",
		Short: "Set, for the current runtime, the value of a given configuration setting",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			return setConfigValue(env, args)
		},
	}
}

// get returns a cobra command to get a runtime config value.
func newGetConfig(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "get [setting]",
		Short: "Get, for the current runtime, the value of a given configuration setting",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			return getConfigValue(env, args)
		},
	}
}

func setConfigValue(env *agent.EnvSettings, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("exactly two parameters are required: the setting name and its value")
	}

	hidden, err := agent.NewSettingsClient(env).Set(args[0], args[1])
	if err != nil {
		return err
	}

	if hidden {
		fmt.Printf("IMPORTANT: you have modified a hidden option, this may incur billing charges or have other unexpected side-effects.\n")
	}

	fmt.Printf("Configuration setting %s is now set to: %s\n", args[0], args[1])

	return nil
}

func getConfigValue(env *agent.EnvSettings, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("a single setting name must be specified")
	}

	value, err := agent.NewSettingsClient(env).Get(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("%s is set to: %v\n", args[0], value)

	return nil
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
