package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery/common/types"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/n9e/n9e-agentd/pkg/version"
	logstypes "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/types"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/snmp/traps"
	snmptypes "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/snmp/types"
	"k8s.io/klog/v2"
)

const (
	authTokenName = "auth_token"

	// DefaultSite is the default site the Agent sends data to.
	//DefaultSite    = "datadoghq.com"
	//infraURLPrefix = "https://app."

	// DefaultNumWorkers default number of workers for our check runner
	DefaultNumWorkers = 4
	// MaxNumWorkers maximum number of workers for our check runner
	MaxNumWorkers = 25

	megaByte = 1024 * 1024

	DefaultBatchWait = 5 * time.Second

	// DefaultBatchMaxConcurrentSend is the default HTTP batch max concurrent send for logs
	DefaultBatchMaxConcurrentSend = 0

	// DefaultAuditorTTL is the default logs auditor TTL in hours
	DefaultAuditorTTL = 23

	// ClusterIDCacheKey is the key name for the orchestrator cluster id in the agent in-mem cache
	ClusterIDCacheKey = "orchestratorClusterID"

	// DefaultRuntimePoliciesDir is the default policies directory used by the runtime security module
	DefaultRuntimePoliciesDir = "/opt/n9e-agentd/runtime-security.d"
)

var (
	// ungly hack, TODO: remove it
	C = NewDefaultConfig()
	// StartTime is the agent startup time
	StartTime = time.Now()

	// Variables to initialize at build time
	DefaultPython      string
	ForceDefaultPython string
)

type Config struct {
	Ident                                string                   `yaml:"ident"`
	Alias                                string                   `yaml:"alias"`
	Lang                                 string                   `yaml:"lang"`
	EnableDocs                           bool                     `yaml:"enableDocs"`
	N9eSeriesFormat                      bool                     `yaml:"n9eSeriesFormat"`                      // the payload format for forwarder
	WorkDir                              string                   `yaml:"-"`                                    // e.g. /etc/n9e-agentd/
	ConfigFilePath                       string                   `yaml:"-"`                                    // e.g. /etc/n9e-agentd/conf.d
	Endpoints                            []string                 `yaml:"endpoints"`                            // site, dd_url
	VerboseReport                        bool                     `yaml:"verboseReport"`                        // collects run in verbose mode, e.g. report both with cpu.used(sys+user), cpu.system & cpu.user
	Listeners                            []Listeners              `yaml:"listeners"`                            // listeners
	AuthTokenFilePath                    string                   `yaml:"authTokenFilePath"`                    // auth_token_file_path
	ApiKey                               string                   `yaml:"apiKey"`                               // api_key
	Hostname                             string                   `yaml:"hostname"`                             //
	HostnameFQDN                         bool                     `yaml:"hostnameFQDN"`                         // hostname_fqdn
	HostnameForceConfigAsCanonical       bool                     `yaml:"hostnameForceConfigAsCanonical"`       // hostname_force_config_as_canonical
	BindHost                             string                   `yaml:"bindHost"`                             // bind_host
	IPCAddress                           string                   `yaml:"ipcAddress"`                           // ipc_address
	CmdPort                              int                      `yaml:"cmdPort"`                              // cmd_port
	PidfilePath                          string                   `yaml:"pidfilePath"`                          //
	MaxProcs                             string                   `yaml:"maxProcs"`                             //
	CoreDump                             bool                     `yaml:"coreDump"`                             // go_core_dump
	ExporterPort                         int                      `yaml:"exporterPort"`                         //
	HealthPort                           int                      `yaml:"healthPort"`                           //
	Pprof                                bool                     `yaml:"pprof"`                                //
	Expvar                               bool                     `yaml:"expvar"`                               //
	ExpvarPort                           int                      `yaml:"expvarPort"`                           // expvar_port
	Metrics                              bool                     `yaml:"metrics"`                              //
	SkipSSLValidation                    bool                     `yaml:"skipSSLValidation"`                    // skip_ssl_validation
	ForceTLS12                           bool                     `yaml:"forceTLS12"`                           // force_tls_12
	ECSMetadataTimeout                   time.Duration            `yaml:"ecsMetadataTimeout"`                   // ecs_metadata_timeout
	MetadataEndpointsMaxHostnameSize     int                      `yaml:"metadataEndpointsMaxHostnameSize"`     // metadata_endpoints_max_hostname_size
	CloudProviderMetadata                []string                 `yaml:"cloudProviderMetadata"`                //cloud_provider_metadata
	GCEMetadataTimeout                   time.Duration            `yaml:"gceMetadataTimeout"`                   // gce_metadata_timeout
	ClusterName                          string                   `yaml:"clusterName"`                          // cluster_name
	RunPath                              string                   `yaml:"runPath"`                              // run_path
	CLCRunnerEnabled                     bool                     `yaml:"clcRunnerEnabled"`                     //
	CLCRunnerHost                        string                   `yaml:"clcRunnerHost"`                        // clc_runner_host
	ConfigProviders                      []ConfigurationProviders `yaml:"configProviders"`                      // config_providers
	ExtraConfigProviders                 []string                 `yaml:"extraConfigProviders"`                 // extra_config_providers
	CloudFoundry                         bool                     `yaml:"cloudFoundry"`                         // cloud_foundry
	BoshID                               string                   `yaml:"boshID"`                               // bosh_id
	CfOSHostnameAliasing                 bool                     `yaml:"cfOSHostnameAliasing"`                 // cf_os_hostname_aliasing
	CollectGCETags                       bool                     `yaml:"collectGCETags"`                       // collect_gce_tags
	CollectEC2Tags                       bool                     `yaml:"collectEC2Tags"`                       // collect_ec2_tags
	DisableClusterNameTagKey             bool                     `yaml:"DisableClusterNameTagKey"`             // disable_cluster_name_tag_key
	Env                                  string                   `yaml:"env"`                                  // env
	Tags                                 []string                 `yaml:"tags"`                                 // tags
	TagValueSplitSeparator               map[string]string        `yaml:"tagValueSplitSeparator"`               // tag_value_split_separator
	NoProxyNonexactMatch                 bool                     `yaml:"noProxyNonexactMatch"`                 // no_proxy_nonexact_match
	EnableGohai                          bool                     `yaml:"enableGohai"`                          // enable_gohai
	ChecksTagCardinality                 string                   `yaml:"checksTagCardinality"`                 // checks_tag_cardinality
	HistogramAggregates                  []string                 `yaml:"histogramAggregates"`                  // histogram_aggregates
	HistogramPercentiles                 []string                 `yaml:"histogramPercentiles"`                 // histogram_percentiles
	AcLoadTimeout                        time.Duration            `yaml:"acLoadTimeout"`                        // ac_load_timeout
	AdConfigPollInterval                 time.Duration            `yaml:"adConfigPollInterval"`                 // ad_config_poll_interval
	AggregatorBufferSize                 int                      `yaml:"aggregatorBufferSize"`                 // aggregator_buffer_size
	IotHost                              bool                     `yaml:"iotHost"`                              // iot_host
	HerokuDyno                           bool                     `yaml:"herokuDyno"`                           // heroku_dyno
	BasicTelemetryAddContainerTags       bool                     `yaml:"basicTelemetryAddContainerTags"`       // basic_telemetry_add_container_tags
	LogPayloads                          bool                     `yaml:"logPayloads"`                          // log_payloads
	AggregatorStopTimeout                time.Duration            `yaml:"aggregatorStopTimeout"`                // aggregator_stop_timeout
	AutoconfTemplateDir                  string                   `yaml:"autoconfTemplateDir"`                  // autoconf_template_dir
	AutoconfTemplateUrlTimeout           bool                     `yaml:"autoconfTemplateUrlTimeout"`           // autoconf_template_url_timeout
	CheckRunners                         int                      `yaml:"checkRunners"`                         // check_runners
	LoggingFrequency                     int64                    `yaml:"loggingFrequency"`                     // logging_frequency
	GUIPort                              bool                     `yaml:"guiPort"`                              // GUI_port
	XAwsEc2MetadataTokenTtlSeconds       bool                     `yaml:"xAwsEc2MetadataTokenTtlSeconds"`       // X-aws-ec2-metadata-token-ttl-seconds
	AcExclude                            bool                     `yaml:"acExclude"`                            // ac_exclude
	AcInclude                            bool                     `yaml:"acInclude"`                            // ac_include
	AdditionalChecksd                    string                   `yaml:"additionalChecksd"`                    // additional_checksd
	AllowArbitraryTags                   bool                     `yaml:"allowArbitraryTags"`                   // allow_arbitrary_tags
	AppKey                               bool                     `yaml:"appKey"`                               // app_key
	CCoreDump                            bool                     `yaml:"cCoreDump"`                            // c_core_dump
	CStacktraceCollection                bool                     `yaml:"cStacktraceCollection"`                // c_stacktrace_collection
	CacheSyncTimeout                     time.Duration            `yaml:"cacheSyncTimeout"`                     // cache_sync_timeout
	ClcRunnerId                          string                   `yaml:"clcRunnerId"`                          // clc_runner_id
	CmdHost                              string                   `yaml:"cmdHost"`                              // cmd_host
	CollectKubernetesEvents              bool                     `yaml:"collectKubernetesEvents"`              // collect_kubernetes_events
	ComplianceConfigDir                  string                   `yaml:"complianceConfigDir"`                  // compliance_config.dir
	ComplianceConfigEnabled              bool                     `yaml:"complianceConfigEnabled"`              // compliance_config.enabled
	ConfdPath                            string                   `yaml:"confdPath"`                            // confd_path
	ContainerCgroupPrefix                string                   `yaml:"containerCgroupPrefix"`                // container_cgroup_prefix
	ContainerCgroupRoot                  string                   `yaml:"containerCgroupRoot"`                  // container_cgroup_root
	ContainerProcRoot                    string                   `yaml:"containerProcRoot"`                    // container_proc_root
	ContainerdNamespace                  string                   `yaml:"containerdNamespace"`                  // containerd_namespace
	CriConnectionTimeout                 time.Duration            `yaml:"criConnectionTimeout"`                 // cri_connection_timeout
	CriQueryTimeout                      time.Duration            `yaml:"criQueryTimeout"`                      // cri_query_timeout
	CriSocketPath                        bool                     `yaml:"criSocketPath"`                        // cri_socket_path
	DatadogCluster                       bool                     `yaml:"datadogCluster"`                       // datadog-cluster
	DisableFileLogging                   bool                     `yaml:"disableFileLogging"`                   // disable_file_logging
	DockerLabelsAsTags                   bool                     `yaml:"dockerLabelsAsTags"`                   // docker_labels_as_tags
	DockerQueryTimeout                   time.Duration            `yaml:"dockerQueryTimeout"`                   // docker_query_timeout
	EC2PreferImdsv2                      bool                     `yaml:"ec2PreferImdsv2"`                      // ec2_prefer_imdsv2
	EC2MetadataTimeout                   time.Duration            `yaml:"ec2MetadataTimeout"`                   // ec2_metadata_timeout
	EC2MetadataTokenLifetime             time.Duration            `yaml:"ec2MetadataTokenLifetime"`             // ec2_metadata_token_lifetime
	EC2UseWindowsPrefixDetection         bool                     `yaml:"ec2UseWindowsPrefixDetection"`         // ec2_use_windows_prefix_detection
	EcsAgentContainerName                string                   `yaml:"ecsAgentContainerName"`                // ecs_agent_container_name
	EcsAgentUrl                          bool                     `yaml:"ecsAgentUrl"`                          // ecs_agent_url
	EcsCollectResourceTagsEc2            bool                     `yaml:"ecsCollectResourceTagsEc2"`            // ecs_collect_resource_tags_ec2
	EKSFargate                           bool                     `yaml:"eksFargate"`                           // eks_fargate
	EnableMetadataCollection             bool                     `yaml:"enableMetadataCollection"`             // enable_metadata_collection
	ExcludeGCETags                       []string                 `yaml:"excludeGCETags"`                       // exclude_gce_tags
	ExcludePauseContainer                bool                     `yaml:"excludePauseContainer"`                // exclude_pause_container
	ExternalMetricsAggregator            string                   `yaml:"externalMetricsAggregator"`            // external_metrics.aggregator
	ExtraListeners                       []string                 `yaml:"extraListeners"`                       // extra_listeners
	ForceTls12                           bool                     `yaml:"forceTls12"`                           // force_tls_12
	FullSketches                         bool                     `yaml:"fullSketches"`                         // full-sketches
	GceSendProjectIdTag                  bool                     `yaml:"gceSendProjectIdTag"`                  // gce_send_project_id_tag
	GoCoreDump                           bool                     `yaml:"goCoreDump"`                           // go_core_dump
	HpaConfigmapName                     string                   `yaml:"hpaConfigmapName"`                     // hpa_configmap_name
	HpaWatcherGcPeriod                   time.Duration            `yaml:"hpaWatcherGcPeriod"`                   // hpa_watcher_gc_period
	IgnoreAutoconf                       []string                 `yaml:"ignoreAutoconf"`                       // ignore_autoconf
	InventoriesEnabled                   bool                     `yaml:"inventoriesEnabled"`                   // inventories_enabled
	InventoriesMaxInterval               time.Duration            `yaml:"inventoriesMaxInterval"`               // inventories_max_interval
	InventoriesMinInterval               time.Duration            `yaml:"inventoriesMinInterval"`               // inventories_min_interval
	KubeResourcesNamespace               bool                     `yaml:"kubeResourcesNamespace"`               // kube_resources_namespace
	KubeletAuthTokenPath                 string                   `yaml:"kubeletAuthTokenPath"`                 // kubelet_auth_token_path
	KubeletCachePodsDuration             time.Duration            `yaml:"kubeletCachePodsDuration"`             // kubelet_cache_pods_duration
	KubeletClientCa                      string                   `yaml:"kubeletClientCa"`                      // kubelet_client_ca
	KubeletClientCrt                     string                   `yaml:"kubeletClientCrt"`                     // kubelet_client_crt
	KubeletClientKey                     string                   `yaml:"kubeletClientKey"`                     // kubelet_client_key
	KubeletListenerPollingInterval       time.Duration            `yaml:"kubeletListenerPollingInterval"`       // kubelet_listener_polling_interval
	KubeletTlsVerify                     bool                     `yaml:"kubeletTlsVerify"`                     // kubelet_tls_verify
	KubeletWaitOnMissingContainer        time.Duration            `yaml:"kubeletWaitOnMissingContainer"`        // kubelet_wait_on_missing_container
	KubernetesApiserverClientTimeout     time.Duration            `yaml:"kubernetesApiserverClientTimeout"`     // kubernetes_apiserver_client_timeout
	KubernetesApiserverUseProtobuf       bool                     `yaml:"kubernetesApiserverUseProtobuf"`       // kubernetes_apiserver_use_protobuf
	KubernetesCollectMetadataTags        bool                     `yaml:"kubernetesCollectMetadataTags"`        // kubernetes_collect_metadata_tags
	KubernetesCollectServiceTags         bool                     `yaml:"kubernetesCollectServiceTags"`         // kubernetes_collect_service_tags
	KubernetesHttpKubeletPort            int                      `yaml:"kubernetesHttpKubeletPort"`            // kubernetes_http_kubelet_port
	KubernetesHttpsKubeletPort           int                      `yaml:"kubernetesHttpsKubeletPort"`           // kubernetes_https_kubelet_port
	KubernetesInformersResyncPeriod      time.Duration            `yaml:"kubernetesInformersResyncPeriod"`      // kubernetes_informers_resync_period
	KubernetesKubeconfigPath             bool                     `yaml:"kubernetesKubeconfigPath"`             // kubernetes_kubeconfig_path
	KubernetesKubeletHost                string                   `yaml:"kubernetesKubeletHost"`                // kubernetes_kubelet_host
	KubernetesKubeletNodename            string                   `yaml:"kubernetesKubeletNodename"`            // kubernetes_kubelet_nodename
	KubernetesMapServicesOnIp            bool                     `yaml:"kubernetesMapServicesOnIp"`            // kubernetes_map_services_on_ip
	KubernetesMetadataTagUpdateFreq      time.Duration            `yaml:"kubernetesMetadataTagUpdateFreq"`      // kubernetes_metadata_tag_update_freq
	KubernetesNamespaceLabelsAsTags      bool                     `yaml:"kubernetesNamespaceLabelsAsTags"`      // kubernetes_namespace_labels_as_tags
	KubernetesNodeLabelsAsTags           bool                     `yaml:"kubernetesNodeLabelsAsTags"`           // kubernetes_node_labels_as_tags
	KubernetesPodAnnotationsAsTags       map[string]string        `yaml:"kubernetesPodAnnotationsAsTags"`       // kubernetes_pod_annotations_as_tags
	KubernetesPodExpirationDuration      time.Duration            `yaml:"kubernetesPodExpirationDuration"`      // kubernetes_pod_expiration_duration
	KubernetesPodLabelsAsTags            map[string]string        `yaml:"kubernetesPodLabelsAsTags"`            // kubernetes_pod_labels_as_tags
	KubernetesServiceTagUpdateFreq       map[string]string        `yaml:"kubernetesServiceTagUpdateFreq"`       // kubernetes_service_tag_update_freq
	LeaderElection                       bool                     `yaml:"leaderElection"`                       // leader_election
	LeaderLeaseDuration                  time.Duration            `yaml:"leaderLeaseDuration"`                  // leader_lease_duration
	LogEnabled                           bool                     `yaml:"logEnabled"`                           // log_enabled
	LogFile                              string                   `yaml:"logFile"`                              // log_file
	LogFormatJson                        bool                     `yaml:"logFormatJson"`                        // log_format_json
	LogFormatRfc3339                     bool                     `yaml:"logFormatRfc3339"`                     // log_format_rfc3339
	LogLevel                             string                   `yaml:"logLevel"`                             // log_level
	LogToConsole                         bool                     `yaml:"logToConsole"`                         // log_to_console
	MemtrackEnabled                      bool                     `yaml:"memtrackEnabled"`                      // memtrack_enabled
	MetricsPort                          int                      `yaml:"metricsPort"`                          // metrics_port
	ProcRoot                             string                   `yaml:"procRoot"`                             // proc_root
	ProcessAgentConfigHostIps            bool                     `yaml:"processAgentConfigHostIps"`            // process_agent_config.host_ips
	ProcessConfigEnabled                 bool                     `yaml:"processConfigEnabled"`                 // process_config.enabled
	ProcfsPath                           string                   `yaml:"procfsPath"`                           // procfs_path
	ProfilingEnabled                     bool                     `yaml:"profilingEnabled"`                     // profiling.enabled
	ProfilingProfileDdUrl                bool                     `yaml:"profilingProfileDdUrl"`                // profiling.profile_dd_url
	PrometheusScrapeEnabled              bool                     `yaml:"prometheusScrapeEnabled"`              // prometheus_scrape.enabled
	PrometheusScrapeServiceEndpoints     bool                     `yaml:"prometheusScrapeServiceEndpoints"`     // prometheus_scrape.service_endpoints
	ProxyHttps                           bool                     `yaml:"proxyHttps"`                           // proxy.https
	ProxyNoProxy                         bool                     `yaml:"proxyNoProxy"`                         // proxy.no_proxy
	Python3LinterTimeout                 time.Duration            `yaml:"python3LinterTimeout"`                 // python3_linter_timeout
	PythonVersion                        string                   `yaml:"pythonVersion"`                        // python_version
	SerializerMaxPayloadSize             int                      `yaml:"serializerMaxPayloadSize"`             // serializer_max_payload_size
	SerializerMaxUncompressedPayloadSize int                      `yaml:"serializerMaxUncompressedPayloadSize"` // serializer_max_uncompressed_payload_size
	ServerTimeout                        time.Duration            `yaml:"serverTimeout"`                        // server_timeout
	SkipSslValidation                    bool                     `yaml:"skipSslValidation"`                    // skip_ssl_validation
	SyslogRfc                            bool                     `yaml:"syslogRfc"`                            // syslog_rfc
	TelemetryEnabled                     bool                     `yaml:"telemetryEnabled"`                     // telemetry.enabled
	TracemallocDebug                     bool                     `yaml:"tracemallocDebug"`                     // tracemalloc_debug
	WindowsUsePythonpath                 bool                     `yaml:"windowsUsePythonpath"`                 // windows_use_pythonpath
	Yaml                                 bool                     `yaml:"yaml"`                                 // yaml
	MetadataProviders                    []MetadataProviders      `yaml:"metadataProviders"`                    // metadata_providers
	Forwarder                            Forwarder                `yaml:"forwarder"`                            // fowarder_*
	PrometheusScrape                     PrometheusScrape         `yaml:"prometheusScrape"`                     // prometheus_scrape
	Autoconfig                           Autoconfig               `yaml:"autoconfig"`                           //
	Container                            Container                `yaml:"container"`                            //
	SnmpTraps                            traps.Config             `yaml:"snmpTraps"`                            // snmp_traps_config
	ClusterAgent                         ClusterAgent             `yaml:"clusterAgent"`                         // cluster_agent
	Network                              Network                  `yaml:"network"`                              // network
	SnmpListener                         snmptypes.ListenerConfig `yaml:"snmpListener"`                         // snmp_listener
	Cmd                                  Cmd                      `yaml:"cmd"`                                  // cmd
	LogsConfig                           LogsConfig               `yaml:"logsConfig"`                           // logs_config
	CloudFoundryGarden                   CloudFoundryGarden       `yaml:"cloudFoundryGarden"`                   // cloud_foundry_garden
	ClusterChecks                        ClusterChecks            `yaml:"clusterChecks"`                        // cluster_checks
	Telemetry                            Telemetry                `yaml:"telemetry"`                            // telemetry
	OrchestratorExplorer                 OrchestratorExplorer     `yaml:"orchestratorExplorer"`                 // orchestrator_explorer
	Statsd                               Statsd                   `yaml:"statsd"`                               // statsd_*, dagstatsd_*
	Apm                                  Apm                      `yaml:"apm"`                                  // apm_config.*
	Jmx                                  Jmx                      `yaml:"jmx"`                                  // jmx_*
	RuntimeSecurity                      RuntimeSecurity          `yaml:"runtimeSecurity"`                      // runtime_security_config.*
	AdminssionController                 AdminssionController     `yaml:"adminssionController"`                 // admission_controller.*
	ExternalMetricsProvider              ExternalMetricsProvider  `yaml:"externalMetricsProvider"`              // external_metrics_provider.*
	EnablePayloads                       EnablePayloads           `yaml:"enablePayloads"`                       // enable_payloads.*
	SystemProbe                          SystemProbe              `yaml:"systemProbe"`                          // system_probe_config.*
	//N9e                                  N9e                      `yaml:"n9e"`
}

func (p Config) String() string {
	return util.Prettify(p)
}

func (p *Config) Validate() error {

	if len(p.Endpoints) == 0 {
		return fmt.Errorf("unable to get agent.endpoints")
	}

	for i, endpoint := range p.Endpoints {
		if _, err := url.Parse(endpoint); err != nil {
			return fmt.Errorf("could not parse agent.endpoint[%d]: %s %s", i, endpoint, err)
		}
	}

	if err := p.Forwarder.Validate(); err != nil {
		return err
	}

	if err := p.SnmpTraps.Validate(p.GetBindHost()); err != nil {
		return err
	}

	if err := p.Statsd.Validate(); err != nil {
		return err
	}

	p.Ident = configEval(p.Ident)
	p.Alias = configEval(p.Alias)

	if strings.Index(p.Ident, "localhost") >= 0 {
		return fmt.Errorf("agent.ident should not include 'localhost'")
	}

	return nil
}

func configEval(value string) string {
	switch strings.ToLower(value) {
	case "$ip":
		return getOutboundIP()
	case "$host":
		host, _ := os.Hostname()
		return host
	default:
		return value
	}
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func (p *Config) Prepare(configFile string) error {
	p.ConfigFilePath = configFile
	p.WorkDir = filepath.Dir(configFile)
	if p.AuthTokenFilePath == "" {
		p.AuthTokenFilePath = filepath.Join(p.WorkDir, authTokenName)
	}

	if ForceDefaultPython == "true" {
		if p.PythonVersion != DefaultPython {
			klog.Warningf("Python version has been forced to %s", DefaultPython)
		}
		p.PythonVersion = DefaultPython
	}

	return nil
}

//type N9e struct {
//	Enabled  bool   `yaml:"enabled"`
//	Endpoint string `yaml:"endpoint"`
//	V5Format bool   `yaml:"v5Format"`
//}

type SystemProbe struct {
	Enabled                      bool                `yaml:"enabled"`                      // system_probe_config.enabled & system_probe
	SysprobeSocket               string              `yaml:"sysprobeSocket"`               // system_probe_config.sysprobe_socket
	BPFDebug                     bool                `yaml:"bpfDebug"`                     // system_probe_config.bpf_debug
	BPFDir                       string              `yaml:"bpfDir"`                       // system_probe_config.bpf_dir
	ExcludedLinuxVersions        []string            `yaml:"excludedLinuxVersions"`        // system_probe_config.excluded_linux_versions
	EnableTracepoints            bool                `yaml:"enableTracepoints"`            // system_probe_config.enable_tracepoints
	EnableRuntimeCompiler        bool                `yaml:"enableRuntimeCompiler"`        // system_probe_config.enable_runtime_compiler
	RuntimeCompilerOutputDir     string              `yaml:"runtimeCompilerOutputDir"`     // system_probe_config.runtime_compiler_output_dir
	KernelHeaderDirs             []string            `yaml:"kernelHeaderDirs"`             // system_probe_config.kernel_header_dirs
	DisableTcp                   bool                `yaml:"disableTcp"`                   // system_probe_config.disable_tcp
	DisableUdp                   bool                `yaml:"disableUdp"`                   // system_probe_config.disable_udp
	DisableIpv6                  bool                `yaml:"disableIpv6"`                  // system_probe_config.disable_ipv6
	OffsetGuessThreshold         int64               `yaml:"offsetGuessThreshold"`         // system_probe_config.offset_guess_threshold
	SourceExcludes               map[string][]string `yaml:"sourceExcludes"`               // system_probe_config.source_excludes
	DestExcludes                 map[string][]string `yaml:"destExcludes"`                 // system_probe_config.dest_excludes
	MaxTrackedConnections        int                 `yaml:"maxTrackedConnections"`        // system_probe_config.max_tracked_connections
	MaxClosedConnectionsBuffered int                 `yaml:"maxClosedConnectionsBuffered"` // system_probe_config.max_closed_connections_buffered
	ClosedChannelSize            int                 `yaml:"closedChannelSize"`            // system_probe_config.closed_channel_size
	MaxConnectionStateBuffered   int                 `yaml:"maxConnectionStateBuffered"`   // system_probe_config.max_connection_state_buffered
	DisableDnsInspection         bool                `yaml:"disableDnsInspection"`         // system_probe_config.disable_dns_inspection
	CollectDnsStats              bool                `yaml:"collectDnsStats"`              // system_probe_config.collect_dns_stats
	CollectLocalDns              bool                `yaml:"collectLocalDns"`              // system_probe_config.collect_local_dns
	CollectDnsDomains            bool                `yaml:"collectDnsDomains"`            // system_probe_config.collect_dns_domains
	MaxDnsStats                  int                 `yaml:"maxDnsStats"`                  // system_probe_config.max_dns_stats
	DnsTimeout                   time.Duration       `yaml:"dnsTimeout"`                   // system_probe_config.dns_timeout_in_s
	EnableConntrack              bool                `yaml:"enableConntrack"`              // system_probe_config.enable_conntrack
	ConntrackMaxStateSize        int                 `yaml:"conntrackMaxStateSize"`        // system_probe_config.conntrack_max_state_size
	ConntrackRateLimit           int                 `yaml:"conntrackRateLimit"`           // system_probe_config.conntrack_rate_limit
	EnableConntrackAllNamespaces bool                `yaml:"enableConntrackAllNamespaces"` // system_probe_config.enable_conntrack_all_namespaces
	WindowsEnableMonotonicCount  bool                `yaml:"windowsEnableMonotonicCount"`  // system_probe_config.windows.enable_monotonic_count
	WindowsDriverBufferSize      int                 `yaml:"windowsDriverBufferSize"`      // system_probe_config.windows.driver_buffer_size
}

type EnablePayloads struct {
	Events              bool `yaml:"events"`              // enable_payloads.events
	Series              bool `yaml:"series"`              // enable_payloads.series
	ServiceChecks       bool `yaml:"serviceChecks"`       // enable_payloads.service_checks
	Sketches            bool `yaml:"sketches"`            // enable_payloads.sketches
	JsonToV1Intake      bool `yaml:"jsonToV1Intake"`      // enable_payloads.json_to_v1_intake
	Metadata            bool `yaml:"metadata"`            //
	HostMetadata        bool `yaml:"hostMetadata"`        //
	AgentchecksMetadata bool `yaml:"agentchecksMetadata"` //
}

type ExternalMetricsProvider struct {
	ApiKey               bool          `yaml:"apiKey"`               // external_metrics_provider.api_key
	AppKey               bool          `yaml:"appKey"`               // external_metrics_provider.app_key
	BucketSize           int           `yaml:"bucketSize"`           // external_metrics_provider.bucket_size
	Enabled              bool          `yaml:"enabled"`              // external_metrics_provider.enabled
	LocalCopyRefreshRate time.Duration `yaml:"localCopyRefreshRate"` // external_metrics_provider.local_copy_refresh_rate
	MaxAge               int           `yaml:"maxAge"`               // external_metrics_provider.max_age
	RefreshPeriod        int           `yaml:"refreshPeriod"`        // external_metrics_provider.refresh_period
	Rollup               int           `yaml:"rollup"`               // external_metrics_provider.rollup
	UseDatadogmetricCrd  bool          `yaml:"useDatadogmetricCrd"`  // external_metrics_provider.use_datadogmetric_crd
	WpaController        bool          `yaml:"wpaController"`        // external_metrics_provider.wpa_controller

}

type AdminssionController struct {
	Enabled                        bool          `yaml:"enabled"`                        // admission_controller.enabled
	CertificateExpirationThreshold time.Duration `yaml:"certificateExpirationThreshold"` // admission_controller.certificate.expiration_threshold
	CertificateSecretName          string        `yaml:"certificateSecretName"`          // admission_controller.certificate.secret_name
	CertificateValidityBound       time.Duration `yaml:"certificateValidityBound"`       // admission_controller.certificate.validity_bound
	InjectConfigEnabled            bool          `yaml:"injectConfigEnabled"`            // admission_controller.inject_config.enabled
	InjectConfigEndpoint           string        `yaml:"injectConfigEndpoint"`           // admission_controller.inject_config.endpoint
	InjectTagsEnabled              bool          `yaml:"injectTagsEnabled"`              // admission_controller.inject_tags.enabled
	InjectTagsEndpoint             string        `yaml:"injectTagsEndpoint"`             // admission_controller.inject_tags.endpoint
	MutateUnlabelled               bool          `yaml:"mutateUnlabelled"`               // admission_controller.mutate_unlabelled
	PodOwnersCacheValidity         int           `yaml:"podOwnersCacheValidity"`         // admission_controller.pod_owners_cache_validity
	ServiceName                    string        `yaml:"serviceName"`                    // admission_controller.service_name
	TimeoutSeconds                 time.Duration `yaml:"timeoutSeconds"`                 // admission_controller.timeout_seconds
	WebhookName                    string        `yaml:"webhookName"`                    // admission_controller.webhook_name

}

type RuntimeSecurity struct {
	Socket                             string `yaml:"socket"`                             // runtime_security_config.socket
	AgentMonitoringEvents              bool   `yaml:"agentMonitoringEvents"`              // runtime_security_config.agent_monitoring_events
	CookieCacheSize                    bool   `yaml:"cookieCacheSize"`                    // runtime_security_config.cookie_cache_size
	CustomSensitiveWords               bool   `yaml:"customSensitiveWords"`               // runtime_security_config.custom_sensitive_words
	EnableApprovers                    bool   `yaml:"enableApprovers"`                    // runtime_security_config.enable_approvers
	EnableDiscarders                   bool   `yaml:"enableDiscarders"`                   // runtime_security_config.enable_discarders
	EnableKernelFilters                bool   `yaml:"enableKernelFilters"`                // runtime_security_config.enable_kernel_filters
	Enabled                            bool   `yaml:"enabled"`                            // runtime_security_config.enabled
	EventServerBurst                   bool   `yaml:"eventServerBurst"`                   // runtime_security_config.event_server.burst
	EventServerRate                    bool   `yaml:"eventServerRate"`                    // runtime_security_config.event_server.rate
	EventsStatsPollingInterval         bool   `yaml:"eventsStatsPollingInterval"`         // runtime_security_config.events_stats.polling_interval
	FimEnabled                         bool   `yaml:"fimEnabled"`                         // runtime_security_config.fim_enabled
	FlushDiscarderWindow               bool   `yaml:"flushDiscarderWindow"`               // runtime_security_config.flush_discarder_window
	LoadControllerControlPeriod        bool   `yaml:"loadControllerControlPeriod"`        // runtime_security_config.load_controller.control_period
	LoadControllerDiscarderTimeout     bool   `yaml:"loadControllerDiscarderTimeout"`     // runtime_security_config.load_controller.discarder_timeout
	LoadControllerEventsCountThreshold bool   `yaml:"loadControllerEventsCountThreshold"` // runtime_security_config.load_controller.events_count_threshold
	PidCacheSize                       bool   `yaml:"pidCacheSize"`                       // runtime_security_config.pid_cache_size
	PoliciesDir                        bool   `yaml:"policiesDir"`                        // runtime_security_config.policies.dir
	SyscallMonitorEnabled              bool   `yaml:"syscallMonitorEnabled"`              // runtime_security_config.syscall_monitor.enabled

}

type Jmx struct {
	CheckPeriod                int           `yaml:"checkPeriod"`                // jmx_check_period
	CollectionTimeout          time.Duration `yaml:"collectionTimeout"`          // jmx_collection_timeout
	CustomJars                 []string      `yaml:"customJars"`                 // jmx_custom_jars
	LogFile                    bool          `yaml:"logFile"`                    // jmx_log_file
	MaxRestarts                int           `yaml:"maxRestarts"`                // jmx_max_restarts
	ReconnectionThreadPoolSize int           `yaml:"reconnectionThreadPoolSize"` // jmx_reconnection_thread_pool_size
	ReconnectionTimeout        time.Duration `yaml:"reconnectionTimeout"`        // jmx_reconnection_timeout
	RestartInterval            time.Duration `yaml:"restartInterval"`            // jmx_restart_interval
	ThreadPoolSize             int           `yaml:"threadPoolSize"`             // jmx_thread_pool_size
	UseCgroupMemoryLimit       bool          `yaml:"useCgroupMemoryLimit"`       // jmx_use_cgroup_memory_limit
	UseContainerSupport        bool          `yaml:"useContainerSupport"`        // jmx_use_container_support

}

type Apm struct {
	AdditionalEndpoints           bool `yaml:"additionalEndpoints"`           // apm_config.additional_endpoints
	AnalyzedRateByService         bool `yaml:"analyzedRateByService"`         // apm_config.analyzed_rate_by_service
	AnalyzedSpans                 bool `yaml:"analyzedSpans"`                 // apm_config.analyzed_spans
	ApiKey                        bool `yaml:"apiKey"`                        // apm_config.api_key
	ApmDdUrl                      bool `yaml:"apmDdUrl"`                      // apm_config.apm_dd_url
	ApmNonLocalTraffic            bool `yaml:"apmNonLocalTraffic"`            // apm_config.apm_non_local_traffic
	ConnectionLimit               bool `yaml:"connectionLimit"`               // apm_config.connection_limit
	ConnectionResetInterval       bool `yaml:"connectionResetInterval"`       // apm_config.connection_reset_interval
	DdAgentBin                    bool `yaml:"ddAgentBin"`                    // apm_config.dd_agent_bin
	Enabled                       bool `yaml:"enabled"`                       // apm_config.enabled
	Env                           bool `yaml:"env"`                           // apm_config.env
	ExtraSampleRate               bool `yaml:"extraSampleRate"`               // apm_config.extra_sample_rate
	FilterTagsReject              bool `yaml:"filterTagsReject"`              // apm_config.filter_tags.reject
	FilterTagsRequire             bool `yaml:"filterTagsRequire"`             // apm_config.filter_tags.require
	IgnoreResources               bool `yaml:"ignoreResources"`               // apm_config.ignore_resources
	LogFile                       bool `yaml:"logFile"`                       // apm_config.log_file
	LogLevel                      bool `yaml:"logLevel"`                      // apm_config.log_level
	LogThrottling                 bool `yaml:"logThrottling"`                 // apm_config.log_throttling
	MaxCpuPercent                 bool `yaml:"maxCpuPercent"`                 // apm_config.max_cpu_percent
	MaxEventsPerSecond            bool `yaml:"maxEventsPerSecond"`            // apm_config.max_events_per_second
	MaxMemory                     bool `yaml:"maxMemory"`                     // apm_config.max_memory
	MaxTracesPerSecond            bool `yaml:"maxTracesPerSecond"`            // apm_config.max_traces_per_second
	Obfuscation                   bool `yaml:"obfuscation"`                   // apm_config.obfuscation
	ProfilingAdditionalEndpoints  bool `yaml:"profilingAdditionalEndpoints"`  // apm_config.profiling_additional_endpoints
	ProfilingDdUrl                bool `yaml:"profilingDdUrl"`                // apm_config.profiling_dd_url
	ReceiverPort                  bool `yaml:"receiverPort"`                  // apm_config.receiver_port
	ReceiverSocket                bool `yaml:"receiverSocket"`                // apm_config.receiver_socket
	ReceiverTimeout               bool `yaml:"receiverTimeout"`               // apm_config.receiver_timeout
	RemoteTagger                  bool `yaml:"remoteTagger"`                  // apm_config.remote_tagger
	SyncFlushing                  bool `yaml:"syncFlushing"`                  // apm_config.sync_flushing
	WindowsPipeBufferSize         bool `yaml:"windowsPipeBufferSize"`         // apm_config.windows_pipe_buffer_size
	WindowsPipeName               bool `yaml:"windowsPipeName"`               // apm_config.windows_pipe_name
	WindowsPipeSecurityDescriptor bool `yaml:"windowsPipeSecurityDescriptor"` // apm_config.windows_pipe_security_descriptor

}

type Statsd struct {
	Enabled                           bool             `yaml:"enabled"`                           // use_dogstatsd
	Host                              string           `yaml:"host"`                              //
	Port                              int              `yaml:"port"`                              // dogstatsd_port
	Socket                            string           `yaml:"socket"`                            // dogstatsd_socket
	PipeName                          string           `yaml:"pipeName"`                          // dogstatsd_pipe_name
	ContextExpirySeconds              time.Duration    `yaml:"contextExpirySeconds"`              // dogstatsd_context_expiry_seconds
	ExpirySeconds                     time.Duration    `yaml:"expirySeconds"`                     // dogstatsd_expiry_seconds
	StatsEnable                       bool             `yaml:"statsEnable"`                       // dogstatsd_stats_enable
	StatsBuffer                       int              `yaml:"statsBuffer"`                       // dogstatsd_stats_buffer
	MetricsStatsEnable                bool             `yaml:"metricsStatsEnable"`                // dogstatsd_metrics_stats_enable - for debug
	BufferSize                        int              `yaml:"bufferSize"`                        // dogstatsd_buffer_size
	MetricNamespace                   string           `yaml:"metricNamespace"`                   // statsd_metric_namespace
	MetricNamespaceBlacklist          []string         `yaml:"metricNamespaceBlacklist"`          // statsd_metric_namespace_blacklist
	Tags                              []string         `yaml:"tags"`                              // dogstatsd_tags
	EntityIdPrecedence                bool             `yaml:"entityIdPrecedence"`                // dogstatsd_entity_id_precedence
	EolRequired                       []string         `yaml:"eolRequired"`                       // dogstatsd_eol_required
	DisableVerboseLogs                bool             `yaml:"disableVerboseLogs"`                // dogstatsd_disable_verbose_logs
	ForwardHost                       string           `yaml:"forwardHost"`                       // statsd_forward_host
	ForwardPort                       int              `yaml:"forwardPort"`                       // statsd_forward_port
	QueueSize                         int              `yaml:"queueSize"`                         // dogstatsd_queue_size
	MapperCacheSize                   int              `yaml:"mapperCacheSize"`                   // dogstatsd_mapper_cache_size
	MapperProfiles                    []MappingProfile `yaml:"mapperProfiles"`                    // dogstatsd_mapper_profiles
	StringInternerSize                int              `yaml:"stringInternerSize"`                // dogstatsd_string_interner_size
	SocketRcvbuf                      int              `yaml:"socektRcvbuf"`                      // dogstatsd_so_rcvbuf
	PacketBufferSize                  int              `yaml:"packetBufferSize"`                  // dogstatsd_packet_buffer_size
	PacketBufferFlushTimeout          time.Duration    `yaml:"packetBufferFlushTimeout"`          // dogstatsd_packet_buffer_flush_timeout
	TagCardinality                    string           `yaml:"tagCardinality"`                    // dogstatsd_tag_cardinality
	NonLocalTraffic                   bool             `yaml:"nonLocalTraffic"`                   // dogstatsd_non_local_traffic
	OriginDetection                   bool             `yaml:"OriginDetection"`                   // dogstatsd_origin_detection
	HistogramCopyToDistribution       bool             `yaml:"histogramCopyToDistribution"`       // histogram_copy_to_distribution
	HistogramCopyToDistributionPrefix string           `yaml:"histogramCopyToDistributionPrefix"` // histogram_copy_to_distribution_prefix

}

func (p *Statsd) Validate() error {
	return nil
}

type AdditionalEndpoint struct {
	Endpoints []string `yaml:"endpoints"`
	ApiKeys   []string `yaml:"apiKeys"`
}

type Forwarder struct {
	//Endpoints                []string             `yaml:"endpoints"`
	AdditionalEndpoints      []AdditionalEndpoint `yaml:"additionalEndpoints"`      // additional_endpoints
	ApikeyValidationInterval time.Duration        `yaml:"apikeyValidationInterval"` // forwarder_apikey_validation_interval
	BackoffBase              float64              `yaml:"backoffBase"`              // forwarder_backoff_base
	BackoffFactor            float64              `yaml:"backoffFactor"`            // forwarder_backoff_factor
	BackoffMax               float64              `yaml:"backoffMax"`               // forwarder_backoff_max
	ConnectionResetInterval  time.Duration        `yaml:"connectionResetInterval"`  // forwarder_connection_reset_interval
	FlushToDiskMemRatio      float64              `yaml:"flushToDiskMemRatio"`      // forwarder_flush_to_disk_mem_ratio
	NumWorkers               int                  `yaml:"numWorkers"`               // forwarder_num_workers
	OutdatedFileInDays       int                  `yaml:"outdatedFileInDays"`       // forwarder_outdated_file_in_days
	RecoveryInterval         int                  `yaml:"recoveryInterval"`         // forwarder_recovery_interval
	RecoveryReset            bool                 `yaml:"recoveryReset"`            // forwarder_recovery_reset
	StopTimeout              time.Duration        `yaml:"stopTimeout"`              // forwarder_stop_timeout
	StorageMaxDiskRatio      float64              `yaml:"storageMaxDiskRatio"`      // forwarder_storage_max_disk_ratio
	StorageMaxSizeInBytes    int64                `yaml:"storageMaxSizeInBytes"`    // forwarder_storage_max_size_in_bytes
	StoragePath              string               `yaml:"storagePath"`              // forwarder_storage_path
	Timeout                  time.Duration        `yaml:"timeout"`                  // forwarder_timeout
	//RetryQueueMaxSize         int           `yaml:"retryQueueMaxSize"`         // forwarder_retry_queue_max_size
	RetryQueuePayloadsMaxSize int `yaml:"retryQueuePayloadsMaxSize"` // forwarder_retry_queue_payloads_max_size

}

func (p *Forwarder) Validate() error {
	for i, addtion := range p.AdditionalEndpoints {
		for j, endpoint := range addtion.Endpoints {
			if _, err := url.Parse(endpoint); err != nil {
				return fmt.Errorf("could not parse agent.forwarder.addtionEndpoints[%d][%d] %s %s", i, j, endpoint, err)
			}
		}
	}

	if p.RecoveryInterval <= 0 {
		return fmt.Errorf("Configured forwarder.recoveryInterval (%v) is not positive", p.RecoveryInterval)
	}
	return nil
}

// MetadataProviders helps unmarshalling `metadata_providers` config param
type MetadataProviders struct {
	Name     string        `yaml:"name"`
	Interval time.Duration `yaml:"interval"`
}

type Cmd struct {
	Check Check `yaml:"check"` // cmd.check
}

type Check struct {
	Fullsketches bool `yaml:"fullsketches"` // cmd.check.fullsketches
}

type CloudFoundryGarden struct {
	ListenNetwork string `yaml:"listenNetwork"` // cloud_foundry_garden.listen_network
	ListenAddress string `yaml:"listenAddress"` // cloud_foundry_garden.listen_address
}

// ProcessingRule defines an exclusion or a masking rule to
// be applied on log lines
type ProcessingRule struct {
	Type               string
	Name               string
	ReplacePlaceholder string `yaml:"replacePlaceholder"`
	Pattern            string
	// TODO: should be moved out
	Regex       *regexp.Regexp
	Placeholder []byte
}

type LogsConfig struct {
	Enabled                     bool                        `yaml:"enabled"`                     // logs_enabled
	AdditionalEndpoints         []logstypes.Endpoint        `yaml:"additionalEndpoints"`         // logs_config.additional_endpoints
	ContainerCollectAll         bool                        `yaml:"containerCollectAll"`         // logs_config.container_collect_all
	ProcessingRules             []*logstypes.ProcessingRule `yaml:"processingRules"`             // logs_config.processing_rules
	APIKey                      string                      `yaml:"apiKey"`                      // logs_config.api_key
	DevModeNoSSL                bool                        `yaml:"devModeNoSSL"`                // logs_config.dev_mode_no_ssl
	ExpectedTagsDuration        time.Duration               `yaml:"expectedTagsDuration"`        // logs_config.expected_tags_duration
	Socks5ProxyAddress          string                      `yaml:"socks5ProxyAddress"`          // logs_config.socks5_proxy_address
	UseTCP                      bool                        `yaml:"useTCP"`                      // logs_config.use_tcp
	UseHTTP                     bool                        `yaml:"useHTTP"`                     // logs_config.use_http
	DevModeUseProto             bool                        `yaml:"devModeUseProto"`             // logs_config.dev_mode_use_proto
	ConnectionResetInterval     time.Duration               `yaml:"connectionResetInterval"`     // logs_config.connection_reset_interval
	LogsUrl                     string                      `yaml:"logsUrl"`                     // logs_config.logs_dd_url, dd_url
	UsePort443                  bool                        `yaml:"usePort443"`                  // logs_config.use_port_443
	UseSSL                      bool                        `yaml:"useSSL"`                      // !logs_config.logs_no_ssl
	Url443                      string                      `yaml:"url443"`                      // logs_config.dd_url_443
	UseCompression              bool                        `yaml:"useCompression"`              // logs_config.use_compression
	CompressionLevel            int                         `yaml:"compressionLevel"`            // logs_config.compression_level
	URL                         string                      `yaml:"url"`                         // logs_config.dd_url (e.g. localhost:8080)
	BatchWait                   time.Duration               `yaml:"batchWait"`                   // logs_config.batch_wait
	BatchMaxConcurrentSend      int                         `yaml:"batchMaxConcurrentSend"`      // logs_config.batch_max_concurrent_send
	TaggerWarmupDuration        time.Duration               `yaml:"taggerWarmupDuration"`        // logs_config.tagger_warmup_duration
	AggregationTimeout          time.Duration               `yaml:"aggregationTimeout"`          // logs_config.aggregation_timeout
	CloseTimeout                time.Duration               `yaml:"closeTimeout"`                // logs_config.close_timeout
	AuditorTTL                  time.Duration               `yaml:"auditorTTL"`                  // logs_config.auditor_ttl
	RunPath                     string                      `yaml:"runPath"`                     // logs_config.run_path
	OpenFilesLimit              int                         `yaml:"openFilesLimit"`              // logs_config.open_files_limit
	K8SContainerUseFile         bool                        `yaml:"k8SContainerUseFile"`         // logs_config.k8s_container_use_file
	DockerContainerUseFile      bool                        `yaml:"dockerContainerUseFile"`      // logs_config.docker_container_use_file
	DockerContainerForceUseFile bool                        `yaml:"dockerContainerForceUseFile"` // logs_config.docker_container_force_use_file
	DockerClientReadTimeout     time.Duration               `yaml:"dockerClientReadTimeout"`     // logs_config.docker_client_read_timeout
	FrameSize                   int                         `yaml:"frameSize"`                   // logs_config.frame_size
	StopGracePeriod             time.Duration               `yaml:"stopGracePeriod"`             // logs_config.stop_grace_period
}

func (p *LogsConfig) Validate() error {
	if p.APIKey != "" {
		p.APIKey = strings.TrimSpace(p.APIKey)
	}

	if p.BatchWait < time.Second || 10*time.Second < p.BatchWait {
		klog.Warningf("Invalid batchWait: %v should be in [1s, 10s], fallback on %v",
			p.BatchWait, DefaultBatchWait.Seconds())
		p.BatchWait = DefaultBatchWait
	}

	if p.BatchMaxConcurrentSend < 0 {
		klog.Warningf("Invalid batchMaxconcurrentSend: %v should be >= 0, fallback on %v",
			p.BatchMaxConcurrentSend, DefaultBatchMaxConcurrentSend)
		p.BatchMaxConcurrentSend = DefaultBatchMaxConcurrentSend
	}

	return nil
}

type Network struct {
	ID                         string `yaml:"id"`                         // network.id
	EnableHttpMonitoring       bool   `yaml:"EnableHttpMonitoring"`       // network_config.enable_http_monitoring
	IgnoreConntrackInitFailure bool   `yaml:"IgnoreConntrackInitFailure"` // network_config.ignore_conntrack_init_failure
	EnableGatewayLookup        bool   `yaml:"EnableGatewayLookup"`        // network_config.enable_gateway_lookup
}

type Telemetry struct {
	Enabled bool     `yaml:"enabled"` // telemetry.enabled
	Checks  []string `yaml:"checks"`  // telemetry.checks
}

type ClusterChecks struct {
	ClcRunnersPort             int           `yaml:"clcRunnersPort"`             // cluster_checks.clc_runners_port
	AdvancedDispatchingEnabled bool          `yaml:"advancedDispatchingEnabled"` // cluster_checks.advanced_dispatching_enabled
	ClusterTagName             string        `yaml:"clusterTagName"`             // cluster_checks.cluster_tag_name
	Enabled                    bool          `yaml:"enabled"`                    // cluster_checks.enabled
	ExtraTags                  []string      `yaml:"extraTags"`                  // cluster_checks.extra_tags
	NodeExpirationTimeout      time.Duration `yaml:"nodeExpirationTimeout"`      // cluster_checks.node_expiration_timeout
	WarmupDuration             time.Duration `yaml:"warmupDuration"`             // cluster_checks.warmup_duration

}

type ClusterAgent struct {
	Url                   string `yaml:"url"`                   // cluster_agent.url
	AuthToken             string `yaml:"authToken"`             // cluster_agent.auth_token
	CmdPort               int    `yaml:"cmdPort"`               // cluster_agent.cmd_port
	Enabled               bool   `yaml:"enabled"`               // cluster_agent.enabled
	KubernetesServiceName string `yaml:"kubernetesServiceName"` // cluster_agent.kubernetes_service_name
	TaggingFallback       string `yaml:"taggingFallback"`       // cluster_agent.tagging_fallback

}

// MappingProfile represent a group of mappings
type MappingProfile struct {
	Name     string          `yaml:"name" json:"name"`
	Prefix   string          `yaml:"prefix" json:"prefix"`
	Mappings []MetricMapping `yaml:"mappings" json:"mappings"`
}

// MetricMapping represent one mapping rule
type MetricMapping struct {
	Match     string            `yaml:"match" json:"match"`
	MatchType string            `yaml:"matchType" json:"matchType"`
	Name      string            `yaml:"name" json:"name"`
	Tags      map[string]string `yaml:"tags" json:"tags"`
}

type OrchestratorExplorer struct { // orchestrator_explorer
	Url                       string   `yaml:"url"`                       // orchestrator_explorer.orchestrator_dd_url
	AdditionalEndpoints       []string `yaml:"additionalEndpoints"`       // orchestrator_explorer.orchestrator_additional_endpoints
	CustomSensitiveWords      []string `yaml:"customSensitiveWords"`      // orchestrator_explorer.custom_sensitive_words
	ContainerScrubbingEnabled bool     `yaml:"containerScrubbingEnabled"` // orchestrator_explorer.container_scrubbing.enabled
	Enabled                   bool     `yaml:"enabled"`                   // orchestrator_explorer.enabled
	ExtraTags                 []string `yaml:"extraTags"`                 // orchestrator_explorer.extra_tags

}

type Autoconfig struct {
	Enabled         bool       `yaml:"enabled"`         // autoconfig_from_environment
	ExcludeFeatures []string   `yaml:"excludeFeatures"` // autoconfig_exclude_features
	features        FeatureMap `yaml:"-"`
}

type PrometheusScrape struct {
	Enabled          bool                     `yaml:"enabled"`          // prometheus_scrape.enabled
	ServiceEndpoints bool                     `yaml:"serviceEndpoints"` // prometheus_scrape.service_endpoints
	Checks           []*types.PrometheusCheck `yaml:"checks"`           // prometheus_scrape.checks
}

// ConfigurationProviders helps unmarshalling `config_providers` config param
type ConfigurationProviders struct {
	Name             string `yaml:"name"`
	Polling          bool   `yaml:"polling"`
	PollInterval     string `yaml:"pollInterval"`
	TemplateURL      string `yaml:"templateUrl"`
	TemplateDir      string `yaml:"templateDir"`
	Username         string `yaml:"username"`
	Password         string `yaml:"password"`
	CAFile           string `yaml:"caFile"`
	CAPath           string `yaml:"caPath"`
	CertFile         string `yaml:"certFile"`
	KeyFile          string `yaml:"keyFile"`
	Token            string `yaml:"token"`
	GraceTimeSeconds int    `yaml:"graceTimeSeconds"`
}

// Listeners helps unmarshalling `listeners` config param
type Listeners struct {
	Name string `mapstructure:"name"`
}

// IsCLCRunner returns whether the Agent is in cluster check runner mode
func (p Config) IsCLCRunner() bool {
	if !p.CLCRunnerEnabled {
		return false
	}

	var cp []ConfigurationProviders
	for _, v := range p.ConfigProviders {
		cp = append(cp, ConfigurationProviders{Name: v.Name})
	}
	for _, name := range p.ExtraConfigProviders {
		cp = append(cp, ConfigurationProviders{Name: name})
	}
	if len(cp) == 1 && cp[0].Name == "clusterchecks" {
		// A cluster check runner is an Agent configured to run clusterchecks only
		return true
	}

	return false
}

// GetIPCAddress returns the IPC address or an error if the address is not local
func (cf *Config) GetIPCAddress() (string, error) {
	address := cf.IPCAddress
	if address == "localhost" {
		return address, nil
	}
	ip := net.ParseIP(address)
	if ip == nil {
		return "", fmt.Errorf("ipcAddress was set to an invalid IP address: %s", address)
	}
	for _, cidr := range []string{
		"127.0.0.0/8", // IPv4 loopback
		"::1/128",     // IPv6 loopback
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			return "", err
		}
		if block.Contains(ip) {
			return address, nil
		}
	}
	return "", fmt.Errorf("ipcAddress was set to a non-loopback IP address: %s", address)
}

// GetBindHost returns `bind_host` variable or default value
// Not using `config.BindEnvAndSetDefault` as some processes need to know
// if value was default one or not (e.g. trace-agent)
func (cf *Config) GetBindHost() string {
	if cf.BindHost != "" {
		return cf.BindHost
	}

	return "localhost"
}

// IsCloudProviderEnabled checks the cloud provider family provided in
// pkg/util/<cloud_provider>.go against the value for cloud_provider: on the
// global config object Datadog
func (cf *Config) IsCloudProviderEnabled(cloudProviderName string) bool {
	cloudProviderFromConfig := cf.CloudProviderMetadata

	for _, cloudName := range cloudProviderFromConfig {
		if strings.ToLower(cloudName) == strings.ToLower(cloudProviderName) {
			klog.V(5).Infof("cloudProviderMetadata is set to %s in agent configuration, trying endpoints for %s Cloud Provider",
				cloudProviderFromConfig,
				cloudProviderName)
			return true
		}
	}

	klog.V(5).Infof("cloudProviderMetadata is set to %s in agent configuration, skipping %s Cloud Provider",
		cloudProviderFromConfig,
		cloudProviderName)
	return false
}

// Warnings represent the warnings in the config
type Warnings struct {
	TraceMallocEnabledWithPy2 bool
}

// GetMultipleEndpoints returns the api keys per domain specified in the main agent config
func GetMultipleEndpoints() (map[string][]string, error) {
	return getMultipleEndpointsWithConfig(C)
}

// getMultipleEndpointsWithConfig implements the logic to extract the api keys per domain from an agent config
func getMultipleEndpointsWithConfig(config *Config) (map[string][]string, error) {
	endpoints := strings.Join(config.Endpoints, ",")

	keysPerDomain := map[string][]string{
		endpoints: {config.ApiKey},
	}

	additionalEndpoints := config.Forwarder.AdditionalEndpoints
	// merge additional endpoints into keysPerDomain
	for _, addition := range additionalEndpoints {
		endpoints := strings.Join(addition.Endpoints, ",")

		if _, ok := keysPerDomain[endpoints]; ok {
			for _, apiKey := range addition.ApiKeys {
				keysPerDomain[endpoints] = append(keysPerDomain[endpoints], apiKey)
			}
		} else {
			keysPerDomain[endpoints] = addition.ApiKeys
		}
	}

	// dedupe api keys and remove domains with no api keys (or empty ones)
	// for endpoints, apiKeys := range keysPerDomain {
	// 	dedupedAPIKeys := make([]string, 0, len(apiKeys))
	// 	seen := make(map[string]bool)
	// 	for _, apiKey := range apiKeys {
	// 		trimmedAPIKey := strings.TrimSpace(apiKey)
	// 		if _, ok := seen[trimmedAPIKey]; !ok && trimmedAPIKey != "" {
	// 			seen[trimmedAPIKey] = true
	// 			dedupedAPIKeys = append(dedupedAPIKeys, trimmedAPIKey)
	// 		}
	// 	}

	// 	if len(dedupedAPIKeys) > 0 {
	// 		keysPerDomain[endpoints] = dedupedAPIKeys
	// 	} else {
	// 		klog.Infof("No API key provided for domain \"%s\", removing domain from endpoints", endpoints)
	// 		delete(keysPerDomain, endpoints)
	// 	}
	// }

	return keysPerDomain, nil
}

// AddAgentVersionToDomain prefixes the domain with the agent version: X-Y-Z.domain
func AddAgentVersionToDomain(DDURL string, app string) (string, error) {
	u, err := url.Parse(DDURL)
	if err != nil {
		return "", err
	}

	subdomain := strings.Split(u.Host, ".")[0]
	newSubdomain := getDomainPrefix(app)

	u.Host = strings.Replace(u.Host, subdomain, newSubdomain, 1)
	return u.String(), nil
}

// getDomainPrefix provides the right prefix for agent X.Y.Z
func getDomainPrefix(app string) string {
	v, _ := version.Agent()
	return fmt.Sprintf("%d-%d-%d-%s.agent", v.Major, v.Minor, v.Patch, app)
}
