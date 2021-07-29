package integrations

import (
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/agentdctl"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

const (
	SVC     = "integrations"
	pkgName = "integrations"
)

func init() {
	agentdctl.RegisterHooks(agentdctl.GenHookOps2(pkgName, agentdctl.PRI_PKG, []agentdctl.HookOps2{
		{testSnmp, agentdctl.CMD_TEST, 0},
	}))
}

func testSnmp(env *agentdctl.EnvSettings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snmp",
		Short: "test snmp collector",
		Example: templates.Examples(`
                        # test snmp
			agentdctl test snmp`),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(env.Out, "hello world")
			return nil
		},
	}
	return cmd
}
