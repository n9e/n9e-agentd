package settings

import (
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/config"
)

// ProfilingGoroutines wraps runtime.SetBlockProfileRate setting
type ProfilingGoroutines (string)

// Name returns the name of the runtime setting
func (r ProfilingGoroutines) Name() string {
	return string(r)
}

// Description returns the runtime setting's description
func (r ProfilingGoroutines) Description() string {
	return "This setting controls whether internal profiling will collect goroutine stacktraces (requires profiling restart)"
}

// Hidden returns whether or not this setting is hidden from the list of runtime settings
func (r ProfilingGoroutines) Hidden() bool {
	return true
}

// Get returns the current value of the runtime setting
func (r ProfilingGoroutines) Get() (interface{}, error) {
	return config.C.InternalProfiling.EnableGoroutineStacktraces, nil
}

// Set changes the value of the runtime setting
func (r ProfilingGoroutines) Set(value interface{}) error {
	return fmt.Errorf("unsupported")
	//enabled, err := GetBool(value)
	//if err != nil {
	//	return err
	//}

	//config.Datadog.Set("internal_profiling.enable_goroutine_stacktraces", enabled)

	//return nil
}
