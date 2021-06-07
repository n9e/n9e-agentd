// +build docker

package util

import (
	"fmt"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/cache"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/docker"
	"k8s.io/klog/v2"
)

// GetAgentNetworkMode retrieves from Docker the network mode of the Agent container
func GetAgentNetworkMode() (string, error) {
	cacheNetworkModeKey := cache.BuildAgentKey("networkMode")
	if cacheNetworkMode, found := cache.Cache.Get(cacheNetworkModeKey); found {
		return cacheNetworkMode.(string), nil
	}

	klog.V(5).Infof("GetAgentNetworkMode trying Docker")
	networkMode, err := docker.GetAgentContainerNetworkMode()
	cache.Cache.Set(cacheNetworkModeKey, networkMode, cache.NoExpiration)
	if err != nil {
		return networkMode, fmt.Errorf("could not detect agent network mode: %v", err)
	}
	klog.V(5).Infof("GetAgentNetworkMode: using network mode from Docker: %s", networkMode)
	return networkMode, nil
}
