// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build python,kubelet

package python

import (
	"errors"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/cache"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/kubernetes/kubelet"
	"k8s.io/klog/v2"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/retry"
)

/*
#include <datadog_agent_rtloader.h>
#cgo !windows LDFLAGS: -ldatadog-agent-rtloader -ldl
#cgo windows LDFLAGS: -ldatadog-agent-rtloader -lstdc++ -static
*/
import "C"

var (
	kubeletCacheKey = cache.BuildAgentKey("py", "kubeutil", "connection_info")

	// for testing
	getConnectionsFunc = getConnections
)

func getConnections() map[string]string {
	kubeutil, err := kubelet.GetKubeUtil()
	if err != nil {
		// Connection to the kubelet fail, return the error
		klog.Errorf("connection to kubelet failed: %v", err)
		var e *retry.Error
		if errors.As(err, &e) {
			return map[string]string{"err": e.Unwrap().Error()}
		}
		return map[string]string{"err": err.Error()}
	}

	// At this point, we have valid credentials to get
	return kubeutil.GetRawConnectionInfo()
}

// GetKubeletConnectionInfo returns a dict containing url and credentials to connect to the kubelet.
// The dict is empty if the kubelet was not detected. The call to kubeutil is cached for 5 minutes.
// See the documentation of kubelet.GetRawConnectionInfo for dict contents.
//export GetKubeletConnectionInfo
func GetKubeletConnectionInfo(payload **C.char) {
	var creds string
	var ok bool

	if cached, hit := cache.Cache.Get(kubeletCacheKey); hit {
		klog.V(5).Info("cache hit for kubelet connection info")
		if creds, ok = cached.(string); !ok {
			klog.Error("invalid cache format, forcing a cache miss")
			creds = ""
		}
	}

	if creds == "" { // Cache miss
		klog.V(5).Info("cache miss for kubelet connection info")
		connections := getConnectionsFunc()
		if connections == nil {
			return
		}

		data, err := yaml.Marshal(connections)
		if err != nil {
			klog.Errorf("could not serialized kubelet connections (%s): %s", connections, err)
			return
		}

		creds = string(data)
		cache.Cache.Set(kubeletCacheKey, creds, 5*time.Minute)
	}

	*payload = TrackedCString(creds)
}