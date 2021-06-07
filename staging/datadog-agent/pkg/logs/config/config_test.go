// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	. "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/types"

	coreConfig "github.com/n9e/n9e-agentd/pkg/config"
)

type ConfigTestSuite struct {
	suite.Suite
	config *coreConfig.Config
}

func (suite *ConfigTestSuite) SetupTest() {
	suite.config = &coreConfig.Config{}
}

func (suite *ConfigTestSuite) TestDefaultDatadogConfig() {
	suite.config = coreConfig.NewDefaultConfig()
}

func (suite *ConfigTestSuite) TestDefaultSources() {
	// container collect all source

	source := ContainerCollectAllSource()
	suite.Nil(source)

	suite.config.LogsConfig.ContainerCollectAll = true

	source = ContainerCollectAllSource()
	suite.NotNil(source)

	suite.Equal("container_collect_all", source.Name)
	suite.Equal(DockerType, source.Config.Type)
	suite.Equal("docker", source.Config.Source)
	suite.Equal("docker", source.Config.Service)
}

func (suite *ConfigTestSuite) TestGlobalProcessingRulesShouldReturnNoRulesWithEmptyValues() {
	var (
		rules []*ProcessingRule
		err   error
	)

	suite.config.LogsConfig.ProcessingRules = nil

	rules, err = GlobalProcessingRules()
	suite.Nil(err)
	suite.Equal(0, len(rules))

	suite.config.LogsConfig.ProcessingRules = nil

	rules, err = GlobalProcessingRules()
	suite.Nil(err)
	suite.Equal(0, len(rules))
}

func (suite *ConfigTestSuite) TestGlobalProcessingRulesShouldReturnRulesWithValidMap() {
	var (
		rules []*ProcessingRule
		rule  *ProcessingRule
		err   error
	)

	suite.config.LogsConfig.ProcessingRules = []*ProcessingRule{
		&ProcessingRule{
			Type:    "exclude_at_match",
			Name:    "exclude_foo",
			Pattern: "foo",
		},
	}

	rules, err = GlobalProcessingRules()
	suite.Nil(err)
	suite.Equal(1, len(rules))

	rule = rules[0]
	suite.Equal(ExcludeAtMatch, rule.Type)
	suite.Equal("exclude_foo", rule.Name)
	suite.Equal("foo", rule.Pattern)
	suite.NotNil(rule.Regex)
}

func (suite *ConfigTestSuite) TestGlobalProcessingRulesShouldReturnRulesWithValidJSONString() {
	var (
		rules []*ProcessingRule
		rule  *ProcessingRule
		err   error
	)

	suite.config.LogsConfig.ProcessingRules = []*ProcessingRule{
		&ProcessingRule{
			Type:               "mask_sequences",
			Name:               "mask_api_keys",
			ReplacePlaceholder: "****************************",
			Pattern:            "([A-Fa-f0-9]{28})",
		}}

	rules, err = GlobalProcessingRules()
	suite.Nil(err)
	suite.Equal(1, len(rules))

	rule = rules[0]
	suite.Equal(MaskSequences, rule.Type)
	suite.Equal("mask_api_keys", rule.Name)
	suite.Equal("([A-Fa-f0-9]{28})", rule.Pattern)
	suite.Equal("****************************", rule.ReplacePlaceholder)
	suite.NotNil(rule.Regex)
}

func (suite *ConfigTestSuite) TestTaggerWarmupDuration() {
	// assert TaggerWarmupDuration is disabled by default
	taggerWarmupDuration := TaggerWarmupDuration()
	suite.Equal(0*time.Second, taggerWarmupDuration)

	// override
	suite.config.LogsConfig.TaggerWarmupDuration = 5 * time.Second
	taggerWarmupDuration = TaggerWarmupDuration()
	suite.Equal(5*time.Second, taggerWarmupDuration)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (suite *ConfigTestSuite) TestMultipleHttpEndpointsEnvVar() {
	suite.config.ApiKey = "123"
	suite.config.LogsConfig.BatchWait = time.Second
	suite.config.LogsConfig.LogsUrl = "agent-http-intake.logs.datadoghq.com:443"
	suite.config.LogsConfig.UseCompression = true
	suite.config.LogsConfig.CompressionLevel = 6
	suite.config.LogsConfig.UseSSL = true

	os.Setenv("DD_LOGS_CONFIG_ADDITIONAL_ENDPOINTS", `[
	{"api_key": "456", "host": "additional.endpoint.1", "port": 1234, "use_compression": true, "compression_level": 2},
	{"api_key": "789", "host": "additional.endpoint.2", "port": 1234, "use_compression": true, "compression_level": 2}]`)
	defer os.Unsetenv("DD_LOGS_CONFIG_ADDITIONAL_ENDPOINTS")

	expectedMainEndpoint := Endpoint{
		APIKey:           "123",
		Host:             "agent-http-intake.logs.datadoghq.com",
		Port:             443,
		UseSSL:           true,
		UseCompression:   true,
		CompressionLevel: 6}
	expectedAdditionalEndpoint1 := Endpoint{
		APIKey:           "456",
		Host:             "additional.endpoint.1",
		Port:             1234,
		UseSSL:           true,
		UseCompression:   true,
		CompressionLevel: 2}
	expectedAdditionalEndpoint2 := Endpoint{
		APIKey:           "789",
		Host:             "additional.endpoint.2",
		Port:             1234,
		UseSSL:           true,
		UseCompression:   true,
		CompressionLevel: 2}

	expectedEndpoints := NewEndpoints(expectedMainEndpoint, []Endpoint{expectedAdditionalEndpoint1, expectedAdditionalEndpoint2}, false, true, time.Second, 0)
	endpoints, err := BuildHTTPEndpoints()

	suite.Nil(err)
	suite.Equal(expectedEndpoints, endpoints)
}

func (suite *ConfigTestSuite) TestMultipleTCPEndpointsEnvVar() {
	suite.config.ApiKey = "123"
	suite.config.LogsConfig.LogsUrl = "agent-http-intake.logs.datadoghq.com:443"
	suite.config.LogsConfig.UseSSL = true
	suite.config.LogsConfig.Socks5ProxyAddress = "proxy.test:3128"
	suite.config.LogsConfig.DevModeUseProto = true

	os.Setenv("DD_LOGS_CONFIG_ADDITIONAL_ENDPOINTS", `[{"api_key": "456      \n", "host": "additional.endpoint", "port": 1234}]`)
	defer os.Unsetenv("DD_LOGS_CONFIG_ADDITIONAL_ENDPOINTS")

	expectedMainEndpoint := Endpoint{
		APIKey:           "123",
		Host:             "agent-http-intake.logs.datadoghq.com",
		Port:             443,
		UseSSL:           true,
		UseCompression:   false,
		CompressionLevel: 0,
		ProxyAddress:     "proxy.test:3128"}
	expectedAdditionalEndpoint := Endpoint{
		APIKey:           "456",
		Host:             "additional.endpoint",
		Port:             1234,
		UseSSL:           true,
		UseCompression:   false,
		CompressionLevel: 0,
		ProxyAddress:     "proxy.test:3128"}

	expectedEndpoints := NewEndpoints(expectedMainEndpoint, []Endpoint{expectedAdditionalEndpoint}, true, false, 0, 0)
	endpoints, err := buildTCPEndpoints()

	suite.Nil(err)
	suite.Equal(expectedEndpoints, endpoints)
}
