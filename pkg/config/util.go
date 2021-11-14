package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/DataDog/datadog-agent/pkg/util/hostname/validate"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/n9e/n9e-agentd/pkg/version"
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

//func GetIPCAddress() (string, error) {
//	return C.BindHost, nil
//}

// FileUsedDir returns the absolute path to the folder containing the config
// file used to populate the registry
//func FileUsedDir() string {
//	return filepath.Join(C.RootDir, "etc")
//}

func ConfigFilesUsed() string {
	return strings.Join(C.ValueFiles, ",")
}

func ConfigFileUsed() string {
	if len(C.ValueFiles) > 0 {
		return C.ValueFiles[0]
	}
	return ""
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
		s := C.Endpoints[0]
		if n := strings.Index(s, "://"); n > 0 {
			s = s[n+3:]
		}
		host, _, _ = net.SplitHostPort(s)
	}
	return
}

// GetMultipleEndpoints returns the api keys per domain specified in the main agent config
func GetMultipleEndpoints() (map[string][]string, error) {
	return getMultipleEndpointsWithConfig(C)
}

// getMultipleEndpointsWithConfig implements the logic to extract the api keys per domain from an agent config
func getMultipleEndpointsWithConfig(config *Config) (map[string][]string, error) {
	endpoints := strings.Join(config.Endpoints, ",")

	keysPerDomain := map[string][]string{
		endpoints: {config.ApiKey},
	}

	additionalEndpoints := config.Forwarder.AdditionalEndpoints
	// merge additional endpoints into keysPerDomain
	for _, addition := range additionalEndpoints {
		endpoints := strings.Join(addition.Endpoints, ",")

		if _, ok := keysPerDomain[endpoints]; ok {
			for _, apiKey := range addition.ApiKeys {
				keysPerDomain[endpoints] = append(keysPerDomain[endpoints], apiKey)
			}
		} else {
			keysPerDomain[endpoints] = addition.ApiKeys
		}
	}

	// dedupe api keys and remove domains with no api keys (or empty ones)
	// for endpoints, apiKeys := range keysPerDomain {
	// 	dedupedAPIKeys := make([]string, 0, len(apiKeys))
	// 	seen := make(map[string]bool)
	// 	for _, apiKey := range apiKeys {
	// 		trimmedAPIKey := strings.TrimSpace(apiKey)
	// 		if _, ok := seen[trimmedAPIKey]; !ok && trimmedAPIKey != "" {
	// 			seen[trimmedAPIKey] = true
	// 			dedupedAPIKeys = append(dedupedAPIKeys, trimmedAPIKey)
	// 		}
	// 	}

	// 	if len(dedupedAPIKeys) > 0 {
	// 		keysPerDomain[endpoints] = dedupedAPIKeys
	// 	} else {
	// 		klog.Infof("No API key provided for domain \"%s\", removing domain from endpoints", endpoints)
	// 		delete(keysPerDomain, endpoints)
	// 	}
	// }

	return keysPerDomain, nil
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

// AddAgentVersionToDomain prefixes the domain with the agent version: X-Y-Z.domain
func AddAgentVersionToDomain(DDURL string, app string) (string, error) {
	u, err := url.Parse(DDURL)
	if err != nil {
		return "", err
	}

	subdomain := strings.Split(u.Host, ".")[0]
	newSubdomain := getDomainPrefix(app)

	u.Host = strings.Replace(u.Host, subdomain, newSubdomain, 1)
	return u.String(), nil
}

// getDomainPrefix provides the right prefix for agent X.Y.Z
func getDomainPrefix(app string) string {
	v, _ := version.Agent()
	return fmt.Sprintf("%d-%d-%d-%s.agent", v.Major, v.Minor, v.Patch, app)
}
