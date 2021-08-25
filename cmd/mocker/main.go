package main

import (
	"flag"
	"fmt"
	"net/http"

	"k8s.io/klog/v2"
)

func main() {

	var port int
	var confd string
	var collectRule bool
	var sendStatsd bool
	flag.IntVar(&port, "port", 8000, "listen port")
	flag.BoolVar(&collectRule, "collect-rule", false, "enable collect rule provider")
	flag.BoolVar(&sendStatsd, "send-statsd", false, "enable send statsd sample data")
	flag.StringVar(&confd, "confd", "./etc/mocker.d", "config dir")
	flag.Parse()

	for _, r := range routes {
		http.HandleFunc(r.pattern, payloadHandle(r.payload))
	}

	if collectRule {
		installCollectRules(confd)
	}

	if sendStatsd {
		if err := sendStart(); err != nil {
			klog.Fatal(err)
		}
	}

	klog.Infof("listen :%d", port)
	klog.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
