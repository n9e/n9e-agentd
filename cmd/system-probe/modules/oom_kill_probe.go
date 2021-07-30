// +build linux

package modules

import (
	"net/http"

	"github.com/n9e/n9e-agentd/cmd/system-probe/api"
	"github.com/n9e/n9e-agentd/cmd/system-probe/config"
	"github.com/n9e/n9e-agentd/cmd/system-probe/utils"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/ebpf/probe"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/ebpf"
	"k8s.io/klog/v2"
	"github.com/pkg/errors"
)

// OOMKillProbe Factory
var OOMKillProbe = api.Factory{
	Name: config.OOMKillProbeModule,
	Fn: func(cfg *config.Config) (api.Module, error) {
		klog.Infof("Starting the OOM Kill probe")
		okp, err := probe.NewOOMKillProbe(ebpf.NewConfig())
		if err != nil {
			return nil, errors.Wrapf(err, "unable to start the OOM kill probe")
		}
		return &oomKillModule{okp}, nil
	},
}

var _ api.Module = &oomKillModule{}

type oomKillModule struct {
	*probe.OOMKillProbe
}

func (o *oomKillModule) Register(httpMux *http.ServeMux) error {
	httpMux.HandleFunc("/check/oom_kill", func(w http.ResponseWriter, req *http.Request) {
		stats := o.OOMKillProbe.GetAndFlush()
		utils.WriteAsJSON(w, stats)
	})

	return nil
}

func (o *oomKillModule) GetStats() map[string]interface{} {
	return nil
}
