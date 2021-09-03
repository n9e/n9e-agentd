package cmds

import (
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/config/settings"
	"github.com/spf13/cobra"
)

// newSettingsCmd returns a cobra command to list the settings that can be changed at runtime.
func newSettingsCmd(env *agent.EnvSettings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "List settings that can be changed at runtime",
		RunE: func(_ *cobra.Command, _ []string) error {
			return listRuntimeConfigurableValue(env)
		},
	}

	cmd.AddCommand(newGetConfig(env))
	cmd.AddCommand(newSetConfig(env))

	return cmd
}

func listRuntimeConfigurableValue(env *agent.EnvSettings) error {

	configs := map[string]settings.RuntimeSettingResponse{}

	if err := env.ApiCall("GET", "/api/v1/config/settings", nil, nil, &configs); err != nil {
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

// get returns a cobra command to get a runtime config value.
func newGetConfig(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "get [setting]",
		Short: "Get, for the current runtime, the value of a given configuration setting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := agent.NewSettingsClient(env).Get(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("%s is set to: %v\n", args[0], value)
			return nil
		},
	}
}

// set returns a cobra command to set a config value at runtime.
func newSetConfig(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "set [setting] [value]",
		Short: "Set, for the current runtime, the value of a given configuration setting",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			hidden, err := agent.NewSettingsClient(env).Set(args[0], args[1])
			if err != nil {
				return err
			}

			if hidden {
				fmt.Printf("IMPORTANT: you have modified a hidden option, this may incur billing charges or have other unexpected side-effects.\n")
			}

			fmt.Printf("Configuration setting %s is now set to: %s\n", args[0], args[1])

			return nil
		},
	}
}
