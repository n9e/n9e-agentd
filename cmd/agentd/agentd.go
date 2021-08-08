package main

import (
	"os"

	"github.com/yubo/golib/staging/logs"
	"k8s.io/klog/v2"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := NewServerCmd().Execute(); err != nil {
		klog.Error(err)
		os.Exit(1)
	}
}
