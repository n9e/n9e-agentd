// +build !linux,!windows

package modules

import "github.com/n9e/n9e-agentd/pkg/system-probe/api/module"

// All System Probe modules should register their factories here
var All = []module.Factory{}
