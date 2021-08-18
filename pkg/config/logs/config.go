package logs

import (
	"strings"
	"time"

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
	Enabled                     bool              `json:"enabled"`                                            // logs_enabled
	AdditionalEndpoints         []Endpoint        `json:"additional_endpoints"`                               // logs_config.additional_endpoints
	ContainerCollectAll         bool              `json:"container_collect_all"`                              // logs_config.container_collect_all
	ProcessingRules             []*ProcessingRule `json:"processing_rules"`                                   // logs_config.processing_rules
	APIKey                      string            `json:"api_key"`                                            // logs_config.api_key
	DevModeNoSSL                bool              `json:"dev_mode_no_ssl"`                                    // logs_config.dev_mode_no_ssl
	FileScanPeriod              int               `json:"file_scan_period" default:"10" description:"second"` // logs_config.file_scan_period
	ValidatePodContainerId      bool              `json:"validate_pod_container_id" default:"false"`
	ExpectedTagsDuration        time.Duration     `json:"-"`
	ExpectedTagsDuration_       int               `json:"expected_tags_duration" flag:"logs-expected-tags-duration" description:"expectedTagsDuration(Second)"` // logs_config.expected_tags_duration
	Socks5ProxyAddress          string            `json:"socks5_proxy_address"`                                                                                 // logs_config.socks5_proxy_address
	UseTCP                      bool              `json:"use_tcp"`                                                                                              // logs_config.use_tcp
	UseHTTP                     bool              `json:"use_http"`                                                                                             // logs_config.use_http
	DevModeUseProto             bool              `json:"dev_mode_use_proto" default:"true"`                                                                    // logs_config.dev_mode_use_proto
	ConnectionResetInterval     time.Duration     `json:"-"`
	ConnectionResetInterval_    int               `json:"connection_reset_interval" flag:"logs-connection-reset-interval" default:"" description:"connectionResetInterval(Second)"` // logs_config.connection_reset_interval
	LogsUrl                     string            `json:"logs_url"`                                                                                                                 // logs_config.logs_dd_url, dd_url
	UsePort443                  bool              `json:"use_port443"`                                                                                                              // logs_config.use_port_443
	UseSSL                      bool              `json:"use_ssl"`                                                                                                                  // !logs_config.logs_no_ssl
	Url443                      string            `json:"url_443"`                                                                                                                  // logs_config.dd_url_443
	UseCompression              bool              `json:"use_compression" default:"true"`                                                                                           // logs_config.use_compression
	CompressionLevel            int               `json:"compression_level" default:"6"`                                                                                            // logs_config.compression_level
	URL                         string            `json:"url" default:"localhost:8080"`                                                                                             // logs_config.dd_url (e.g. localhost:8080)
	BatchWait                   time.Duration     `json:"-"`
	BatchWait_                  int               `json:"batch_wait" flag:"logs-batch-wait" default:"5" description:"batchWait(Second)"` // logs_config.batch_wait
	BatchMaxConcurrentSend      int               `json:"batch_max_concurrent_send"`                                                     // logs_config.batch_max_concurrent_send
	BatchMaxSize                int               `json:"batch_max_size" default:"100"`
	BatchMaxContentSize         int               `json:"batch_max_content_size" default:"1000000"`
	SenderBackoffFactor         float64           `json:"sender_backoff_factor" default:"2.0"`
	SenderBackoffBase           float64           `json:"sender_backoff_base" default:"1.0"`
	SenderBackoffMax            float64           `json:"sender_backoff_max" default:"120.0"`
	SenderRecoveryInterval      int               `json:"sender_recovery_interval" default:"2"`
	SenderRecoveryReset         bool              `json:"sender_recovery_reset"`
	TaggerWarmupDuration        time.Duration     `json:"-"`
	TaggerWarmupDuration_       int               `json:"tagger_warmup_duration" flag:"logs-tagger-warmup-duration" description:"taggerWarmupDuration(Second)"` // logs_config.tagger_warmup_duration
	AggregationTimeout          time.Duration     `json:"-"`
	AggregationTimeout_         int               `json:"aggregation_timeout" flag:"logs-aggregation-timeout" default:"1000" description:"aggregationTimeout(Millisecond)"` // logs_config.aggregation_timeout
	CloseTimeout                time.Duration     `json:"-"`
	CloseTimeout_               int               `json:"close_timeout" flag:"logs-close-timeout" default:"60" description:"closeTimeout(Second)"` // logs_config.close_timeout
	AuditorTTL                  time.Duration     `json:"-"`
	AuditorTTL_                 int               `json:"auditor_ttl" flag:"logs-auditor-ttl" description:"auditorTTL(Second)"` // logs_config.auditor_ttl
	RunPath                     string            `json:"run_path"`                                                             // logs_config.run_path
	OpenFilesLimit              int               `json:"open_files_limit" flag:"logs-open-files-limit" default:"100"`          // logs_config.open_files_limit
	K8SContainerUseFile         bool              `json:"k8s_container_use_file"`                                               // logs_config.k8s_container_use_file
	DockerContainerUseFile      bool              `json:"docker_container_use_file"`                                            // logs_config.docker_container_use_file
	DockerContainerForceUseFile bool              `json:"docker_container_force_use_file"`                                      // logs_config.docker_container_force_use_file
	DockerClientReadTimeout     time.Duration     `json:"-"`
	DockerClientReadTimeout_    int               `json:"docker_client_read_timeout" flag:"logs-docker-client-read-timeout" default:"30" description:"dockerClientReadTimeout(Second)"` // logs_config.docker_client_read_timeout
	FrameSize                   int               `json:"frame_size" default:"9000"`                                                                                                    // logs_config.frame_size
	StopGracePeriod             time.Duration     `json:"-"`
	StopGracePeriod_            int               `json:"stop_grace_period" flag:"logs-stop-grace-period" default:"30" description:"stopGracePeriod(Second)"` // logs_config.stop_grace_period
	DDPort                      int               `json:"dd_port"`                                                                                            // logs_config.dd_port
	UseV2Api                    bool              `json:"use_v2_api"`                                                                                         // logs_config.use_v2_api
}

func (p *Config) Validate() error {
	p.ExpectedTagsDuration = time.Second * time.Duration(p.ExpectedTagsDuration_)
	p.ConnectionResetInterval = time.Second * time.Duration(p.ConnectionResetInterval_)
	p.BatchWait = time.Second * time.Duration(p.BatchWait_)
	p.TaggerWarmupDuration = time.Second * time.Duration(p.TaggerWarmupDuration_)
	p.AggregationTimeout = time.Millisecond * time.Duration(p.AggregationTimeout_)
	p.CloseTimeout = time.Second * time.Duration(p.CloseTimeout_)
	p.AuditorTTL = time.Second * time.Duration(p.AuditorTTL_)
	p.DockerClientReadTimeout = time.Second * time.Duration(p.DockerClientReadTimeout_)
	p.StopGracePeriod = time.Second * time.Duration(p.StopGracePeriod_)

	if p.APIKey != "" {
		p.APIKey = strings.TrimSpace(p.APIKey)
	}

	if p.BatchWait < time.Second || 10*time.Second < p.BatchWait {
		klog.V(6).Infof("Invalid batchWait: %v should be in [1, 10], fallback on %v",
			p.BatchWait_, DefaultBatchWait.Seconds())
		p.BatchWait = DefaultBatchWait
	}

	if p.BatchMaxConcurrentSend < 0 {
		klog.Warningf("Invalid batchMaxconcurrentSend: %v should be >= 0, fallback on %v",
			p.BatchMaxConcurrentSend, DefaultBatchMaxConcurrentSend)
		p.BatchMaxConcurrentSend = DefaultBatchMaxConcurrentSend
	}

	return nil
}
