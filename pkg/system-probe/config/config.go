package config

import (
	"time"

	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// ModuleName is a typed alias for string, used only for module names
type ModuleName string

const (
	// Namespace is the top-level configuration key that all system-probe settings are nested underneath
	Namespace                    = "system_probe_config"
	spNS                         = Namespace
	defaultConfigFileName        = "system-probe.yaml"
	defaultConnsMessageBatchSize = 600
	maxConnsMessageBatchSize     = 1000

	// system-probe module names
	NetworkTracerModule        ModuleName = "network_tracer"
	OOMKillProbeModule         ModuleName = "oom_kill_probe"
	TCPQueueLengthTracerModule ModuleName = "tcp_queue_length_tracer"
	SecurityRuntimeModule      ModuleName = "security_runtime"
	ProcessModule              ModuleName = "process"
)

type Config struct {
	Enabled        bool            `json:"enabled" default:"false" env:"DD_SYSTEM_PROBE_ENABLED"`
	EnabledModules map[string]bool `json:"enabled_modules"`

	// When the system-probe is enabled in a separate container, we need a way to also disable the system-probe
	// packaged in the main agent container (without disabling network collection on the process-agent).
	ExternalSystemProbe bool   `json:"external" default:"false" env:"DD_SYSTEM_PROBE_EXTERNAL"`                              // external
	SocketAddress       string `json:"sysprobe_socket" default:"/opt/n9e/agentd/run/sysprobe.sock" env:"DD_SYSPROBE_SOCKET"` // sysprobe_socket
	MaxConnsPerMessage  int    `json:"max_conns_per_message" default:"600"`                                                  // max_conns_per_message
	LogFile             string `json:"log_file" default:"/opt/n9e/agentd/logs/system-probe.log"`                             // log_file
	LogLevel            string `json:"log_level" default:"info" env:"DD_LOG_LEVEL"`                                          // log_level
	DebugPort           int    `json:"debug_port" default:"0"`                                                               // debug_port
	StatsdHost          string `json:"-" default:"127.0.0.1"`                                                                // GetBindHost()
	StatsdPort          int    `json:"dogstatsd_port" default:"8125"`                                                        // dogstatsd_port

	SysprobeSocket               string              `json:"sysprobe_socket"`                 // system_probe_config.sysprobe_socket
	BPFDebug                     bool                `json:"bpf_debug"`                       // system_probe_config.bpf_debug
	BPFDir                       string              `json:"bpf_dir"`                         // system_probe_config.bpf_dir
	ExcludedLinuxVersions        []string            `json:"excluded_linux_versions"`         // system_probe_config.excluded_linux_versions
	EnableTracepoints            bool                `json:"enable_tracepoints"`              // system_probe_config.enable_tracepoints
	EnableRuntimeCompiler        bool                `json:"enable_runtime_compiler"`         // system_probe_config.enable_runtime_compiler
	RuntimeCompilerOutputDir     string              `json:"runtime_compiler_output_dir"`     // system_probe_config.runtime_compiler_output_dir
	KernelHeaderDirs             []string            `json:"kernel_header_dirs"`              // system_probe_config.kernel_header_dirs
	DisableTcp                   bool                `json:"disable_tcp"`                     // system_probe_config.disable_tcp
	DisableUdp                   bool                `json:"disable_udp"`                     // system_probe_config.disable_udp
	DisableIpv6                  bool                `json:"disable_ipv6"`                    // system_probe_config.disable_ipv6
	OffsetGuessThreshold         int64               `json:"offset_guess_threshold"`          // system_probe_config.offset_guess_threshold
	SourceExcludes               map[string][]string `json:"source_excludes"`                 // system_probe_config.source_excludes
	DestExcludes                 map[string][]string `json:"dest_excludes"`                   // system_probe_config.dest_excludes
	MaxTrackedConnections        int                 `json:"max_tracked_connections"`         // system_probe_config.max_tracked_connections
	MaxClosedConnectionsBuffered int                 `json:"max_closed_connections_buffered"` // system_probe_config.max_closed_connections_buffered
	ClosedChannelSize            int                 `json:"closed_channel_size"`             // system_probe_config.closed_channel_size
	MaxConnectionStateBuffered   int                 `json:"max_connection_state_buffered"`   // system_probe_config.max_connection_state_buffered
	DisableDnsInspection         bool                `json:"disable_dns_inspection"`          // system_probe_config.disable_dns_inspection
	CollectDnsStats              bool                `json:"collect_dns_stats"`               // system_probe_config.collect_dns_stats
	CollectLocalDns              bool                `json:"collect_local_dns"`               // system_probe_config.collect_local_dns
	CollectDnsDomains            bool                `json:"collect_dns_domains"`             // system_probe_config.collect_dns_domains
	MaxDnsStats                  int                 `json:"max_dns_stats"`                   // system_probe_config.max_dns_stats
	DnsTimeout                   time.Duration       `json:"-"`
	DnsTimeout_                  int                 `json:"dns_timeout" flag:"system-probe-dns-timeout" default:"15" description:"dnsTimeout(Second)"` // system_probe_config.dns_timeout_in_s
	EnableConntrack              bool                `json:"enable_conntrack"`                                                                          // system_probe_config.enable_conntrack
	ConntrackMaxStateSize        int                 `json:"conntrack_max_state_size"`                                                                  // system_probe_config.conntrack_max_state_size
	ConntrackRateLimit           int                 `json:"conntrack_rate_limit"`                                                                      // system_probe_config.conntrack_rate_limit
	EnableConntrackAllNamespaces bool                `json:"enable_conntrack_all_namespaces"`                                                           // system_probe_config.enable_conntrack_all_namespaces
	WindowsEnableMonotonicCount  bool                `json:"windows_enable_monotonic_count"`                                                            // system_probe_config.windows.enable_monotonic_count
	WindowsDriverBufferSize      int                 `json:"windows_driver_buffer_size"`                                                                // system_probe_config.windows.driver_buffer_size
	KernelHeadersDownloadDir     string              `json:"kernel_header_download_dir"`                                                                // system_probe_config.kernel_header_download_dir
}

func (p *Config) Validate() error {
	p.DnsTimeout = time.Second * time.Duration(p.DnsTimeout_)

	if p.MaxConnsPerMessage > maxConnsMessageBatchSize {
		log.Warn("Overriding the configured connections count per message limit because it exceeds maximum")
		p.MaxConnsPerMessage = defaultConnsMessageBatchSize
	}

	return nil
}

// ModuleIsEnabled returns a bool indicating if the given module name is enabled.
func (c Config) ModuleIsEnabled(modName string) bool {
	_, ok := c.EnabledModules[modName]
	return ok
}
