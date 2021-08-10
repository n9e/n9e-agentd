package cmds

import (
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/agentctl"
	"github.com/spf13/cobra"
)

const completionDesc = `
Generate autocompletions script for fks for the specified shell (bash).

This command can generate shell autocompletions. e.g.

	$ agentctl completion

Can be sourced as such

	$ source <(agentctl completion)
`

func newCompletionCmd(env *ctl.EnvSettings) *cobra.Command {
	var typ string

	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate autocompletions script for the specified shell [bash/zsh] (defualt:bash)",
		Long:  completionDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				typ = args[0]
			}
			switch typ {
			case "zsh":
				return env.TopCmd.GenZshCompletion(env.Out)
			case "bash", "":
				return env.TopCmd.GenBashCompletion(env.Out)
			default:
				return fmt.Errorf("unsupported %s", typ)
			}
		},
	}
	return cmd
}
