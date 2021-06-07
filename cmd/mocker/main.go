package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	agentpayload "github.com/n9e/agent-payload/gogen"
	"github.com/n9e/n9e-agentd/pkg/api"
	"k8s.io/klog/v2"
)

var (
	routes = []struct {
		pattern string
		payload interface{}
	}{
		{api.RoutePathSeries, agentpayload.MetricsPayload{}},
		{api.RoutePathSketchSeries, agentpayload.SketchPayload{}},
		{api.RoutePathEvents, agentpayload.EventsPayload{}},
		{api.RoutePathServiceChecks, agentpayload.ServiceChecksPayload{}},

		{api.RoutePathLogs, nil},
		{api.RoutePathCheckRuns, nil},
		{api.RoutePathIntake, nil},
		{api.RoutePathValidate, nil},
		{api.RoutePathHostMetadata, nil},
		{api.RoutePathMetadata, nil},
		{api.RoutePathCollector, nil},
		{api.RoutePathContainer, nil},
		{api.RoutePathOrchestrator, nil},
	}
)

func main() {
	if err := sendStart(); err != nil {
		klog.Fatal(err)
	}

	var port int
	flag.IntVar(&port, "port", 8080, "listen port")
	flag.Parse()

	for _, r := range routes {
		http.HandleFunc(r.pattern, payloadHandle(r.payload))
	}

	installCollectRules()

	klog.Infof("listen :%d", port)

	klog.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func payloadHandle(payload interface{}) func(http.ResponseWriter, *http.Request) {
	return func(_ http.ResponseWriter, r *http.Request) {
		if payload != nil {
			readAll(r, reflect.New(reflect.TypeOf(payload)).Interface())
			return
		}

		readAll(r, nil)
	}
}

type decode interface {
	Unmarshal(b []byte) error
}

func readAll(r *http.Request, payload interface{}) {
	reader := r.Body

	if r.Header.Get("Content-Encoding") == "gzip" {
		var err error
		if reader, err = gzip.NewReader(r.Body); err != nil {
			klog.Error(err)
			return
		}
		defer reader.Close()
	}

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		klog.Error(err)
		return
	}

	for k, v := range r.Header {
		klog.Infof("%s %v", k, v)
	}

	if payload == nil {
		klog.Infof("%s %s [%d] %s", r.Method, r.URL, len(b), string(b))
		return
	}

	switch r.Header.Get("Content-Type") {
	case "application/x-protobuf":
		d, ok := payload.(decode)
		if !ok {
			klog.Infof("invalid payload")
			return
		}

		if err := d.Unmarshal(b); err != nil {
			klog.Errorf("%s %s Unmarshal err %s", r.Method, r.URL, err)
			return
		}
	case "application/json":
		fallthrough
	default:
		if err := json.Unmarshal(b, payload); err != nil {
			klog.Errorf("%s %s Unmarshal err %s", r.Method, r.URL, err)
			return
		}
	}
	buf, _ := json.Marshal(payload)
	klog.Infof("%s %s [%d] %s", r.Method, r.URL, len(b), string(buf))
}
