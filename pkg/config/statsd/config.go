package statsd

import "github.com/yubo/golib/api"

type Config struct {
	Enabled                  bool             `json:"enabled"`                                                                                        // use_dogstatsd
	Host                     string           `json:"host"`                                                                                           //
	Port                     int              `json:"port"`                                                                                           // dogstatsd_port
	Socket                   string           `json:"socket"`                                                                                         // dogstatsd_socket
	PipeName                 string           `json:"pipe_name"`                                                                                      // dogstatsd_pipe_name
	ContextExpirySeconds     api.Duration     `json:"context_expiry_seconds" flag:"statsd-context-expiry-seconds" description:"contextExpirySeconds"` // dogstatsd_context_expiry_seconds
	ExpirySeconds            api.Duration     `json:"expiry_seconds" flag:"statsd-expiry-seconds" description:"expirySeconds"`                        // dogstatsd_expiry_seconds
	StatsEnable              bool             `json:"stats_enable"`                                                                                   // dogstatsd_stats_enable
	StatsBuffer              int              `json:"stats_buffer"`                                                                                   // dogstatsd_stats_buffer
	MetricsStatsEnable       bool             `json:"metrics_stats_enable"`                                                                           // dogstatsd_metrics_stats_enable - for debug - move to settings
	BufferSize               int              `json:"buffer_size"`                                                                                    // dogstatsd_buffer_size
	MetricNamespace          string           `json:"metric_namespace"`                                                                               // statsd_metric_namespace
	MetricNamespaceBlacklist []string         `json:"metric_namespace_blacklist"`                                                                     // statsd_metric_namespace_blacklist
	Tags                     []string         `json:"tags"`                                                                                           // dogstatsd_tags
	EntityIdPrecedence       bool             `json:"entity_id_precedence"`                                                                           // dogstatsd_entity_id_precedence
	EolRequired              []string         `json:"eol_required"`                                                                                   // dogstatsd_eol_required
	DisableVerboseLogs       bool             `json:"disable_verbose_logs"`                                                                           // dogstatsd_disable_verbose_logs
	ForwardHost              string           `json:"forward_host"`                                                                                   // statsd_forward_host
	ForwardPort              int              `json:"forward_port"`                                                                                   // statsd_forward_port
	QueueSize                int              `json:"queue_size"`                                                                                     // dogstatsd_queue_size
	MapperCacheSize          int              `json:"mapper_cache_size"`                                                                              // dogstatsd_mapper_cache_size
	MapperProfiles           []MappingProfile `json:"mapper_profiles"`                                                                                // dogstatsd_mapper_profiles
	StringInternerSize       int              `json:"string_interner_size"`                                                                           // dogstatsd_string_interner_size
	SocketRcvbuf             int              `json:"socekt_rcvbuf"`                                                                                  // dogstatsd_so_rcvbuf
	PacketBufferSize         int              `json:"packet_buffer_size"`                                                                             // dogstatsd_packet_buffer_size

	PacketBufferFlushTimeout api.Duration `json:"packet_buffer_flush_timeout" flag:"statsd-packet-buffer-flush-timeout" description:"packetBufferFlushTimeout"` // dogstatsd_packet_buffer_flush_timeout

	TagCardinality                    string `json:"tag_cardinality"`                       // dogstatsd_tag_cardinality
	NonLocalTraffic                   bool   `json:"non_local_traffic"`                     // dogstatsd_non_local_traffic
	OriginDetection                   bool   `json:"OriginDetection"`                       // dogstatsd_origin_detection
	HistogramCopyToDistribution       bool   `json:"histogram_copy_to_distribution"`        // histogram_copy_to_distribution
	HistogramCopyToDistributionPrefix string `json:"histogram_copy_to_distribution_prefix"` // histogram_copy_to_distribution_prefix
	CaptureDepth                      int    `json:"capture_depth"`                         //dogstatsd_capture_depth
}

// MappingProfile represent a group of mappings
type MappingProfile struct {
	Name     string          `json:"name" json:"name"`
	Prefix   string          `json:"prefix" json:"prefix"`
	Mappings []MetricMapping `json:"mappings" json:"mappings"`
}

// MetricMapping represent one mapping rule
type MetricMapping struct {
	Match     string            `json:"match" json:"match"`
	MatchType string            `json:"match_type" json:"match_type"`
	Name      string            `json:"name" json:"name"`
	Tags      map[string]string `json:"tags" json:"tags"`
}

func (p *Config) Validate() error {
	return nil
}
