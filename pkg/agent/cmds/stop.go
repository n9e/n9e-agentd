package cmds

import (
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/spf13/cobra"
)

func newStopCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stops a running Agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := env.ApiCall("POST", "/api/v1/stop", nil, nil, nil)
			if err != nil {
				return err
			}
			fmt.Fprintf(env.Out, "Agent successfully stopped")
			return nil
		},
	}
}
