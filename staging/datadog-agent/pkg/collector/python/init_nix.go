// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build python
// +build !windows

package python

import (
	"unsafe"

	"github.com/DataDog/datadog-agent/pkg/util/log"

	"github.com/n9e/n9e-agentd/pkg/config"
)

/*
#cgo !windows LDFLAGS: -ldatadog-agent-rtloader -ldl

#include <datadog_agent_rtloader.h>
#include <rtloader_mem.h>
*/
import "C"

// Any platform-specific initialization belongs here.
func initializePlatform() error {
	// Setup crash handling specifics - *NIX-only
	if config.C.CStacktraceCollection {
		var cCoreDump int

		if config.C.CCoreDump {
			cCoreDump = 1
		}

		var handlerErr *C.char = nil
		if C.handle_crashes(C.int(cCoreDump), &handlerErr) == 0 {
			log.Errorf("Unable to install crash handler, C-land stacktraces and dumps will be unavailable: %s", C.GoString(handlerErr))
			if handlerErr != nil {
				C._free(unsafe.Pointer(handlerErr))
			}
		}
	}

	return nil
}
