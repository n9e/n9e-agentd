package main

import (
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/yubo/golib/staging/logs"
	cliflag "k8s.io/component-base/cli/flag"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
