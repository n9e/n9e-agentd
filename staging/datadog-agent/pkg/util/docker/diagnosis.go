// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build docker

package docker

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/diagnose/diagnosis"
	"k8s.io/klog/v2"
)

func init() {
	diagnosis.Register("Docker availability", diagnose)
}

// diagnose the docker availability on the system
func diagnose() error {
	_, err := GetDockerUtil()
	if err != nil {
		klog.Error(err)
	} else {
		klog.Info("successfully connected to docker")
	}

	hostname, err := HostnameProvider()
	if err != nil {
		klog.Errorf("returned hostname %q with error: %s", hostname, err)
	} else {
		klog.Infof("successfully got hostname %q from docker", hostname)
	}
	return err
}
