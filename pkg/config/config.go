package config

import (
	"context"
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
	"github.com/yubo/golib/proc"
	"k8s.io/klog/v2"
)

const (
	//authTokenName = "auth_token"

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
	// DefaultRuntimePoliciesDir = "/opt/n9e/agentd/runtime-security.d"
)

var (
	Context context.Context

	// Deprecated
	Configfile string
	TestConfig bool

	// ungly hack, TODO: remove it
	C = NewDefaultConfig()
	// StartTime is the agent startup time
	StartTime = time.Now()

	// Variables to initialize at build time
	DefaultPython      string
	ForceDefaultPython string
)

func AddFlags() {
	proc.RegisterFlags("agent", "agent generic", &Config{})
	proc.RegisterFlags("agent.forwarder", "forwarder", &Forwarder{})
	proc.RegisterFlags("agent.logsConfig", "logs", &LogsConfig{})
	proc.RegisterFlags("agent.clusterChecks", "cluster checks", &ClusterChecks{})
	proc.RegisterFlags("agent.statsd", "statsd", &Statsd{})
	proc.RegisterFlags("agent.jmx", "jmx", &Jmx{})
	proc.RegisterFlags("agent.adminssionController", "adminssion controller", &AdminssionController{})

	fs := proc.NamedFlagSets().FlagSet("global")
	fs.StringVarP(&Configfile, "config", "c", "", "Config file path of n9e agentd server.(Deprecated, use -f instead of it)")
	fs.BoolVarP(&TestConfig, "test-config", "t", false, "test configuratioin and exit")

}

type Config struct {
	//path
	WorkDir           string `json:"work_dir" flag:"workdir,w" env:"AGENTD_WORK_DIR" description:"workdir"` // e.g. /opt/n9e/agentd
	PidfilePath       string `json:"pidfile_path"`                                                          //
	AdditionalChecksd string `json:"additional_checksd"`                                                    // additional_checksd
	//AuthTokenFilePath              string `json:"auth_token_file_path"`                                    // auth_token_file_path // move to apiserver
	RunPath                  string `json:"run_path"`                   // run_path
	ConfdPath                string `json:"confd_path"`                 // confd_path
	CriSocketPath            string `json:"cri_socket_path"`            // cri_socket_path
	KubeletAuthTokenPath     string `json:"kubelet_auth_token_path"`    // kubelet_auth_token_path
	KubernetesKubeconfigPath string `json:"kubernetes_kubeconfig_path"` // kubernetes_kubeconfig_path
	ProcfsPath               string `json:"procfs_path"`                // procfs_path
	WindowsUsePythonpath     string `json:"windows_use_pythonpath"`     // windows_use_pythonpath
	DistPath                 string `json:"-"`                          // {workdir}/dist
	PyChecksPath             string `json:"-"`                          // {workdir}/checks.d

	Ident             string   `json:"ident" flag:"ident" default:"$ip" description:"Ident of this host"`
	Alias             string   `json:"alias" flag:"alias" default:"$hostname" description:"Alias of the host"`
	Lang              string   `json:"lang" flag:"lang" default:"zh" description:"Default lang(zh, en)"`
	EnableN9eProvider bool     `json:"enable_n9e_provider" flag:"enable-n9e-provider" default:"true" description:"enable n9e server api as autodiscovery provider"`
	N9eSeriesFormat   bool     `json:"n9e_series_format" default:"true"`                                                                            // the payload format for forwarder
	Endpoints         []string `json:"endpoints" flag:"endpoints" default:"http://localhost:8000"  description:"endpoints addresses of n9e server"` // site, dd_url

	MetadataProviders              []MetadataProviders      `json:"metadata_providers"`                                      // metadata_providers
	Forwarder                      Forwarder                `json:"forwarder,inline"`                                        // fowarder_*
	PrometheusScrape               PrometheusScrape         `json:"prometheus_scrape,inline"`                                // prometheus_scrape
	Autoconfig                     Autoconfig               `json:"autoconfig,inline"`                                       //
	Container                      Container                `json:"container,inline"`                                        //
	SnmpTraps                      traps.Config             `json:"snmp_traps,inline"`                                       // snmp_traps_config
	ClusterAgent                   ClusterAgent             `json:"cluster_agent,inline"`                                    // cluster_agent
	Network                        Network                  `json:"network,inline"`                                          // network
	SnmpListener                   snmptypes.ListenerConfig `json:"snmp_listener,inline"`                                    // snmp_listener
	Cmd                            Cmd                      `json:"cmd,inline"`                                              // cmd
	LogsConfig                     LogsConfig               `json:"logs_config"`                                             // logs_config
	CloudFoundryGarden             CloudFoundryGarden       `json:"cloud_foundry_garden,inline"`                             // cloud_foundry_garden
	ClusterChecks                  ClusterChecks            `json:"cluster_checks"`                                          // cluster_checks
	Exporter                       Exporter                 `json:"exporter,inline"`                                         // telemetry
	OrchestratorExplorer           OrchestratorExplorer     `json:"orchestrator_explorer,inline"`                            // orchestrator_explorer
	Statsd                         Statsd                   `json:"statsd"`                                                  // statsd_*, dagstatsd_*
	Apm                            Apm                      `json:"apm,inline"`                                              // apm_config.*
	Jmx                            Jmx                      `json:"jmx"`                                                     // jmx_*
	RuntimeSecurity                RuntimeSecurity          `json:"runtime_security,inline"`                                 // runtime_security_config.*
	AdminssionController           AdminssionController     `json:"adminssion_controller"`                                   // admission_controller.*
	ExternalMetricsProvider        ExternalMetricsProvider  `json:"external_metrics_provider,inline"`                        // external_metrics_provider.*
	EnablePayloads                 EnablePayloads           `json:"enable_payloads,inline"`                                  // enable_payloads.*
	SystemProbe                    SystemProbe              `json:"system_probe,inline"`                                     // system_probe_config.*
	Listeners                      []Listeners              `json:"listeners,inline"`                                        // listeners
	ConfigProviders                []ConfigurationProviders `json:"config_providers,inline"`                                 // config_providers
	VerboseReport                  bool                     `json:"verbose_report" default:"true"`                           // collects run in verbose mode, e.g. report both with cpu.used(sys+user), cpu.system & cpu.user
	ApiKey                         string                   `json:"api_key"`                                                 // api_key
	Hostname                       string                   `json:"hostname" flag:"hostname" description:"custom host name"` //
	HostnameFQDN                   bool                     `json:"hostname_fqdn"`                                           // hostname_fqdn
	HostnameForceConfigAsCanonical bool                     `json:"hostname_force_config_as_canonical"`                      // hostname_force_config_as_canonical
	BindHost                       string                   `json:"bind_host"`                                               // bind_host
	IPCAddress                     string                   `json:"ipc_address" default:"localhost"`                         // ipc_address
	//CmdPort                          int               `json:"cmd_port" default:"5001"`                                                                                           // cmd_port, move to apiserver
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
	AggregatorBufferSize             int               `json:"aggregator_buffer_size" default:"100"`                                                                              // aggregator_buffer_size
	IotHost                          bool              `json:"iot_host"`                                                                                                          // iot_host
	HerokuDyno                       bool              `json:"heroku_dyno"`                                                                                                       // heroku_dyno
	BasicTelemetryAddContainerTags   bool              `json:"basic_telemetry_add_container_tags"`                                                                                // basic_telemetry_add_container_tags
	LogPayloads                      bool              `json:"log_payloads"`                                                                                                      // log_payloads
	AggregatorStopTimeout            time.Duration     `json:"-"`                                                                                                                 // aggregator_stop_timeout
	AggregatorStopTimeout_           int               `json:"aggregator_stop_timeout" flag:"aggregator-stop-timeout" default:"2" description:"aggregator stop timeout(Second)"`  // aggregator_stop_timeout
	AutoconfTemplateDir              string            `json:"autoconf_template_dir"`                                                                                             // autoconf_template_dir
	AutoconfTemplateUrlTimeout       bool              `json:"autoconf_template_url_timeout"`                                                                                     // autoconf_template_url_timeout
	CheckRunners                     int               `json:"check_runners" default:"4"`                                                                                         // check_runners
	LoggingFrequency                 int64             `json:"logging_frequency" default:"500"`                                                                                   // logging_frequency
	GUIPort                          bool              `json:"gui_port"`                                                                                                          // GUI_port
	XAwsEc2MetadataTokenTtlSeconds   bool              `json:"x_aws_ec2_metadata_token_ttl_seconds"`                                                                              // X-aws-ec2-metadata-token-ttl-seconds
	AcExclude                        bool              `json:"ac_exclude"`                                                                                                        // ac_exclude
	AcInclude                        bool              `json:"ac_include"`                                                                                                        // ac_include
	AllowArbitraryTags               bool              `json:"allow_arbitrary_tags"`                                                                                              // allow_arbitrary_tags
	AppKey                           bool              `json:"app_key"`                                                                                                           // app_key
	CCoreDump                        bool              `json:"c_core_dump"`                                                                                                       // c_core_dump
	CStacktraceCollection            bool              `json:"c_stacktrace_collection"`                                                                                           // c_stacktrace_collection
	CacheSyncTimeout                 time.Duration     `json:"-"`
	CacheSyncTimeout_                int               `json:"cache_sync_timeout" flag:"cache-sync-timeout" default:"2" description:"cache sync timeout(Second)"`                                // cache_sync_timeout
	ClcRunnerId                      string            `json:"clc_runner_id"`                                                                                                                    // clc_runner_id
	CmdHost                          string            `json:"cmd_host" default:"localhost"`                                                                                                     // cmd_host
	CollectKubernetesEvents          bool              `json:"collect_kubernetes_events"`                                                                                                        // collect_kubernetes_events
	ComplianceConfigDir              string            `json:"compliance_config_dir" default:"/opt/n9e/agentd/compliance.d"`                                                                     // compliance_config.dir
	ComplianceConfigEnabled          bool              `json:"compliance_config_enabled"`                                                                                                        // compliance_config.enabled
	ContainerCgroupPrefix            string            `json:"container_cgroup_prefix"`                                                                                                          // container_cgroup_prefix
	ContainerCgroupRoot              string            `json:"container_cgroup_root"`                                                                                                            // container_cgroup_root
	ContainerProcRoot                string            `json:"container_proc_root"`                                                                                                              // container_proc_root
	ContainerdNamespace              string            `json:"containerd_namespace" default:"k8s.io"`                                                                                            // containerd_namespace
	CriConnectionTimeout             time.Duration     `json:"-"`                                                                                                                                // cri_connection_timeout
	CriConnectionTimeout_            int               `json:"cri_connection_timeout" flag:"cri-connection-timeout" default:"1" description:"cri connection timeout(Second)"`                    // cri_connection_timeout
	CriQueryTimeout                  time.Duration     `json:"-"`                                                                                                                                // cri_query_timeout
	CriQueryTimeout_                 int               `json:"cri_query_timeout" flag:"cri-query-timeout" default:"5" description:"cri query timeout(Second)"`                                   // cri_query_timeout
	DatadogCluster                   bool              `json:"datadog_cluster"`                                                                                                                  // datadog-cluster
	DisableFileLogging               bool              `json:"disable_file_logging"`                                                                                                             // disable_file_logging
	DockerLabelsAsTags               bool              `json:"docker_labels_as_tags"`                                                                                                            // docker_labels_as_tags
	DockerQueryTimeout               time.Duration     `json:"-"`                                                                                                                                // docker_query_timeout
	DockerQueryTimeout_              int               `json:"docker_query_timeout" flag:"docker-query-timeout" default:"5" description:"docker query timeout(Second)"`                          // docker_query_timeout
	EC2PreferImdsv2                  bool              `json:"ec2_prefer_imdsv2"`                                                                                                                // ec2_prefer_imdsv2
	EC2MetadataTimeout               time.Duration     `json:"-"`                                                                                                                                // ec2_metadata_timeout
	EC2MetadataTimeout_              int               `json:"ec2_metadata_timeout" falg:"ec2-metadata-timeout" default:"300" description:"ec2 metadata timeout(Millisecond)"`                   // ec2_metadata_timeout
	EC2MetadataTokenLifetime         time.Duration     `json:"-"`                                                                                                                                // ec2_metadata_token_lifetime
	EC2MetadataTokenLifetime_        int               `json:"ec2_metadata_token_lifetime" falg:"ec2-metadata-token-lifetime" default:"21600" description:"ec2 metadata token lifetime(Second)"` // ec2_metadata_token_lifetime
	EC2UseWindowsPrefixDetection     bool              `json:"ec2_use_windows_prefix_detection"`                                                                                                 // ec2_use_windows_prefix_detection
	EcsAgentContainerName            string            `json:"ecs_agent_container_name" default:"ecs-agent"`                                                                                     // ecs_agent_container_name
	EcsAgentUrl                      bool              `json:"ecs_agent_url"`                                                                                                                    // ecs_agent_url
	EcsCollectResourceTagsEc2        bool              `json:"ecs_collect_resource_tags_ec2"`                                                                                                    // ecs_collect_resource_tags_ec2
	EKSFargate                       bool              `json:"eks_fargate"`                                                                                                                      // eks_fargate
	EnableMetadataCollection         bool              `json:"enable_metadata_collection" default:"true"`                                                                                        // enable_metadata_collection

	ExcludeGCETags []string `json:"exclude_gce_tags" default:"kube-env,kubelet-config,containerd-configure-sh,startup-script,shutdown-script,configure-sh,sshKeys,ssh-keys,user-data,cli-cert,ipsec-cert,ssl-cert,google-container-manifest,boshSettings,windows-startup-script-ps1,common-psm1,k8s-node-setup-psm1,serial-port-logging-enable,enable-oslogin,disable-address-manager,disable-legacy-endpoints,windows-keys,kubeconfig"` // exclude_gce_tags

	ExcludePauseContainer                bool              `json:"exclude_pause_container"`                         // exclude_pause_container
	ExternalMetricsAggregator            string            `json:"external_metrics_aggregator" default:"avg"`       // external_metrics.aggregator
	ExtraListeners                       []string          `json:"extra_listeners"`                                 // extra_listeners
	ForceTls12                           bool              `json:"force_tls_12"`                                    // force_tls_12
	FullSketches                         bool              `json:"full_sketches"`                                   // full-sketches
	GceSendProjectIdTag                  bool              `json:"gce_send_project_id_tag"`                         // gce_send_project_id_tag
	GoCoreDump                           bool              `json:"go_core_dump"`                                    // go_core_dump
	HpaConfigmapName                     string            `json:"hpa_configmap_name" default:"n9e-custom-metrics"` // hpa_configmap_name
	HpaWatcherGcPeriod                   time.Duration     `json:"-"`
	HpaWatcherGcPeriod_                  int               `json:"hpa_watcher_gc_period" flag:"hpa-watcher-gc-period" default:"300" description:"hpa_watcher_gcPeriod(Second)"` // hpa_watcher_gc_period
	IgnoreAutoconf                       []string          `json:"ignore_autoconf"`                                                                                             // ignore_autoconf
	InventoriesEnabled                   bool              `json:"inventories_enabled" default:"true"`                                                                          // inventories_enabled
	InventoriesMaxInterval               time.Duration     `json:"-"`
	InventoriesMaxInterval_              int               `json:"inventories_max_interval" flag:"inventories-max-interval" default:"600" description:"inventoriesMaxInterval(Second)"` // inventories_max_interval
	InventoriesMinInterval               time.Duration     `json:"-"`
	InventoriesMinInterval_              int               `json:"inventories_min_interval" flag:"inventories-min-interval" default:"300" description:"inventoriesMinInterval(Second)"` // inventories_min_interval
	KubeResourcesNamespace               bool              `json:"kube_resources_namespace"`                                                                                            // kube_resources_namespace
	KubeletCachePodsDuration             time.Duration     `json:"-"`
	KubeletCachePodsDuration_            int               `json:"kubelet_cache_pods_duration" flag:"kubelet-cache-pods-duration" default:"5" description:"kubeletCachePodsDuration(Second)"` // kubelet_cache_pods_duration
	KubeletClientCa                      string            `json:"kubelet_client_ca"`                                                                                                         // kubelet_client_ca
	KubeletClientCrt                     string            `json:"kubelet_client_crt"`                                                                                                        // kubelet_client_crt
	KubeletClientKey                     string            `json:"kubelet_client_key"`                                                                                                        // kubelet_client_key
	KubeletListenerPollingInterval       time.Duration     `json:"-"`
	KubeletListenerPollingInterval_      int               `json:"kubelet_listener_polling_interval" flag:"kubelet-listener-polling-interval" default:"5" description:"kubeletListenerPollingInterval(Second)"` // kubelet_listener_polling_interval
	KubeletTlsVerify                     bool              `json:"kubelet_tls_verify" default:"true"`                                                                                                           // kubelet_tls_verify
	KubeletWaitOnMissingContainer        time.Duration     `json:"-"`
	KubeletWaitOnMissingContainer_       int               `json:"kubelet_wait_on_missing_container" flag:"kubelet-wait-on-missing-container" description:"kubeletWaitOnMissingContainer(Second)"` // kubelet_wait_on_missing_container
	KubernetesApiserverClientTimeout     time.Duration     `json:"-"`
	KubernetesApiserverClientTimeout_    int               `json:"kubernetes_apiserver_client_timeout" flag:"kubernetes-apiserver-client-timeout" default:"10" description:"kubernetes_apiserverClientTimeout(Seconde)"` // kubernetes_apiserver_client_timeout
	KubernetesApiserverUseProtobuf       bool              `json:"kubernetes_apiserver_use_protobuf"`                                                                                                                    // kubernetes_apiserver_use_protobuf
	KubernetesCollectMetadataTags        bool              `json:"kubernetes_collect_metadata_tags" default:"true"`                                                                                                      // kubernetes_collect_metadata_tags
	KubernetesCollectServiceTags         bool              `json:"kubernetes_collect_service_tags"`                                                                                                                      // kubernetes_collect_service_tags
	KubernetesHttpKubeletPort            int               `json:"kubernetes_http_kubelet_port" default:"10255"`                                                                                                         // kubernetes_http_kubelet_port
	KubernetesHttpsKubeletPort           int               `json:"kubernetes_https_kubelet_port" default:"10250"`                                                                                                        // kubernetes_https_kubelet_port
	KubernetesInformersResyncPeriod      time.Duration     `json:"-"`
	KubernetesInformersResyncPeriod_     int               `json:"kubernetes_informers_resync_period" flag:"kubernetes-informers-resync-period" default:"300" description:"kubernetesInformersResyncPeriod(Second)"` // kubernetes_informers_resync_period
	KubernetesKubeletHost                string            `json:"kubernetes_kubelet_host"`                                                                                                                          // kubernetes_kubelet_host
	KubernetesKubeletNodename            string            `json:"kubernetes_kubelet_nodename"`                                                                                                                      // kubernetes_kubelet_nodename
	KubernetesMapServicesOnIp            bool              `json:"kubernetes_map_services_on_ip"`                                                                                                                    // kubernetes_map_services_on_ip
	KubernetesMetadataTagUpdateFreq      time.Duration     `json:"-"`
	KubernetesMetadataTagUpdateFreq_     int               `json:"kubernetes_metadata_tag_update_freq" flag:"kubernetes-metadata-tag-update-freq" default:"60" description:"kubernetesMetadataTagUpdateFreq(Second)"` // kubernetes_metadata_tag_update_freq
	KubernetesNamespaceLabelsAsTags      bool              `json:"kubernetes_namespace_labels_as_tags"`                                                                                                               // kubernetes_namespace_labels_as_tags
	KubernetesNodeLabelsAsTags           bool              `json:"kubernetes_node_labels_as_tags"`                                                                                                                    // kubernetes_node_labels_as_tags
	KubernetesPodAnnotationsAsTags       map[string]string `json:"kubernetes_pod_annotations_as_tags"`                                                                                                                // kubernetes_pod_annotations_as_tags
	KubernetesPodExpirationDuration      time.Duration     `json:"-"`
	KubernetesPodExpirationDuration_     int               `json:"kubernetes_pod_expiration_duration" flag:"kubernetes-pod-expiration-duration" default:"900" description:"kubernetes_podExpirationDuration(Second)"` // kubernetes_pod_expiration_duration
	KubernetesPodLabelsAsTags            map[string]string `json:"kubernetes_pod_labels_as_tags"`                                                                                                                     // kubernetes_pod_labels_as_tags
	KubernetesServiceTagUpdateFreq       map[string]string `json:"kubernetes_service_tag_update_freq"`                                                                                                                // kubernetes_service_tag_update_freq
	LeaderElection                       bool              `json:"leader_election"`                                                                                                                                   // leader_election
	LeaderLeaseDuration                  time.Duration     `json:"-"`
	LeaderLeaseDuration_                 int               `json:"leader_lease_duration" flag:"leader-lease-duration" default:"60" description:"leader lease duration(second)"` // leader_lease_duration
	LogEnabled                           bool              `json:"log_enabled"`                                                                                                 // log_enabled
	LogFile                              string            `json:"log_file"`                                                                                                    // log_file
	LogFormatJson                        bool              `json:"log_format_json"`                                                                                             // log_format_json
	LogFormatRfc3339                     bool              `json:"log_format_rfc3339"`                                                                                          // log_format_rfc3339
	LogLevel                             string            `json:"log_level"`                                                                                                   // log_level
	LogToConsole                         bool              `json:"log_to_console"`                                                                                              // log_to_console
	MemtrackEnabled                      bool              `json:"memtrack_enabled"`                                                                                            // memtrack_enabled
	MetricsPort                          int               `json:"metrics_port" default:"5000"`                                                                                 // metrics_port
	ProcRoot                             string            `json:"proc_root" default:"/proc"`                                                                                   // proc_root
	ProcessAgentConfigHostIps            bool              `json:"process_agent_config_host_ips"`                                                                               // process_agent_config.host_ips
	ProcessConfigEnabled                 bool              `json:"process_config_enabled"`                                                                                      // process_config.enabled
	ProfilingEnabled                     bool              `json:"profiling_enabled"`                                                                                           // profiling.enabled
	ProfilingProfileDdUrl                bool              `json:"profiling_profile_dd_url"`                                                                                    // profiling.profile_dd_url
	PrometheusScrapeEnabled              bool              `json:"prometheus_scrape_enabled"`                                                                                   // prometheus_scrape.enabled
	PrometheusScrapeServiceEndpoints     bool              `json:"prometheus_scrape_service_endpoints"`                                                                         // prometheus_scrape.service_endpoints
	ProxyHttps                           bool              `json:"proxy_https"`                                                                                                 // proxy.https
	ProxyNoProxy                         bool              `json:"proxy_no_proxy"`                                                                                              // proxy.no_proxy
	Python3LinterTimeout                 time.Duration     `json:"-"`
	Python3LinterTimeout_                int               `json:"python3_linter_timeout" flag:"python3-linter-timeout" default:"120" description:"python3LinterTimeout(Second)"` // python3_linter_timeout
	PythonVersion                        string            `json:"python_version"`                                                                                                // python_version
	SerializerMaxPayloadSize             int               `json:"serializer_max_payload_size" default:"2621440" description:"2.5mb"`                                             // serializer_max_payload_size
	SerializerMaxUncompressedPayloadSize int               `json:"serializer_max_uncompressed_payload_size" default:"4194304" description:"4mb"`                                  // serializer_max_uncompressed_payload_size
	//ServerTimeout                        time.Duration     `json:"-"`
	//ServerTimeout_                       int               `json:"server_timeout" flag:"server-timeout" default:"15" description:"server_timeout(Second)"` // server_timeout, move to apiserver
	SkipSslValidation   bool   `json:"skip_ssl_validation"`              // skip_ssl_validation
	SyslogRfc           bool   `json:"syslog_rfc"`                       // syslog_rfc
	TelemetryEnabled    bool   `json:"telemetry_enabled" default:"true"` // telemetry.enabled
	TracemallocDebug    bool   `json:"tracemalloc_debug"`                // tracemalloc_debug
	Yaml                bool   `json:"yaml"`                             // yaml
	MetricTransformFile string `json:"metric_transform_file"`
	//N9e                                  N9e                      `json:"n9e"`
}

func (p Config) String() string {
	return util.Prettify(p)
}

func (p *Config) prepareTimeDuration() error {
	p.ECSMetadataTimeout = time.Millisecond * time.Duration(p.ECSMetadataTimeout_)
	p.GCEMetadataTimeout = time.Millisecond * time.Duration(p.GCEMetadataTimeout_)
	p.AdConfigPollInterval = time.Second * time.Duration(p.AdConfigPollInterval_)
	p.AggregatorStopTimeout = time.Second * time.Duration(p.AggregatorStopTimeout_)
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

// Validate will be called by configer.Read()
func (p *Config) Validate() error {
	if err := p.prepareTimeDuration(); err != nil {
		return err
	}

	if len(p.Endpoints) == 0 {
		return fmt.Errorf("unable to get agent.endpoints")
	}

	for i, endpoint := range p.Endpoints {
		if _, err := url.Parse(endpoint); err != nil {
			return fmt.Errorf("could not parse agent.endpoint[%d]: %s %s", i, endpoint, err)
		}
	}

	for i := range p.MetadataProviders {
		if err := p.MetadataProviders[i].Validate(); err != nil {
			return err
		}
	}

	if err := p.Forwarder.Validate(); err != nil {
		return err
	}
	if err := p.PrometheusScrape.Validate(); err != nil {
		return err
	}
	if err := p.Autoconfig.Validate(); err != nil {
		return err
	}
	if err := p.Container.Validate(); err != nil {
		return err
	}
	if err := p.SnmpTraps.Validate(p.GetBindHost()); err != nil {
		return err
	}
	if err := p.ClusterAgent.Validate(); err != nil {
		return err
	}
	if err := p.Network.Validate(); err != nil {
		return err
	}
	if err := p.SnmpListener.Validate(); err != nil {
		return err
	}
	if err := p.LogsConfig.Validate(); err != nil {
		return err
	}
	if err := p.ClusterChecks.Validate(); err != nil {
		return err
	}
	if err := p.Exporter.Validate(); err != nil {
		return err
	}
	if err := p.OrchestratorExplorer.Validate(); err != nil {
		return err
	}
	if err := p.Statsd.Validate(); err != nil {
		return err
	}
	if err := p.Apm.Validate(); err != nil {
		return err
	}
	if err := p.Jmx.Validate(); err != nil {
		return err
	}
	if err := p.RuntimeSecurity.Validate(); err != nil {
		return err
	}
	if err := p.AdminssionController.Validate(); err != nil {
		return err
	}
	if err := p.ExternalMetricsProvider.Validate(); err != nil {
		return err
	}
	if err := p.EnablePayloads.Validate(); err != nil {
		return err
	}
	if err := p.SystemProbe.Validate(); err != nil {
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

	return nil
}

func (p *Config) ValidatePath() error {
	if p.WorkDir == "" {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return err
		}

		// strip bin/
		klog.Infof("dir %s %s", dir, filepath.Base(dir))
		if filepath.Base(dir) == "bin" {
			dir = filepath.Dir(dir)
		}

		p.WorkDir = dir
	}
	if !IsDir(p.WorkDir) {
		return fmt.Errorf("agent.workDir %s does not exist, please create it", p.WorkDir)
	}

	klog.Infof("Chdir to ${workdir} %s", p.WorkDir)
	os.Chdir(p.WorkDir)

	// {prefix}/conf.d
	if p.ConfdPath == "" {
		p.ConfdPath = filepath.Join(p.WorkDir, "conf.d")
	}
	if !IsDir(p.ConfdPath) {
		klog.Warningf("agent.confdPath %s does not exist, please create it", p.ConfdPath)
	}
	klog.Infof("agent.confdPath %s", p.ConfdPath)

	// {prefix}/checks.d
	if p.AdditionalChecksd == "" {
		p.AdditionalChecksd = filepath.Join(p.WorkDir, "checks.d")
	}
	//if !IsDir(p.AdditionalChecksd) {
	//	return fmt.Errorf("agent.additionalChecks %s does not exist, please create it", p.AdditionalChecksd)
	//}
	klog.Infof("agent.additionalChecks %s", p.AdditionalChecksd)

	// {prefix}/run
	if p.RunPath == "" {
		p.RunPath = filepath.Join(p.WorkDir, "run")
	}
	if !IsDir(p.RunPath) {
		return fmt.Errorf("agent.runPath %s does not exist, please create it", p.RunPath)
	}

	// LogsConfig
	if p.LogsConfig.RunPath == "" {
		p.LogsConfig.RunPath = p.RunPath
	}
	klog.Infof("agent.runPath %s", p.RunPath)

	// {prefix}/run/transactions_to_retry
	if p.Forwarder.StoragePath == "" {
		p.Forwarder.StoragePath = filepath.Join(p.RunPath, "transactions_to_retry")
	}
	klog.Infof("agent.forwarder.storagePath %s", p.Forwarder.StoragePath)

	p.DistPath = filepath.Join(p.WorkDir, "dist")
	p.PyChecksPath = filepath.Join(p.WorkDir, "checks.d")

	// {prefix}/{name}.sock

	return nil
}

func configEval(value string) string {
	switch strings.ToLower(value) {
	case "$ip":
		return getOutboundIP()
	case "$host", "$hostname":
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

//type N9e struct {
//	Enabled  bool   `json:"enabled"`
//	Endpoint string `json:"endpoint"`
//	V5Format bool   `json:"v5_format"`
//}

type SystemProbe struct {
	Enabled                      bool                `json:"enabled"`                         // system_probe_config.enabled & system_probe
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
}

func (p *SystemProbe) Validate() error {
	p.DnsTimeout = time.Second * time.Duration(p.DnsTimeout_)
	return nil
}

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
	Socket                             string `json:"socket"`                                                        // runtime_security_config.socket
	AgentMonitoringEvents              bool   `json:"agent_monitoring_events"`                                       // runtime_security_config.agent_monitoring_events
	CookieCacheSize                    bool   `json:"cookie_cache_size"`                                             // runtime_security_config.cookie_cache_size
	CustomSensitiveWords               bool   `json:"custom_sensitive_words"`                                        // runtime_security_config.custom_sensitive_words
	EnableApprovers                    bool   `json:"enable_approvers"`                                              // runtime_security_config.enable_approvers
	EnableDiscarders                   bool   `json:"enable_discarders"`                                             // runtime_security_config.enable_discarders
	EnableKernelFilters                bool   `json:"enable_kernel_filters"`                                         // runtime_security_config.enable_kernel_filters
	Enabled                            bool   `json:"enabled"`                                                       // runtime_security_config.enabled
	EventServerBurst                   bool   `json:"event_server_burst"`                                            // runtime_security_config.event_server.burst
	EventServerRate                    bool   `json:"event_server_rate"`                                             // runtime_security_config.event_server.rate
	EventsStatsPollingInterval         bool   `json:"events_stats_polling_interval"`                                 // runtime_security_config.events_stats.polling_interval
	FimEnabled                         bool   `json:"fim_enabled"`                                                   // runtime_security_config.fim_enabled
	FlushDiscarderWindow               bool   `json:"flush_discarder_window"`                                        // runtime_security_config.flush_discarder_window
	LoadControllerControlPeriod        bool   `json:"load_controller_control_period"`                                // runtime_security_config.load_controller.control_period
	LoadControllerDiscarderTimeout     bool   `json:"load_controller_discarder_timeout"`                             // runtime_security_config.load_controller.discarder_timeout
	LoadControllerEventsCountThreshold bool   `json:"load_controller_events_count_threshold"`                        // runtime_security_config.load_controller.events_count_threshold
	PidCacheSize                       bool   `json:"pid_cache_size"`                                                // runtime_security_config.pid_cache_size
	PoliciesDir                        string `json:"policies_dir" default:"/opt/n9e/agentd/etc/runtime-security.d"` // runtime_security_config.policies.dir
	SyscallMonitorEnabled              bool   `json:"syscall_monitor_enabled"`                                       // runtime_security_config.syscall_monitor.enabled
}

func (p *RuntimeSecurity) Validate() error {
	return nil
}

type Jmx struct {
	CheckPeriod                int           `json:"check_period"` // jmx_check_period
	CollectionTimeout          time.Duration `json:"-"`
	CollectionTimeout_         int           `json:"collection_timeout" flag:"jmx-collection-timeout" default:"60" description:"collectionTimeout"` // jmx_collection_timeout
	CustomJars                 []string      `json:"custom_jars"`                                                                                   // jmx_custom_jars
	LogFile                    bool          `json:"log_file"`                                                                                      // jmx_log_file
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

type Apm struct {
	AdditionalEndpoints           bool   `json:"additional_endpoints"`             // apm_config.additional_endpoints
	AnalyzedRateByService         bool   `json:"analyzed_rate_by_service"`         // apm_config.analyzed_rate_by_service
	AnalyzedSpans                 bool   `json:"analyzed_spans"`                   // apm_config.analyzed_spans
	ApiKey                        bool   `json:"api_key"`                          // apm_config.api_key
	ApmDdUrl                      bool   `json:"apm_dd_url"`                       // apm_config.apm_dd_url
	ApmNonLocalTraffic            bool   `json:"apm_non_local_traffic"`            // apm_config.apm_non_local_traffic
	ConnectionLimit               bool   `json:"connection_limit"`                 // apm_config.connection_limit
	ConnectionResetInterval       bool   `json:"connection_reset_interval"`        // apm_config.connection_reset_interval
	DdAgentBin                    bool   `json:"dd_agent_bin"`                     // apm_config.dd_agent_bin
	Enabled                       bool   `json:"enabled"`                          // apm_config.enabled
	Env                           bool   `json:"env"`                              // apm_config.env
	ExtraSampleRate               bool   `json:"extra_sample_rate"`                // apm_config.extra_sample_rate
	FilterTagsReject              bool   `json:"filter_tags_reject"`               // apm_config.filter_tags.reject
	FilterTagsRequire             bool   `json:"filter_tags_require"`              // apm_config.filter_tags.require
	IgnoreResources               bool   `json:"ignore_resources"`                 // apm_config.ignore_resources
	LogFile                       bool   `json:"log_file"`                         // apm_config.log_file
	LogLevel                      bool   `json:"log_level"`                        // apm_config.log_level
	LogThrottling                 bool   `json:"log_throttling"`                   // apm_config.log_throttling
	MaxCpuPercent                 bool   `json:"max_cpu_percent"`                  // apm_config.max_cpu_percent
	MaxEventsPerSecond            bool   `json:"max_events_per_second"`            // apm_config.max_events_per_second
	MaxMemory                     bool   `json:"max_memory"`                       // apm_config.max_memory
	MaxTracesPerSecond            bool   `json:"max_traces_per_second"`            // apm_config.max_traces_per_second
	Obfuscation                   bool   `json:"obfuscation"`                      // apm_config.obfuscation
	ProfilingAdditionalEndpoints  bool   `json:"profiling_additional_endpoints"`   // apm_config.profiling_additional_endpoints
	ProfilingDdUrl                bool   `json:"profiling_dd_url"`                 // apm_config.profiling_dd_url
	ReceiverPort                  string `json:"receiver_port"`                    // apm_config.receiver_port
	ReceiverSocket                bool   `json:"receiver_socket"`                  // apm_config.receiver_socket
	ReceiverTimeout               bool   `json:"receiver_timeout"`                 // apm_config.receiver_timeout
	RemoteTagger                  bool   `json:"remote_tagger"`                    // apm_config.remote_tagger
	SyncFlushing                  bool   `json:"sync_flushing"`                    // apm_config.sync_flushing
	WindowsPipeBufferSize         bool   `json:"windows_pipe_buffer_size"`         // apm_config.windows_pipe_buffer_size
	WindowsPipeName               bool   `json:"windows_pipe_name"`                // apm_config.windows_pipe_name
	WindowsPipeSecurityDescriptor bool   `json:"windows_pipe_security_descriptor"` // apm_config.windows_pipe_security_descriptor
}

func (p *Apm) Validate() error {
	return nil
}

type Statsd struct {
	Enabled                           bool             `json:"enabled" default:"false"` // use_dogstatsd
	Host                              string           `json:"host"`                    //
	Port                              int              `json:"port" default:"8125"`     // dogstatsd_port
	Socket                            string           `json:"socket"`                  // dogstatsd_socket
	PipeName                          string           `json:"pipe_name"`               // dogstatsd_pipe_name
	ContextExpirySeconds              time.Duration    `json:"-"`
	ContextExpirySeconds_             int              `json:"context_expiry_seconds" flag:"statsd-context-expiry-seconds" default:"300" description:"contextExpirySeconds(Second)"` // dogstatsd_context_expiry_seconds
	ExpirySeconds                     time.Duration    `json:"-"`
	ExpirySeconds_                    int              `json:"expiry_seconds" flag:"statsd-expiry-seconds" default:"300" description:"expirySeconds(Second)"` // dogstatsd_expiry_seconds
	StatsEnable                       bool             `json:"stats_enable" default:"true"`                                                                   // dogstatsd_stats_enable
	StatsBuffer                       int              `json:"stats_buffer" default:"10"`                                                                     // dogstatsd_stats_buffer
	MetricsStatsEnable                bool             `json:"metrics_stats_enable" default:"false"`                                                          // dogstatsd_metrics_stats_enable - for debug
	BufferSize                        int              `json:"buffer_size" default:"8192"`                                                                    // dogstatsd_buffer_size
	MetricNamespace                   string           `json:"metric_namespace"`                                                                              // statsd_metric_namespace
	MetricNamespaceBlacklist          []string         `json:"metric_namespace_blacklist"`                                                                    // statsd_metric_namespace_blacklist
	Tags                              []string         `json:"tags"`                                                                                          // dogstatsd_tags
	EntityIdPrecedence                bool             `json:"entity_id_precedence"`                                                                          // dogstatsd_entity_id_precedence
	EolRequired                       []string         `json:"eol_required"`                                                                                  // dogstatsd_eol_required
	DisableVerboseLogs                bool             `json:"disable_verbose_logs"`                                                                          // dogstatsd_disable_verbose_logs
	ForwardHost                       string           `json:"forward_host"`                                                                                  // statsd_forward_host
	ForwardPort                       int              `json:"forward_port"`                                                                                  // statsd_forward_port
	QueueSize                         int              `json:"queue_size" default:"1024"`                                                                     // dogstatsd_queue_size
	MapperCacheSize                   int              `json:"mapper_cache_size" default:"1000"`                                                              // dogstatsd_mapper_cache_size
	MapperProfiles                    []MappingProfile `json:"mapper_profiles"`                                                                               // dogstatsd_mapper_profiles
	StringInternerSize                int              `json:"string_interner_size" default:"4096"`                                                           // dogstatsd_string_interner_size
	SocketRcvbuf                      int              `json:"socekt_rcvbuf"`                                                                                 // dogstatsd_so_rcvbuf
	PacketBufferSize                  int              `json:"packet_buffer_size" default:"32"`                                                               // dogstatsd_packet_buffer_size
	PacketBufferFlushTimeout          time.Duration    `json:"-"`
	PacketBufferFlushTimeout_         int              `json:"packet_buffer_flush_timeout" flag:"statsd-packet-buffer-flush-timeout" default:"100" description:"packetBufferFlushTimeout(Millisecond)"` // dogstatsd_packet_buffer_flush_timeout
	TagCardinality                    string           `json:"tag_cardinality" default:"low"`                                                                                                           // dogstatsd_tag_cardinality
	NonLocalTraffic                   bool             `json:"non_local_traffic"`                                                                                                                       // dogstatsd_non_local_traffic
	OriginDetection                   bool             `json:"OriginDetection"`                                                                                                                         // dogstatsd_origin_detection
	HistogramCopyToDistribution       bool             `json:"histogram_copy_to_distribution"`                                                                                                          // histogram_copy_to_distribution
	HistogramCopyToDistributionPrefix string           `json:"histogram_copy_to_distribution_prefix"`                                                                                                   // histogram_copy_to_distribution_prefix
}

func (p *Statsd) Validate() error {
	p.ContextExpirySeconds = time.Second * time.Duration(p.ContextExpirySeconds_)
	p.ExpirySeconds = time.Second * time.Duration(p.ExpirySeconds_)
	p.PacketBufferFlushTimeout = time.Millisecond * time.Duration(p.PacketBufferFlushTimeout_)
	return nil
}

type AdditionalEndpoint struct {
	Endpoints []string `json:"endpoints"`
	ApiKeys   []string `json:"api_keys"`
}

type Forwarder struct {
	AdditionalEndpoints       []AdditionalEndpoint `json:"additional_endpoints"` // additional_endpoints
	ApikeyValidationInterval  time.Duration        `json:"-"`
	ApikeyValidationInterval_ int                  `json:"apikey_validation_interval" flag:"forwarder-apikey-validation-interval" default:"3600" description:"apikeyValidationInterval(Second)"` // forwarder_apikey_validation_interval
	BackoffBase               float64              `json:"backoff_base" default:"2"`                                                                                                             // forwarder_backoff_base
	BackoffFactor             float64              `json:"backoff_factor" default:"2"`                                                                                                           // forwarder_backoff_factor
	BackoffMax                float64              `json:"backoff_max" default:"64"`                                                                                                             // forwarder_backoff_max
	ConnectionResetInterval   time.Duration        `json:"-"`
	ConnectionResetInterval_  int                  `json:"connection_reset_interval" flag:"forwarder-connection-reset-interval" description:"connectionResetInterval(Second)"` // forwarder_connection_reset_interval
	FlushToDiskMemRatio       float64              `json:"flush_to_disk_mem_ratio" default:"0.5"`                                                                              // forwarder_flush_to_disk_mem_ratio
	NumWorkers                int                  `json:"num_workers" default:"1"`                                                                                            // forwarder_num_workers
	OutdatedFileInDays        int                  `json:"outdated_file_in_days" default:"10"`                                                                                 // forwarder_outdated_file_in_days
	RecoveryInterval          int                  `json:"recovery_interval" default:"2"`                                                                                      // forwarder_recovery_interval
	RecoveryReset             bool                 `json:"recovery_reset"`                                                                                                     // forwarder_recovery_reset
	StopTimeout               time.Duration        `json:"-"`
	StopTimeout_              int                  `json:"stop_timeout" flag:"forwarder-stop-timeout" default:"2" description:"stopTimeout(Second)"` // forwarder_stop_timeout
	StorageMaxDiskRatio       float64              `json:"storage_max_disk_ratio" default:"0.95"`                                                    // forwarder_storage_max_disk_ratio
	StorageMaxSizeInBytes     int64                `json:"storage_max_size_in_bytes"`                                                                // forwarder_storage_max_size_in_bytes
	StoragePath               string               `json:"storage_path"`                                                                             // forwarder_storage_path
	Timeout                   time.Duration        `json:"-"`
	Timeout_                  int                  `json:"timeout" flag:"forwarder-timeout" default:"20" description:"timeout(Second)"` // forwarder_timeout
	RetryQueuePayloadsMaxSize int                  `json:"retry_queue_payloads_max_size" default:"15728640" description:"15m"`          // forwarder_retry_queue_payloads_max_size
	//RetryQueueMaxSize         int           `json:"retry_queue_max_size"`         // forwarder_retry_queue_max_size

}

func (p *Forwarder) Validate() error {
	p.ApikeyValidationInterval = time.Second * time.Duration(p.ApikeyValidationInterval_)
	p.ConnectionResetInterval = time.Second * time.Duration(p.ConnectionResetInterval_)
	p.StopTimeout = time.Second * time.Duration(p.StopTimeout_)
	p.Timeout = time.Second * time.Duration(p.Timeout_)
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

type LogsConfig struct {
	Enabled                     bool                        `json:"enabled"`               // logs_enabled
	AdditionalEndpoints         []logstypes.Endpoint        `json:"additional_endpoints"`  // logs_config.additional_endpoints
	ContainerCollectAll         bool                        `json:"container_collect_all"` // logs_config.container_collect_all
	ProcessingRules             []*logstypes.ProcessingRule `json:"processing_rules"`      // logs_config.processing_rules
	APIKey                      string                      `json:"api_key"`               // logs_config.api_key
	DevModeNoSSL                bool                        `json:"dev_mode_no_ssl"`       // logs_config.dev_mode_no_ssl
	ExpectedTagsDuration        time.Duration               `json:"-"`
	ExpectedTagsDuration_       int                         `json:"expected_tags_duration" flag:"logs-expected-tags-duration" description:"expectedTagsDuration(Second)"` // logs_config.expected_tags_duration
	Socks5ProxyAddress          string                      `json:"socks5_proxy_address"`                                                                                 // logs_config.socks5_proxy_address
	UseTCP                      bool                        `json:"use_tcp"`                                                                                              // logs_config.use_tcp
	UseHTTP                     bool                        `json:"use_http"`                                                                                             // logs_config.use_http
	DevModeUseProto             bool                        `json:"dev_mode_use_proto" default:"true"`                                                                    // logs_config.dev_mode_use_proto
	ConnectionResetInterval     time.Duration               `json:"-"`
	ConnectionResetInterval_    int                         `json:"connection_reset_interval" flag:"logs-connection-reset-interval" default:"" description:"connectionResetInterval(Second)"` // logs_config.connection_reset_interval
	LogsUrl                     string                      `json:"logs_url"`                                                                                                                 // logs_config.logs_dd_url, dd_url
	UsePort443                  bool                        `json:"use_port443"`                                                                                                              // logs_config.use_port_443
	UseSSL                      bool                        `json:"use_ssl"`                                                                                                                  // !logs_config.logs_no_ssl
	Url443                      string                      `json:"url_443"`                                                                                                                  // logs_config.dd_url_443
	UseCompression              bool                        `json:"use_compression" default:"true"`                                                                                           // logs_config.use_compression
	CompressionLevel            int                         `json:"compression_level" default:"6"`                                                                                            // logs_config.compression_level
	URL                         string                      `json:"url" default:"localhost:8080"`                                                                                             // logs_config.dd_url (e.g. localhost:8080)
	BatchWait                   time.Duration               `json:"-"`
	BatchWait_                  int                         `json:"batch_wait" flag:"logs-batch-wait" default:"5" description:"batchWait(Second)"` // logs_config.batch_wait
	BatchMaxConcurrentSend      int                         `json:"batch_max_concurrent_send"`                                                     // logs_config.batch_max_concurrent_send
	TaggerWarmupDuration        time.Duration               `json:"-"`
	TaggerWarmupDuration_       int                         `json:"tagger_warmup_duration" flag:"logs-tagger-warmup-duration" description:"taggerWarmupDuration(Second)"` // logs_config.tagger_warmup_duration
	AggregationTimeout          time.Duration               `json:"-"`
	AggregationTimeout_         int                         `json:"aggregation_timeout" flag:"logs-aggregation-timeout" default:"1000" description:"aggregationTimeout(Millisecond)"` // logs_config.aggregation_timeout
	CloseTimeout                time.Duration               `json:"-"`
	CloseTimeout_               int                         `json:"close_timeout" flag:"logs-close-timeout" default:"60" description:"closeTimeout(Second)"` // logs_config.close_timeout
	AuditorTTL                  time.Duration               `json:"-"`
	AuditorTTL_                 int                         `json:"auditor_ttl" flag:"logs-auditor-ttl" description:"auditorTTL(Second)"` // logs_config.auditor_ttl
	RunPath                     string                      `json:"run_path"`                                                             // logs_config.run_path
	OpenFilesLimit              int                         `json:"open_files_limit" flag:"logs-open-files-limit" default:"100"`          // logs_config.open_files_limit
	K8SContainerUseFile         bool                        `json:"k8s_container_use_file"`                                               // logs_config.k8s_container_use_file
	DockerContainerUseFile      bool                        `json:"docker_container_use_file"`                                            // logs_config.docker_container_use_file
	DockerContainerForceUseFile bool                        `json:"docker_container_force_use_file"`                                      // logs_config.docker_container_force_use_file
	DockerClientReadTimeout     time.Duration               `json:"-"`
	DockerClientReadTimeout_    int                         `json:"docker_client_read_timeout" flag:"logs-docker-client-read-timeout" default:"30" description:"dockerClientReadTimeout(Second)"` // logs_config.docker_client_read_timeout
	FrameSize                   int                         `json:"frame_size" default:"9000"`                                                                                                    // logs_config.frame_size
	StopGracePeriod             time.Duration               `json:"-"`
	StopGracePeriod_            int                         `json:"stop_grace_period" flag:"logs-stop-grace-period" default:"30" description:"stopGracePeriod(Second)"` // logs_config.stop_grace_period
}

func (p *LogsConfig) Validate() error {
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
		klog.Warningf("Invalid batchWait: %v should be in [1, 10], fallback on %v",
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

type Network struct {
	ID                         string `json:"id"`                            // network.id
	EnableHttpMonitoring       bool   `json:"enable_http_monitoring"`        // network_config.enable_http_monitoring
	IgnoreConntrackInitFailure bool   `json:"ignore_conntrack_init_failure"` // network_config.ignore_conntrack_init_failure
	EnableGatewayLookup        bool   `json:"enable_gateway_lookup"`         // network_config.enable_gateway_lookup
}

func (p *Network) Validate() error {
	return nil
}

type Exporter struct {
	Port    int      `json:"port" default:"8070"`    // expvar_port
	Docs    bool     `json:"docs" default:"true"`    // /docs/*
	Metrics bool     `json:"metrics" default:"true"` // /metrics/
	Expvar  bool     `json:"expvar" default:"true"`  // /vars
	Pprof   bool     `json:"pprof" default:"true"`   // /debug/pprof
	Checks  []string `json:"checks"`                 // telemetry.checks
}

func (p *Exporter) Validate() error {
	return nil
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

type OrchestratorExplorer struct { // orchestrator_explorer
	Url                       string   `json:"url"`                         // orchestrator_explorer.orchestrator_dd_url
	AdditionalEndpoints       []string `json:"additional_endpoints"`        // orchestrator_explorer.orchestrator_additional_endpoints
	CustomSensitiveWords      []string `json:"custom_sensitive_words"`      // orchestrator_explorer.custom_sensitive_words
	ContainerScrubbingEnabled bool     `json:"container_scrubbing_enabled"` // orchestrator_explorer.container_scrubbing.enabled
	Enabled                   bool     `json:"enabled" default:"true"`      // orchestrator_explorer.enabled
	ExtraTags                 []string `json:"extra_tags"`                  // orchestrator_explorer.extra_tags
}

func (p *OrchestratorExplorer) Validate() error {
	return nil
}

type Autoconfig struct {
	Enabled         bool       `json:"enabled" default:"true"` // autoconfig_from_environment
	ExcludeFeatures []string   `json:"exclude_features"`       // autoconfig_exclude_features
	features        FeatureMap `json:"-"`
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
