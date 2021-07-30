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
	traceconfig "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/config"
	"github.com/yubo/golib/configer"
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
	//
)

var (
	Context context.Context

	// Deprecated
	Configfile string
	TestConfig bool

	// ungly hack, TODO: remove it
	C        = NewDefaultConfig()
	Configer *configer.Configer
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
	WorkDir           string `json:"workDir" flag:"workdir,w" env:"AGENTD_WORK_DIR" description:"workdir"` // e.g. /opt/n9e/agentd
	PidfilePath       string `json:"pidfilePath"`                                                          //
	AdditionalChecksd string `json:"additionalChecksd"`                                                    // additional_checksd
	//AuthTokenFilePath              string `json:"authTokenFilePath"`                                    // auth_token_file_path // move to apiserver
	RunPath                  string `json:"runPath"`                  // run_path
	ConfdPath                string `json:"confdPath"`                // confd_path
	CriSocketPath            string `json:"criSocketPath"`            // cri_socket_path
	KubeletAuthTokenPath     string `json:"kubeletAuthTokenPath"`     // kubelet_auth_token_path
	KubernetesKubeconfigPath string `json:"kubernetesKubeconfigPath"` // kubernetes_kubeconfig_path
	ProcfsPath               string `json:"procfsPath"`               // procfs_path
	WindowsUsePythonpath     string `json:"windowsUsePythonpath"`     // windows_use_pythonpath
	DistPath                 string `json:"-"`                        // {workdir}/dist
	PyChecksPath             string `json:"-"`                        // {workdir}/checks.d

	Ident             string   `json:"ident" flag:"ident" default:"$ip" description:"Ident of this host"`
	Alias             string   `json:"alias" flag:"alias" default:"$hostname" description:"Alias of the host"`
	Lang              string   `json:"lang" flag:"lang" default:"zh" description:"Default lang(zh, en)"`
	EnableN9eProvider bool     `json:"enableN9eProvider" flag:"enable-n9e-provider" default:"true" description:"enable n9e server api as autodiscovery provider"`
	N9eSeriesFormat   bool     `json:"n9eSeriesFormat" default:"true"`                                                                              // the payload format for forwarder
	Endpoints         []string `json:"endpoints" flag:"endpoints" default:"http://localhost:8000"  description:"endpoints addresses of n9e server"` // site, dd_url

	MetadataProviders              []MetadataProviders      `json:"metadataProviders"`                                       // metadata_providers
	Forwarder                      Forwarder                `json:"forwarder,inline"`                                        // fowarder_*
	PrometheusScrape               PrometheusScrape         `json:"prometheusScrape,inline"`                                 // prometheus_scrape
	Autoconfig                     Autoconfig               `json:"autoconfig,inline"`                                       //
	Container                      Container                `json:"container,inline"`                                        //
	SnmpTraps                      traps.Config             `json:"snmpTraps,inline"`                                        // snmp_traps_config
	ClusterAgent                   ClusterAgent             `json:"clusterAgent,inline"`                                     // cluster_agent
	Network                        Network                  `json:"network,inline"`                                          // network
	SnmpListener                   snmptypes.ListenerConfig `json:"snmpListener,inline"`                                     // snmp_listener
	Cmd                            Cmd                      `json:"cmd,inline"`                                              // cmd
	LogsConfig                     LogsConfig               `json:"logsConfig"`                                              // logs_config
	CloudFoundryGarden             CloudFoundryGarden       `json:"cloudFoundryGarden,inline"`                               // cloud_foundry_garden
	ClusterChecks                  ClusterChecks            `json:"clusterChecks"`                                           // cluster_checks
	Exporter                       Exporter                 `json:"exporter,inline"`                                         // telemetry
	OrchestratorExplorer           OrchestratorExplorer     `json:"orchestratorExplorer,inline"`                             // orchestrator_explorer
	Statsd                         Statsd                   `json:"statsd"`                                                  // statsd_*, dagstatsd_*
	Apm                            Apm                      `json:"apm,inline"`                                              // apm_config.*
	Jmx                            Jmx                      `json:"jmx"`                                                     // jmx_*
	RuntimeSecurity                RuntimeSecurity          `json:"runtimeSecurity,inline"`                                  // runtime_security_config.*
	AdminssionController           AdminssionController     `json:"adminssionController"`                                    // admission_controller.*
	ExternalMetricsProvider        ExternalMetricsProvider  `json:"externalMetricsProvider,inline"`                          // external_metrics_provider.*
	EnablePayloads                 EnablePayloads           `json:"enablePayloads,inline"`                                   // enable_payloads.*
	SystemProbe                    SystemProbe              `json:"systemProbe,inline"`                                      // system_probe_config.*
	Listeners                      []Listeners              `json:"listeners,inline"`                                        // listeners
	ConfigProviders                []ConfigurationProviders `json:"configProviders,inline"`                                  // config_providers
	VerboseReport                  bool                     `json:"verboseReport" default:"true"`                            // collects run in verbose mode, e.g. report both with cpu.used(sys+user), cpu.system & cpu.user
	ApiKey                         string                   `json:"apiKey"`                                                  // api_key
	Hostname                       string                   `json:"hostname" flag:"hostname" description:"custom host name"` //
	HostnameFQDN                   bool                     `json:"hostnameFqdn"`                                            // hostname_fqdn
	HostnameForceConfigAsCanonical bool                     `json:"hostnameForceConfigAsCanonical"`                          // hostname_force_config_as_canonical
	BindHost                       string                   `json:"bindHost"`                                                // bind_host
	IPCAddress                     string                   `json:"ipcAddress" default:"localhost"`                          // ipc_address
	//CmdPort                          int               `json:"cmdPort" default:"5001"`                                                                                           // cmd_port, move to apiserver
	MaxProcs                         string            `json:"maxProcs" default:"4"`                                                                                           //
	CoreDump                         bool              `json:"coreDump" default:"true"`                                                                                        // go_core_dump
	HealthPort                       int               `json:"healthPort" default:"0"`                                                                                         // health_port
	SkipSSLValidation                bool              `json:"skipSslValidation"`                                                                                              // skip_ssl_validation
	ForceTLS12                       bool              `json:"forceTls_12"`                                                                                                    // force_tls_12
	ECSMetadataTimeout               time.Duration     `json:"-"`                                                                                                              // ecs_metadata_timeout
	ECSMetadataTimeout_              int               `json:"ecsMetadataTimeout" flag:"ecs-metadata-timeout" default:"500" description:"ecs metadata timeout (Millisecond)"`  // ecs_metadata_timeout
	MetadataEndpointsMaxHostnameSize int               `json:"metadataEndpointsMaxHostnameSize" default:"255"`                                                                 // metadata_endpoints_max_hostname_size
	CloudProviderMetadata            []string          `json:"cloudProviderMetadata"`                                                                                          //cloud_provider_metadata
	GCEMetadataTimeout               time.Duration     `json:"-"`                                                                                                              // gce_metadata_timeout
	GCEMetadataTimeout_              int               `json:"gceMetadataTimeout" flag:"gce-metadata-timeout" default:"1000" description:"gce metadata timeout (Millisecond)"` // gce_metadata_timeout
	ClusterName                      string            `json:"clusterName"`                                                                                                    // cluster_name
	CLCRunnerEnabled                 bool              `json:"clcRunnerEnabled"`                                                                                               //
	CLCRunnerHost                    string            `json:"clcRunnerHost"`                                                                                                  // clc_runner_host
	ExtraConfigProviders             []string          `json:"extraConfigProviders"`                                                                                           // extra_config_providers
	CloudFoundry                     bool              `json:"cloudFoundry"`                                                                                                   // cloud_foundry
	BoshID                           string            `json:"boshID"`                                                                                                         // bosh_id
	CfOSHostnameAliasing             bool              `json:"cfOsHostnameAliasing"`                                                                                           // cf_os_hostname_aliasing
	CollectGCETags                   bool              `json:"collectGceTags" default:"true"`                                                                                  // collect_gce_tags
	CollectEC2Tags                   bool              `json:"collectEc2Tags"`                                                                                                 // collect_ec2_tags
	DisableClusterNameTagKey         bool              `json:"disableClusterNameTagKey"`                                                                                       // disable_cluster_name_tag_key
	Env                              string            `json:"env"`                                                                                                            // env
	Tags                             []string          `json:"tags"`                                                                                                           // tags
	TagValueSplitSeparator           map[string]string `json:"tagValueSplitSeparator"`                                                                                         // tag_value_split_separator
	NoProxyNonexactMatch             bool              `json:"noProxyNonexactMatch"`                                                                                           // no_proxy_nonexact_match
	EnableGohai                      bool              `json:"enableGohai" default:"true"`                                                                                     // enable_gohai
	ChecksTagCardinality             string            `json:"checksTagCardinality" default:"low"`                                                                             // checks_tag_cardinality
	HistogramAggregates              []string          `json:"histogramAggregates" default:"max median avg count"`                                                             // histogram_aggregates
	HistogramPercentiles             []string          `json:"histogramPercentiles" default:"0.95"`                                                                            // histogram_percentiles
	AcLoadTimeout                    time.Duration     `json:"-"`                                                                                                              // ac_load_timeout
	AcLoadTimeout_                   int               `json:"acLoadTimeout" flag:"ac-load-timeout" default:"30000" description:"ac load timeout(Millisecond)"`                // ac_load_timeout
	AdConfigPollInterval             time.Duration     `json:"-"`                                                                                                              // ad_config_poll_interval
	AdConfigPollInterval_            int               `json:"adConfigPollInterval" flag:"ac-config-poll-interval" default:"10" description:"ac config poll interval(Second)"` // ad_config_poll_interval
	AggregatorBufferSize             int               `json:"aggregatorBufferSize" default:"100"`                                                                             // aggregator_buffer_size
	IotHost                          bool              `json:"iotHost"`                                                                                                        // iot_host
	HerokuDyno                       bool              `json:"herokuDyno"`                                                                                                     // heroku_dyno
	BasicTelemetryAddContainerTags   bool              `json:"basicTelemetryAddContainerTags"`                                                                                 // basic_telemetry_add_container_tags
	LogPayloads                      bool              `json:"logPayloads"`                                                                                                    // log_payloads
	AggregatorStopTimeout            time.Duration     `json:"-"`                                                                                                              // aggregator_stop_timeout
	AggregatorStopTimeout_           int               `json:"aggregatorStopTimeout" flag:"aggregator-stop-timeout" default:"2" description:"aggregator stop timeout(Second)"` // aggregator_stop_timeout
	AutoconfTemplateDir              string            `json:"autoconfTemplateDir"`                                                                                            // autoconf_template_dir
	AutoconfTemplateUrlTimeout       bool              `json:"autoconfTemplateUrlTimeout"`                                                                                     // autoconf_template_url_timeout
	CheckRunners                     int               `json:"checkRunners" default:"4"`                                                                                       // check_runners
	LoggingFrequency                 int64             `json:"loggingFrequency" default:"500"`                                                                                 // logging_frequency
	GUIPort                          bool              `json:"guiPort"`                                                                                                        // GUI_port
	XAwsEc2MetadataTokenTtlSeconds   bool              `json:"xAwsEc2MetadataTokenTtlSeconds"`                                                                                 // X-aws-ec2-metadata-token-ttl-seconds
	AcExclude                        bool              `json:"acExclude"`                                                                                                      // ac_exclude
	AcInclude                        bool              `json:"acInclude"`                                                                                                      // ac_include
	AllowArbitraryTags               bool              `json:"allowArbitraryTags"`                                                                                             // allow_arbitrary_tags
	AppKey                           bool              `json:"appKey"`                                                                                                         // app_key
	CCoreDump                        bool              `json:"cCoreDump"`                                                                                                      // c_core_dump
	CStacktraceCollection            bool              `json:"cStacktraceCollection"`                                                                                          // c_stacktrace_collection
	CacheSyncTimeout                 time.Duration     `json:"-"`
	CacheSyncTimeout_                int               `json:"cacheSyncTimeout" flag:"cache-sync-timeout" default:"2" description:"cache sync timeout(Second)"`                               // cache_sync_timeout
	ClcRunnerId                      string            `json:"clcRunnerId"`                                                                                                                   // clc_runner_id
	CmdHost                          string            `json:"cmdHost" default:"localhost"`                                                                                                   // cmd_host
	CollectKubernetesEvents          bool              `json:"collectKubernetesEvents"`                                                                                                       // collect_kubernetes_events
	ComplianceConfigDir              string            `json:"complianceConfigDir" default:"/opt/n9e/agentd/compliance.d"`                                                                    // compliance_config.dir
	ComplianceConfigEnabled          bool              `json:"complianceConfigEnabled"`                                                                                                       // compliance_config.enabled
	ContainerCgroupPrefix            string            `json:"containerCgroupPrefix"`                                                                                                         // container_cgroup_prefix
	ContainerCgroupRoot              string            `json:"containerCgroupRoot"`                                                                                                           // container_cgroup_root
	ContainerProcRoot                string            `json:"containerProcRoot"`                                                                                                             // container_proc_root
	ContainerdNamespace              string            `json:"containerdNamespace" default:"k8s.io"`                                                                                          // containerd_namespace
	CriConnectionTimeout             time.Duration     `json:"-"`                                                                                                                             // cri_connection_timeout
	CriConnectionTimeout_            int               `json:"criConnectionTimeout" flag:"cri-connection-timeout" default:"1" description:"cri connection timeout(Second)"`                   // cri_connection_timeout
	CriQueryTimeout                  time.Duration     `json:"-"`                                                                                                                             // cri_query_timeout
	CriQueryTimeout_                 int               `json:"criQueryTimeout" flag:"cri-query-timeout" default:"5" description:"cri query timeout(Second)"`                                  // cri_query_timeout
	DatadogCluster                   bool              `json:"datadogCluster"`                                                                                                                // datadog-cluster
	DisableFileLogging               bool              `json:"disableFileLogging"`                                                                                                            // disable_file_logging
	DockerLabelsAsTags               bool              `json:"dockerLabelsAsTags"`                                                                                                            // docker_labels_as_tags
	DockerQueryTimeout               time.Duration     `json:"-"`                                                                                                                             // docker_query_timeout
	DockerQueryTimeout_              int               `json:"dockerQueryTimeout" flag:"docker-query-timeout" default:"5" description:"docker query timeout(Second)"`                         // docker_query_timeout
	EC2PreferImdsv2                  bool              `json:"ec2PreferImdsv2"`                                                                                                               // ec2_prefer_imdsv2
	EC2MetadataTimeout               time.Duration     `json:"-"`                                                                                                                             // ec2_metadata_timeout
	EC2MetadataTimeout_              int               `json:"ec2MetadataTimeout" falg:"ec2-metadata-timeout" default:"300" description:"ec2 metadata timeout(Millisecond)"`                  // ec2_metadata_timeout
	EC2MetadataTokenLifetime         time.Duration     `json:"-"`                                                                                                                             // ec2_metadata_token_lifetime
	EC2MetadataTokenLifetime_        int               `json:"ec2MetadataTokenLifetime" falg:"ec2-metadata-token-lifetime" default:"21600" description:"ec2 metadata token lifetime(Second)"` // ec2_metadata_token_lifetime
	EC2UseWindowsPrefixDetection     bool              `json:"ec2UseWindowsPrefixDetection"`                                                                                                  // ec2_use_windows_prefix_detection
	EcsAgentContainerName            string            `json:"ecsAgentContainerName" default:"ecs-agent"`                                                                                     // ecs_agent_container_name
	EcsAgentUrl                      bool              `json:"ecsAgentUrl"`                                                                                                                   // ecs_agent_url
	EcsCollectResourceTagsEc2        bool              `json:"ecsCollectResourceTagsEc2"`                                                                                                     // ecs_collect_resource_tags_ec2
	EKSFargate                       bool              `json:"eksFargate"`                                                                                                                    // eks_fargate
	EnableMetadataCollection         bool              `json:"enableMetadataCollection" default:"true"`                                                                                       // enable_metadata_collection

	ExcludeGCETags []string `json:"excludeGceTags" default:"kube-env,kubelet-config,containerd-configure-sh,startup-script,shutdown-script,configure-sh,sshKeys,ssh-keys,user-data,cli-cert,ipsec-cert,ssl-cert,google-container-manifest,boshSettings,windows-startup-script-ps1,common-psm1,k8s-node-setup-psm1,serial-port-logging-enable,enable-oslogin,disable-address-manager,disable-legacy-endpoints,windows-keys,kubeconfig"` // exclude_gce_tags

	ExcludePauseContainer                bool              `json:"excludePauseContainer"`                         // exclude_pause_container
	ExternalMetricsAggregator            string            `json:"externalMetricsAggregator" default:"avg"`       // external_metrics.aggregator
	ExtraListeners                       []string          `json:"extraListeners"`                                // extra_listeners
	ForceTls12                           bool              `json:"forceTls_12"`                                   // force_tls_12
	FullSketches                         bool              `json:"fullSketches"`                                  // full-sketches
	GceSendProjectIdTag                  bool              `json:"gceSendProjectIdTag"`                           // gce_send_project_id_tag
	GoCoreDump                           bool              `json:"goCoreDump"`                                    // go_core_dump
	HpaConfigmapName                     string            `json:"hpaConfigmapName" default:"n9e-custom-metrics"` // hpa_configmap_name
	HpaWatcherGcPeriod                   time.Duration     `json:"-"`
	HpaWatcherGcPeriod_                  int               `json:"hpaWatcherGcPeriod" flag:"hpa-watcher-gc-period" default:"300" description:"hpa_watcher_gcPeriod(Second)"` // hpa_watcher_gc_period
	IgnoreAutoconf                       []string          `json:"ignoreAutoconf"`                                                                                           // ignore_autoconf
	InventoriesEnabled                   bool              `json:"inventoriesEnabled" default:"true"`                                                                        // inventories_enabled
	InventoriesMaxInterval               time.Duration     `json:"-"`
	InventoriesMaxInterval_              int               `json:"inventoriesMaxInterval" flag:"inventories-max-interval" default:"600" description:"inventoriesMaxInterval(Second)"` // inventories_max_interval
	InventoriesMinInterval               time.Duration     `json:"-"`
	InventoriesMinInterval_              int               `json:"inventoriesMinInterval" flag:"inventories-min-interval" default:"300" description:"inventoriesMinInterval(Second)"` // inventories_min_interval
	KubeResourcesNamespace               bool              `json:"kubeResourcesNamespace"`                                                                                            // kube_resources_namespace
	KubeletCachePodsDuration             time.Duration     `json:"-"`
	KubeletCachePodsDuration_            int               `json:"kubeletCachePodsDuration" flag:"kubelet-cache-pods-duration" default:"5" description:"kubeletCachePodsDuration(Second)"` // kubelet_cache_pods_duration
	KubeletClientCa                      string            `json:"kubeletClientCa"`                                                                                                        // kubelet_client_ca
	KubeletClientCrt                     string            `json:"kubeletClientCrt"`                                                                                                       // kubelet_client_crt
	KubeletClientKey                     string            `json:"kubeletClientKey"`                                                                                                       // kubelet_client_key
	KubeletListenerPollingInterval       time.Duration     `json:"-"`
	KubeletListenerPollingInterval_      int               `json:"kubeletListenerPollingInterval" flag:"kubelet-listener-polling-interval" default:"5" description:"kubeletListenerPollingInterval(Second)"` // kubelet_listener_polling_interval
	KubeletTlsVerify                     bool              `json:"kubeletTlsVerify" default:"true"`                                                                                                          // kubelet_tls_verify
	KubeletWaitOnMissingContainer        time.Duration     `json:"-"`
	KubeletWaitOnMissingContainer_       int               `json:"kubeletWaitOnMissingContainer" flag:"kubelet-wait-on-missing-container" description:"kubeletWaitOnMissingContainer(Second)"` // kubelet_wait_on_missing_container
	KubernetesApiserverClientTimeout     time.Duration     `json:"-"`
	KubernetesApiserverClientTimeout_    int               `json:"kubernetesApiserverClientTimeout" flag:"kubernetes-apiserver-client-timeout" default:"10" description:"kubernetes_apiserverClientTimeout(Seconde)"` // kubernetes_apiserver_client_timeout
	KubernetesApiserverUseProtobuf       bool              `json:"kubernetesApiserverUseProtobuf"`                                                                                                                    // kubernetes_apiserver_use_protobuf
	KubernetesCollectMetadataTags        bool              `json:"kubernetesCollectMetadataTags" default:"true"`                                                                                                      // kubernetes_collect_metadata_tags
	KubernetesCollectServiceTags         bool              `json:"kubernetesCollectServiceTags"`                                                                                                                      // kubernetes_collect_service_tags
	KubernetesHttpKubeletPort            int               `json:"kubernetesHttpKubeletPort" default:"10255"`                                                                                                         // kubernetes_http_kubelet_port
	KubernetesHttpsKubeletPort           int               `json:"kubernetesHttpsKubeletPort" default:"10250"`                                                                                                        // kubernetes_https_kubelet_port
	KubernetesInformersResyncPeriod      time.Duration     `json:"-"`
	KubernetesInformersResyncPeriod_     int               `json:"kubernetesInformersResyncPeriod" flag:"kubernetes-informers-resync-period" default:"300" description:"kubernetesInformersResyncPeriod(Second)"` // kubernetes_informers_resync_period
	KubernetesKubeletHost                string            `json:"kubernetesKubeletHost"`                                                                                                                         // kubernetes_kubelet_host
	KubernetesKubeletNodename            string            `json:"kubernetesKubeletNodename"`                                                                                                                     // kubernetes_kubelet_nodename
	KubernetesMapServicesOnIp            bool              `json:"kubernetesMapServicesOnIp"`                                                                                                                     // kubernetes_map_services_on_ip
	KubernetesMetadataTagUpdateFreq      time.Duration     `json:"-"`
	KubernetesMetadataTagUpdateFreq_     int               `json:"kubernetesMetadataTagUpdateFreq" flag:"kubernetes-metadata-tag-update-freq" default:"60" description:"kubernetesMetadataTagUpdateFreq(Second)"` // kubernetes_metadata_tag_update_freq
	KubernetesNamespaceLabelsAsTags      bool              `json:"kubernetesNamespaceLabelsAsTags"`                                                                                                               // kubernetes_namespace_labels_as_tags
	KubernetesNodeLabelsAsTags           bool              `json:"kubernetesNodeLabelsAsTags"`                                                                                                                    // kubernetes_node_labels_as_tags
	KubernetesPodAnnotationsAsTags       map[string]string `json:"kubernetesPodAnnotationsAsTags"`                                                                                                                // kubernetes_pod_annotations_as_tags
	KubernetesPodExpirationDuration      time.Duration     `json:"-"`
	KubernetesPodExpirationDuration_     int               `json:"kubernetesPodExpirationDuration" flag:"kubernetes-pod-expiration-duration" default:"900" description:"kubernetes_podExpirationDuration(Second)"` // kubernetes_pod_expiration_duration
	KubernetesPodLabelsAsTags            map[string]string `json:"kubernetesPodLabelsAsTags"`                                                                                                                      // kubernetes_pod_labels_as_tags
	KubernetesServiceTagUpdateFreq       map[string]string `json:"kubernetesServiceTagUpdateFreq"`                                                                                                                 // kubernetes_service_tag_update_freq
	LeaderElection                       bool              `json:"leaderElection"`                                                                                                                                 // leader_election
	LeaderLeaseDuration                  time.Duration     `json:"-"`
	LeaderLeaseDuration_                 int               `json:"leaderLeaseDuration" flag:"leader-lease-duration" default:"60" description:"leader lease duration(second)"` // leader_lease_duration
	LogEnabled                           bool              `json:"logEnabled"`                                                                                                // log_enabled
	LogFile                              string            `json:"logFile"`                                                                                                   // log_file
	LogFormatJson                        bool              `json:"logFormatJson"`                                                                                             // log_format_json
	LogFormatRfc3339                     bool              `json:"logFormatRfc3339"`                                                                                          // log_format_rfc3339
	LogLevel                             string            `json:"logLevel"`                                                                                                  // log_level
	LogToConsole                         bool              `json:"logToConsole"`                                                                                              // log_to_console
	MemtrackEnabled                      bool              `json:"memtrackEnabled"`                                                                                           // memtrack_enabled
	MetricsPort                          int               `json:"metricsPort" default:"5000"`                                                                                // metrics_port
	ProcRoot                             string            `json:"procRoot" default:"/proc"`                                                                                  // proc_root
	ProcessAgentConfigHostIps            bool              `json:"processAgentConfigHostIps"`                                                                                 // process_agent_config.host_ips
	ProcessConfigEnabled                 bool              `json:"processConfigEnabled"`                                                                                      // process_config.enabled
	ProfilingEnabled                     bool              `json:"profilingEnabled"`                                                                                          // profiling.enabled
	ProfilingProfileDdUrl                bool              `json:"profilingProfileDdUrl"`                                                                                     // profiling.profile_dd_url
	PrometheusScrapeEnabled              bool              `json:"prometheusScrapeEnabled"`                                                                                   // prometheus_scrape.enabled
	PrometheusScrapeServiceEndpoints     bool              `json:"prometheusScrapeServiceEndpoints"`                                                                          // prometheus_scrape.service_endpoints
	ProxyHttps                           bool              `json:"proxyHttps"`                                                                                                // proxy.https
	ProxyNoProxy                         bool              `json:"proxyNoProxy"`                                                                                              // proxy.no_proxy
	Python3LinterTimeout                 time.Duration     `json:"-"`
	Python3LinterTimeout_                int               `json:"python3LinterTimeout" flag:"python3-linter-timeout" default:"120" description:"python3LinterTimeout(Second)"` // python3_linter_timeout
	PythonVersion                        string            `json:"pythonVersion" default:"3"`                                                                                   // python_version
	SerializerMaxPayloadSize             int               `json:"serializerMaxPayloadSize" default:"2621440" description:"2.5mb"`                                              // serializer_max_payload_size
	SerializerMaxUncompressedPayloadSize int               `json:"serializerMaxUncompressedPayloadSize" default:"4194304" description:"4mb"`                                    // serializer_max_uncompressed_payload_size
	//ServerTimeout                        time.Duration     `json:"-"`
	//ServerTimeout_                       int               `json:"serverTimeout" flag:"server-timeout" default:"15" description:"server_timeout(Second)"` // server_timeout, move to apiserver
	SkipSslValidation   bool   `json:"skipSslValidation"`               // skip_ssl_validation
	SyslogRfc           bool   `json:"syslogRfc"`                       // syslog_rfc
	TelemetryEnabled    bool   `json:"telemetryEnabled" default:"true"` // telemetry.enabled
	TracemallocDebug    bool   `json:"tracemallocDebug"`                // tracemalloc_debug
	Yaml                bool   `json:"yaml"`                            // yaml
	MetricTransformFile string `json:"metricTransformFile"`
	//N9e                                  N9e                      `json:"n9e"`
	//

	WinSkipComInit       bool `json:"winSkipComInit"`       // win_skip_com_init
	DisablePy3Validation bool `json:"disablePy3Validation"` // disable_py3_validation

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
//	V5Format bool   `json:"v5Format"`
//}

type SystemProbe struct {
	Enabled                      bool                `json:"enabled"`                      // system_probe_config.enabled & system_probe
	SysprobeSocket               string              `json:"sysprobeSocket"`               // system_probe_config.sysprobe_socket
	BPFDebug                     bool                `json:"bpfDebug"`                     // system_probe_config.bpf_debug
	BPFDir                       string              `json:"bpfDir"`                       // system_probe_config.bpf_dir
	ExcludedLinuxVersions        []string            `json:"excludedLinuxVersions"`        // system_probe_config.excluded_linux_versions
	EnableTracepoints            bool                `json:"enableTracepoints"`            // system_probe_config.enable_tracepoints
	EnableRuntimeCompiler        bool                `json:"enableRuntimeCompiler"`        // system_probe_config.enable_runtime_compiler
	RuntimeCompilerOutputDir     string              `json:"runtimeCompilerOutputDir"`     // system_probe_config.runtime_compiler_output_dir
	KernelHeaderDirs             []string            `json:"kernelHeaderDirs"`             // system_probe_config.kernel_header_dirs
	DisableTcp                   bool                `json:"disableTcp"`                   // system_probe_config.disable_tcp
	DisableUdp                   bool                `json:"disableUdp"`                   // system_probe_config.disable_udp
	DisableIpv6                  bool                `json:"disableIpv6"`                  // system_probe_config.disable_ipv6
	OffsetGuessThreshold         int64               `json:"offsetGuessThreshold"`         // system_probe_config.offset_guess_threshold
	SourceExcludes               map[string][]string `json:"sourceExcludes"`               // system_probe_config.source_excludes
	DestExcludes                 map[string][]string `json:"destExcludes"`                 // system_probe_config.dest_excludes
	MaxTrackedConnections        int                 `json:"maxTrackedConnections"`        // system_probe_config.max_tracked_connections
	MaxClosedConnectionsBuffered int                 `json:"maxClosedConnectionsBuffered"` // system_probe_config.max_closed_connections_buffered
	ClosedChannelSize            int                 `json:"closedChannelSize"`            // system_probe_config.closed_channel_size
	MaxConnectionStateBuffered   int                 `json:"maxConnectionStateBuffered"`   // system_probe_config.max_connection_state_buffered
	DisableDnsInspection         bool                `json:"disableDnsInspection"`         // system_probe_config.disable_dns_inspection
	CollectDnsStats              bool                `json:"collectDnsStats"`              // system_probe_config.collect_dns_stats
	CollectLocalDns              bool                `json:"collectLocalDns"`              // system_probe_config.collect_local_dns
	CollectDnsDomains            bool                `json:"collectDnsDomains"`            // system_probe_config.collect_dns_domains
	MaxDnsStats                  int                 `json:"maxDnsStats"`                  // system_probe_config.max_dns_stats
	DnsTimeout                   time.Duration       `json:"-"`
	DnsTimeout_                  int                 `json:"dnsTimeout" flag:"system-probe-dns-timeout" default:"15" description:"dnsTimeout(Second)"` // system_probe_config.dns_timeout_in_s
	EnableConntrack              bool                `json:"enableConntrack"`                                                                          // system_probe_config.enable_conntrack
	ConntrackMaxStateSize        int                 `json:"conntrackMaxStateSize"`                                                                    // system_probe_config.conntrack_max_state_size
	ConntrackRateLimit           int                 `json:"conntrackRateLimit"`                                                                       // system_probe_config.conntrack_rate_limit
	EnableConntrackAllNamespaces bool                `json:"enableConntrackAllNamespaces"`                                                             // system_probe_config.enable_conntrack_all_namespaces
	WindowsEnableMonotonicCount  bool                `json:"windowsEnableMonotonicCount"`                                                              // system_probe_config.windows.enable_monotonic_count
	WindowsDriverBufferSize      int                 `json:"windowsDriverBufferSize"`                                                                  // system_probe_config.windows.driver_buffer_size
}

func (p *SystemProbe) Validate() error {
	p.DnsTimeout = time.Second * time.Duration(p.DnsTimeout_)
	return nil
}

type EnablePayloads struct {
	Series              bool `json:"series" default:"true"`               // enable_payloads.series
	Events              bool `json:"events" default:"false"`              // enable_payloads.events
	ServiceChecks       bool `json:"serviceChecks" default:"false"`       // enable_payloads.service_checks
	Sketches            bool `json:"sketches" default:"false"`            // enable_payloads.sketches
	JsonToV1Intake      bool `json:"jsonToV1Intake" default:"false"`      // enable_payloads.json_to_v1_intake
	Metadata            bool `json:"metadata" default:"false"`            //
	HostMetadata        bool `json:"hostMetadata" default:"false"`        //
	AgentchecksMetadata bool `json:"agentchecksMetadata" default:"false"` //
}

func (p *EnablePayloads) Validate() error {
	return nil
}

type ExternalMetricsProvider struct {
	ApiKey                bool          `json:"apiKey"`                   // external_metrics_provider.api_key
	AppKey                bool          `json:"appKey"`                   // external_metrics_provider.app_key
	BucketSize            int           `json:"bucketSize" default:"300"` // external_metrics_provider.bucket_size
	Enabled               bool          `json:"enabled"`                  // external_metrics_provider.enabled
	LocalCopyRefreshRate  time.Duration `json:"-"`
	LocalCopyRefreshRate_ int           `json:"localCopyRefreshRate" flag:"external-metrics-provider-local-copy-refresh-rate" default:"30" description:"localCopyRefreshRate(Second)"` // external_metrics_provider.local_copy_refresh_rate
	MaxAge                int           `json:"maxAge" default:"20"`                                                                                                                   // external_metrics_provider.max_age
	RefreshPeriod         int           `json:"refreshPeriod" default:"30"`                                                                                                            // external_metrics_provider.refresh_period
	Rollup                int           `json:"rollup" default:"30"`                                                                                                                   // external_metrics_provider.rollup
	UseDatadogmetricCrd   bool          `json:"useDatadogmetricCrd"`                                                                                                                   // external_metrics_provider.use_datadogmetric_crd
	WpaController         bool          `json:"wpaController"`                                                                                                                         // external_metrics_provider.wpa_controller
}

func (p *ExternalMetricsProvider) Validate() error {
	p.LocalCopyRefreshRate = time.Second * time.Duration(p.LocalCopyRefreshRate_)
	return nil
}

type AdminssionController struct {
	Enabled                         bool          `json:"enabled"` // admission_controller.enabled
	CertificateExpirationThreshold  time.Duration `json:"-"`
	CertificateExpirationThreshold_ int           `json:"certificateExpirationThreshold" flag:"admission-controller-certificate-expiration-threshold" default:"30" description:"certificateExpirationThreshold(Day)"` // admission_controller.certificate.expiration_threshold
	CertificateSecretName           string        `json:"certificateSecretName" default:"webhook-certificate"`                                                                                                        // admission_controller.certificate.secret_name
	CertificateValidityBound        time.Duration `json:"-"`
	CertificateValidityBound_       int           `json:"certificateValidityBound" flag:"admission-controller-certificate-validity-bound" default:"365" description:"certificateValidityBound(Day)"` // admission_controller.certificate.validity_bound
	InjectConfigEnabled             bool          `json:"injectConfigEnabled" default:"true"`                                                                                                        // admission_controller.inject_config.enabled
	InjectConfigEndpoint            string        `json:"injectConfigEndpoint" default:"/injectconfig"`                                                                                              // admission_controller.inject_config.endpoint
	InjectTagsEnabled               bool          `json:"injectTagsEnabled" default:"true"`                                                                                                          // admission_controller.inject_tags.enabled
	InjectTagsEndpoint              string        `json:"injectTagsEndpoint" default:"/injecttags"`                                                                                                  // admission_controller.inject_tags.endpoint
	MutateUnlabelled                bool          `json:"mutateUnlabelled"`                                                                                                                          // admission_controller.mutate_unlabelled
	PodOwnersCacheValidity          int           `json:"podOwnersCacheValidity" default:"10" description:"Minute"`                                                                                  // admission_controller.pod_owners_cache_validity
	ServiceName                     string        `json:"serviceName" default:"admission-controller"`                                                                                                // admission_controller.service_name
	TimeoutSeconds                  time.Duration `json:"-"`
	TimeoutSeconds_                 int           `json:"timeoutSeconds" flag:"adminssion-controller-timeout" default:"30" description:"timeoutSeconds(Second)"` // admission_controller.timeout_seconds
	WebhookName                     string        `json:"webhookName" default:"n9e-webhook"`                                                                     // admission_controller.webhook_name

}

func (p *AdminssionController) Validate() error {
	p.CertificateExpirationThreshold = 24 * time.Hour * time.Duration(p.CertificateExpirationThreshold_)
	p.CertificateValidityBound = 24 * time.Hour * time.Duration(p.CertificateValidityBound_)
	p.TimeoutSeconds = time.Second * time.Duration(p.TimeoutSeconds_)
	return nil
}

type RuntimeSecurity struct {
	Socket                             string `json:"socket"`                                                       // runtime_security_config.socket
	AgentMonitoringEvents              bool   `json:"agentMonitoringEvents"`                                        // runtime_security_config.agent_monitoring_events
	CookieCacheSize                    bool   `json:"cookieCacheSize"`                                              // runtime_security_config.cookie_cache_size
	CustomSensitiveWords               bool   `json:"customSensitiveWords"`                                         // runtime_security_config.custom_sensitive_words
	EnableApprovers                    bool   `json:"enableApprovers"`                                              // runtime_security_config.enable_approvers
	EnableDiscarders                   bool   `json:"enableDiscarders"`                                             // runtime_security_config.enable_discarders
	EnableKernelFilters                bool   `json:"enableKernelFilters"`                                          // runtime_security_config.enable_kernel_filters
	Enabled                            bool   `json:"enabled"`                                                      // runtime_security_config.enabled
	EventServerBurst                   bool   `json:"eventServerBurst"`                                             // runtime_security_config.event_server.burst
	EventServerRate                    bool   `json:"eventServerRate"`                                              // runtime_security_config.event_server.rate
	EventsStatsPollingInterval         bool   `json:"eventsStatsPollingInterval"`                                   // runtime_security_config.events_stats.polling_interval
	FimEnabled                         bool   `json:"fimEnabled"`                                                   // runtime_security_config.fim_enabled
	FlushDiscarderWindow               bool   `json:"flushDiscarderWindow"`                                         // runtime_security_config.flush_discarder_window
	LoadControllerControlPeriod        bool   `json:"loadControllerControlPeriod"`                                  // runtime_security_config.load_controller.control_period
	LoadControllerDiscarderTimeout     bool   `json:"loadControllerDiscarderTimeout"`                               // runtime_security_config.load_controller.discarder_timeout
	LoadControllerEventsCountThreshold bool   `json:"loadControllerEventsCountThreshold"`                           // runtime_security_config.load_controller.events_count_threshold
	PidCacheSize                       bool   `json:"pidCacheSize"`                                                 // runtime_security_config.pid_cache_size
	PoliciesDir                        string `json:"policiesDir" default:"/opt/n9e/agentd/etc/runtime-security.d"` // runtime_security_config.policies.dir
	SyscallMonitorEnabled              bool   `json:"syscallMonitorEnabled"`                                        // runtime_security_config.syscall_monitor.enabled
}

func (p *RuntimeSecurity) Validate() error {
	return nil
}

type Jmx struct {
	CheckPeriod                int           `json:"checkPeriod"` // jmx_check_period
	CollectionTimeout          time.Duration `json:"-"`
	CollectionTimeout_         int           `json:"collectionTimeout" flag:"jmx-collection-timeout" default:"60" description:"collectionTimeout"` // jmx_collection_timeout
	CustomJars                 []string      `json:"customJars"`                                                                                   // jmx_custom_jars
	LogFile                    bool          `json:"logFile"`                                                                                      // jmx_log_file
	MaxRestarts                int           `json:"maxRestarts" default:"3"`                                                                      // jmx_max_restarts
	ReconnectionThreadPoolSize int           `json:"reconnectionThreadPoolSize" default:"3"`                                                       // jmx_reconnection_thread_pool_size
	ReconnectionTimeout        time.Duration `json:"-"`
	ReconnectionTimeout_       int           `json:"reconnectionTimeout" flag:"jmx-reconnection-timeout" default:"50" description:"reconnectionTimeout(Second)"` // jmx_reconnection_timeout
	RestartInterval            time.Duration `json:"-"`
	RestartInterval_           int           `json:"restartInterval" flag:"jmx-restart-interval" default:"5" description:"restartInterval(Second)"` // jmx_restart_interval
	ThreadPoolSize             int           `json:"threadPoolSize" default:"3"`                                                                    // jmx_thread_pool_size
	UseCgroupMemoryLimit       bool          `json:"useCgroupMemoryLimit"`                                                                          // jmx_use_cgroup_memory_limit
	UseContainerSupport        bool          `json:"useContainerSupport"`                                                                           // jmx_use_container_support
}

func (p *Jmx) Validate() error {
	p.CollectionTimeout = time.Second * time.Duration(p.CollectionTimeout_)
	p.ReconnectionTimeout = time.Second * time.Duration(p.ReconnectionTimeout_)
	p.RestartInterval = time.Second * time.Duration(p.RestartInterval_)
	return nil
}

type Apm struct {
	AdditionalEndpoints           bool                          `json:"additionalEndpoints"`           // apm_config.additional_endpoints
	AnalyzedRateByService         bool                          `json:"analyzedRateByService"`         // apm_config.analyzed_rate_by_service
	AnalyzedSpans                 bool                          `json:"analyzedSpans"`                 // apm_config.analyzed_spans
	ApiKey                        bool                          `json:"apiKey"`                        // apm_config.api_key
	ApmDdUrl                      bool                          `json:"apmDdUrl"`                      // apm_config.apm_dd_url
	ApmNonLocalTraffic            bool                          `json:"apmNonLocalTraffic"`            // apm_config.apm_non_local_traffic
	ConnectionLimit               bool                          `json:"connectionLimit"`               // apm_config.connection_limit
	ConnectionResetInterval       bool                          `json:"connectionResetInterval"`       // apm_config.connection_reset_interval
	DdAgentBin                    bool                          `json:"ddAgentBin"`                    // apm_config.dd_agent_bin
	Enabled                       bool                          `json:"enabled"`                       // apm_config.enabled
	Env                           bool                          `json:"env"`                           // apm_config.env
	ExtraSampleRate               bool                          `json:"extraSampleRate"`               // apm_config.extra_sample_rate
	FilterTagsReject              bool                          `json:"filterTagsReject"`              // apm_config.filter_tags.reject
	FilterTagsRequire             bool                          `json:"filterTagsRequire"`             // apm_config.filter_tags.require
	IgnoreResources               bool                          `json:"ignoreResources"`               // apm_config.ignore_resources
	LogFile                       bool                          `json:"logFile"`                       // apm_config.log_file
	LogLevel                      bool                          `json:"logLevel"`                      // apm_config.log_level
	LogThrottling                 bool                          `json:"logThrottling"`                 // apm_config.log_throttling
	MaxCpuPercent                 bool                          `json:"maxCpuPercent"`                 // apm_config.max_cpu_percent
	MaxEventsPerSecond            bool                          `json:"maxEventsPerSecond"`            // apm_config.max_events_per_second
	MaxMemory                     bool                          `json:"maxMemory"`                     // apm_config.max_memory
	MaxTracesPerSecond            bool                          `json:"maxTracesPerSecond"`            // apm_config.max_traces_per_second
	Obfuscation                   traceconfig.ObfuscationConfig `json:"obfuscation"`                   // apm_config.obfuscation
	ProfilingAdditionalEndpoints  bool                          `json:"profilingAdditionalEndpoints"`  // apm_config.profiling_additional_endpoints
	ProfilingDdUrl                bool                          `json:"profilingDdUrl"`                // apm_config.profiling_dd_url
	ReceiverPort                  string                        `json:"receiverPort"`                  // apm_config.receiver_port
	ReceiverSocket                bool                          `json:"receiverSocket"`                // apm_config.receiver_socket
	ReceiverTimeout               bool                          `json:"receiverTimeout"`               // apm_config.receiver_timeout
	RemoteTagger                  bool                          `json:"remoteTagger"`                  // apm_config.remote_tagger
	SyncFlushing                  bool                          `json:"syncFlushing"`                  // apm_config.sync_flushing
	WindowsPipeBufferSize         bool                          `json:"windowsPipeBufferSize"`         // apm_config.windows_pipe_buffer_size
	WindowsPipeName               bool                          `json:"windowsPipeName"`               // apm_config.windows_pipe_name
	WindowsPipeSecurityDescriptor bool                          `json:"windowsPipeSecurityDescriptor"` // apm_config.windows_pipe_security_descriptor
}

func (p *Apm) Validate() error {
	return nil
}

type Statsd struct {
	Enabled                           bool             `json:"enabled" default:"false"` // use_dogstatsd
	Host                              string           `json:"host"`                    //
	Port                              int              `json:"port" default:"8125"`     // dogstatsd_port
	Socket                            string           `json:"socket"`                  // dogstatsd_socket
	PipeName                          string           `json:"pipeName"`                // dogstatsd_pipe_name
	ContextExpirySeconds              time.Duration    `json:"-"`
	ContextExpirySeconds_             int              `json:"contextExpirySeconds" flag:"statsd-context-expiry-seconds" default:"300" description:"contextExpirySeconds(Second)"` // dogstatsd_context_expiry_seconds
	ExpirySeconds                     time.Duration    `json:"-"`
	ExpirySeconds_                    int              `json:"expirySeconds" flag:"statsd-expiry-seconds" default:"300" description:"expirySeconds(Second)"` // dogstatsd_expiry_seconds
	StatsEnable                       bool             `json:"statsEnable" default:"true"`                                                                   // dogstatsd_stats_enable
	StatsBuffer                       int              `json:"statsBuffer" default:"10"`                                                                     // dogstatsd_stats_buffer
	MetricsStatsEnable                bool             `json:"metricsStatsEnable" default:"false"`                                                           // dogstatsd_metrics_stats_enable - for debug
	BufferSize                        int              `json:"bufferSize" default:"8192"`                                                                    // dogstatsd_buffer_size
	MetricNamespace                   string           `json:"metricNamespace"`                                                                              // statsd_metric_namespace
	MetricNamespaceBlacklist          []string         `json:"metricNamespaceBlacklist"`                                                                     // statsd_metric_namespace_blacklist
	Tags                              []string         `json:"tags"`                                                                                         // dogstatsd_tags
	EntityIdPrecedence                bool             `json:"entityIdPrecedence"`                                                                           // dogstatsd_entity_id_precedence
	EolRequired                       []string         `json:"eolRequired"`                                                                                  // dogstatsd_eol_required
	DisableVerboseLogs                bool             `json:"disableVerboseLogs"`                                                                           // dogstatsd_disable_verbose_logs
	ForwardHost                       string           `json:"forwardHost"`                                                                                  // statsd_forward_host
	ForwardPort                       int              `json:"forwardPort"`                                                                                  // statsd_forward_port
	QueueSize                         int              `json:"queueSize" default:"1024"`                                                                     // dogstatsd_queue_size
	MapperCacheSize                   int              `json:"mapperCacheSize" default:"1000"`                                                               // dogstatsd_mapper_cache_size
	MapperProfiles                    []MappingProfile `json:"mapperProfiles"`                                                                               // dogstatsd_mapper_profiles
	StringInternerSize                int              `json:"stringInternerSize" default:"4096"`                                                            // dogstatsd_string_interner_size
	SocketRcvbuf                      int              `json:"socektRcvbuf"`                                                                                 // dogstatsd_so_rcvbuf
	PacketBufferSize                  int              `json:"packetBufferSize" default:"32"`                                                                // dogstatsd_packet_buffer_size
	PacketBufferFlushTimeout          time.Duration    `json:"-"`
	PacketBufferFlushTimeout_         int              `json:"packetBufferFlushTimeout" flag:"statsd-packet-buffer-flush-timeout" default:"100" description:"packetBufferFlushTimeout(Millisecond)"` // dogstatsd_packet_buffer_flush_timeout
	TagCardinality                    string           `json:"tagCardinality" default:"low"`                                                                                                         // dogstatsd_tag_cardinality
	NonLocalTraffic                   bool             `json:"nonLocalTraffic"`                                                                                                                      // dogstatsd_non_local_traffic
	OriginDetection                   bool             `json:"OriginDetection"`                                                                                                                      // dogstatsd_origin_detection
	HistogramCopyToDistribution       bool             `json:"histogramCopyToDistribution"`                                                                                                          // histogram_copy_to_distribution
	HistogramCopyToDistributionPrefix string           `json:"histogramCopyToDistributionPrefix"`                                                                                                    // histogram_copy_to_distribution_prefix
}

func (p *Statsd) Validate() error {
	p.ContextExpirySeconds = time.Second * time.Duration(p.ContextExpirySeconds_)
	p.ExpirySeconds = time.Second * time.Duration(p.ExpirySeconds_)
	p.PacketBufferFlushTimeout = time.Millisecond * time.Duration(p.PacketBufferFlushTimeout_)
	return nil
}

type AdditionalEndpoint struct {
	Endpoints []string `json:"endpoints"`
	ApiKeys   []string `json:"apiKeys"`
}

type Forwarder struct {
	AdditionalEndpoints       []AdditionalEndpoint `json:"additionalEndpoints"` // additional_endpoints
	ApikeyValidationInterval  time.Duration        `json:"-"`
	ApikeyValidationInterval_ int                  `json:"apikeyValidationInterval" flag:"forwarder-apikey-validation-interval" default:"3600" description:"apikeyValidationInterval(Second)"` // forwarder_apikey_validation_interval
	BackoffBase               float64              `json:"backoffBase" default:"2"`                                                                                                            // forwarder_backoff_base
	BackoffFactor             float64              `json:"backoffFactor" default:"2"`                                                                                                          // forwarder_backoff_factor
	BackoffMax                float64              `json:"backoffMax" default:"64"`                                                                                                            // forwarder_backoff_max
	ConnectionResetInterval   time.Duration        `json:"-"`
	ConnectionResetInterval_  int                  `json:"connectionResetInterval" flag:"forwarder-connection-reset-interval" description:"connectionResetInterval(Second)"` // forwarder_connection_reset_interval
	FlushToDiskMemRatio       float64              `json:"flushToDiskMemRatio" default:"0.5"`                                                                                // forwarder_flush_to_disk_mem_ratio
	NumWorkers                int                  `json:"numWorkers" default:"1"`                                                                                           // forwarder_num_workers
	OutdatedFileInDays        int                  `json:"outdatedFileInDays" default:"10"`                                                                                  // forwarder_outdated_file_in_days
	RecoveryInterval          int                  `json:"recoveryInterval" default:"2"`                                                                                     // forwarder_recovery_interval
	RecoveryReset             bool                 `json:"recoveryReset"`                                                                                                    // forwarder_recovery_reset
	StopTimeout               time.Duration        `json:"-"`
	StopTimeout_              int                  `json:"stopTimeout" flag:"forwarder-stop-timeout" default:"2" description:"stopTimeout(Second)"` // forwarder_stop_timeout
	StorageMaxDiskRatio       float64              `json:"storageMaxDiskRatio" default:"0.95"`                                                      // forwarder_storage_max_disk_ratio
	StorageMaxSizeInBytes     int64                `json:"storageMaxSizeInBytes"`                                                                   // forwarder_storage_max_size_in_bytes
	StoragePath               string               `json:"storagePath"`                                                                             // forwarder_storage_path
	Timeout                   time.Duration        `json:"-"`
	Timeout_                  int                  `json:"timeout" flag:"forwarder-timeout" default:"20" description:"timeout(Second)"` // forwarder_timeout
	RetryQueuePayloadsMaxSize int                  `json:"retryQueuePayloadsMaxSize" default:"15728640" description:"15m"`              // forwarder_retry_queue_payloads_max_size
	//RetryQueueMaxSize         int           `json:"retryQueueMaxSize"`         // forwarder_retry_queue_max_size

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
	ListenNetwork string `json:"listenNetwork" default:"unix"`                              // cloud_foundry_garden.listen_network
	ListenAddress string `json:"listenAddress" default:"/var/vcap/data/garden/garden.sock"` // cloud_foundry_garden.listen_address
}

// ProcessingRule defines an exclusion or a masking rule to
// be applied on log lines
type ProcessingRule struct {
	Type               string
	Name               string
	ReplacePlaceholder string `json:"replacePlaceholder"`
	Pattern            string
	// TODO: should be moved out
	Regex       *regexp.Regexp
	Placeholder []byte
}

func (p *ProcessingRule) Validate() error {
	return nil
}

type LogsConfig struct {
	Enabled                     bool                        `json:"enabled"`             // logs_enabled
	AdditionalEndpoints         []logstypes.Endpoint        `json:"additionalEndpoints"` // logs_config.additional_endpoints
	ContainerCollectAll         bool                        `json:"containerCollectAll"` // logs_config.container_collect_all
	ProcessingRules             []*logstypes.ProcessingRule `json:"processingRules"`     // logs_config.processing_rules
	APIKey                      string                      `json:"apiKey"`              // logs_config.api_key
	DevModeNoSSL                bool                        `json:"devModeNoSsl"`        // logs_config.dev_mode_no_ssl
	ExpectedTagsDuration        time.Duration               `json:"-"`
	ExpectedTagsDuration_       int                         `json:"expectedTagsDuration" flag:"logs-expected-tags-duration" description:"expectedTagsDuration(Second)"` // logs_config.expected_tags_duration
	Socks5ProxyAddress          string                      `json:"socks5ProxyAddress"`                                                                                 // logs_config.socks5_proxy_address
	UseTCP                      bool                        `json:"useTcp"`                                                                                             // logs_config.use_tcp
	UseHTTP                     bool                        `json:"useHttp"`                                                                                            // logs_config.use_http
	DevModeUseProto             bool                        `json:"devModeUseProto" default:"true"`                                                                     // logs_config.dev_mode_use_proto
	ConnectionResetInterval     time.Duration               `json:"-"`
	ConnectionResetInterval_    int                         `json:"connectionResetInterval" flag:"logs-connection-reset-interval" default:"" description:"connectionResetInterval(Second)"` // logs_config.connection_reset_interval
	LogsUrl                     string                      `json:"logsUrl"`                                                                                                                // logs_config.logs_dd_url, dd_url
	UsePort443                  bool                        `json:"usePort443"`                                                                                                             // logs_config.use_port_443
	UseSSL                      bool                        `json:"useSsl"`                                                                                                                 // !logs_config.logs_no_ssl
	Url443                      string                      `json:"url_443"`                                                                                                                // logs_config.dd_url_443
	UseCompression              bool                        `json:"useCompression" default:"true"`                                                                                          // logs_config.use_compression
	CompressionLevel            int                         `json:"compressionLevel" default:"6"`                                                                                           // logs_config.compression_level
	URL                         string                      `json:"url" default:"localhost:8080"`                                                                                           // logs_config.dd_url (e.g. localhost:8080)
	BatchWait                   time.Duration               `json:"-"`
	BatchWait_                  int                         `json:"batchWait" flag:"logs-batch-wait" default:"5" description:"batchWait(Second)"` // logs_config.batch_wait
	BatchMaxConcurrentSend      int                         `json:"batchMaxConcurrentSend"`                                                       // logs_config.batch_max_concurrent_send
	TaggerWarmupDuration        time.Duration               `json:"-"`
	TaggerWarmupDuration_       int                         `json:"taggerWarmupDuration" flag:"logs-tagger-warmup-duration" description:"taggerWarmupDuration(Second)"` // logs_config.tagger_warmup_duration
	AggregationTimeout          time.Duration               `json:"-"`
	AggregationTimeout_         int                         `json:"aggregationTimeout" flag:"logs-aggregation-timeout" default:"1000" description:"aggregationTimeout(Millisecond)"` // logs_config.aggregation_timeout
	CloseTimeout                time.Duration               `json:"-"`
	CloseTimeout_               int                         `json:"closeTimeout" flag:"logs-close-timeout" default:"60" description:"closeTimeout(Second)"` // logs_config.close_timeout
	AuditorTTL                  time.Duration               `json:"-"`
	AuditorTTL_                 int                         `json:"auditorTtl" flag:"logs-auditor-ttl" description:"auditorTTL(Second)"` // logs_config.auditor_ttl
	RunPath                     string                      `json:"runPath"`                                                             // logs_config.run_path
	OpenFilesLimit              int                         `json:"openFilesLimit" flag:"logs-open-files-limit" default:"100"`           // logs_config.open_files_limit
	K8SContainerUseFile         bool                        `json:"k8sContainerUseFile"`                                                 // logs_config.k8s_container_use_file
	DockerContainerUseFile      bool                        `json:"dockerContainerUseFile"`                                              // logs_config.docker_container_use_file
	DockerContainerForceUseFile bool                        `json:"dockerContainerForceUseFile"`                                         // logs_config.docker_container_force_use_file
	DockerClientReadTimeout     time.Duration               `json:"-"`
	DockerClientReadTimeout_    int                         `json:"dockerClientReadTimeout" flag:"logs-docker-client-read-timeout" default:"30" description:"dockerClientReadTimeout(Second)"` // logs_config.docker_client_read_timeout
	FrameSize                   int                         `json:"frameSize" default:"9000"`                                                                                                  // logs_config.frame_size
	StopGracePeriod             time.Duration               `json:"-"`
	StopGracePeriod_            int                         `json:"stopGracePeriod" flag:"logs-stop-grace-period" default:"30" description:"stopGracePeriod(Second)"` // logs_config.stop_grace_period
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
	ID                         string `json:"id"`                         // network.id
	EnableHttpMonitoring       bool   `json:"enableHttpMonitoring"`       // network_config.enable_http_monitoring
	IgnoreConntrackInitFailure bool   `json:"ignoreConntrackInitFailure"` // network_config.ignore_conntrack_init_failure
	EnableGatewayLookup        bool   `json:"enableGatewayLookup"`        // network_config.enable_gateway_lookup
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
	ClcRunnersPort             int           `json:"clcRunnersPort" default:"5005"`         // cluster_checks.clc_runners_port
	AdvancedDispatchingEnabled bool          `json:"advancedDispatchingEnabled"`            // cluster_checks.advanced_dispatching_enabled
	ClusterTagName             string        `json:"clusterTagName" default:"cluisterName"` // cluster_checks.cluster_tag_name
	Enabled                    bool          `json:"enabled"`                               // cluster_checks.enabled
	ExtraTags                  []string      `json:"extraTags"`                             // cluster_checks.extra_tags
	NodeExpirationTimeout      time.Duration `json:"-"`
	NodeExpirationTimeout_     int           `json:"nodeExpirationTimeout" flag:"clc-node-expiration-timeout" default:"30" description:"nodeExpirationTimeout(Second)"` // cluster_checks.node_expiration_timeout
	WarmupDuration             time.Duration `json:"-"`
	WarmupDuration_            int           `json:"warmupDuration" flag:"clc-warmup-duration" default:"30" description:"warmupDuration(Second)"` // cluster_checks.warmup_duration

}

func (p *ClusterChecks) Validate() error {
	p.NodeExpirationTimeout = time.Second * time.Duration(p.NodeExpirationTimeout_)
	p.WarmupDuration = time.Second * time.Duration(p.WarmupDuration_)

	return nil
}

type ClusterAgent struct {
	Url                   string `json:"url"`                                               // cluster_agent.url
	AuthToken             string `json:"authToken"`                                         // cluster_agent.auth_token
	CmdPort               int    `json:"cmdPort" default:"5005"`                            // cluster_agent.cmd_port
	Enabled               bool   `json:"enabled"`                                           // cluster_agent.enabled
	KubernetesServiceName string `json:"kubernetesServiceName" default:"n9e-cluster-agent"` // cluster_agent.kubernetes_service_name
	TaggingFallback       string `json:"taggingFallback"`                                   // cluster_agent.tagging_fallback
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
	MatchType string            `json:"matchType" json:"matchType"`
	Name      string            `json:"name" json:"name"`
	Tags      map[string]string `json:"tags" json:"tags"`
}

type OrchestratorExplorer struct { // orchestrator_explorer
	Url                       string   `json:"url"`                       // orchestrator_explorer.orchestrator_dd_url
	AdditionalEndpoints       []string `json:"additionalEndpoints"`       // orchestrator_explorer.orchestrator_additional_endpoints
	CustomSensitiveWords      []string `json:"customSensitiveWords"`      // orchestrator_explorer.custom_sensitive_words
	ContainerScrubbingEnabled bool     `json:"containerScrubbingEnabled"` // orchestrator_explorer.container_scrubbing.enabled
	Enabled                   bool     `json:"enabled" default:"true"`    // orchestrator_explorer.enabled
	ExtraTags                 []string `json:"extraTags"`                 // orchestrator_explorer.extra_tags
}

func (p *OrchestratorExplorer) Validate() error {
	return nil
}

type Autoconfig struct {
	Enabled         bool       `json:"enabled" default:"true"` // autoconfig_from_environment
	ExcludeFeatures []string   `json:"excludeFeatures"`        // autoconfig_exclude_features
	features        FeatureMap `json:"-"`
}

func (p *Autoconfig) Validate() error {
	return nil
}

type PrometheusScrape struct {
	Enabled          bool                     `json:"enabled"`          // prometheus_scrape.enabled
	ServiceEndpoints bool                     `json:"serviceEndpoints"` // prometheus_scrape.service_endpoints
	Checks           []*types.PrometheusCheck `json:"checks"`           // prometheus_scrape.checks
}

func (p *PrometheusScrape) Validate() error {
	return nil
}

// ConfigurationProviders helps unmarshalling `config_providers` config param
type ConfigurationProviders struct {
	Name             string `json:"name"`
	Polling          bool   `json:"polling"`
	PollInterval     string `json:"pollInterval"`
	TemplateURL      string `json:"templateUrl"`
	TemplateDir      string `json:"templateDir"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	CAFile           string `json:"caFile"`
	CAPath           string `json:"caPath"`
	CertFile         string `json:"certFile"`
	KeyFile          string `json:"keyFile"`
	Token            string `json:"token"`
	GraceTimeSeconds int    `json:"graceTimeSeconds"`
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
