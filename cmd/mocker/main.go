package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/yubo/golib/cli/globalflag"
	"github.com/yubo/golib/configer"
	"github.com/yubo/golib/logs"
	"k8s.io/klog/v2"
)

type config struct {
	Port        int    `flag:"port" default:"8000" description:"listen port"`
	Confd       string `flag:"collect-rule" description:"enable collect rule provider"`
	CollectRule bool   `flag:"" description:"enable send statsd sample data"`
	SendStatsd  bool   `flag:"" default:"./etc/mocker.d" description:"config dir"`
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := newRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cf := &config{}

	cmd := &cobra.Command{
		Use:   "watcher",
		Short: "watcher is a tool which watch files change and execute some command",
		RunE: func(cmd *cobra.Command, args []string) error {

			for _, r := range routes {
				http.HandleFunc(r.pattern, payloadHandle(r.payload))
			}

			if cf.CollectRule {
				installCollectRules(cf.Confd)
			}

			if cf.SendStatsd {
				if err := sendStart(); err != nil {
					klog.Fatal(err)
				}
			}

			klog.Infof("listen :%d", cf.Port)
			klog.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

		},
	}

	configer.AddFlags(cmd.Flags(), cf)
	globalflag.AddGlobalFlags(cmd.Flags(), "watcher")
	return cmd
}
