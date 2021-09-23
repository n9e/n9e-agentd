package main

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/n9e/n9e-agentd/pkg/util/templates"
	"github.com/spf13/cobra"
	"github.com/yubo/golib/cli/globalflag"
	"github.com/yubo/golib/configer"
	"github.com/yubo/golib/proc"

	_ "github.com/n9e/n9e-agentd/pkg/agent/cmds"
	_ "github.com/n9e/n9e-agentd/pkg/agent/server"
	_ "github.com/n9e/n9e-agentd/pkg/apiserver"
	_ "github.com/n9e/n9e-agentd/plugins/all"
)

const (
	AppName    = "agent"
	moduleName = "agentd.main"
)

func init() {
	options.InstallReporter()
}

func newRootCmd() *cobra.Command {
	rand.Seed(time.Now().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	ctx := newContext()
	settings := agent.NewSettings(ctx)

	proc.WithContext(ctx)

	cmd := &cobra.Command{
		Use:   AppName,
		Short: "n9e agentd",
		Long: templates.LongDesc(`
			agentd controls the n9e agentd.

			Find more information at:
			https://github.com/n9e/n9e-agentd`),
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			override := map[string]string{}
			switch cmd.Use {
			case "version":
				return nil
			case "start":
				override["agent"] = fmt.Sprintf("is_cli_runner: true")
			}

			return settings.Init(override)
		},
	}

	// add flags
	setupCommand(cmd, settings)

	return cmd
}

func newContext() context.Context {
	ctx := proc.WithName(context.Background(), AppName)
	ctx = proc.WithAttr(ctx, make(map[interface{}]interface{}))

	return ctx
}

func setupCommand(cmd *cobra.Command, settings *agent.EnvSettings) {
	settings.TopCmd = cmd
	fs := cmd.PersistentFlags()

	// add flags
	configer.SetOptions(true, false, 5, fs)
	namedFlagSets := proc.NamedFlagSets()
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), AppName)
	configer.GlobalOptions.AddFlags(namedFlagSets.FlagSet("global"))
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	otherCmds, groups := agent.GetHookGroups(settings)
	groups.Add(cmd)
	filters := []string{"options"}
	templates.ActsAsRootCommand(cmd, filters, groups...)

	cmd.AddCommand(otherCmds...)
}
