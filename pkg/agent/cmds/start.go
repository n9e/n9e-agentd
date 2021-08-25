package cmds

import (
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/spf13/cobra"
	"github.com/yubo/golib/proc"
	"k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

func newStartCmd(env *agent.EnvSettings) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "start agent deamon",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if klog.V(5).Enabled() {
				fs := cmd.Flags()
				flag.PrintFlags(fs)
			}
			return proc.Start()
		},
	}
	env.ServerCmd = cmd
	return cmd
}
