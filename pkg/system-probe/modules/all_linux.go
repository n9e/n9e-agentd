// +build linux

package modules

import (
	"time"

	"github.com/n9e/n9e-agentd/pkg/system-probe/api/module"
)

// All System Probe modules should register their factories here
var All = []module.Factory{
	NetworkTracer,
	TCPQueueLength,
	OOMKillProbe,
	SecurityRuntime,
	Process,
}

func inactivityEventLog(duration time.Duration) {

}
