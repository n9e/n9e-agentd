// +build linux

package modules

import (
	"net/http"

	"github.com/n9e/n9e-agentd/cmd/system-probe/api"
	"github.com/n9e/n9e-agentd/cmd/system-probe/config"
	"github.com/n9e/n9e-agentd/cmd/system-probe/utils"
	"github.com/n9e/n9e-agentd/pkg/collector/corechecks/ebpf/probe"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/ebpf"
	"github.com/pkg/errors"
)

// TCPQueueLength Factory
var TCPQueueLength = api.Factory{
	Name: config.TCPQueueLengthTracerModule,
	Fn: func(cfg *config.Config) (api.Module, error) {
		t, err := probe.NewTCPQueueLengthTracer(ebpf.NewConfig())
		if err != nil {
			return nil, errors.Wrapf(err, "unable to start the TCP queue length tracer")
		}

		return &tcpQueueLengthModule{t}, nil
	},
}

var _ api.Module = &tcpQueueLengthModule{}

type tcpQueueLengthModule struct {
	*probe.TCPQueueLengthTracer
}

func (t *tcpQueueLengthModule) Register(httpMux *http.ServeMux) error {
	httpMux.HandleFunc("/check/tcp_queue_length", func(w http.ResponseWriter, req *http.Request) {
		stats := t.TCPQueueLengthTracer.GetAndFlush()
		utils.WriteAsJSON(w, stats)
	})

	return nil
}

func (t *tcpQueueLengthModule) GetStats() map[string]interface{} {
	return nil
}
