package cmds

import (
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/spf13/cobra"
	"github.com/yubo/apiserver/pkg/cmdcli"
)

func newEnvCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "env",
		Short: "show grab env information",
		RunE: func(cmd *cobra.Command, args []string) error {
			env.Out.Write(cmdcli.Table(env.Output()))
			return nil
		},
	}
}
