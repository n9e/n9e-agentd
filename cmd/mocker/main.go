package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yubo/golib/logs"
	"github.com/yubo/golib/proc"

	_ "github.com/n9e/n9e-agentd/pkg/mocker"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	ctx := proc.WithName(context.Background(), "mocker")
	if err := proc.NewRootCmd(ctx).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
