package statsd

import "time"

type Config struct {
	Enabled                  bool             `json:"enabled" default:"false"` // use_dogstatsd
	Host                     string           `json:"host"`                    //
	Port                     int              `json:"port" default:"8125"`     // dogstatsd_port
	Socket                   string           `json:"socket"`                  // dogstatsd_socket
	PipeName                 string           `json:"pipe_name"`               // dogstatsd_pipe_name
	ContextExpirySeconds     time.Duration    `json:"-"`
	ContextExpirySeconds_    int              `json:"context_expiry_seconds" flag:"statsd-context-expiry-seconds" default:"300" description:"contextExpirySeconds(Second)"` // dogstatsd_context_expiry_seconds
	ExpirySeconds            time.Duration    `json:"-"`
	ExpirySeconds_           int              `json:"expiry_seconds" flag:"statsd-expiry-seconds" default:"300" description:"expirySeconds(Second)"` // dogstatsd_expiry_seconds
	StatsEnable              bool             `json:"stats_enable" default:"true"`                                                                   // dogstatsd_stats_enable
	StatsBuffer              int              `json:"stats_buffer" default:"10"`                                                                     // dogstatsd_stats_buffer
	MetricsStatsEnable       bool             `json:"metrics_stats_enable" default:"false"`                                                          // dogstatsd_metrics_stats_enable - for debug
	BufferSize               int              `json:"buffer_size" default:"8192"`                                                                    // dogstatsd_buffer_size
	MetricNamespace          string           `json:"metric_namespace"`                                                                              // statsd_metric_namespace
	MetricNamespaceBlacklist []string         `json:"metric_namespace_blacklist"`                                                                    // statsd_metric_namespace_blacklist
	Tags                     []string         `json:"tags"`                                                                                          // dogstatsd_tags
	EntityIdPrecedence       bool             `json:"entity_id_precedence"`                                                                          // dogstatsd_entity_id_precedence
	EolRequired              []string         `json:"eol_required"`                                                                                  // dogstatsd_eol_required
	DisableVerboseLogs       bool             `json:"disable_verbose_logs"`                                                                          // dogstatsd_disable_verbose_logs
	ForwardHost              string           `json:"forward_host"`                                                                                  // statsd_forward_host
	ForwardPort              int              `json:"forward_port"`                                                                                  // statsd_forward_port
	QueueSize                int              `json:"queue_size" default:"1024"`                                                                     // dogstatsd_queue_size
	MapperCacheSize          int              `json:"mapper_cache_size" default:"1000"`                                                              // dogstatsd_mapper_cache_size
	MapperProfiles           []MappingProfile `json:"mapper_profiles"`                                                                               // dogstatsd_mapper_profiles
	StringInternerSize       int              `json:"string_interner_size" default:"4096"`                                                           // dogstatsd_string_interner_size
	SocketRcvbuf             int              `json:"socekt_rcvbuf"`                                                                                 // dogstatsd_so_rcvbuf
	PacketBufferSize         int              `json:"packet_buffer_size" default:"32"`                                                               // dogstatsd_packet_buffer_size
	PacketBufferFlushTimeout time.Duration    `json:"-"`

	PacketBufferFlushTimeout_ int `json:"packet_buffer_flush_timeout" flag:"statsd-packet-buffer-flush-timeout" default:"100" description:"packetBufferFlushTimeout(Millisecond)"` // dogstatsd_packet_buffer_flush_timeout

	TagCardinality                    string `json:"tag_cardinality" default:"low"`         // dogstatsd_tag_cardinality
	NonLocalTraffic                   bool   `json:"non_local_traffic"`                     // dogstatsd_non_local_traffic
	OriginDetection                   bool   `json:"OriginDetection"`                       // dogstatsd_origin_detection
	HistogramCopyToDistribution       bool   `json:"histogram_copy_to_distribution"`        // histogram_copy_to_distribution
	HistogramCopyToDistributionPrefix string `json:"histogram_copy_to_distribution_prefix"` // histogram_copy_to_distribution_prefix
	CaptureDepth                      int    `json:"capture_depth" default:"0"`             //dogstatsd_capture_depth
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
	p.ContextExpirySeconds = time.Second * time.Duration(p.ContextExpirySeconds_)
	p.ExpirySeconds = time.Second * time.Duration(p.ExpirySeconds_)
	p.PacketBufferFlushTimeout = time.Millisecond * time.Duration(p.PacketBufferFlushTimeout_)
	return nil
}
