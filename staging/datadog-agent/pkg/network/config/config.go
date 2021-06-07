package config

import (
	"strings"
	"time"

	ddconfig "github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/ebpf"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/kernel"
	"k8s.io/klog/v2"
)

const (
	defaultUDPTimeoutSeconds       = 30
	defaultUDPStreamTimeoutSeconds = 120

	defaultOffsetThreshold = 400
	maxOffsetThreshold     = 3000
)

// Config stores all flags used by the network eBPF tracer
type Config struct {
	ebpf.Config

	// CollectTCPConns specifies whether the tracer should collect traffic statistics for TCP connections
	CollectTCPConns bool

	// CollectUDPConns specifies whether the tracer should collect traffic statistics for UDP connections
	CollectUDPConns bool

	// CollectIPv6Conns specifics whether the tracer should capture traffic for IPv6 TCP/UDP connections
	CollectIPv6Conns bool

	// CollectLocalDNS specifies whether the tracer should capture traffic for local DNS calls
	CollectLocalDNS bool

	// DNSInspection specifies whether the tracer should enhance connection data with domain names by inspecting DNS traffic
	// Notice this does *not* depend on CollectLocalDNS
	DNSInspection bool

	// CollectDNSStats specifies whether the tracer should enhance connection data with relevant DNS stats
	// It is relevant *only* when DNSInspection is enabled.
	CollectDNSStats bool

	// CollectDNSDomains specifies whether collected DNS stats would be scoped by domain
	// It is relevant *only* when DNSInspection and CollectDNSStats is enabled.
	CollectDNSDomains bool

	// DNSTimeout determines the length of time to wait before considering a DNS Query to have timed out
	DNSTimeout time.Duration

	// MaxDNSStats determines the number of separate DNS Stats objects DNSStatkeeper can have at any given time
	// These stats objects get flushed on every client request (default 30s check interval)
	MaxDNSStats int

	// EnableHTTPMonitoring specifies whether the tracer should monitor HTTP traffic
	EnableHTTPMonitoring bool

	// UDPConnTimeout determines the length of traffic inactivity between two
	// (IP, port)-pairs before declaring a UDP connection as inactive. This is
	// set to /proc/sys/net/netfilter/nf_conntrack_udp_timeout on Linux by
	// default.
	UDPConnTimeout time.Duration

	// UDPStreamTimeout is the timeout for udp streams. This is set to
	// /proc/sys/net/netfilter/nf_conntrack_udp_timeout_stream on Linux by
	// default.
	UDPStreamTimeout time.Duration

	// TCPConnTimeout is like UDPConnTimeout, but for TCP connections. TCP connections are cleared when
	// the BPF module receives a tcp_close call, but TCP connections also age out to catch cases where
	// tcp_close is not intercepted for some reason.
	TCPConnTimeout time.Duration

	// TCPClosedTimeout represents the maximum amount of time a closed TCP connection can remain buffered in eBPF before
	// being marked as idle and flushed to the perf ring.
	TCPClosedTimeout time.Duration

	// MaxTrackedConnections specifies the maximum number of connections we can track. This determines the size of the eBPF Maps
	MaxTrackedConnections uint

	// MaxClosedConnectionsBuffered represents the maximum number of closed connections we'll buffer in memory. These closed connections
	// get flushed on every client request (default 30s check interval)
	MaxClosedConnectionsBuffered int

	// MaxDNSStatsBuffered represents the maximum number of DNS stats we'll buffer in memory. These stats
	// get flushed on every client request (default 30s check interval)
	MaxDNSStatsBuffered int

	// MaxHTTPStatsBuffered represents the maximum number of HTTP stats we'll buffer in memory. These stats
	// get flushed on every client request (default 30s check interval)
	MaxHTTPStatsBuffered int

	// MaxConnectionsStateBuffered represents the maximum number of state objects that we'll store in memory. These state objects store
	// the stats for a connection so we can accurately determine traffic change between client requests.
	MaxConnectionsStateBuffered int

	// ClientStateExpiry specifies the max time a client (e.g. process-agent)'s state will be stored in memory before being evicted.
	ClientStateExpiry time.Duration

	// EnableConntrack enables probing conntrack for network address translation
	EnableConntrack bool

	// IgnoreConntrackInitFailure will ignore any conntrack initialization failiures during system-probe load. If this is set to false, system-probe
	// will fail to start if there is a conntrack initialization failure.
	IgnoreConntrackInitFailure bool

	// ConntrackMaxStateSize specifies the maximum number of connections with NAT we can track
	ConntrackMaxStateSize int

	// ConntrackRateLimit specifies the maximum number of netlink messages *per second* that can be processed
	// Setting it to -1 disables the limit and can result in a high CPU usage.
	ConntrackRateLimit int

	// EnableConntrackAllNamespaces enables network address translation via netlink for all namespaces that are peers of the root namespace.
	// default is true
	EnableConntrackAllNamespaces bool

	// ClosedChannelSize specifies the size for closed channel for the tracer
	ClosedChannelSize int

	// ExcludedSourceConnections is a map of source connections to blacklist
	ExcludedSourceConnections map[string][]string

	// ExcludedDestinationConnections is a map of destination connections to blacklist
	ExcludedDestinationConnections map[string][]string

	// OffsetGuessThreshold is the size of the byte threshold we will iterate over when guessing offsets
	OffsetGuessThreshold uint64

	// EnableMonotonicCount (Windows only) determines if we will calculate send/recv bytes of connections with headers and retransmits
	EnableMonotonicCount bool

	// DriverBufferSize (Windows only) determines the size (in bytes) of the buffer we pass to the driver when reading flows
	DriverBufferSize int

	// EnableGatewayLookup enables looking up gateway information for connection destinations
	EnableGatewayLookup bool
}

func join(pieces ...string) string {
	return strings.Join(pieces, ".")
}

// New creates a config for the network tracer
func New() *Config {
	sp := ddconfig.C.SystemProbe
	net := ddconfig.C.Network

	c := &Config{
		Config: *ebpf.NewConfig(),

		CollectTCPConns:  !sp.DisableTcp,
		TCPConnTimeout:   2 * time.Minute,
		TCPClosedTimeout: 1 * time.Second,

		CollectUDPConns:  !sp.DisableUdp,
		UDPConnTimeout:   defaultUDPTimeoutSeconds * time.Second,
		UDPStreamTimeout: defaultUDPStreamTimeoutSeconds * time.Second,

		CollectIPv6Conns:               !sp.DisableIpv6,
		OffsetGuessThreshold:           uint64(sp.OffsetGuessThreshold),
		ExcludedSourceConnections:      sp.SourceExcludes,
		ExcludedDestinationConnections: sp.DestExcludes,

		MaxTrackedConnections:        uint(sp.MaxTrackedConnections),
		MaxClosedConnectionsBuffered: sp.MaxClosedConnectionsBuffered,
		ClosedChannelSize:            sp.ClosedChannelSize,
		MaxConnectionsStateBuffered:  sp.MaxConnectionStateBuffered,
		ClientStateExpiry:            2 * time.Minute,

		DNSInspection:       !sp.DisableDnsInspection,
		CollectDNSStats:     sp.CollectDnsStats,
		CollectLocalDNS:     sp.CollectLocalDns,
		CollectDNSDomains:   sp.CollectDnsDomains,
		MaxDNSStats:         sp.MaxDnsStats,
		MaxDNSStatsBuffered: 75000,
		DNSTimeout:          sp.DnsTimeout,

		EnableHTTPMonitoring: net.EnableHttpMonitoring,
		MaxHTTPStatsBuffered: 100000,

		EnableConntrack:              sp.EnableConntrack,
		ConntrackMaxStateSize:        sp.ConntrackMaxStateSize,
		ConntrackRateLimit:           sp.ConntrackRateLimit,
		EnableConntrackAllNamespaces: sp.EnableConntrackAllNamespaces,
		IgnoreConntrackInitFailure:   net.IgnoreConntrackInitFailure,

		EnableGatewayLookup: net.EnableGatewayLookup,

		EnableMonotonicCount: sp.WindowsEnableMonotonicCount,
		DriverBufferSize:     sp.WindowsDriverBufferSize,
	}

	if c.OffsetGuessThreshold > maxOffsetThreshold {
		klog.Warning("offset_guess_threshold exceeds maximum of 3000. Setting it to the default of 400")
		c.OffsetGuessThreshold = defaultOffsetThreshold
	}

	if !kernel.IsIPv6Enabled() {
		c.CollectIPv6Conns = false
		klog.Info("network tracer IPv6 tracing disabled by system")
	} else if !c.CollectIPv6Conns {
		klog.Info("network tracer IPv6 tracing disabled by configuration")
	}

	if !c.CollectUDPConns {
		klog.Info("network tracer UDP tracing disabled by configuration")
	}
	if !c.CollectTCPConns {
		klog.Info("network tracer TCP tracing disabled by configuration")
	}
	if !c.DNSInspection {
		klog.Info("network tracer DNS inspection disabled by configuration")
	}

	return c
}
