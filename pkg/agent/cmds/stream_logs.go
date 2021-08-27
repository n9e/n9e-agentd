package cmds

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-agent/pkg/logs/diagnostic"
	"github.com/n9e/n9e-agentd/pkg/agent"

	"github.com/spf13/cobra"
)

func newStreamLogsCmd(env *agent.EnvSettings) *cobra.Command {
	var filters diagnostic.Filters

	cmd := &cobra.Command{
		Use:   "stream-logs",
		Short: "Stream the logs being processed by a running agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return streamLogs(env, &filters)
		},
	}

	fs := cmd.Flags()
	fs.StringVar(&filters.Name, "name", "", "Filter by name")
	fs.StringVar(&filters.Type, "type", "", "Filter by type")
	fs.StringVar(&filters.Source, "source", "", "Filter by source")

	return cmd
}

func streamLogs(env *agent.EnvSettings, filters *diagnostic.Filters) error {
	watching, err := env.Client.Get().
		Prefix("/api/v1/stream-logs").
		Timeout(0).
		Watch(context.Background(), new(string))
	if err != nil {
		return err
	}

	ch := watching.ResultChan()
	for {
		e, ok := <-ch
		if !ok {
			return nil
		}
		fmt.Print(*e.Object.(*string))
	}
}
