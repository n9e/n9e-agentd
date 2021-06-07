package app

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/n9e/n9e-agentd/pkg/options"
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
	ctx = proc.WithConfigOps(ctx) //config.WithBaseBytes2("http", app.DefaultOptions),

	cmd, err := newRootCmd(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	// version
	cmd.AddCommand(options.NewVersionCmd())

	return cmd
}

func newRootCmd(ctx context.Context) (*cobra.Command, error) {
	rand.Seed(time.Now().UnixNano())

	cmd := &cobra.Command{
		Use:          AppName,
		Short:        fmt.Sprintf("%s service", AppName),
		SilenceUsage: true,
	}

	if err := proc.ApplyToCmd(ctx, cmd); err != nil {
		return nil, err
	}

	return cmd, nil
}

func init() {
	options.InstallReporter()
}
