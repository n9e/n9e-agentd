package cmds

import (
	"github.com/n9e/n9e-agentd/pkg/ctl"
	"github.com/spf13/cobra"
	"github.com/yubo/apiserver/pkg/cmdcli"
)

func newEnvCmd(env *ctl.EnvSettings) *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "env",
		Short: "show grab env information",
		RunE: func(cmd *cobra.Command, args []string) error {
			env.Out.Write(cmdcli.Table(env.Output(verbose)))
			return nil
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&verbose, "verbvose", "V", false, "show secret infomation")
	return cmd
}
