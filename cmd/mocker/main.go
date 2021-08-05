package main

import (
	"flag"
	"fmt"
	"net/http"

	"k8s.io/klog/v2"
)

func main() {
	if err := sendStart(); err != nil {
		klog.Fatal(err)
	}

	var port int
	var confd string
	var collectRule bool
	flag.IntVar(&port, "port", 8000, "listen port")
	flag.BoolVar(&collectRule, "collect-rule", false, "enable collect rule provider")
	flag.StringVar(&confd, "confd", "./etc/mocker.d", "config dir")
	flag.Parse()

	for _, r := range routes {
		http.HandleFunc(r.pattern, payloadHandle(r.payload))
	}

	if collectRule {
		installCollectRules(confd)
	}

	klog.Infof("listen :%d", port)
	klog.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
