package config

import (
	"os"
	"time"

	"github.com/n9e/n9e-agentd/pkg/collector/check/defaults"
)

func NewDefaultConfig() *Config {
	cf := &Config{}

	if IsContainerized() {
		// In serverless-containerized environments (e.g Fargate)
		// it's impossible to mount host volumes.
		// Make sure the host paths exist before setting-up the default values.
		// Fallback to the container paths if host paths aren't mounted.
		if pathExists("/host/proc") {
			cf.ProcfsPath = "/host/proc"
			cf.ContainerProcRoot = "/host/proc"

			// Used by some librairies (like gopsutil)
			if v := os.Getenv("HOST_PROC"); v == "" {
				os.Setenv("HOST_PROC", "/host/proc")
			}
		} else {
			cf.ProcfsPath = "/proc"
			cf.ContainerProcRoot = "/proc"
		}
		if pathExists("/host/sys/fs/cgroup/") {
			cf.ContainerCgroupRoot = "/host/sys/fs/cgroup/"
		} else {
			cf.ContainerCgroupRoot = "/sys/fs/cgroup/"
		}
	} else {
		cf.ContainerProcRoot = "/proc"
		// for amazon linux the cgroup directory on host is /cgroup/
		// we pick memory.stat to make sure it exists and not empty
		if _, err := os.Stat("/cgroup/memory/memory.stat"); !os.IsNotExist(err) {
			cf.ContainerCgroupRoot = "/cgroup/"
		} else {
			cf.ContainerCgroupRoot = "/sys/fs/cgroup/"
		}

	}

	cf.Statsd.MetricNamespaceBlacklist = StandardStatsdPrefixes
	cf.Jmx.CheckPeriod = int(defaults.DefaultCheckInterval / time.Millisecond)
	cf.LogsConfig.AuditorTTL = DefaultAuditorTTL
	cf.PythonVersion = DefaultPython

	return cf
}
