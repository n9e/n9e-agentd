package snmp

import (
	"errors"
	"fmt"
	"hash/fnv"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/gosnmp/gosnmp"
	"github.com/n9e/n9e-agentd/pkg/api"
)

const (
	// snmp
	defaultPort    = 161
	defaultTimeout = 5 * time.Second
	defaultRetries = 3

	// traps
	defaultTrapsPort   = uint16(162) // Standard UDP port for traps.
	defaultStopTimeout = 5
	packetsChanSize    = 100
)

// ListenerConfig holds global configuration for SNMP discovery
type ListenerConfig struct {
	Workers               int      `json:"workers"`
	DiscoveryInterval     int      `json:"discovery_interval"`
	AllowedFailures       int      `json:"discovery_allowed_failures"`
	Loader                string   `json:"loader"`
	CollectDeviceMetadata bool     `json:"collect_device_metadata"`
	Configs               []Config `json:"configs"`

	// legacy
	AllowedFailuresLegacy int `json:"allowed_failures"`
}

func (c *ListenerConfig) Validate() error {
	if c.AllowedFailures == 0 && c.AllowedFailuresLegacy != 0 {
		c.AllowedFailures = c.AllowedFailuresLegacy
	}

	// Set the default values, we can't otherwise on an array
	for i := range c.Configs {
		// We need to modify the struct in place
		config := &c.Configs[i]
		if config.Port == 0 {
			config.Port = defaultPort
		}
		if config.Timeout.Duration == 0 {
			config.Timeout.Duration = defaultTimeout
		}
		if config.Retries == 0 {
			config.Retries = defaultRetries
		}
		if config.CollectDeviceMetadataConfig != nil {
			config.CollectDeviceMetadata = *config.CollectDeviceMetadataConfig
		} else {
			config.CollectDeviceMetadata = c.CollectDeviceMetadata
		}
		if config.Loader == "" {
			config.Loader = c.Loader
		}
		config.Community = firstNonEmpty(config.Community, config.CommunityLegacy)
		config.AuthKey = firstNonEmpty(config.AuthKey, config.AuthKeyLegacy)
		config.AuthProtocol = firstNonEmpty(config.AuthProtocol, config.AuthProtocolLegacy)
		config.PrivKey = firstNonEmpty(config.PrivKey, config.PrivKeyLegacy)
		config.PrivProtocol = firstNonEmpty(config.PrivProtocol, config.PrivProtocolLegacy)
		config.Network = firstNonEmpty(config.Network, config.NetworkLegacy)
		config.Version = firstNonEmpty(config.Version, config.VersionLegacy)
	}

	return nil
}

// Config holds configuration for a particular subnet
type Config struct {
	Network                     string          `json:"network_address"`
	Port                        uint16          `json:"port"`
	Version                     string          `json:"snmp_version"`
	Timeout                     api.Duration    `json:"timeout" description:"dial timeout"`
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

// Digest returns an hash value representing the data stored in this configuration, minus the network address
func (c *Config) Digest(address string) string {
	h := fnv.New64()
	// Hash write never returns an error
	h.Write([]byte(address))                   //nolint:errcheck
	h.Write([]byte(fmt.Sprintf("%d", c.Port))) //nolint:errcheck
	h.Write([]byte(c.Version))                 //nolint:errcheck
	h.Write([]byte(c.Community))               //nolint:errcheck
	h.Write([]byte(c.User))                    //nolint:errcheck
	h.Write([]byte(c.AuthKey))                 //nolint:errcheck
	h.Write([]byte(c.AuthProtocol))            //nolint:errcheck
	h.Write([]byte(c.PrivKey))                 //nolint:errcheck
	h.Write([]byte(c.PrivProtocol))            //nolint:errcheck
	h.Write([]byte(c.ContextEngineID))         //nolint:errcheck
	h.Write([]byte(c.ContextName))             //nolint:errcheck
	h.Write([]byte(c.Loader))                  //nolint:errcheck

	// Sort the addresses to get a stable digest
	addresses := make([]string, 0, len(c.IgnoredIPAddresses))
	for ip := range c.IgnoredIPAddresses {
		addresses = append(addresses, ip)
	}
	sort.Strings(addresses)
	for _, ip := range addresses {
		h.Write([]byte(ip)) //nolint:errcheck
	}

	return strconv.FormatUint(h.Sum64(), 16)
}

// BuildSNMPParams returns a valid GoSNMP struct to start making queries
func (c *Config) BuildSNMPParams(deviceIP string) (*gosnmp.GoSNMP, error) {
	if c.Community == "" && c.User == "" {
		return nil, errors.New("No authentication mechanism specified")
	}

	var version gosnmp.SnmpVersion
	if c.Version == "1" {
		version = gosnmp.Version1
	} else if c.Version == "2" || (c.Version == "" && c.Community != "") {
		version = gosnmp.Version2c
	} else if c.Version == "3" || (c.Version == "" && c.User != "") {
		version = gosnmp.Version3
	} else {
		return nil, fmt.Errorf("SNMP version not supported: %s", c.Version)
	}

	var authProtocol gosnmp.SnmpV3AuthProtocol
	lowerAuthProtocol := strings.ToLower(c.AuthProtocol)
	if lowerAuthProtocol == "" {
		authProtocol = gosnmp.NoAuth
	} else if lowerAuthProtocol == "md5" {
		authProtocol = gosnmp.MD5
	} else if lowerAuthProtocol == "sha" {
		authProtocol = gosnmp.SHA
	} else {
		return nil, fmt.Errorf("Unsupported authentication protocol: %s", c.AuthProtocol)
	}

	var privProtocol gosnmp.SnmpV3PrivProtocol
	lowerPrivProtocol := strings.ToLower(c.PrivProtocol)
	if lowerPrivProtocol == "" {
		privProtocol = gosnmp.NoPriv
	} else if lowerPrivProtocol == "des" {
		privProtocol = gosnmp.DES
	} else if lowerPrivProtocol == "aes" {
		privProtocol = gosnmp.AES
	} else if lowerPrivProtocol == "aes192" {
		privProtocol = gosnmp.AES192
	} else if lowerPrivProtocol == "aes192c" {
		privProtocol = gosnmp.AES192C
	} else if lowerPrivProtocol == "aes256" {
		privProtocol = gosnmp.AES256
	} else if lowerPrivProtocol == "aes256c" {
		privProtocol = gosnmp.AES256C
	} else {
		return nil, fmt.Errorf("Unsupported privacy protocol: %s", c.PrivProtocol)
	}

	msgFlags := gosnmp.NoAuthNoPriv
	if c.PrivKey != "" {
		msgFlags = gosnmp.AuthPriv
	} else if c.AuthKey != "" {
		msgFlags = gosnmp.AuthNoPriv
	}

	return &gosnmp.GoSNMP{
		Target:          deviceIP,
		Port:            c.Port,
		Community:       c.Community,
		Transport:       "udp",
		Version:         version,
		Timeout:         c.Timeout.Duration,
		Retries:         c.Retries,
		SecurityModel:   gosnmp.UserSecurityModel,
		MsgFlags:        msgFlags,
		ContextEngineID: c.ContextEngineID,
		ContextName:     c.ContextName,
		SecurityParameters: &gosnmp.UsmSecurityParameters{
			UserName:                 c.User,
			AuthenticationProtocol:   authProtocol,
			AuthenticationPassphrase: c.AuthKey,
			PrivacyProtocol:          privProtocol,
			PrivacyPassphrase:        c.PrivKey,
		},
	}, nil
}

// IsIPIgnored checks the given IP against IgnoredIPAddresses
func (c *Config) IsIPIgnored(ip net.IP) bool {
	ipString := ip.String()
	_, present := c.IgnoredIPAddresses[ipString]
	return present
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// snmp.traps

// Config contains configuration for SNMP trap listeners.
// YAML field tags provided for test marshalling purposes.
type TrapsConfig struct {
	Enabled          bool     `json:"enabled"`
	Port             uint16   `json:"port" yaml:"port"`
	CommunityStrings []string `json:"community_strings" yaml:"community_strings"`
	BindHost         string   `json:"bind_host" yaml:"bind_host"`
	StopTimeout      int      `json:"stop_timeout" yaml:"stop_timeout"`
}

func (c *TrapsConfig) Validate(bindHost string) error {
	if !c.Enabled {
		return nil
	}

	// Validate required fields.
	if c.CommunityStrings == nil || len(c.CommunityStrings) == 0 {
		return errors.New("`community_strings` is required and must be non-empty")
	}

	// Set defaults.
	if c.Port == 0 {
		c.Port = defaultTrapsPort
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

func (c *TrapsConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.BindHost, c.Port)
}

// BuildV2Params returns a valid GoSNMP SNMPv2 params structure from configuration.
func (c *TrapsConfig) BuildV2Params() *gosnmp.GoSNMP {
	return &gosnmp.GoSNMP{
		Port:      c.Port,
		Transport: "udp",
		Version:   gosnmp.Version2c,
		Logger:    gosnmp.NewLogger(&trapLogger{}),
	}
}

// trapLogger is a GoSNMP logger interface implementation.
type trapLogger struct {
	gosnmp.Logger
}

// NOTE: GoSNMP logs show the full content of decoded trap packets. Logging as DEBUG would be too noisy.
func (logger *trapLogger) Print(v ...interface{}) {
	log.Trace(v...)
}
func (logger *trapLogger) Printf(format string, v ...interface{}) {
	log.Tracef(format, v...)
}
