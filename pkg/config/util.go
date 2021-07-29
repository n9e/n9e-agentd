package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yubo/apiserver/pkg/options"
)

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

var _ipcAddress string

func GetIPCAddress() (string, error) {
	if _ipcAddress != "" {
		return _ipcAddress, nil
	}

	server, ok := options.ApiServerFrom(Context)
	if !ok {
		return "", fmt.Errorf("unable to get ipc address")
	}
	_ipcAddress = server.Address()
	return _ipcAddress, nil
}

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
