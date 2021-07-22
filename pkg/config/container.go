package config

import (
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/system"
	"k8s.io/klog/v2"
)

const (
	// Docker socket present
	Docker Feature = "docker"
	// Containerd socket present
	Containerd Feature = "containerd"
	// Cri is any cri socket present
	Cri Feature = "cri"
	// Kubernetes environment
	Kubernetes Feature = "kubernetes"
	// ECSFargate environment
	ECSFargate Feature = "ecsfargate"
	// EKSFargate environment
	EKSFargate Feature = "eksfargate"
	// KubeOrchestratorExplorer can be enabled
	KubeOrchestratorExplorer Feature = "orchestratorExplorer"

	defaultLinuxDockerSocket       = "/var/run/docker.sock"
	defaultWindowsDockerSocketPath = "//./pipe/docker_engine"
	defaultLinuxContainerdSocket   = "/var/run/containerd/containerd.sock"
	defaultLinuxCrioSocket         = "/var/run/crio/crio.sock"
	defaultHostMountPrefix         = "/host"
	unixSocketPrefix               = "unix://"
	winNamedPipePrefix             = "npipe://"

	socketTimeout = 500 * time.Millisecond
)

type Container struct {
	DockerHost            string   `yaml:"-"`
	EksFargate            bool     `yaml:"eksFargate"`
	CriSocketPath         string   `yaml:"criSocketPath"`
	IncludeMetrics        []string `yaml:"includeMetrics"`        // container_include, container_include_metrics, ac_include
	ExcludeMetrics        []string `yaml:"excludeMetrics"`        // container_exclude, container_exclude_metrics, ac_exclude
	IncludeLogs           []string `yaml:"includeLogs"`           // container_include_logs
	ExcludeLogs           []string `yaml:"excludeLogs"`           // container_exclude_logs
	ExcludePauseContainer bool     `yaml:"excludeParseContainer"` // exclude_pause_container
}

func (p *Container) Validate() error {
	return nil
}

func (cf *Config) detectContainerFeatures(features FeatureMap) {
	cf.detectKubernetes(features)
	cf.detectDocker(features)
	cf.detectContainerd(features)
	cf.detectFargate(features)
}

func (cf *Config) detectKubernetes(features FeatureMap) {
	if IsKubernetes() {
		features[Kubernetes] = struct{}{}
		if cf.OrchestratorExplorer.Enabled {
			features[KubeOrchestratorExplorer] = struct{}{}
		}
	}
}

func (cf *Config) detectDocker(features FeatureMap) {
	if cf.Container.DockerHost != "" {
		features[Docker] = struct{}{}
	} else {
		for _, defaultDockerSocketPath := range getDefaultDockerPaths() {
			exists, reachable := system.CheckSocketAvailable(defaultDockerSocketPath, socketTimeout)
			if exists && !reachable {
				klog.Infof("Agent found Docker socket at: %s but socket not reachable (permissions?)", defaultDockerSocketPath)
				continue
			}

			if exists && reachable {
				features[Docker] = struct{}{}

				// Even though it does not modify configuration, using the OverrideFunc mechanism for uniformity
				cf.Container.DockerHost = getDefaultDockerSocketType() + defaultDockerSocketPath
				break
			}
		}
	}
}

func (cf *Config) detectContainerd(features FeatureMap) {
	// CRI Socket - Do not automatically default socket path if the Agent runs in Docker
	// as we'll very likely discover the containerd instance wrapped by Docker.
	criSocket := cf.Container.CriSocketPath
	if criSocket == "" && !IsDockerRuntime() {
		for _, defaultCriPath := range getDefaultCriPaths() {
			exists, reachable := system.CheckSocketAvailable(defaultCriPath, socketTimeout)
			if exists && !reachable {
				klog.Infof("Agent found cri socket at: %s but socket not reachable (permissions?)", defaultCriPath)
				continue
			}

			if exists && reachable {
				criSocket = defaultCriPath
				cf.Container.CriSocketPath = defaultCriPath
				// Currently we do not support multiple CRI paths
				break
			}
		}
	}

	if criSocket != "" {
		// Containerd support was historically meant for K8S
		// However, containerd is now used standalone elsewhere.
		// TODO: Consider having a dedicated setting for containerd standalone
		if IsKubernetes() {
			features[Cri] = struct{}{}
		}

		if strings.Contains(criSocket, "containerd") {
			features[Containerd] = struct{}{}
		}
	}
}

func (cf *Config) detectFargate(features FeatureMap) {
	if IsECSFargate() {
		features[ECSFargate] = struct{}{}
		return
	}

	if cf.Container.EksFargate {
		features[EKSFargate] = struct{}{}
		features[Kubernetes] = struct{}{}
	}
}

func getHostMountPrefixes() []string {
	if IsContainerized() {
		return []string{"", defaultHostMountPrefix}
	}
	return []string{""}
}

func getDefaultDockerSocketType() string {
	if runtime.GOOS == "windows" {
		return winNamedPipePrefix
	}

	return unixSocketPrefix
}

func getDefaultDockerPaths() []string {
	if runtime.GOOS == "windows" {
		return []string{defaultWindowsDockerSocketPath}
	}

	paths := []string{}
	for _, prefix := range getHostMountPrefixes() {
		paths = append(paths, path.Join(prefix, defaultLinuxDockerSocket))
	}
	return paths
}

func getDefaultCriPaths() []string {
	paths := []string{}
	for _, prefix := range getHostMountPrefixes() {
		paths = append(paths, path.Join(prefix, defaultLinuxContainerdSocket), path.Join(prefix, defaultLinuxCrioSocket))
	}
	return paths
}
