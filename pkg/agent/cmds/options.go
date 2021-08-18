package cmds

import (
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/util/templates"
	"github.com/spf13/cobra"
	"github.com/yubo/golib/cli/flag"
	"github.com/yubo/golib/proc"
	"github.com/yubo/golib/util/term"
	"k8s.io/kubectl/pkg/util/i18n"
)

// newOptionsCmd implements the options command
func newOptionsCmd(env *agent.EnvSettings) *cobra.Command {
	namedFlagSets := proc.NamedFlagSets()
	cmd := &cobra.Command{
		Use:   "options",
		Short: i18n.T("Print the list of flags inherited by all commands"),
		Long:  i18n.T("Print the list of flags inherited by all commands"),
		Example: templates.Examples(i18n.T(`
		# Print flags inherited by all commands
		agentd options`)),
		Run: func(cmd *cobra.Command, args []string) {
			cols, _, _ := term.GetTerminalSize(cmd.OutOrStdout())
			fmt.Fprintf(cmd.OutOrStderr(), "Usage:\n %s\n", cmd.UseLine())
			flag.PrintSections(cmd.OutOrStderr(), *namedFlagSets, cols)
		},
	}

	// The `options` command needs write its output to the `out` stream
	// (typically stdout). Without calling SetOutput here, the Usage()
	// function call will fall back to stderr.
	//
	// See https://github.com/kubernetes/kubernetes/pull/46394 for details.
	cmd.SetOutput(env.Out)

	templates.UseOptionsTemplates(cmd)
	return cmd
}
