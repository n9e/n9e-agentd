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

	// apiserver
	_ "github.com/yubo/apiserver/pkg/authentication/register"
	_ "github.com/yubo/apiserver/pkg/authorization/register"
	_ "github.com/yubo/apiserver/pkg/rest/swagger/register"
	_ "github.com/yubo/apiserver/pkg/server/register"
	_ "github.com/yubo/apiserver/plugin/authorizer/abac/register"

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

	settings := agent.NewSettings(context.TODO())

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

			return settings.Init(cmd, override)
		},
	}

	// add flags
	setupCommand(cmd, settings)

	return cmd
}

func setupCommand(cmd *cobra.Command, settings *agent.EnvSettings) {
	settings.TopCmd = cmd
	fs := cmd.PersistentFlags()

	// add flags
	nfs := proc.NamedFlagSets()
	globalflag.AddGlobalFlags(nfs.FlagSet("global"))
	configer.AddFlags(nfs.FlagSet("global"))
	proc.AddFlags(nfs.FlagSet("global"))

	for _, f := range nfs.FlagSets {
		fs.AddFlagSet(f)
	}

	otherCmds, groups := agent.GetHookGroups(settings)
	groups.Add(cmd)
	filters := []string{"options"}
	templates.ActsAsRootCommand(cmd, filters, groups...)

	cmd.AddCommand(otherCmds...)
}
