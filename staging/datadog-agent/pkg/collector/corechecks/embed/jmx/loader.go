// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build jmx

package jmx

import (
	"errors"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/check"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/loaders"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

// JMXCheckLoader is a specific loader for checks living in this package
type JMXCheckLoader struct{}

// NewJMXCheckLoader creates a loader for go checks
func NewJMXCheckLoader() (*JMXCheckLoader, error) {
	state.runner.initRunner()
	return &JMXCheckLoader{}, nil
}

// Load returns JMX loader name
func (jl *JMXCheckLoader) Name() string {
	return "jmx"
}

// Load returns a JMX check
func (jl *JMXCheckLoader) Load(config integration.Config, instance integration.Data) (check.Check, error) {
	var c check.Check

	if !check.IsJMXInstance(config.Name, instance, config.InitConfig) {
		return c, errors.New("check is not a jmx check, or unable to determine if it's so")
	}

	if err := state.runner.configureRunner(instance, config.InitConfig); err != nil {
		klog.Errorf("jmx.loader: could not configure check: %s", err)
		return c, err
	}

	// Validate common instance structure
	commonOptions := integration.CommonInstanceConfig{}
	err := yaml.Unmarshal(instance, &commonOptions)
	if err != nil {
		klog.V(5).Infof("jmx.loader: invalid instance for check %s: %s", config.Name, err)
	}

	cf := integration.Config{
		ADIdentifiers: config.ADIdentifiers,
		Entity:        config.Entity,
		InitConfig:    config.InitConfig,
		Instances:     []integration.Data{instance},
		LogsConfig:    config.LogsConfig,
		MetricConfig:  config.MetricConfig,
		Name:          config.Name,
		Provider:      config.Provider,
	}
	c = newJMXCheck(cf, config.Source)

	return c, err
}

func (jl *JMXCheckLoader) String() string {
	return "JMX Check Loader"
}

func (jl *JMXCheckLoader) SampleConfig() string {
	return "TODO"
}

func init() {
	factory := func() (check.Loader, error) {
		return NewJMXCheckLoader()
	}

	loaders.RegisterLoader(10, factory)
}