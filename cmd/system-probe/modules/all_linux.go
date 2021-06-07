// +build linux

package modules

import (
	"github.com/n9e/n9e-agentd/cmd/system-probe/api"
)

// All System Probe modules should register their factories here
var All = []api.Factory{
	NetworkTracer,
	TCPQueueLength,
	OOMKillProbe,
	SecurityRuntime,
	Process,
}
