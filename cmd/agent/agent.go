package main

import (
	"fmt"
	"os"

	"github.com/yubo/golib/logs"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := newRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
