package logs

import (
	"strings"
	"time"

	"github.com/n9e/n9e-agentd/pkg/api"
	"k8s.io/klog/v2"
)

const (
	DefaultBatchWait = 5 * time.Second

	// DefaultBatchMaxConcurrentSend is the default HTTP batch max concurrent send for logs
	DefaultBatchMaxConcurrentSend = 0

	// DefaultBatchMaxSize is the default HTTP batch max size (maximum number of events in a single batch) for logs
	DefaultBatchMaxSize = 100

	// DefaultBatchMaxContentSize is the default HTTP batch max content size (before compression) for logs
	// It is also the maximum possible size of a single event. Events exceeding this limit are dropped.
	DefaultBatchMaxContentSize = 1000000
)

type Config struct {
	Enabled                     bool              `json:"enabled"`                                                       // logs_enabled
	AdditionalEndpoints         []Endpoint        `json:"additional_endpoints"`                                          // logs_config.additional_endpoints
	ContainerCollectAll         bool              `json:"container_collect_all"`                                         // logs_config.container_collect_all
	ProcessingRules             []*ProcessingRule `json:"processing_rules"`                                              // logs_config.processing_rules
	APIKey                      string            `json:"api_key"`                                                       // logs_config.api_key
	DevModeNoSSL                bool              `json:"dev_mode_no_ssl"`                                               // logs_config.dev_mode_no_ssl
	FileScanPeriod              api.Duration      `json:"file_scan_period" default:"10s" description:"file scan period"` // logs_config.file_scan_period
	ValidatePodContainerId      bool              `json:"validate_pod_container_id" default:"false"`
	ExpectedTagsDuration        api.Duration      `json:"expected_tags_duration" flag:"logs-expected-tags-duration" description:"expectedTagsDuration(Second)"`                     // logs_config.expected_tags_duration
	Socks5ProxyAddress          string            `json:"socks5_proxy_address"`                                                                                                     // logs_config.socks5_proxy_address
	UseTCP                      bool              `json:"use_tcp"`                                                                                                                  // logs_config.use_tcp
	UseHTTP                     bool              `json:"use_http"`                                                                                                                 // logs_config.use_http
	DevModeUseProto             bool              `json:"dev_mode_use_proto" default:"true"`                                                                                        // logs_config.dev_mode_use_proto
	ConnectionResetInterval     api.Duration      `json:"connection_reset_interval" flag:"logs-connection-reset-interval" default:"" description:"connectionResetInterval(Second)"` // logs_config.connection_reset_interval
	LogsUrl                     string            `json:"logs_url"`                                                                                                                 // logs_config.logs_dd_url, dd_url
	UsePort443                  bool              `json:"use_port443"`                                                                                                              // logs_config.use_port_443
	UseSSL                      bool              `json:"use_ssl"`                                                                                                                  // !logs_config.logs_no_ssl
	Url443                      string            `json:"url_443"`                                                                                                                  // logs_config.dd_url_443
	UseCompression              bool              `json:"use_compression" default:"true"`                                                                                           // logs_config.use_compression
	CompressionLevel            int               `json:"compression_level" default:"6"`                                                                                            // logs_config.compression_level
	URL                         string            `json:"url" default:"localhost:8080"`                                                                                             // logs_config.dd_url (e.g. localhost:8080)
	BatchWait                   api.Duration      `json:"batch_wait" flag:"logs-batch-wait" default:"5s" description:"batchWait"`                                                   // logs_config.batch_wait
	BatchMaxConcurrentSend      int               `json:"batch_max_concurrent_send"`                                                                                                // logs_config.batch_max_concurrent_send
	BatchMaxSize                int               `json:"batch_max_size" default:"100"`
	BatchMaxContentSize         int               `json:"batch_max_content_size" default:"1000000"`
	SenderBackoffFactor         float64           `json:"sender_backoff_factor" default:"2.0"`
	SenderBackoffBase           float64           `json:"sender_backoff_base" default:"1.0"`
	SenderBackoffMax            float64           `json:"sender_backoff_max" default:"120.0"`
	SenderRecoveryInterval      int               `json:"sender_recovery_interval" default:"2"`
	SenderRecoveryReset         bool              `json:"sender_recovery_reset"`
	TaggerWarmupDuration        api.Duration      `json:"tagger_warmup_duration" flag:"logs-tagger-warmup-duration" description:"taggerWarmupDuration"`                          // logs_config.tagger_warmup_duration
	AggregationTimeout          api.Duration      `json:"aggregation_timeout" flag:"logs-aggregation-timeout" default:"1000ms" description:"aggregationTimeout"`                 // logs_config.aggregation_timeout
	CloseTimeout                api.Duration      `json:"close_timeout" flag:"logs-close-timeout" default:"60s" description:"closeTimeout"`                                      // logs_config.close_timeout
	AuditorTTL                  api.Duration      `json:"auditor_ttl" flag:"logs-auditor-ttl" description:"auditorTTL"`                                                          // logs_config.auditor_ttl
	RunPath                     string            `json:"run_path" description:"default {root}/{run_path}"`                                                                      // logs_config.run_path
	OpenFilesLimit              int               `json:"open_files_limit" flag:"logs-open-files-limit" default:"100"`                                                           // logs_config.open_files_limit
	K8SContainerUseFile         bool              `json:"k8s_container_use_file"`                                                                                                // logs_config.k8s_container_use_file
	DockerContainerUseFile      bool              `json:"docker_container_use_file"`                                                                                             // logs_config.docker_container_use_file
	DockerContainerForceUseFile bool              `json:"docker_container_force_use_file"`                                                                                       // logs_config.docker_container_force_use_file
	DockerClientReadTimeout     api.Duration      `json:"docker_client_read_timeout" flag:"logs-docker-client-read-timeout" default:"30s" description:"dockerClientReadTimeout"` // logs_config.docker_client_read_timeout
	FrameSize                   int               `json:"frame_size" default:"9000"`                                                                                             // logs_config.frame_size
	StopGracePeriod             api.Duration      `json:"stop_grace_period" flag:"logs-stop-grace-period" default:"30s" description:"stopGracePeriod"`                           // logs_config.stop_grace_period
	DDPort                      int               `json:"dd_port"`                                                                                                               // logs_config.dd_port
	UseV2Api                    bool              `json:"use_v2_api"`                                                                                                            // logs_config.use_v2_api
}

func (p *Config) Validate() error {
	if p.APIKey != "" {
		p.APIKey = strings.TrimSpace(p.APIKey)
	}

	if p.BatchWait.Duration < time.Second || 10*time.Second < p.BatchWait.Duration {
		klog.V(6).Infof("Invalid batchWait: %v should be in [1, 10], fallback on %v",
			p.BatchWait.Duration, DefaultBatchWait)
		p.BatchWait.Duration = DefaultBatchWait
	}

	if p.BatchMaxConcurrentSend < 0 {
		klog.Warningf("Invalid batchMaxconcurrentSend: %v should be >= 0, fallback on %v",
			p.BatchMaxConcurrentSend, DefaultBatchMaxConcurrentSend)
		p.BatchMaxConcurrentSend = DefaultBatchMaxConcurrentSend
	}

	return nil
}
