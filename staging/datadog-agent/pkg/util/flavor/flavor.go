// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package flavor

const (
	// DefaultAgent is the default Agent flavor
	DefaultAgent = "agent"
	// IotAgent is the IoT Agent flavor
	IotAgent = "iot_agent"
	// ClusterAgent is the Cluster Agent flavor
	ClusterAgent = "cluster_agent"
	// Dogstatsd is the DogStatsD flavor
	Dogstatsd = "statsd"
	// SecurityAgent is the Security Agent flavor
	SecurityAgent = "security_agent"
	// ServerlessAgent is an Agent running in a serverless environment
	ServerlessAgent = "serverless_agent"
	// HerokuAgent is the Heroku Agent flavor
	HerokuAgent = "heroku_agent"
)

var agentFlavor = DefaultAgent

// SetFlavor sets the Agent flavor
//func SetFlavor(flavor string) {
//	agentFlavor = flavor
//
//	if agentFlavor == IotAgent {
//		config.Datadog.SetDefault("iot_host", true)
//	}
//}

// GetFlavor gets the running Agent flavor
// it MUST NOT be called before the main package is initialized;
// e.g. in init functions or to initialize package constants or variables.
func GetFlavor() string {
	return agentFlavor
}
