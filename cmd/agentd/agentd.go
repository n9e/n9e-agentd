package main

import (
	"github.com/n9e/n9e-agentd/cmd/agentd/app"
	"github.com/yubo/golib/staging/logs"
	"k8s.io/klog/v2"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := app.NewServerCmd().Execute(); err != nil {
		klog.Fatal(err)
	}
}
