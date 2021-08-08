package cmds

import (
	"github.com/n9e/n9e-agentd/pkg/ctl"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/spf13/cobra"
)

func newVersionCmd(env *ctl.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "show version information",
		SilenceUsage: true,
		RunE:         options.VersionCmd,
	}
}
