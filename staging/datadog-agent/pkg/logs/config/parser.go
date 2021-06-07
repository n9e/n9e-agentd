// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package config

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

// ParseJSON parses the data formatted in JSON
// returns an error if the parsing failed.
func ParseJSON(data []byte) ([]*LogsConfig, error) {
	var configs []*LogsConfig
	err := json.Unmarshal(data, &configs)
	if err != nil {
		return nil, fmt.Errorf("could not parse JSON logs config: %v", err)
	}
	return configs, nil
}

type logsConfig struct {
	Logs []*LogsConfig `yaml:"logs"`
}

// ParseYAML parses the data formatted in YAML,
// returns an error if the parsing failed.
func ParseYAML(data []byte) ([]*LogsConfig, error) {
	var config logsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("could not decode YAML logs config: %v", err)
	}

	return config.Logs, nil
}
