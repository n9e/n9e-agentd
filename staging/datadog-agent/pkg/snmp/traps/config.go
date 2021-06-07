// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2020-present Datadog, Inc.

package traps

import (
	"errors"
	"fmt"
	"time"

	"github.com/soniah/gosnmp"
)

// IsEnabled returns whether SNMP trap collection is enabled in the Agent configuration.
//func IsEnabled() bool {
//	return config.C.SnmpTraps.Enabled
//}

// Config contains configuration for SNMP trap listeners.
// YAML field tags provided for test marshalling purposes.
type Config struct {
	Enabled          bool          `yaml:"enabled"`          // snmp_traps_enabled
	Port             uint16        `yaml:"port"`             // snmp_traps_config.port
	CommunityStrings []string      `yaml:"communityStrings"` // snmp_traps_config.community_strings
	BindHost         string        `yaml:"bindHost"`         // snmp_traps_config.bind_host
	StopTimeout      time.Duration `yaml:"stopTimeout"`      // snmp_traps_config.stop_timeout
}

// ReadConfig builds and returns configuration from Agent configuration.
func (c *Config) Validate(bindhost string) error {
	if !c.Enabled {
		return nil
	}

	// Validate required fields.
	if c.CommunityStrings == nil || len(c.CommunityStrings) == 0 {
		return errors.New("`community_strings` is required and must be non-empty")
	}

	// Set defaults.
	if c.Port == 0 {
		c.Port = defaultPort
	}
	if c.BindHost == "" {
		// Default to global bind_host option.
		c.BindHost = bindhost
	}
	if c.StopTimeout == 0 {
		c.StopTimeout = defaultStopTimeout
	}

	return nil
}

// Addr returns the host:port address to listen on.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.BindHost, c.Port)
}

// BuildV2Params returns a valid GoSNMP SNMPv2 params structure from configuration.
func (c *Config) BuildV2Params() *gosnmp.GoSNMP {
	return &gosnmp.GoSNMP{
		Port:      c.Port,
		Transport: "udp",
		Version:   gosnmp.Version2c,
		Logger:    &trapLogger{},
	}
}
