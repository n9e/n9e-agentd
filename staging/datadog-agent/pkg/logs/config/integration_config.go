// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package config

import (
	"fmt"
	"strings"

	. "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/types"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/serverless/aws"
)

// Logs source types
const (
	TCPType           = "tcp"
	UDPType           = "udp"
	FileType          = "file"
	DockerType        = "docker"
	JournaldType      = "journald"
	WindowsEventType  = "windows_event"
	SnmpTrapsType     = "snmp_traps"
	StringChannelType = "string_channel"

	// UTF16BE for UTF-16 Big endian encoding
	UTF16BE string = "utf-16-be"
	// UTF16LE for UTF-16 Little Endian encoding
	UTF16LE string = "utf-16-le"
)

// LogsConfig represents a log source config, which can be for instance
// a file to tail or a port to listen to.
type LogsConfig struct {
	Type          string   `yaml:"type"`
	Port          int      `yaml:"port"`          // Network
	Path          string   `yaml:"path"`          // File, Journald
	Encoding      string   `yaml:"encoding"`      // File
	ExcludePaths  []string `yaml:"excludePaths"`  // File
	TailingMode   string   `yaml:"startPosition"` // File
	IncludeUnits  []string `yaml:"includeUnits"`  // Journald
	ExcludeUnits  []string `yaml:"excludeUnits"`  // Journald
	ContainerMode bool     `yaml:"containerMode"` // Journald
	Image         string   `yaml:"image"`         // Docker
	Label         string   `yaml:"label"`         // Docker
	Name          string   `yaml:"name"`          // Docker Name contains the container name
	Identifier    string   `yaml:"identifier"`    // Docker Identifier contains the container ID
	ChannelPath   string   `yaml:"channelPath"`   // Windows Event
	Query         string   `yaml:"query"`         // Windows Event

	// used as input only by the Channel tailer.
	// could have been unidirectional but the tailer could not close it in this case.
	// TODO(remy): strongly typed to an AWS Lambda LogMessage, we should probably use
	// a more generic type here.
	Channel chan aws.LogMessage `yaml:"channel"`

	Service         string            `yaml:"service"`
	Source          string            `yaml:"source"`
	SourceCategory  string            `yaml:"sourceCategory"`
	Tags            []string          `yaml:"tags"`
	ProcessingRules []*ProcessingRule `yaml:"logProcessingRules"`
}

// TailingMode type
type TailingMode uint8

// Tailing Modes
const (
	ForceBeginning = iota
	ForceEnd
	Beginning
	End
)

var tailingModeTuples = []struct {
	s string
	m TailingMode
}{
	{"forceBeginning", ForceBeginning},
	{"forceEnd", ForceEnd},
	{"beginning", Beginning},
	{"end", End},
}

// TailingModeFromString parses a string and returns a corresponding tailing mode, default to End if not found
func TailingModeFromString(mode string) (TailingMode, bool) {
	for _, t := range tailingModeTuples {
		if t.s == mode {
			return t.m, true
		}
	}
	return End, false
}

// TailingModeToString returns seelog string representation for a specified tailing mode. Returns "" for invalid tailing mode.
func (mode TailingMode) String() string {
	for _, t := range tailingModeTuples {
		if t.m == mode {
			return t.s
		}
	}
	return ""
}

// Validate returns an error if the config is misconfigured
func (c *LogsConfig) Validate() error {
	switch {
	case c.Type == "":
		// user don't have to specify a logs-config type when defining
		// an autodiscovery label because so we must override it at some point,
		// this check is mostly used for sanity purposed to detect an override miss.
		return fmt.Errorf("a config must have a type")
	case c.Type == FileType:
		if c.Path == "" {
			return fmt.Errorf("file source must have a path")
		}
		err := c.validateTailingMode()
		if err != nil {
			return err
		}
	case c.Type == TCPType && c.Port == 0:
		return fmt.Errorf("tcp source must have a port")
	case c.Type == UDPType && c.Port == 0:
		return fmt.Errorf("udp source must have a port")
	}
	err := ValidateProcessingRules(c.ProcessingRules)
	if err != nil {
		return err
	}
	return CompileProcessingRules(c.ProcessingRules)
}

func (c *LogsConfig) validateTailingMode() error {
	mode, found := TailingModeFromString(c.TailingMode)
	if !found && c.TailingMode != "" {
		return fmt.Errorf("invalid tailing mode '%v' for %v", c.TailingMode, c.Path)
	}
	if ContainsWildcard(c.Path) && (mode == Beginning || mode == ForceBeginning) {
		return fmt.Errorf("tailing from the beginning is not supported for wildcard path %v", c.Path)
	}
	return nil
}

// ContainsWildcard returns true if the path contains any wildcard character
func ContainsWildcard(path string) bool {
	return strings.ContainsAny(path, "*?[")
}
