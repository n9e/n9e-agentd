package main

import (
	"fmt"
	"os"

	server "github.com/yubo/apiserver/pkg/server/module"
	"github.com/yubo/golib/logs"
	"github.com/yubo/golib/proc"

	_ "github.com/n9e/n9e-agentd/pkg/mocker"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := proc.NewRootCmd(
		server.WithInsecureServing(),
	).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
