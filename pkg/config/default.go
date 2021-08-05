package config

import (
	"os"
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/pkg/collector/check/defaults"
	"github.com/yubo/golib/configer"
)

const (
	//authTokenName = "auth_token"

	// DefaultSite is the default site the Agent sends data to.
	//DefaultSite    = "datadoghq.com"
	//infraURLPrefix = "https://app."

	// DefaultNumWorkers default number of workers for our check runner
	DefaultNumWorkers = 4
	// MaxNumWorkers maximum number of workers for our check runner
	MaxNumWorkers = 25
	// DefaultAPIKeyValidationInterval is the default interval of api key validation checks
	DefaultAPIKeyValidationInterval = 3600 * time.Second

	// DefaultForwarderRecoveryInterval is the default recovery interval,
	// also used if the user-provided value is invalid.
	DefaultForwarderRecoveryInterval = 2

	megaByte = 1024 * 1024

	// DefaultAuditorTTL is the default logs auditor TTL in hours
	DefaultAuditorTTL = 23

	// ClusterIDCacheKey is the key name for the orchestrator cluster id in the agent in-mem cache
	ClusterIDCacheKey = "orchestratorClusterID"

	// DefaultRuntimePoliciesDir is the default policies directory used by the runtime security module
	DefaultRuntimePoliciesDir = "/etc/datadog-agent/runtime-security.d"

	// DefaultLogsSenderBackoffFactor is the default logs sender backoff randomness factor
	DefaultLogsSenderBackoffFactor = 2.0

	// DefaultLogsSenderBackoffBase is the default logs sender base backoff time, seconds
	DefaultLogsSenderBackoffBase = 1.0

	// DefaultLogsSenderBackoffMax is the default logs sender maximum backoff time, seconds
	DefaultLogsSenderBackoffMax = 120.0

	// DefaultLogsSenderBackoffRecoveryInterval is the default logs sender backoff recovery interval
	DefaultLogsSenderBackoffRecoveryInterval = 2
)

func NewDefaultConfig(configer *configer.Configer) *Config {
	cf := &Config{
		m:        new(sync.RWMutex),
		configer: configer,
	}

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
	cf.Logs.AuditorTTL = DefaultAuditorTTL
	cf.PythonVersion = DefaultPython

	return cf
}
