package main

import (
	"github.com/n9e/n9e-agentd/pkg/agentdctl"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/spf13/cobra"
	cmdoptions "k8s.io/kubectl/pkg/cmd/options"
	"k8s.io/kubectl/pkg/util/templates"

	_ "github.com/n9e/n9e-agentd/pkg/agentdctl/integrations"
)

type rootCmdOps struct {
	hookNum  int
	groupNum int
	use      string
	short    string
	aliases  []string
	run      func(env *agentdctl.EnvSettings) func(cmd *cobra.Command, args []string) error
}

const (
	pkgName = "agentdctl.main"
)

func init() {
	agentdctl.RegisterHooks(genRootHookOps([]rootCmdOps{{
		hookNum:  agentdctl.CMD_VERSION,
		groupNum: agentdctl.CMD_G_OTHER,
		use:      "version",
		short:    "show version information",
		run: func(env *agentdctl.EnvSettings) func(cmd *cobra.Command, args []string) error {
			return options.VersionCmd
		},
	}, {
		hookNum:  agentdctl.CMD_TEST,
		groupNum: agentdctl.CMD_G_OTHER,
		use:      "test",
		short:    "This command test api/checks",
	}}))
}

func genRootHookOps(in []rootCmdOps) (ret []agentdctl.HookOps) {
	for k, _ := range in {
		ops := &in[k]
		ret = append(ret, agentdctl.HookOps{
			Hook: func(env *agentdctl.EnvSettings) *cobra.Command {
				cmd := &cobra.Command{
					Use:     ops.use,
					Short:   ops.short,
					Aliases: ops.aliases,
				}

				if ops.run != nil {
					cmd.RunE = ops.run(env)
				}

				cmd.AddCommand(agentdctl.GetHookCmds(env, ops.hookNum)...)
				return cmd
			},
			Owner:    pkgName,
			HookNum:  agentdctl.CMD_ROOT,
			GroupNum: ops.groupNum,
			Priority: agentdctl.PRI_PKG,
		})
	}
	return
}

func newRootCmd() *cobra.Command {
	settings := agentdctl.NewSettings()

	cmd := &cobra.Command{
		Use:   "agentdctl",
		Short: "n9e agentd controller",
		Long: templates.LongDesc(`
			agentd controls the n9e agentd.

			Find more information at:
			https://github.com/n9e/n9e-agentd`),
		SilenceUsage: true,
		Args:         agentdctl.NoArgs,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if err := settings.Init(); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	settings.TopCmd = cmd

	flags := cmd.PersistentFlags()
	settings.AddFlags(flags)

	otherCmds, groups := agentdctl.GetHookGroups(settings)
	groups.Add(cmd)
	filters := []string{"options"}
	templates.ActsAsRootCommand(cmd, filters, groups...)

	cmd.AddCommand(otherCmds...)
	cmd.AddCommand(cmdoptions.NewCmdOptions(settings.Out))

	return cmd
}
