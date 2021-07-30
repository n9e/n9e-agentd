package config

import "errors"

// SnmpListenerConfig holds global configuration for SNMP discovery
type SnmpListenerConfig struct {
	Workers               int          `json:"workers"`
	DiscoveryInterval     int          `json:"discovery_interval"`
	AllowedFailures       int          `json:"discovery_allowed_failures"`
	Loader                string       `json:"loader"`
	CollectDeviceMetadata bool         `json:"collect_device_metadata"`
	Configs               []SnmpConfig `json:"configs"`

	// legacy
	AllowedFailuresLegacy int `json:"allowed_failures"`
}

func (c *SnmpListenerConfig) Validate() error {
	return nil
}

// SnmpConfig holds configuration for a particular subnet
type SnmpConfig struct {
	Network                     string          `json:"network_address"`
	Port                        uint16          `json:"port"`
	Version                     string          `json:"snmp_version"`
	Timeout                     int             `json:"timeout"`
	Retries                     int             `json:"retries"`
	OidBatchSize                int             `json:"oid_batch_size"`
	Community                   string          `json:"community_string"`
	User                        string          `json:"user"`
	AuthKey                     string          `json:"auth_key"`
	AuthProtocol                string          `json:"auth_Protocol"`
	PrivKey                     string          `json:"priv_key"`
	PrivProtocol                string          `json:"priv_protocol"`
	ContextEngineID             string          `json:"context_engine_id"`
	ContextName                 string          `json:"context_name"`
	IgnoredIPAddresses          map[string]bool `json:"ignored_ip_addresses"`
	ADIdentifier                string          `json:"ad_identifier"`
	Loader                      string          `json:"loader"`
	CollectDeviceMetadataConfig *bool           `json:"collect_device_metadata"`
	CollectDeviceMetadata       bool
	Tags                        []string `json:"tags"`

	// Legacy
	NetworkLegacy      string `json:"network"`
	VersionLegacy      string `json:"version"`
	CommunityLegacy    string `json:"community"`
	AuthKeyLegacy      string `json:"authentication_key"`
	AuthProtocolLegacy string `json:"authentication_protocol"`
	PrivKeyLegacy      string `json:"privacy_key"`
	PrivProtocolLegacy string `json:"privacy_protocol"`
}

// snmp.traps

// Config contains configuration for SNMP trap listeners.
// YAML field tags provided for test marshalling purposes.
type SnmpTrapsConfig struct {
	Enabled          bool     `json:"enabled"`
	Port             uint16   `json:"port" yaml:"port"`
	CommunityStrings []string `json:"community_strings" yaml:"community_strings"`
	BindHost         string   `json:"bind_host" yaml:"bind_host"`
	StopTimeout      int      `json:"stop_timeout" yaml:"stop_timeout"`
}

const (
	defaultPort        = uint16(162) // Standard UDP port for traps.
	defaultStopTimeout = 5
	packetsChanSize    = 100
)

func (c *SnmpTrapsConfig) Validate(bindHost string) error {
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
		c.BindHost = bindHost
	}
	if c.StopTimeout == 0 {
		c.StopTimeout = defaultStopTimeout
	}

	return nil
}
