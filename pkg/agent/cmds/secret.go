package cmds

import (
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/spf13/cobra"

	"github.com/DataDog/datadog-agent/pkg/secrets"
)

func newSecretCmd(env *agent.EnvSettings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Print information about decrypted secrets in configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return env.ApiCallDone("GET", "/api/v1/secrets", nil, nil, &secrets.SecretInfo{})
		},
	}
	return cmd
}
