// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package config

import (
	"fmt"
	"net"
	"strconv"
	"time"

	coreConfig "github.com/n9e/n9e-agentd/pkg/config"
	. "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/types"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/snmp/traps"
	"k8s.io/klog/v2"
)

// ContainerCollectAll is the name of the docker integration that collect logs from all containers
const ContainerCollectAll = "container_collect_all"

// SnmpTraps is the name of the integration that collects logs from SNMP traps received by the Agent
const SnmpTraps = "snmp_traps"

// logs-intake endpoint prefix.
const (
//tcpEndpointPrefix  = "agent-intake.logs."
//httpEndpointPrefix = "agent-http-intake.logs."
)

// logs-intake endpoints depending on the site and environment.
var logsEndpoints = map[string]int{
	"agent-intake.logs.datadoghq.com": 10516,
	"agent-intake.logs.datadoghq.eu":  443,
	"agent-intake.logs.datad0g.com":   10516,
	"agent-intake.logs.datad0g.eu":    443,
}

// HTTPConnectivity is the status of the HTTP connectivity
type HTTPConnectivity bool

var (
	// HTTPConnectivitySuccess is the status for successful HTTP connectivity
	HTTPConnectivitySuccess HTTPConnectivity = true
	// HTTPConnectivityFailure is the status for failed HTTP connectivity
	HTTPConnectivityFailure HTTPConnectivity = false
)

// ContainerCollectAllSource returns a source to collect all logs from all containers.
func ContainerCollectAllSource() *LogSource {
	if coreConfig.C.LogsConfig.ContainerCollectAll {
		// source to collect all logs from all containers
		return NewLogSource(ContainerCollectAll, &LogsConfig{
			Type:    DockerType,
			Service: "docker",
			Source:  "docker",
		})
	}
	return nil
}

// SNMPTrapsSource returs a source to forward SNMP traps as logs.
func SNMPTrapsSource() *LogSource {
	if coreConfig.C.SnmpTraps.Enabled && traps.IsRunning() {
		// source to forward SNMP traps as logs.
		return NewLogSource(SnmpTraps, &LogsConfig{
			Type:    SnmpTrapsType,
			Service: "snmp",
			Source:  "snmp",
		})
	}
	return nil
}

// GlobalProcessingRules returns the global processing rules to apply to all logs.
func GlobalProcessingRules() ([]*ProcessingRule, error) {
	rules := coreConfig.C.LogsConfig.ProcessingRules
	if err := ValidateProcessingRules(rules); err != nil {
		return nil, err
	}
	if err := CompileProcessingRules(rules); err != nil {
		return nil, err
	}
	return rules, nil
}

// BuildEndpoints returns the endpoints to send logs.
func BuildEndpoints(httpConnectivity HTTPConnectivity) (*Endpoints, error) {
	if coreConfig.C.LogsConfig.DevModeNoSSL {
		klog.Warningf("Use of illegal configuration parameter, if you need to send your logs to a proxy, please use 'logs_config.logs_dd_url' and 'logs_config.logs_no_ssl' instead")
	}
	if isForceHTTPUse() || (bool(httpConnectivity) && !(isForceTCPUse() || isSocks5ProxySet() || hasAdditionalEndpoints())) {
		return BuildHTTPEndpoints()
	}
	klog.Warning("You are currently sending Logs to Datadog through TCP (either because logs_config.use_tcp or logs_config.socks5_proxy_address is set or the HTTP connectivity test has failed) " +
		"To benefit from increased reliability and better network performances, " +
		"we strongly encourage switching over to compressed HTTPS which is now the default protocol.")
	return buildTCPEndpoints()
}

// ExpectedTagsDuration returns a duration of the time expected tags will be submitted for.
func ExpectedTagsDuration() time.Duration {
	return coreConfig.C.LogsConfig.ExpectedTagsDuration
}

// IsExpectedTagsSet returns boolean showing if expected tags feature is enabled.
func IsExpectedTagsSet() bool {
	return ExpectedTagsDuration() > 0
}

func isSocks5ProxySet() bool {
	return len(coreConfig.C.LogsConfig.Socks5ProxyAddress) > 0
}

func isForceTCPUse() bool {
	return coreConfig.C.LogsConfig.UseTCP
}

func isForceHTTPUse() bool {
	return coreConfig.C.LogsConfig.UseHTTP
}

func hasAdditionalEndpoints() bool {
	return len(coreConfig.C.LogsConfig.AdditionalEndpoints) > 0
}

func buildTCPEndpoints() (*Endpoints, error) {
	cf := coreConfig.C.LogsConfig
	useProto := cf.DevModeUseProto
	proxyAddress := cf.Socks5ProxyAddress
	main := Endpoint{
		APIKey:                  getLogsAPIKey(),
		ProxyAddress:            proxyAddress,
		ConnectionResetInterval: cf.ConnectionResetInterval,
	}
	if len(cf.LogsUrl) > 0 {
		// Proxy settings, expect 'logs_config.logs_dd_url' to respect the format '<HOST>:<PORT>'
		// and '<PORT>' to be an integer.
		// By default ssl is enabled ; to disable ssl set 'logs_config.logs_no_ssl' to true.
		host, port, err := parseAddress(cf.LogsUrl)
		if err != nil {
			return nil, fmt.Errorf("could not parse logs_dd_url: %v", err)
		}
		main.Host = host
		main.Port = port
		main.UseSSL = cf.UseSSL
	} else if cf.UsePort443 {
		main.Host = cf.Url443
		main.Port = 443
		main.UseSSL = true
	} else {
		return nil, fmt.Errorf("could not get logs url")
	}

	additionals := cf.AdditionalEndpoints
	for i := 0; i < len(additionals); i++ {
		additionals[i].UseSSL = main.UseSSL
		additionals[i].ProxyAddress = proxyAddress
		additionals[i].APIKey = additionals[i].APIKey
	}
	return NewEndpoints(main, additionals, useProto, false, 0, 0), nil
}

// BuildHTTPEndpoints returns the HTTP endpoints to send logs to.
func BuildHTTPEndpoints() (*Endpoints, error) {
	cf := coreConfig.C.LogsConfig
	defaultUseSSL := cf.UseSSL

	main := Endpoint{
		APIKey:                  getLogsAPIKey(),
		UseCompression:          cf.UseCompression,
		CompressionLevel:        cf.CompressionLevel,
		ConnectionResetInterval: cf.ConnectionResetInterval,
	}

	host, port, err := parseAddress(cf.URL)
	if err != nil {
		return nil, fmt.Errorf("could not parse logs_dd_url: %v", err)
	}
	main.Host = host
	main.Port = port
	main.UseSSL = defaultUseSSL

	additionals := cf.AdditionalEndpoints
	for i := 0; i < len(additionals); i++ {
		additionals[i].UseSSL = main.UseSSL
		additionals[i].APIKey = additionals[i].APIKey
	}

	batchWait := cf.BatchWait
	batchMaxConcurrentSend := cf.BatchMaxConcurrentSend
	klog.V(10).Infof("batchWait %f", batchWait.Seconds())

	return NewEndpoints(main, additionals, false, true,
		batchWait, batchMaxConcurrentSend), nil
}

// getLogsAPIKey provides the dd api key used by the main logs agent sender.
func getLogsAPIKey() string {
	if key := coreConfig.C.LogsConfig.APIKey; len(key) > 0 {
		return key
	}
	return coreConfig.C.ApiKey
}

// parseAddress returns the host and the port of the address.
func parseAddress(address string) (string, int, error) {
	host, portString, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, err
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

// TaggerWarmupDuration is used to configure the tag providers
func TaggerWarmupDuration() time.Duration {
	return coreConfig.C.LogsConfig.TaggerWarmupDuration
}

// AggregationTimeout is used when performing aggregation operations
func AggregationTimeout() time.Duration {
	return coreConfig.C.LogsConfig.AggregationTimeout
}
