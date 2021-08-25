package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/pkg/autodiscovery/common/types"
	apm "github.com/n9e/n9e-agentd/pkg/config/apm"
	forwarder "github.com/n9e/n9e-agentd/pkg/config/forwarder"
	"github.com/n9e/n9e-agentd/pkg/config/internalprofiling"
	logs "github.com/n9e/n9e-agentd/pkg/config/logs"
	snmp "github.com/n9e/n9e-agentd/pkg/config/snmp"
	statsd "github.com/n9e/n9e-agentd/pkg/config/statsd"
	systemprobe "github.com/n9e/n9e-agentd/pkg/system-probe/config"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/n9e/n9e-agentd/pkg/version"
	"github.com/yubo/golib/configer"
	"github.com/yubo/golib/proc"
	"k8s.io/klog/v2"
)

var (
	Context context.Context

	// Deprecated
	//Configfile string
	TestConfig bool

	// ungly hack, TODO: remove it
	C = NewConfig(nil)
	// StartTime is the agent startup time
	StartTime = time.Now()

	// Variables to initialize at build time
	DefaultPython      string
	ForceDefaultPython string
)

func AddFlags() {
	proc.RegisterFlags("agent", "agent generic", &Config{})

	fs := proc.NamedFlagSets().FlagSet("global")
	//fs.StringVarP(&Configfile, "config", "c", "", "Config file path of n9e agentd server.(Deprecated, use -f instead of it)")
	fs.BoolVarP(&TestConfig, "test-config", "t", false, "test configuratioin and exit")
}

type Config struct {
	m        *sync.RWMutex
	configer *configer.Configer

	IsCliRunner bool `json:"is_cli_runner"`

	ValueFiles []string `json:"-"` // from golib.configer.Setting.valueFiles

	//path
	RootDir           string `json:"root_dir" flag:"root" env:"N9E_ROOT_DIR" description:"root dir path"` // e.g. /opt/n9e/agentd
	PidfilePath       string `json:"pidfile_path"`                                                        //
	AdditionalChecksd string `json:"additional_checksd" description:"custom py checks dir"`               // additional_checksd
	CheckFlareDir     string `json:"check_flare_dir" description:"check flare directory"`
	//AuthTokenFilePath              string `json:"auth_token_file_path"`                                    // auth_token_file_path // move to apiserver

	// apiserver
	BindHost string `json:"-"` // bind_host -> apiserver.bind_host
	BindPort int    `json:"-"` // bind_host -> apiserver.bind_host
	//IPCAddress string `json:"ipc_address" default:"localhost"` //  -> apiserver.bind_host
	//CmdHost    string `json:"cmd_host" default:"localhost"` // cmd_host // -> apiserver.bind_host
	//CmdPort    int    `json:"cmd_port" default:"5001"`      // cmd_port, move to apiserver -> apiserver.bind_port

	// client
	CliQueryTimeout  time.Duration `json:"-"`
	CliQueryTimeout_ int           `json:"cli_query_timeout" flag:"cli-query-timeout" default:"5" description:"cli query timeout(Second)"`
	DisablePage      bool          `json:"disable_page" flag:"disable-page" default:"false" env:"AGENTD_DISABLE_PAGE"`
	PageSize         int           `json:"page_size" flag:"page-size" default:"10" env:"AGENTD_PAGE_SIZE"`
	NoColor          bool          `json:"no_color" flag:"no-color,n" default:"false" env:"AGENTD_NO_COLOR" description:"disable color output"`

	RunPath                  string `json:"run_path"`                                             // run_path
	JmxPath                  string `json:"jmx_path" description:"default {root}/misc/jmx"`       // jmx_path
	ConfdPath                string `json:"confd_path" description:"default {root}/conf.d"`       // confd_path
	CriSocketPath            string `json:"cri_socket_path"`                                      // cri_socket_path
	KubeletAuthTokenPath     string `json:"kubelet_auth_token_path"`                              // kubelet_auth_token_path
	KubernetesKubeconfigPath string `json:"kubernetes_kubeconfig_path"`                           // kubernetes_kubeconfig_path
	ProcfsPath               string `json:"procfs_path"`                                          // procfs_path
	WindowsUsePythonpath     string `json:"windows_use_pythonpath"`                               // windows_use_pythonpath
	DistPath                 string `json:"dist_path" description:"default {root}/dis"`           // {root}/dist
	PyChecksPath             string `json:"py_checks_path" description:"default {root}/checks.d"` // {root}/checks.d
	HostnameFile             string `json:"hostname_file"`                                        // hostname_file

	Ident             string   `json:"ident" flag:"ident" default:"$ip" description:"Ident of this host"`
	Alias             string   `json:"alias" flag:"alias" default:"$hostname" description:"Alias of the host"`
	Lang              string   `json:"lang" flag:"lang" default:"zh" description:"Default lang(zh, en)"`
	EnableN9eProvider bool     `json:"enable_n9e_provider" flag:"enable-n9e-provider" default:"true" description:"enable n9e server api as autodiscovery provider"`
	N9eSeriesFormat   bool     `json:"n9e_series_format" default:"true"`                                                                            // the payload format for forwarder
	Endpoints         []string `json:"endpoints" flag:"endpoints" default:"http://localhost:8000"  description:"endpoints addresses of n9e server"` // site, dd_url

	MetadataProviders       []MetadataProviders                 `json:"metadata_providers"`            // metadata_providers
	Forwarder               forwarder.Config                    `json:"forwarder"`                     // fowarder_*
	PrometheusScrape        PrometheusScrape                    `json:"prometheus_scrape"`             // prometheus_scrape
	Autoconfig              Autoconfig                          `json:"autoconfig"`                    //
	Container               Container                           `json:"container"`                     //
	SnmpTraps               snmp.TrapsConfig                    `json:"snmp_traps"`                    // snmp_traps_config
	SnmpListener            snmp.ListenerConfig                 `json:"snmp_listener"`                 // snmp_listener
	ClusterAgent            ClusterAgent                        `json:"cluster_agent"`                 // cluster_agent
	Network                 Network                             `json:"network"`                       // network
	NetworkConfig           NetworkConfig                       `json:"network_config"`                // network_config
	Cmd                     Cmd                                 `json:"cmd"`                           // cmd
	Logs                    logs.Config                         `json:"logs_config"`                   // logs_config
	CloudFoundryGarden      CloudFoundryGarden                  `json:"cloud_foundry_garden"`          // cloud_foundry_garden
	ClusterChecks           ClusterChecks                       `json:"cluster_checks"`                // cluster_checks
	Telemetry               Telemetry                           `json:"telemetry"`                     // telemetry
	OrchestratorExplorer    OrchestratorExplorer                `json:"orchestrator_explorer"`         // orchestrator_explorer
	Statsd                  statsd.Config                       `json:"statsd"`                        // statsd_*, dagstatsd_*
	Apm                     apm.Config                          `json:"apm_config"`                    // apm_config.*
	Jmx                     Jmx                                 `json:"jmx"`                           // jmx_*
	RuntimeSecurity         RuntimeSecurity                     `json:"runtime_security"`              // runtime_security_config.*
	AdminssionController    AdminssionController                `json:"adminssion_controller"`         // admission_controller.*
	ExternalMetricsProvider ExternalMetricsProvider             `json:"external_metrics_provider"`     // external_metrics_provider.*
	InternalProfiling       internalprofiling.InternalProfiling `json:"internal_profiling"`            // internal_profiling
	SystemProbe             systemprobe.Config                  `json:"system_probe"`                  // system_probe_config.*
	Listeners               []Listeners                         `json:"listeners"`                     // listeners
	ConfigProviders         []ConfigurationProviders            `json:"config_providers"`              // config_providers
	VerboseReport           bool                                `json:"verbose_report" default:"true"` // collects run in verbose mode, e.g. report both with cpu.used(sys+user), cpu.system & cpu.user

	ApiKey                         string   `json:"api_key"`                                                 // api_key
	Hostname                       string   `json:"hostname" flag:"hostname" description:"custom host name"` //
	HostAliases                    []string `json:"host_aliases" flag:"host-aliases"`                        // host_aliases
	HostnameFQDN                   bool     `json:"hostname_fqdn"`                                           // hostname_fqdn
	HostnameForceConfigAsCanonical bool     `json:"hostname_force_config_as_canonical"`                      // hostname_force_config_as_canonical

	MaxProcs                         string            `json:"max_procs" default:"4"`                                                                                             //
	CoreDump                         bool              `json:"core_dump" default:"true"`                                                                                          // go_core_dump
	HealthPort                       int               `json:"health_port" default:"0"`                                                                                           // health_port
	SkipSSLValidation                bool              `json:"skip_ssl_validation"`                                                                                               // skip_ssl_validation
	ForceTLS12                       bool              `json:"force_tls_12"`                                                                                                      // force_tls_12
	ECSMetadataTimeout               time.Duration     `json:"-"`                                                                                                                 // ecs_metadata_timeout
	ECSMetadataTimeout_              int               `json:"ecs_metadata_timeout" flag:"ecs-metadata-timeout" default:"500" description:"ecs metadata timeout (Millisecond)"`   // ecs_metadata_timeout
	MetadataEndpointsMaxHostnameSize int               `json:"metadata_endpoints_max_hostname_size" default:"255"`                                                                // metadata_endpoints_max_hostname_size
	CloudProviderMetadata            []string          `json:"cloud_provider_metadata"`                                                                                           //cloud_provider_metadata
	GCEMetadataTimeout               time.Duration     `json:"-"`                                                                                                                 // gce_metadata_timeout
	GCEMetadataTimeout_              int               `json:"gce_metadata_timeout" flag:"gce-metadata-timeout" default:"1000" description:"gce metadata timeout (Millisecond)"`  // gce_metadata_timeout
	ClusterName                      string            `json:"cluster_name"`                                                                                                      // cluster_name
	CLCRunnerEnabled                 bool              `json:"clc_runner_enabled"`                                                                                                //
	CLCRunnerHost                    string            `json:"clc_runner_host"`                                                                                                   // clc_runner_host
	ExtraConfigProviders             []string          `json:"extra_config_providers"`                                                                                            // extra_config_providers
	CloudFoundry                     bool              `json:"cloud_foundry"`                                                                                                     // cloud_foundry
	BoshID                           string            `json:"bosh_i_d"`                                                                                                          // bosh_id
	CfOSHostnameAliasing             bool              `json:"cf_os_hostname_aliasing"`                                                                                           // cf_os_hostname_aliasing
	CollectGCETags                   bool              `json:"collect_gce_tags" default:"true"`                                                                                   // collect_gce_tags
	CollectEC2Tags                   bool              `json:"collect_ec2_tags"`                                                                                                  // collect_ec2_tags
	DisableClusterNameTagKey         bool              `json:"disable_cluster_name_tag_key"`                                                                                      // disable_cluster_name_tag_key
	Env                              string            `json:"env"`                                                                                                               // env
	Tags                             []string          `json:"tags"`                                                                                                              // tags
	ExtraTags                        []string          `json:"extra_tags"`                                                                                                        // extra_tags
	TagValueSplitSeparator           map[string]string `json:"tag_value_split_separator"`                                                                                         // tag_value_split_separator
	NoProxyNonexactMatch             bool              `json:"no_proxy_nonexact_match"`                                                                                           // no_proxy_nonexact_match
	EnableGohai                      bool              `json:"enable_gohai" default:"true"`                                                                                       // enable_gohai
	ChecksTagCardinality             string            `json:"checks_tag_cardinality" default:"low"`                                                                              // checks_tag_cardinality
	HistogramAggregates              []string          `json:"histogram_aggregates" default:"max median avg count"`                                                               // histogram_aggregates
	HistogramPercentiles             []string          `json:"histogram_percentiles" default:"0.95"`                                                                              // histogram_percentiles
	AcLoadTimeout                    time.Duration     `json:"-"`                                                                                                                 // ac_load_timeout
	AcLoadTimeout_                   int               `json:"ac_load_timeout" flag:"ac-load-timeout" default:"30000" description:"ac load timeout(Millisecond)"`                 // ac_load_timeout
	AdConfigPollInterval             time.Duration     `json:"-"`                                                                                                                 // ad_config_poll_interval
	AdConfigPollInterval_            int               `json:"ad_config_poll_interval" flag:"ac-config-poll-interval" default:"10" description:"ac config poll interval(Second)"` // ad_config_poll_interval

	// aggregator
	AggregatorBufferSize   int           `json:"aggregator_buffer_size" default:"100"`                                                                             // aggregator_buffer_size
	AggregatorStopTimeout  time.Duration `json:"-"`                                                                                                                // aggregator_stop_timeout
	AggregatorStopTimeout_ int           `json:"aggregator_stop_timeout" flag:"aggregator-stop-timeout" default:"2" description:"aggregator stop timeout(Second)"` // aggregator_stop_timeout

	IotHost                        bool   `json:"iot_host"`                                                         // iot_host
	HerokuDyno                     bool   `json:"heroku_dyno"`                                                      // heroku_dyno
	BasicTelemetryAddContainerTags bool   `json:"basic_telemetry_add_container_tags"`                               // basic_telemetry_add_container_tags
	LogPayloads                    bool   `json:"log_payloads"`                                                     // log_payloads
	AutoconfTemplateDir            string `json:"autoconf_template_dir" description:"default {root}/check_configs"` // autoconf_template_dir
	AutoconfTemplateUrlTimeout     int    `json:"autoconf_template_url_timeout" default:"5"`                        // autoconf_template_url_timeout
	CheckRunners                   int    `json:"check_runners" default:"4"`                                        // check_runners
	GUIPort                        bool   `json:"gui_port"`                                                         // GUI_port
	XAwsEc2MetadataTokenTtlSeconds bool   `json:"x_aws_ec2_metadata_token_ttl_seconds"`                             // X-aws-ec2-metadata-token-ttl-seconds
	AcExclude                      bool   `json:"ac_exclude"`                                                       // ac_exclude
	AcInclude                      bool   `json:"ac_include"`                                                       // ac_include
	AllowArbitraryTags             bool   `json:"allow_arbitrary_tags"`                                             // allow_arbitrary_tags
	AppKey                         bool   `json:"app_key"`                                                          // app_key

	CacheSyncTimeout  time.Duration `json:"-"`
	CacheSyncTimeout_ int           `json:"cache_sync_timeout" flag:"cache-sync-timeout" default:"2" description:"cache sync timeout(Second)"` // cache_sync_timeout
	ClcRunnerId       string        `json:"clc_runner_id"`                                                                                     // clc_runner_id

	CollectKubernetesEvents      bool          `json:"collect_kubernetes_events"`                                                                                                        // collect_kubernetes_events
	ComplianceConfigDir          string        `json:"compliance_config_dir" description:"default {root}/compliance.d"`                                                                  // compliance_config.dir
	ComplianceConfigEnabled      bool          `json:"compliance_config_enabled"`                                                                                                        // compliance_config.enabled
	ContainerCgroupPrefix        string        `json:"container_cgroup_prefix"`                                                                                                          // container_cgroup_prefix
	ContainerCgroupRoot          string        `json:"container_cgroup_root"`                                                                                                            // container_cgroup_root
	ContainerProcRoot            string        `json:"container_proc_root"`                                                                                                              // container_proc_root
	ContainerdNamespace          string        `json:"containerd_namespace" default:"k8s.io"`                                                                                            // containerd_namespace
	CriConnectionTimeout         time.Duration `json:"-"`                                                                                                                                // cri_connection_timeout
	CriConnectionTimeout_        int           `json:"cri_connection_timeout" flag:"cri-connection-timeout" default:"1" description:"cri connection timeout(Second)"`                    // cri_connection_timeout
	CriQueryTimeout              time.Duration `json:"-"`                                                                                                                                // cri_query_timeout
	CriQueryTimeout_             int           `json:"cri_query_timeout" flag:"cri-query-timeout" default:"5" description:"cri query timeout(Second)"`                                   // cri_query_timeout
	DatadogCluster               bool          `json:"datadog_cluster"`                                                                                                                  // datadog-cluster
	DockerLabelsAsTags           bool          `json:"docker_labels_as_tags"`                                                                                                            // docker_labels_as_tags
	DockerQueryTimeout           time.Duration `json:"-"`                                                                                                                                // docker_query_timeout
	DockerQueryTimeout_          int           `json:"docker_query_timeout" flag:"docker-query-timeout" default:"5" description:"docker query timeout(Second)"`                          // docker_query_timeout
	EC2PreferImdsv2              bool          `json:"ec2_prefer_imdsv2"`                                                                                                                // ec2_prefer_imdsv2
	EC2MetadataTimeout           time.Duration `json:"-"`                                                                                                                                // ec2_metadata_timeout
	EC2MetadataTimeout_          int           `json:"ec2_metadata_timeout" falg:"ec2-metadata-timeout" default:"300" description:"ec2 metadata timeout(Millisecond)"`                   // ec2_metadata_timeout
	EC2MetadataTokenLifetime     time.Duration `json:"-"`                                                                                                                                // ec2_metadata_token_lifetime
	EC2MetadataTokenLifetime_    int           `json:"ec2_metadata_token_lifetime" falg:"ec2-metadata-token-lifetime" default:"21600" description:"ec2 metadata token lifetime(Second)"` // ec2_metadata_token_lifetime
	EC2UseWindowsPrefixDetection bool          `json:"ec2_use_windows_prefix_detection"`                                                                                                 // ec2_use_windows_prefix_detection
	EcsAgentContainerName        string        `json:"ecs_agent_container_name" default:"ecs-agent"`                                                                                     // ecs_agent_container_name
	EcsAgentUrl                  bool          `json:"ecs_agent_url"`                                                                                                                    // ecs_agent_url
	EcsCollectResourceTagsEc2    bool          `json:"ecs_collect_resource_tags_ec2"`                                                                                                    // ecs_collect_resource_tags_ec2
	EcsResourceTagsReplaceColon  bool          `json:"ecs_resource_tags_replace_colon"`                                                                                                  // ecs_resource_tags_replace_colon
	EKSFargate                   bool          `json:"eks_fargate"`                                                                                                                      // eks_fargate
	EnableMetadataCollection     bool          `json:"enable_metadata_collection" default:"true"`                                                                                        // enable_metadata_collection

	ExcludeGCETags []string `json:"exclude_gce_tags" default:"kube-env,kubelet-config,containerd-configure-sh,startup-script,shutdown-script,configure-sh,sshKeys,ssh-keys,user-data,cli-cert,ipsec-cert,ssl-cert,google-container-manifest,boshSettings,windows-startup-script-ps1,common-psm1,k8s-node-setup-psm1,serial-port-logging-enable,enable-oslogin,disable-address-manager,disable-legacy-endpoints,windows-keys,kubeconfig"` // exclude_gce_tags

	ExcludePauseContainer             bool              `json:"exclude_pause_container"`                         // exclude_pause_container
	ExternalMetricsAggregator         string            `json:"external_metrics_aggregator" default:"avg"`       // external_metrics.aggregator
	ExtraListeners                    []string          `json:"extra_listeners"`                                 // extra_listeners
	FullSketches                      bool              `json:"full_sketches"`                                   // full-sketches
	GceSendProjectIdTag               bool              `json:"gce_send_project_id_tag"`                         // gce_send_project_id_tag
	GoCoreDump                        bool              `json:"go_core_dump"`                                    // go_core_dump
	HpaConfigmapName                  string            `json:"hpa_configmap_name" default:"n9e-custom-metrics"` // hpa_configmap_name
	HpaWatcherGcPeriod                time.Duration     `json:"-"`
	HpaWatcherGcPeriod_               int               `json:"hpa_watcher_gc_period" flag:"hpa-watcher-gc-period" default:"300" description:"hpa_watcher_gcPeriod(Second)"` // hpa_watcher_gc_period
	IgnoreAutoconf                    []string          `json:"ignore_autoconf"`                                                                                             // ignore_autoconf
	InventoriesEnabled                bool              `json:"inventories_enabled" default:"true"`                                                                          // inventories_enabled
	InventoriesMaxInterval            time.Duration     `json:"-"`
	InventoriesMaxInterval_           int               `json:"inventories_max_interval" flag:"inventories-max-interval" default:"600" description:"inventoriesMaxInterval(Second)"` // inventories_max_interval
	InventoriesMinInterval            time.Duration     `json:"-"`
	InventoriesMinInterval_           int               `json:"inventories_min_interval" flag:"inventories-min-interval" default:"300" description:"inventoriesMinInterval(Second)"` // inventories_min_interval
	KubeResourcesNamespace            bool              `json:"kube_resources_namespace"`                                                                                            // kube_resources_namespace
	KubeletCachePodsDuration          time.Duration     `json:"-"`
	KubeletCachePodsDuration_         int               `json:"kubelet_cache_pods_duration" flag:"kubelet-cache-pods-duration" default:"5" description:"kubeletCachePodsDuration(Second)"` // kubelet_cache_pods_duration
	KubeletClientCa                   string            `json:"kubelet_client_ca"`                                                                                                         // kubelet_client_ca
	KubeletClientCrt                  string            `json:"kubelet_client_crt"`                                                                                                        // kubelet_client_crt
	KubeletClientKey                  string            `json:"kubelet_client_key"`                                                                                                        // kubelet_client_key
	KubeletListenerPollingInterval    time.Duration     `json:"-"`
	KubeletListenerPollingInterval_   int               `json:"kubelet_listener_polling_interval" flag:"kubelet-listener-polling-interval" default:"5" description:"kubeletListenerPollingInterval(Second)"` // kubelet_listener_polling_interval
	KubeletTlsVerify                  bool              `json:"kubelet_tls_verify" default:"true"`                                                                                                           // kubelet_tls_verify
	KubeletWaitOnMissingContainer     time.Duration     `json:"-"`
	KubeletWaitOnMissingContainer_    int               `json:"kubelet_wait_on_missing_container" flag:"kubelet-wait-on-missing-container" description:"kubeletWaitOnMissingContainer(Second)"` // kubelet_wait_on_missing_container
	KubernetesApiserverClientTimeout  time.Duration     `json:"-"`
	KubernetesApiserverClientTimeout_ int               `json:"kubernetes_apiserver_client_timeout" flag:"kubernetes-apiserver-client-timeout" default:"10" description:"kubernetes_apiserverClientTimeout(Seconde)"` // kubernetes_apiserver_client_timeout
	KubernetesApiserverUseProtobuf    bool              `json:"kubernetes_apiserver_use_protobuf"`                                                                                                                    // kubernetes_apiserver_use_protobuf
	KubernetesCollectMetadataTags     bool              `json:"kubernetes_collect_metadata_tags" default:"true"`                                                                                                      // kubernetes_collect_metadata_tags
	KubernetesCollectServiceTags      bool              `json:"kubernetes_collect_service_tags"`                                                                                                                      // kubernetes_collect_service_tags
	KubernetesHttpKubeletPort         int               `json:"kubernetes_http_kubelet_port" default:"10255"`                                                                                                         // kubernetes_http_kubelet_port
	KubernetesHttpsKubeletPort        int               `json:"kubernetes_https_kubelet_port" default:"10250"`                                                                                                        // kubernetes_https_kubelet_port
	KubernetesInformersResyncPeriod   time.Duration     `json:"-"`
	KubernetesInformersResyncPeriod_  int               `json:"kubernetes_informers_resync_period" flag:"kubernetes-informers-resync-period" default:"300" description:"kubernetesInformersResyncPeriod(Second)"` // kubernetes_informers_resync_period
	KubernetesKubeletHost             string            `json:"kubernetes_kubelet_host"`                                                                                                                          // kubernetes_kubelet_host
	KubernetesKubeletNodename         string            `json:"kubernetes_kubelet_nodename"`                                                                                                                      // kubernetes_kubelet_nodename
	KubernetesMapServicesOnIp         bool              `json:"kubernetes_map_services_on_ip"`                                                                                                                    // kubernetes_map_services_on_ip
	KubernetesMetadataTagUpdateFreq   time.Duration     `json:"-"`
	KubernetesMetadataTagUpdateFreq_  int               `json:"kubernetes_metadata_tag_update_freq" flag:"kubernetes-metadata-tag-update-freq" default:"60" description:"kubernetesMetadataTagUpdateFreq(Second)"` // kubernetes_metadata_tag_update_freq
	KubernetesNamespaceLabelsAsTags   bool              `json:"kubernetes_namespace_labels_as_tags"`                                                                                                               // kubernetes_namespace_labels_as_tags
	KubernetesNodeLabelsAsTags        bool              `json:"kubernetes_node_labels_as_tags"`                                                                                                                    // kubernetes_node_labels_as_tags
	KubernetesPodAnnotationsAsTags    map[string]string `json:"kubernetes_pod_annotations_as_tags"`                                                                                                                // kubernetes_pod_annotations_as_tags
	KubernetesPodExpirationDuration   time.Duration     `json:"-"`
	KubernetesPodExpirationDuration_  int               `json:"kubernetes_pod_expiration_duration" flag:"kubernetes-pod-expiration-duration" default:"900" description:"kubernetes_podExpirationDuration(Second)"` // kubernetes_pod_expiration_duration
	KubernetesPodLabelsAsTags         map[string]string `json:"kubernetes_pod_labels_as_tags"`                                                                                                                     // kubernetes_pod_labels_as_tags
	KubernetesServiceTagUpdateFreq    map[string]string `json:"kubernetes_service_tag_update_freq"`                                                                                                                // kubernetes_service_tag_update_freq
	LeaderElection                    bool              `json:"leader_election"`                                                                                                                                   // leader_election
	LeaderLeaseDuration               time.Duration     `json:"-"`
	LeaderLeaseDuration_              int               `json:"leader_lease_duration" flag:"leader-lease-duration" default:"60" description:"leader lease duration(second)"` // leader_lease_duration

	// log
	LoggingFrequency   int64    `json:"logging_frequency" default:"500"` // logging_frequency
	LogFile            string   `json:"log_file" default:"./logs/agent.log"`
	DisableFileLogging bool     `json:"disable_file_logging" flag:"disable-file-logging"` // disable_file_logging
	LogFormatJson      bool     `json:"log_format_json"`
	SyslogRfc          bool     `json:"syslog_rfc"` // syslog_rfc
	LogFormatRfc3339   bool     `json:"log_format_rfc3339"`
	LogLevel           string   `json:"log_level" flag:"log-level" default:"info" description:"info, debug"`
	LogToConsole       bool     `json:"log_to_console" flag:"log-to-console" default:"true"`
	LogToSyslog        bool     `json:"log_to_syslog" flag:"log-to-syslog"`
	SyslogUri          string   `json:"syslog_uri"`
	SyslogPem          string   `json:"syslog_pem"`
	SyslogKey          string   `json:"syslog_key"`
	SyslogTlsVerify    bool     `json:"syslog_tls_verify"`
	FlareStrippedKeys  []string `json:"flare_stripped_keys"`
	LogFileMaxSize     int      `json:"log_file_max_size" default:"10485760"`
	LogFileMaxRolls    int      `json:"log_file_max_rolls" default:"1"`

	MemtrackEnabled           bool          `json:"memtrack_enabled"`              // memtrack_enabled
	MetricsPort               int           `json:"metrics_port" default:"5000"`   // metrics_port
	ProcRoot                  string        `json:"proc_root" default:"/proc"`     // proc_root
	ProcessAgentConfigHostIps bool          `json:"process_agent_config_host_ips"` // process_agent_config.host_ips
	ProcessConfig             ProcessConfig `json:"process_config"`                // process_config

	PrometheusScrapeEnabled          bool  `json:"prometheus_scrape_enabled"`           // prometheus_scrape.enabled
	PrometheusScrapeServiceEndpoints bool  `json:"prometheus_scrape_service_endpoints"` // prometheus_scrape.service_endpoints
	Proxy                            Proxy `json:"proxy"`

	// python
	Python3LinterTimeout             time.Duration `json:"-"`
	Python3LinterTimeout_            int           `json:"python3_linter_timeout" flag:"python3-linter-timeout" default:"120" description:"python3LinterTimeout(Second)"` // python3_linter_timeout
	PythonVersion                    string        `json:"python_version" default:"3"`                                                                                    // python_version
	PythonHome                       string        `json:"python_home" flag:"python-home" description:"default {root}/embedded"`
	AllowPythonPathHeuristicsFailure bool          `json:"allow_python_path_heuristics_failure"`
	CCoreDump                        bool          `json:"c_core_dump"`             // c_core_dump
	CStacktraceCollection            bool          `json:"c_stacktrace_collection"` // c_stacktrace_collection
	DisablePy3Validation             bool          `json:"disable_py3_validation"`  // disable_py3_validation
	WinSkipComInit                   bool          `json:"win_skip_com_init"`       // win_skip_com_init

	// serializer
	EnableJsonStreamSharedCompressorBuffers       bool           `json:"enable_json_stream_shared_compressor_buffers" default:"true"`                  //enable_json_stream_shared_compressor_buffers
	EnablePayloads                                EnablePayloads `json:"enable_payloads"`                                                              // enable_payloads.*
	SerializerMaxPayloadSize                      int            `json:"serializer_max_payload_size" default:"2621440" description:"2.5mb"`            // serializer_max_payload_size
	SerializerMaxUncompressedPayloadSize          int            `json:"serializer_max_uncompressed_payload_size" default:"4194304" description:"4mb"` // serializer_max_uncompressed_payload_size
	EnableStreamPayloadSerialization              bool           `json:"enable_stream_payload_serialization" default:"false"`
	EnableServiceChecksStreamPayloadSerialization bool           `json:"enable_service_checks_stream_payload_serialization" default:"true"`
	EnableEventsStreamPayloadSerialization        bool           `json:"enable_events_stream_payload_serialization" default:"true"`
	EnableSketchStreamPayloadSerialization        bool           `json:"enable_sketch_stream_payload_serialization" default:"true"`

	//ServerTimeout                        time.Duration     `json:"-"`
	//ServerTimeout_                       int               `json:"server_timeout" flag:"server-timeout" default:"15" description:"server_timeout(Second)"` // server_timeout, move to apiserver
	TracemallocDebug    bool   `json:"tracemalloc_debug"` // tracemalloc_debug
	Yaml                bool   `json:"yaml"`              // yaml
	MetricTransformFile string `json:"metric_transform_file"`
	//N9e                                  N9e                      `json:"n9e"`

	LogAllGoroutinesWhenUnhealthy bool         `json:"log_all_goroutines_when_unhealthy"`
	AzureHostnameStyle            string       `json:"azure_hostname_style" default:"os"`
	Experimental                  Experimental `json:"experimental"`
	AutoconfigFromEnvironment     bool         `json:"autoconfig_from_environment" env:"AUTOCONFIG_FROM_ENVIRONMENT"`
	AutoconfigExcludeFeatures     []string     `json:"autoconfig_exclude_features"`
	AutoconfigIncludeFeatures     []string     `json:"autoconfig_include_features"`
	UseV2Api                      UseV2Api     `json:"use_v2_api"`

	SecretBackendSkipChecks              bool `json:"secret_backend_skip_checks"`
	CheckSamplerBucketCommitsCountExpiry int  `json:"check_sampler_bucket_commits_count_expiry"`
} // end of Config

func (p *Config) IsSet(path string) bool {
	p.m.RLock()
	defer p.m.RUnlock()
	return p.configer.IsSet("agent." + path)
}

func (p *Config) Get(path string) interface{} {
	p.m.RLock()
	defer p.m.RUnlock()
	return p.configer.GetRaw("agent." + path)
}

func (p *Config) Set(k string, v interface{}) error {
	p.m.Lock()
	defer p.m.Unlock()
	return p.configer.Set(k, v)
}

func (p *Config) Read(path string, dst interface{}) error {
	p.m.Lock()
	defer p.m.Unlock()
	return p.configer.Read(path, dst)
}

func (p Config) String() string {
	p.m.RLock()
	defer p.m.RUnlock()
	return util.Prettify(p)
}

func (p *Config) prepareTimeDuration() error {
	p.CliQueryTimeout = time.Second * time.Duration(p.CliQueryTimeout_)
	p.ECSMetadataTimeout = time.Millisecond * time.Duration(p.ECSMetadataTimeout_)
	p.GCEMetadataTimeout = time.Millisecond * time.Duration(p.GCEMetadataTimeout_)
	p.AdConfigPollInterval = time.Second * time.Duration(p.AdConfigPollInterval_)
	p.AggregatorStopTimeout = time.Second * time.Duration(p.AggregatorStopTimeout_)
	p.AcLoadTimeout = time.Millisecond * time.Duration(p.AcLoadTimeout_)
	p.CacheSyncTimeout = time.Second * time.Duration(p.CacheSyncTimeout_)
	p.CriConnectionTimeout = time.Second * time.Duration(p.CriConnectionTimeout_)
	p.CriQueryTimeout = time.Second * time.Duration(p.CriQueryTimeout_)
	p.DockerQueryTimeout = time.Second * time.Duration(p.DockerQueryTimeout_)
	p.EC2MetadataTimeout = time.Millisecond * time.Duration(p.EC2MetadataTimeout_)
	p.EC2MetadataTokenLifetime = time.Second * time.Duration(p.EC2MetadataTokenLifetime_)
	p.HpaWatcherGcPeriod = time.Second * time.Duration(p.HpaWatcherGcPeriod_)
	p.InventoriesMaxInterval = time.Second * time.Duration(p.InventoriesMaxInterval_)
	p.InventoriesMinInterval = time.Second * time.Duration(p.InventoriesMinInterval_)
	p.KubeletCachePodsDuration = time.Second * time.Duration(p.KubeletCachePodsDuration_)
	p.KubeletWaitOnMissingContainer = time.Second * time.Duration(p.KubeletWaitOnMissingContainer_)
	p.KubernetesApiserverClientTimeout = time.Second * time.Duration(p.KubernetesApiserverClientTimeout_)
	p.KubernetesInformersResyncPeriod = time.Second * time.Duration(p.KubernetesInformersResyncPeriod_)
	p.KubernetesMetadataTagUpdateFreq = time.Second * time.Duration(p.KubernetesMetadataTagUpdateFreq_)
	p.KubernetesPodExpirationDuration = time.Second * time.Duration(p.KubernetesPodExpirationDuration_)
	p.LeaderLeaseDuration = time.Second * time.Duration(p.LeaderLeaseDuration_)
	p.Python3LinterTimeout = time.Second * time.Duration(p.Python3LinterTimeout_)

	return nil
}

type validator interface {
	Validate() error
}

// Validate will be called by configer.Read()
func (p *Config) Validate() error {
	if err := p.prepareTimeDuration(); err != nil {
		return err
	}

	if len(p.Endpoints) == 0 && !p.IsCliRunner {
		return fmt.Errorf("unable to get agent.endpoints")
	}

	for i, endpoint := range p.Endpoints {
		if _, err := url.Parse(endpoint); err != nil {
			return fmt.Errorf("could not parse agent.endpoint[%d]: %s %s", i, endpoint, err)
		}
	}

	if err := util.FieldsValidate(p); err != nil {
		return err
	}

	if err := p.SnmpTraps.Validate(p.GetBindHost()); err != nil {
		return err
	}

	// ClusterCheck
	p.Ident = configEval(p.Ident)
	p.Alias = configEval(p.Alias)

	if strings.Contains(p.Ident, "localhost") || strings.Contains(p.Ident, "127.0.0.1") {
		return fmt.Errorf("agent.ident should not include 'localhost'")
	}

	if err := p.ValidatePath(); err != nil {
		return err
	}

	if p.EnableN9eProvider {
		p.UseV2Api.Series = true
	}

	if ForceDefaultPython == "true" {
		if p.PythonVersion != DefaultPython {
			klog.Warningf("Python version has been forced to %s", DefaultPython)
		}
		p.PythonVersion = DefaultPython
	}

	// transformer
	if err := defaultTransformer.SetMetricFromFile(p.MetricTransformFile); err != nil {
		return err
	}

	// apiserver
	p.BindHost = p.configer.GetString("apiserver.bind_host")
	p.BindPort, _ = p.configer.GetInt("apiserver.bind_port")

	// TODO
	DetectFeatures()
	//applyOverrideFuncs(p)
	// setTracemallocEnabled *must* be called before setNumWorkers
	//warnings.TraceMallocEnabledWithPy2 = setTracemallocEnabled(config)

	return nil
}

func (p *Config) ValidatePath() (err error) {
	if p.RootDir, err = util.ResolveRootPath(p.RootDir); err != nil {
		return err
	}
	if !util.IsDir(p.RootDir) {
		return fmt.Errorf("agent.workDir %s does not exist, please create it", p.RootDir)
	}
	os.Chdir(p.RootDir)
	klog.V(1).InfoS("agent", "root_dir", p.RootDir, "chdir", p.RootDir)

	root := util.NewRootDir(p.RootDir)

	// {root}/conf.d
	p.ConfdPath = root.Abs(p.ConfdPath, "conf.d")
	if !util.IsDir(p.ConfdPath) {
		klog.Warningf("agent.confd_path %s does not exist, please create it", p.ConfdPath)
	}
	klog.V(1).InfoS("agent", "confd_path", p.ConfdPath)

	// {root}/run
	p.RunPath = root.Abs(p.RunPath, "run")
	if !util.IsDir(p.RunPath) {
		return fmt.Errorf("agent.run_path %s does not exist, please create it", p.RunPath)
	}

	// {root}/misc/jmx
	p.JmxPath = root.Abs(p.JmxPath, "misc", "jmx")
	klog.V(1).InfoS("agent", "jmx_path", p.JmxPath)

	// {root}/run
	p.Logs.RunPath = root.Abs(p.Logs.RunPath, p.RunPath)
	klog.V(1).InfoS("agent", "logs_config.run_path", p.Logs.RunPath)

	// {root}/run/transactions_to_retry
	p.Forwarder.StoragePath = root.Abs(p.Forwarder.StoragePath, p.RunPath, "transactions_to_retry")
	klog.V(1).InfoS("agent", "forwarder.storage_path", p.Forwarder.StoragePath)

	p.DistPath = root.Abs(p.DistPath, "dist")
	klog.V(1).InfoS("agent", "dist_path", p.DistPath)

	p.PyChecksPath = root.Abs(p.PyChecksPath, "checks.d")
	klog.V(1).InfoS("agent", "py_checks_path", p.PyChecksPath)

	p.PythonHome = root.Abs(p.PythonHome, "embedded")
	klog.V(1).InfoS("agent", "python_home", p.PythonHome)

	// {root}/{name}.sock

	// {root}/logs/checks for check flare
	p.CheckFlareDir = root.Abs(p.PythonHome, "logs", "checks")
	klog.V(1).InfoS("agent", "check_flare_dir", p.CheckFlareDir)

	p.RuntimeSecurity.PoliciesDir = root.Abs(p.RuntimeSecurity.PoliciesDir, "etc", "runtime-security.d")
	p.SystemProbe.SocketAddress = root.Abs(p.SystemProbe.SocketAddress, "run", "sysprobe.sock")
	p.SystemProbe.LogFile = root.Abs(p.SystemProbe.LogFile, "logs", "system-probe.log")
	p.ComplianceConfigDir = root.Abs(p.ComplianceConfigDir, "compliance.d")
	p.AutoconfTemplateDir = root.Abs(p.AutoconfTemplateDir, "check_configs")

	return nil
}

type UseV2Api struct {
	Series        bool `json:"series" default:"true"`
	Events        bool `json:"events" default:"false"`
	ServiceChecks bool `json:"service_checks" default:"false"`
}

type Container struct {
	DockerHost            string   `json:"-"`
	EksFargate            bool     `json:"eks_fargate"`
	CriSocketPath         string   `json:"cri_socket_path"`
	IncludeMetrics        []string `json:"include_metrics"`         // container_include, container_include_metrics, ac_include
	ExcludeMetrics        []string `json:"exclude_metrics"`         // container_exclude, container_exclude_metrics, ac_exclude
	IncludeLogs           []string `json:"include_logs"`            // container_include_logs
	ExcludeLogs           []string `json:"exclude_logs"`            // container_exclude_logs
	ExcludePauseContainer bool     `json:"exclude_parse_container"` // exclude_pause_container
}

func (p *Container) Validate() error {
	return nil
}

type ProcessConfig struct {
	Enabled                         bool                `json:"enabled"`                           // process_config.enabled
	OrchestratorAdditionalEndpoints map[string][]string `json:"orchestrator_additional_endpoints"` // process_config.orchestrator_additional_endpoints
	Url                             string              `json:"url"`                               // process_config.orchestrator_dd_url
}

type Experimental struct {
	OTLP OTLP `json:"otlp"`
}

type OTLP struct {
	HTTPPort int `json:"http_port"`
	GRPCPort int `json:"grpc_port"`
}
type Proxy struct {
	HTTP    string   `json:"http"`     // proxy.http
	HTTPS   string   `json:"https"`    // proxy.https
	NoProxy []string `json:"no_proxy"` // proxy.no_proxy
}

//type N9e struct {
//	Enabled  bool   `json:"enabled"`
//	Endpoint string `json:"endpoint"`
//	V5Format bool   `json:"v5_format"`
//}

type EnablePayloads struct {
	Series              bool `json:"series" default:"true"`                // enable_payloads.series
	Events              bool `json:"events" default:"false"`               // enable_payloads.events
	ServiceChecks       bool `json:"service_checks" default:"false"`       // enable_payloads.service_checks
	Sketches            bool `json:"sketches" default:"false"`             // enable_payloads.sketches
	JsonToV1Intake      bool `json:"json_to_v1_intake" default:"false"`    // enable_payloads.json_to_v1_intake
	Metadata            bool `json:"metadata" default:"false"`             //
	HostMetadata        bool `json:"host_metadata" default:"false"`        //
	AgentchecksMetadata bool `json:"agentchecks_metadata" default:"false"` //
}

func (p *EnablePayloads) Validate() error {
	return nil
}

type ExternalMetricsProvider struct {
	ApiKey                bool          `json:"api_key"`                   // external_metrics_provider.api_key
	AppKey                bool          `json:"app_key"`                   // external_metrics_provider.app_key
	BucketSize            int           `json:"bucket_size" default:"300"` // external_metrics_provider.bucket_size
	Enabled               bool          `json:"enabled"`                   // external_metrics_provider.enabled
	LocalCopyRefreshRate  time.Duration `json:"-"`
	LocalCopyRefreshRate_ int           `json:"local_copy_refresh_rate" flag:"external-metrics-provider-local-copy-refresh-rate" default:"30" description:"localCopyRefreshRate(Second)"` // external_metrics_provider.local_copy_refresh_rate
	MaxAge                int           `json:"max_age" default:"20"`                                                                                                                     // external_metrics_provider.max_age
	RefreshPeriod         int           `json:"refresh_period" default:"30"`                                                                                                              // external_metrics_provider.refresh_period
	Rollup                int           `json:"rollup" default:"30"`                                                                                                                      // external_metrics_provider.rollup
	UseDatadogmetricCrd   bool          `json:"use_datadogmetric_crd"`                                                                                                                    // external_metrics_provider.use_datadogmetric_crd
	WpaController         bool          `json:"wpa_controller"`                                                                                                                           // external_metrics_provider.wpa_controller
}

func (p *ExternalMetricsProvider) Validate() error {
	p.LocalCopyRefreshRate = time.Second * time.Duration(p.LocalCopyRefreshRate_)
	return nil
}

type AdminssionController struct {
	Enabled                         bool          `json:"enabled"` // admission_controller.enabled
	CertificateExpirationThreshold  time.Duration `json:"-"`
	CertificateExpirationThreshold_ int           `json:"certificate_expiration_threshold" flag:"admission-controller-certificate-expiration-threshold" default:"30" description:"certificateExpirationThreshold(Day)"` // admission_controller.certificate.expiration_threshold
	CertificateSecretName           string        `json:"certificate_secret_name" default:"webhook-certificate"`                                                                                                        // admission_controller.certificate.secret_name
	CertificateValidityBound        time.Duration `json:"-"`
	CertificateValidityBound_       int           `json:"certificate_validity_bound" flag:"admission-controller-certificate-validity-bound" default:"365" description:"certificateValidityBound(Day)"` // admission_controller.certificate.validity_bound
	InjectConfigEnabled             bool          `json:"inject_config_enabled" default:"true"`                                                                                                        // admission_controller.inject_config.enabled
	InjectConfigEndpoint            string        `json:"inject_config_endpoint" default:"/injectconfig"`                                                                                              // admission_controller.inject_config.endpoint
	InjectTagsEnabled               bool          `json:"inject_tags_enabled" default:"true"`                                                                                                          // admission_controller.inject_tags.enabled
	InjectTagsEndpoint              string        `json:"inject_tags_endpoint" default:"/injecttags"`                                                                                                  // admission_controller.inject_tags.endpoint
	MutateUnlabelled                bool          `json:"mutate_unlabelled"`                                                                                                                           // admission_controller.mutate_unlabelled
	PodOwnersCacheValidity          int           `json:"pod_owners_cache_validity" default:"10" description:"Minute"`                                                                                 // admission_controller.pod_owners_cache_validity
	ServiceName                     string        `json:"service_name" default:"admission-controller"`                                                                                                 // admission_controller.service_name
	TimeoutSeconds                  time.Duration `json:"-"`
	TimeoutSeconds_                 int           `json:"timeout_seconds" flag:"adminssion-controller-timeout" default:"30" description:"timeoutSeconds(Second)"` // admission_controller.timeout_seconds
	WebhookName                     string        `json:"webhook_name" default:"n9e-webhook"`                                                                     // admission_controller.webhook_name

}

func (p *AdminssionController) Validate() error {
	p.CertificateExpirationThreshold = 24 * time.Hour * time.Duration(p.CertificateExpirationThreshold_)
	p.CertificateValidityBound = 24 * time.Hour * time.Duration(p.CertificateValidityBound_)
	p.TimeoutSeconds = time.Second * time.Duration(p.TimeoutSeconds_)
	return nil
}

type RuntimeSecurity struct {
	Enabled                            bool   `json:"enabled"`                                                          // runtime_security_config.enabled
	Socket                             string `json:"socket"`                                                           // runtime_security_config.socket
	AgentMonitoringEvents              bool   `json:"agent_monitoring_events"`                                          // runtime_security_config.agent_monitoring_events
	CookieCacheSize                    bool   `json:"cookie_cache_size"`                                                // runtime_security_config.cookie_cache_size
	CustomSensitiveWords               bool   `json:"custom_sensitive_words"`                                           // runtime_security_config.custom_sensitive_words
	EnableApprovers                    bool   `json:"enable_approvers"`                                                 // runtime_security_config.enable_approvers
	EnableDiscarders                   bool   `json:"enable_discarders"`                                                // runtime_security_config.enable_discarders
	EnableKernelFilters                bool   `json:"enable_kernel_filters"`                                            // runtime_security_config.enable_kernel_filters
	EventServerBurst                   bool   `json:"event_server_burst"`                                               // runtime_security_config.event_server.burst
	EventServerRate                    bool   `json:"event_server_rate"`                                                // runtime_security_config.event_server.rate
	EventsStatsPollingInterval         bool   `json:"events_stats_polling_interval"`                                    // runtime_security_config.events_stats.polling_interval
	FimEnabled                         bool   `json:"fim_enabled"`                                                      // runtime_security_config.fim_enabled
	FlushDiscarderWindow               bool   `json:"flush_discarder_window"`                                           // runtime_security_config.flush_discarder_window
	LoadControllerControlPeriod        bool   `json:"load_controller_control_period"`                                   // runtime_security_config.load_controller.control_period
	LoadControllerDiscarderTimeout     bool   `json:"load_controller_discarder_timeout"`                                // runtime_security_config.load_controller.discarder_timeout
	LoadControllerEventsCountThreshold bool   `json:"load_controller_events_count_threshold"`                           // runtime_security_config.load_controller.events_count_threshold
	PidCacheSize                       bool   `json:"pid_cache_size"`                                                   // runtime_security_config.pid_cache_size
	PoliciesDir                        string `json:"policies_dir" description:"default {root}/etc/runtime-security.d"` // runtime_security_config.policies.dir
	SyscallMonitorEnabled              bool   `json:"syscall_monitor_enabled"`                                          // runtime_security_config.syscall_monitor.enabled
}

func (p *RuntimeSecurity) Validate() error {
	return nil
}

type Jmx struct {
	CheckPeriod                int           `json:"check_period"` // jmx_check_period
	CollectionTimeout          time.Duration `json:"-"`
	CollectionTimeout_         int           `json:"collection_timeout" flag:"jmx-collection-timeout" default:"60" description:"collectionTimeout"` // jmx_collection_timeout
	CustomJars                 []string      `json:"custom_jars"`                                                                                   // jmx_custom_jars
	LogFile                    string        `json:"log_file" default:"./logs/jmxfetch.log"`                                                        // jmx_log_file
	MaxRestarts                int           `json:"max_restarts" default:"3"`                                                                      // jmx_max_restarts
	ReconnectionThreadPoolSize int           `json:"reconnection_thread_pool_size" default:"3"`                                                     // jmx_reconnection_thread_pool_size
	ReconnectionTimeout        time.Duration `json:"-"`
	ReconnectionTimeout_       int           `json:"reconnection_timeout" flag:"jmx-reconnection-timeout" default:"50" description:"reconnectionTimeout(Second)"` // jmx_reconnection_timeout
	RestartInterval            time.Duration `json:"-"`
	RestartInterval_           int           `json:"restart_interval" flag:"jmx-restart-interval" default:"5" description:"restartInterval(Second)"` // jmx_restart_interval
	ThreadPoolSize             int           `json:"thread_pool_size" default:"3"`                                                                   // jmx_thread_pool_size
	UseCgroupMemoryLimit       bool          `json:"use_cgroup_memory_limit"`                                                                        // jmx_use_cgroup_memory_limit
	UseContainerSupport        bool          `json:"use_container_support"`                                                                          // jmx_use_container_support
}

func (p *Jmx) Validate() error {
	p.CollectionTimeout = time.Second * time.Duration(p.CollectionTimeout_)
	p.ReconnectionTimeout = time.Second * time.Duration(p.ReconnectionTimeout_)
	p.RestartInterval = time.Second * time.Duration(p.RestartInterval_)
	return nil
}

// MetadataProviders helps unmarshalling `metadata_providers` config param
type MetadataProviders struct {
	Name      string        `json:"name"`
	Interval  time.Duration `json:"-"`
	Interval_ int           `json:"interval" description:"interval(Second)"`
}

func (p *MetadataProviders) Validate() error {
	p.Interval = time.Second * time.Duration(p.Interval_)
	return nil
}

type Cmd struct {
	Check Check `json:"check"` // cmd.check
}

type Check struct {
	Fullsketches bool `json:"fullsketches"` // cmd.check.fullsketches
}

type CloudFoundryGarden struct {
	ListenNetwork string `json:"listen_network" default:"unix"`                              // cloud_foundry_garden.listen_network
	ListenAddress string `json:"listen_address" default:"/var/vcap/data/garden/garden.sock"` // cloud_foundry_garden.listen_address
}

// ProcessingRule defines an exclusion or a masking rule to
// be applied on log lines
type ProcessingRule struct {
	Type               string
	Name               string
	ReplacePlaceholder string `json:"replace_placeholder"`
	Pattern            string
	// TODO: should be moved out
	Regex       *regexp.Regexp
	Placeholder []byte
}

func (p *ProcessingRule) Validate() error {
	return nil
}

type Network struct {
	ID string `json:"id"` // network.id
}

func (p *Network) Validate() error {
	return nil
}

type NetworkConfig struct {
	Enabled              bool `json:"enabled" default:"true"`  // network_config.enabled"
	EnableDnsByQuerytype bool `json:"enable_dns_by_querytype"` // network_config.enable_dns_by_querytype"
	//EnableHttpMonitoring       bool   `json:"enable_http_monitoring"`                 // network_config.enable_http_monitoring
	//IgnoreConntrackInitFailure bool   `json:"ignore_conntrack_init_failure"`          // network_config.ignore_conntrack_init_failure
	//EnableGatewayLookup        bool   `json:"enable_gateway_lookup"`                  // network_config.enable_gateway_lookup
}

func (p *NetworkConfig) Validate() error {
	return nil
}

type Telemetry struct {
	Enabled bool     `json:"enabled" default:"true"` // telemetry.enabled
	Port    int      `json:"port" default:"8011"`    // expvar_port
	Docs    bool     `json:"docs" default:"true"`    // /docs/*
	Metrics bool     `json:"metrics" default:"true"` // /metrics/
	Expvar  bool     `json:"expvar" default:"true"`  // /vars
	Pprof   bool     `json:"pprof" default:"true"`   // /debug/pprof
	Checks  []string `json:"checks"`                 // telemetry.checks

	Statsd TelemetryStatsd `json:"statsd"` // telemetry.dogstatsd
}

func (p *Telemetry) Validate() error {
	return nil
}

type TelemetryStatsd struct {
	AggregatorChannelLatencyBuckets []float64 `json:"aggregator_channel_latency_buckets" defualt:"100 250 500 1000 10000"`
	ListenersLatencyBuckets         []float64 `json:"listeners_latency_buckets" defualt:""`
	ListenersChannelLatencyBuckets  []float64 `json:"listeners_channel_latency_buckets" defualt:""`
}
type ClusterChecks struct {
	ClcRunnersPort             int           `json:"clc_runners_port" default:"5005"`         // cluster_checks.clc_runners_port
	AdvancedDispatchingEnabled bool          `json:"advanced_dispatching_enabled"`            // cluster_checks.advanced_dispatching_enabled
	ClusterTagName             string        `json:"cluster_tag_name" default:"cluisterName"` // cluster_checks.cluster_tag_name
	Enabled                    bool          `json:"enabled"`                                 // cluster_checks.enabled
	ExtraTags                  []string      `json:"extra_tags"`                              // cluster_checks.extra_tags
	NodeExpirationTimeout      time.Duration `json:"-"`
	NodeExpirationTimeout_     int           `json:"node_expiration_timeout" flag:"clc-node-expiration-timeout" default:"30" description:"nodeExpirationTimeout(Second)"` // cluster_checks.node_expiration_timeout
	WarmupDuration             time.Duration `json:"-"`
	WarmupDuration_            int           `json:"warmup_duration" flag:"clc-warmup-duration" default:"30" description:"warmupDuration(Second)"` // cluster_checks.warmup_duration

}

func (p *ClusterChecks) Validate() error {
	p.NodeExpirationTimeout = time.Second * time.Duration(p.NodeExpirationTimeout_)
	p.WarmupDuration = time.Second * time.Duration(p.WarmupDuration_)

	return nil
}

type ClusterAgent struct {
	Url                   string `json:"url"`                                                 // cluster_agent.url
	AuthToken             string `json:"auth_token"`                                          // cluster_agent.auth_token
	CmdPort               int    `json:"cmd_port" default:"5005"`                             // cluster_agent.cmd_port
	Enabled               bool   `json:"enabled"`                                             // cluster_agent.enabled
	KubernetesServiceName string `json:"kubernetes_service_name" default:"n9e-cluster-agent"` // cluster_agent.kubernetes_service_name
	TaggingFallback       string `json:"tagging_fallback"`                                    // cluster_agent.tagging_fallback
}

func (p *ClusterAgent) Validate() error {
	return nil
}

type OrchestratorExplorer struct { // orchestrator_explorer
	Url                       string              `json:"url"`                         // orchestrator_explorer.orchestrator_dd_url
	AdditionalEndpoints       map[string][]string `json:"additional_endpoints"`        // orchestrator_explorer.orchestrator_additional_endpoints
	CustomSensitiveWords      []string            `json:"custom_sensitive_words"`      // orchestrator_explorer.custom_sensitive_words
	ContainerScrubbingEnabled bool                `json:"container_scrubbing_enabled"` // orchestrator_explorer.container_scrubbing.enabled
	Enabled                   bool                `json:"enabled" default:"true"`      // orchestrator_explorer.enabled
	ExtraTags                 []string            `json:"extra_tags"`                  // orchestrator_explorer.extra_tags
	MaxPerMessage             int                 `json:"max_per_message"`
	PodQueueBytes             int                 `json:"pod_queue_bytes"`
}

func (p *OrchestratorExplorer) Validate() error {
	return nil
}

type Autoconfig struct {
	Enabled         bool     `json:"enabled" default:"true"` // autoconfig_from_environment
	ExcludeFeatures []string `json:"exclude_features"`       // autoconfig_exclude_features
	features        FeatureMap
}

func (p *Autoconfig) Validate() error {
	return nil
}

type PrometheusScrape struct {
	Enabled          bool                     `json:"enabled"`           // prometheus_scrape.enabled
	ServiceEndpoints bool                     `json:"service_endpoints"` // prometheus_scrape.service_endpoints
	Checks           []*types.PrometheusCheck `json:"checks"`            // prometheus_scrape.checks
}

func (p *PrometheusScrape) Validate() error {
	return nil
}

// ConfigurationProviders helps unmarshalling `config_providers` config param
type ConfigurationProviders struct {
	Name             string `json:"name"`
	Polling          bool   `json:"polling"`
	PollInterval     string `json:"poll_interval"`
	TemplateURL      string `json:"template_url"`
	TemplateDir      string `json:"template_dir"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	CAFile           string `json:"ca_file"`
	CAPath           string `json:"ca_path"`
	CertFile         string `json:"cert_file"`
	KeyFile          string `json:"key_file"`
	Token            string `json:"token"`
	GraceTimeSeconds int    `json:"grace_time_seconds"`
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

func (cf *Config) GetIPCAddress() (string, error) {
	return cf.GetBindHost(), nil
}

// GetBindHost returns `bind_host` variable or default value
// Not using `config.BindEnvAndSetDefault` as some processes need to know
// if value was default one or not (e.g. trace-agent)
func (cf *Config) GetBindHost() string {
	return cf.BindHost
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
