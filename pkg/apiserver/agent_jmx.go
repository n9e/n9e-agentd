// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package agent implements the api endpoints for the `/agent` prefix.
// This group of endpoints is meant to provide high-level functionalities
// at the agent level.

//go:build jmx
// +build jmx

package apiserver

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/embed/jmx"
	"github.com/DataDog/datadog-agent/pkg/status"
	"github.com/DataDog/datadog-agent/pkg/util"
	"github.com/yubo/apiserver/pkg/rest"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

type JMXConfig struct {
	Instances  []integration.JSONMap `json:"instances"`
	InitConfig interface{}           `json:"init_config"`
	CheckName  string                `json:"check_name"`
}

type JMXConfigs struct {
	Configs   map[string]JMXConfig `json:"configs"`
	Timestamp int64                `json:"timestamp"`
}

func getJMXConfigs(w http.ResponseWriter, r *http.Request) (*JMXConfigs, error) {
	var ts int
	queries := r.URL.Query()
	if timestamps, ok := queries["timestamp"]; ok {
		ts, _ = strconv.Atoi(timestamps[0])
	}

	if int64(ts) > jmx.GetScheduledConfigsModificationTimestamp() {
		w.WriteHeader(http.StatusNoContent)
		return nil, nil
	}

	w.Header().Set("Content-Type", "application/json")
	klog.V(5).Infof("Getting latest JMX Configs as of: %#v", ts)

	configs := map[string]JMXConfig{}

	for name, config := range jmx.GetScheduledConfigs() {
		var rawInitConfig integration.RawMap
		err := yaml.Unmarshal(config.InitConfig, &rawInitConfig)
		if err != nil {
			return nil, fmt.Errorf("unable to parse JMX configuration: %s", err)
		}

		c := JMXConfig{}
		c.InitConfig = util.GetJSONSerializableMap(rawInitConfig)
		instances := []integration.JSONMap{}
		for _, instance := range config.Instances {
			var rawInstanceConfig integration.JSONMap
			err := yaml.Unmarshal(instance, &rawInstanceConfig)
			if err != nil {
				return nil, fmt.Errorf("unable to parse JMX configuration: %s", err)
			}
			instances = append(instances, util.GetJSONSerializableMap(rawInstanceConfig).(integration.JSONMap))
		}

		c.Instances = instances
		c.CheckName = config.Name

		configs[name] = c
	}
	return &JMXConfigs{Configs: configs, Timestamp: time.Now().Unix()}, nil
}

func setJMXStatus(w http.ResponseWriter, r *http.Request, _ *rest.NonParam, jmxStatus *status.JMXStatus) {
	status.SetJMXStatus(*jmxStatus)
}
