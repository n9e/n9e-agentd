package cmds

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-agent/pkg/util"
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/spf13/cobra"
)

func newHostnameCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:          "hostname",
		Short:        "Print the hostname used by the Agent",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			hname, err := util.GetHostname(context.TODO())
			if err != nil {
				return fmt.Errorf("Error getting the hostname: %v", err)
			}
			fmt.Println(hname)
			return nil
		},
	}
}
