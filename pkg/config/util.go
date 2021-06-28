package config

import "os"

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
