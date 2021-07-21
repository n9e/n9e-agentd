package app

import (
	"context"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/spf13/cobra"
	"github.com/yubo/golib/configer"
	"github.com/yubo/golib/proc"

	_ "github.com/n9e/n9e-agentd/pkg/agent"
	_ "github.com/n9e/n9e-agentd/plugins/all"
)

const (
	AppName    = "agentd"
	moduleName = "agentd.main"
)

func NewServerCmd() *cobra.Command {
	ctx := context.Background()
	ctx = proc.WithName(ctx, AppName)
	ctx = proc.WithConfigOps(ctx,
		configer.WithFlagOptions(true, false, 5),
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

func init() {
	options.InstallReporter()
}
