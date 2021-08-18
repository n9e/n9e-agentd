package cmds

import (
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/spf13/cobra"
)

func newVersionCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "show version information",
		SilenceUsage: true,
		RunE:         options.VersionCmd,
	}
}
