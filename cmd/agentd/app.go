package main

import (
	"context"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/spf13/cobra"
	"github.com/yubo/golib/configer"
	"github.com/yubo/golib/proc"

	_ "github.com/n9e/n9e-agentd/pkg/agent"
	_ "github.com/n9e/n9e-agentd/pkg/apiserver"
	_ "github.com/n9e/n9e-agentd/plugins/all"
)

const (
	AppName    = "agentd"
	moduleName = "agentd.main"
)

func init() {
	options.InstallReporter()
}

func NewServerCmd() *cobra.Command {
	ctx := context.Background()
	ctx = proc.WithName(ctx, AppName)
	ctx = proc.WithConfigOps(ctx,
		// Configfile Deprecated, remove it later
		configer.WithCallback(func(o configer.Options) {
			if config.Configfile != "" {
				o.AppendValueFile(config.Configfile)
			}
		}),
	)

	cmd := proc.NewRootCmd(ctx)

	// version
	cmd.AddCommand(options.NewVersionCmd())

	return cmd
}
