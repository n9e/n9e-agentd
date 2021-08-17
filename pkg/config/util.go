package config

import (
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/DataDog/datadog-agent/pkg/util/hostname/validate"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

func configEval(value string) string {
	switch strings.ToLower(value) {
	case "$ip":
		return getOutboundIP()
	case "$host", "$hostname":
		host, _ := os.Hostname()
		return host
	default:
		return value
	}
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
func NewSystemProbeConfig() (*Config, error) {
	return &Config{}, nil
}

func MergeSystemProbeConfig() (*Config, error) {
	return &Config{}, nil
}

// GetEnvDefault retrieves a value from the environment named by the key or return def if not set.
func GetEnvDefault(key, def string) string {
	value, found := os.LookupEnv(key)
	if !found {
		return def
	}
	return value
}

// IsContainerized returns whether the Agent is running on a Docker container
// DOCKER_DD_AGENT is set in our official Dockerfile
func IsContainerized() bool {
	return os.Getenv("DOCKER_AGENT") != ""
}

// IsDockerRuntime returns true if we are to find the /.dockerenv file
// which is typically only set by Docker
func IsDockerRuntime() bool {
	_, err := os.Stat("/.dockerenv")
	if err == nil {
		return true
	}

	return false
}

// IsKubernetes returns whether the Agent is running on a kubernetes cluster
func IsKubernetes() bool {
	// Injected by Kubernetes itself
	if os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		return true
	}
	// support of Datadog environment variable for Kubernetes
	if os.Getenv("KUBERNETES") != "" {
		return true
	}
	return false
}

// IsECSFargate returns whether the Agent is running in ECS Fargate
func IsECSFargate() bool {
	return os.Getenv("ECS_FARGATE") != ""
}

// pathExists returns true if the given path exists
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

//func GetIPCAddress() (string, error) {
//	return C.BindHost, nil
//}

// FileUsedDir returns the absolute path to the folder containing the config
// file used to populate the registry
func FileUsedDir() string {
	return filepath.Dir(ConfigFileUsed())
}

func ConfigFileUsed() string {
	return Configfile
}

// SanitizeAPIKey strips newlines and other control characters from a given string.
func SanitizeAPIKey(key string) string {
	return strings.TrimSpace(key)
}

// GetDistPath returns the fully qualified path to the 'dist' directory
func GetDistPath() string {
	return C.DistPath
}

// IsCloudProviderEnabled checks the cloud provider family provided in
// pkg/util/<cloud_provider>.go against the value for cloud_provider: on the
// global config object Datadog
func IsCloudProviderEnabled(cloudProviderName string) bool {
	cloudProviderFromConfig := C.CloudProviderMetadata

	for _, cloudName := range cloudProviderFromConfig {
		if strings.ToLower(cloudName) == strings.ToLower(cloudProviderName) {
			log.Debugf("cloud_provider_metadata is set to %s in agent configuration, trying endpoints for %s Cloud Provider",
				cloudProviderFromConfig,
				cloudProviderName)
			return true
		}
	}

	log.Debugf("cloud_provider_metadata is set to %s in agent configuration, skipping %s Cloud Provider",
		cloudProviderFromConfig,
		cloudProviderName)
	return false
}

func GetProxies() *Proxy {
	return &C.Proxy
}

// GetConfiguredTags returns complete list of user configured tags
func GetConfiguredTags(includeDogstatsd bool) []string {
	tags := C.Tags
	extraTags := C.ExtraTags

	var dsdTags []string
	if includeDogstatsd {
		dsdTags = C.Statsd.Tags
	}

	combined := make([]string, 0, len(tags)+len(extraTags)+len(dsdTags))
	combined = append(combined, tags...)
	combined = append(combined, extraTags...)
	combined = append(combined, dsdTags...)

	return combined
}

// IsCLCRunner returns whether the Agent is in cluster check runner mode
func IsCLCRunner() bool {
	if !C.CLCRunnerEnabled {
		return false
	}

	cps := C.ConfigProviders

	for _, name := range C.ExtraConfigProviders {
		cps = append(cps, ConfigurationProviders{Name: name})
	}

	// A cluster check runner is an Agent configured to run clusterchecks only
	// We want exactly one ConfigProvider named clusterchecks
	if len(cps) == 0 {
		return false
	}

	for _, cp := range cps {
		if cp.Name != "clusterchecks" {
			return false
		}
	}

	return true
}

func GetMainEndpoint() (host string) {
	if len(C.Endpoints) > 0 {
		host, _, _ = net.SplitHostPort(C.Endpoints[0])
	}
	return
}

// GetValidHostAliases validates host aliases set in `host_aliases` variable and returns
// only valid ones.
func GetValidHostAliases() []string {
	return getValidHostAliasesWithConfig()
}

func getValidHostAliasesWithConfig() []string {
	aliases := []string{}
	for _, alias := range C.HostAliases {
		if err := validate.ValidHostname(alias); err == nil {
			aliases = append(aliases, alias)
		} else {
			log.Warnf("skipping invalid host alias '%s': %s", alias, err)
		}
	}

	return aliases
}
