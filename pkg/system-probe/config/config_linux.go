//+build linux

package config

import (
	"fmt"
	"path/filepath"
)

const (
	// defaultSystemProbeAddress is the default unix socket path to be used for connecting to the system probe
	defaultSystemProbeAddress = "/opt/n9e/agentd/run/sysprobe.sock"

	defaultConfigDir = "/opt/n9e/agentd/etcd"
)

// ValidateSocketAddress validates that the sysprobe socket config option is of the correct format.
func ValidateSocketAddress(sockPath string) error {
	if !filepath.IsAbs(sockPath) {
		return fmt.Errorf("socket path must be an absolute file path: %s", sockPath)
	}
	return nil
}
