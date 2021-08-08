package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/n9e/n9e-agentd/pkg/ctl"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/yubo/golib/staging/logs"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubectl/pkg/cmd/options"
	"k8s.io/kubectl/pkg/util/templates"

	_ "github.com/n9e/n9e-agentd/pkg/ctl/cmds"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	settings := ctl.NewSettings()

	cmd := &cobra.Command{
		Use:   "agentdctl",
		Short: "n9e agentd controller",
		Long: templates.LongDesc(`
			agentd controls the n9e agentd.

			Find more information at:
			https://github.com/n9e/n9e-agentd`),
		SilenceUsage: true,
		Args:         ctl.NoArgs,
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

	otherCmds, groups := ctl.GetHookGroups(settings)
	groups.Add(cmd)
	filters := []string{"options"}
	templates.ActsAsRootCommand(cmd, filters, groups...)

	cmd.AddCommand(otherCmds...)
	cmd.AddCommand(options.NewCmdOptions(settings.Out))

	return cmd
}
