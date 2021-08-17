package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/n9e/n9e-agentd/pkg/agentctl"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/yubo/golib/configer"
	"github.com/yubo/golib/logs"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubectl/pkg/cmd/options"
	"k8s.io/kubectl/pkg/util/templates"

	_ "github.com/n9e/n9e-agentd/pkg/agentctl/cmds"
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
	settings := agentctl.NewSettings()

	cmd := &cobra.Command{
		Use:   "agentctl",
		Short: "n9e agentd controller",
		Long: templates.LongDesc(`
			agentd controls the n9e agentd.

			Find more information at:
			https://github.com/n9e/n9e-agentd`),
		SilenceUsage: true,
		Args:         agentctl.NoArgs,
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

	fs := cmd.PersistentFlags()
	configer.SetOptions(true, false, 5, fs)
	settings.AddFlags(fs)

	otherCmds, groups := agentctl.GetHookGroups(settings)
	groups.Add(cmd)
	filters := []string{"options"}
	templates.ActsAsRootCommand(cmd, filters, groups...)

	cmd.AddCommand(otherCmds...)
	cmd.AddCommand(options.NewCmdOptions(settings.Out))

	return cmd
}
