// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package agent implements the api endpoints for the `/agent` prefix.
// This group of endpoints is meant to provide high-level functionalities
// at the agent level.

// +build jmx

package apiserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/embed/jmx"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/status"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

func getJMXConfigs(w http.ResponseWriter, r *http.Request) {
	var ts int
	queries := r.URL.Query()
	if timestamps, ok := queries["timestamp"]; ok {
		ts, _ = strconv.Atoi(timestamps[0])
	}

	if int64(ts) > jmx.GetScheduledConfigsModificationTimestamp() {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	klog.V(5).Infof("Getting latest JMX Configs as of: %#v", ts)

	j := map[string]interface{}{}
	configs := map[string]integration.JSONMap{}

	for name, config := range jmx.GetScheduledConfigs() {
		var rawInitConfig integration.RawMap
		err := yaml.Unmarshal(config.InitConfig, &rawInitConfig)
		if err != nil {
			klog.Errorf("unable to parse JMX configuration: %s", err)
			http.Error(w, err.Error(), 500)
			return
		}

		c := map[string]interface{}{}
		c["init_config"] = util.GetJSONSerializableMap(rawInitConfig)
		instances := []integration.JSONMap{}
		for _, instance := range config.Instances {
			var rawInstanceConfig integration.JSONMap
			err := yaml.Unmarshal(instance, &rawInstanceConfig)
			if err != nil {
				klog.Errorf("unable to parse JMX configuration: %s", err)
				http.Error(w, err.Error(), 500)
				return
			}
			instances = append(instances, util.GetJSONSerializableMap(rawInstanceConfig).(integration.JSONMap))
		}

		c["instances"] = instances
		c["check_name"] = config.Name

		configs[name] = c
	}
	j["configs"] = configs
	j["timestamp"] = time.Now().Unix()
	jsonPayload, err := json.Marshal(util.GetJSONSerializableMap(j))
	if err != nil {
		klog.Errorf("unable to parse JMX configuration: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(jsonPayload) //nolint:errcheck
}

func setJMXStatus(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var jmxStatus status.JMXStatus
	err := decoder.Decode(&jmxStatus)
	if err != nil {
		klog.Errorf("unable to parse jmx status: %s", err)
		http.Error(w, err.Error(), 500)
	}

	status.SetJMXStatus(jmxStatus)
}
