// +build docker

package util

import (
	"fmt"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/cache"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/containers"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/docker"
	"k8s.io/klog/v2"
)

// GetAgentUTSMode retrieves from Docker the UTS mode of the Agent container
func GetAgentUTSMode() (containers.UTSMode, error) {
	cacheUTSModeKey := cache.BuildAgentKey("utsMode")
	if cacheUTSMode, found := cache.Cache.Get(cacheUTSModeKey); found {
		return cacheUTSMode.(containers.UTSMode), nil
	}

	klog.V(5).Infof("GetAgentUTSMode trying docker")
	utsMode, err := docker.GetAgentContainerUTSMode()
	cache.Cache.Set(cacheUTSModeKey, utsMode, cache.NoExpiration)
	if err != nil {
		return utsMode, fmt.Errorf("could not detect agent UTS mode: %v", err)
	}
	klog.V(5).Infof("GetAgentUTSMode: using UTS mode from Docker: %s", utsMode)
	return utsMode, nil
}
