package utils

import (
	"github.com/n9e/n9e-agentd/pkg/config"
)

// IsCheckEnabled returns if we want telemetry for the given check.
// Returns true if a * is present in the telemetry.checks list.
func IsCheckEnabled(checkName string) bool {
	// false if telemetry is disabled
	if !IsEnabled() {
		return false
	}

	// by default, we don't enable telemetry for every checks stats
	for _, check := range config.C.Telemetry.Checks {
		if check == "*" {
			return true
		} else if check == checkName {
			return true
		}
	}
	return false
}

// IsEnabled returns whether or not telemetry is enabled
func IsEnabled() bool {
	return config.C.Telemetry.Enabled
}
