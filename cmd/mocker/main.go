package main

import (
	"fmt"
	"os"

	"github.com/yubo/golib/configer"
	"github.com/yubo/golib/logs"
	"github.com/yubo/golib/proc"

	_ "github.com/n9e/n9e-agentd/pkg/mocker"
)

const (
	defaultConfig = `apiserver:
  secureServing:
    enabled: false
  insecureServing:
    enabled: true
    bindAddress: 127.0.0.1
    bindPort: 8000
`
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := proc.NewRootCmd(proc.WithConfigOptions(
		configer.WithDefaultYaml("", defaultConfig),
	)).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
