package util

import (
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/cache"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/ec2"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/gce"
	"k8s.io/klog/v2"
)

// GetNetworkID retrieves the network_id which can be used to improve network
// connection resolution. This can be configured or detected.  The
// following sources will be queried:
// * configuration
// * GCE
// * EC2
func GetNetworkID() (string, error) {
	cacheNetworkIDKey := cache.BuildAgentKey("networkID")
	if cacheNetworkID, found := cache.Cache.Get(cacheNetworkIDKey); found {
		return cacheNetworkID.(string), nil
	}

	// the the id from configuration
	if networkID := config.C.Network.ID; networkID != "" {
		cache.Cache.Set(cacheNetworkIDKey, networkID, cache.NoExpiration)
		klog.V(5).Infof("GetNetworkID: using configured network ID: %s", networkID)
		return networkID, nil
	}

	klog.V(5).Infof("GetNetworkID trying GCE")
	if networkID, err := gce.GetNetworkID(); err == nil {
		cache.Cache.Set(cacheNetworkIDKey, networkID, cache.NoExpiration)
		klog.V(5).Infof("GetNetworkID: using network ID from GCE metadata: %s", networkID)
		return networkID, nil
	}

	klog.V(5).Infof("GetNetworkID trying EC2")
	if networkID, err := ec2.GetNetworkID(); err == nil {
		cache.Cache.Set(cacheNetworkIDKey, networkID, cache.NoExpiration)
		klog.V(5).Infof("GetNetworkID: using network ID from EC2 metadata: %s", networkID)
		return networkID, nil
	}

	return "", fmt.Errorf("could not detect network ID")
}
