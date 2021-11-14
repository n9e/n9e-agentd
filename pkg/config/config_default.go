package config

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/pkg/collector/check/defaults"
	forwarder "github.com/n9e/n9e-agentd/pkg/config/forwarder"
	"github.com/n9e/n9e-agentd/pkg/config/internalprofiling"
	logs "github.com/n9e/n9e-agentd/pkg/config/logs"
	statsd "github.com/n9e/n9e-agentd/pkg/config/statsd"
	systemprobe "github.com/n9e/n9e-agentd/pkg/system-probe/config"
	"github.com/yubo/golib/api"
	"github.com/yubo/golib/api/resource"
)

const (
	//authTokenName = "auth_token"

	// DefaultSite is the default site the Agent sends data to.
	DefaultSite    = "datadoghq.com"
	infraURLPrefix = "https://app."

	// DefaultNumWorkers default number of workers for our check runner
	DefaultNumWorkers = 4
	// MaxNumWorkers maximum number of workers for our check runner
	MaxNumWorkers = 25
	// DefaultAPIKeyValidationInterval is the default interval of api key validation checks
	DefaultAPIKeyValidationInterval = 3600 * time.Second

	// DefaultForwarderRecoveryInterval is the default recovery interval,
	// also used if the user-provided value is invalid.
	DefaultForwarderRecoveryInterval = 2

	megaByte = 1024 * 1024

	// DefaultAuditorTTL is the default logs auditor TTL in hours
	DefaultAuditorTTL = "23s"

	// ClusterIDCacheKey is the key name for the orchestrator cluster id in the agent in-mem cache
	ClusterIDCacheKey = "orchestratorClusterID"

	// DefaultRuntimePoliciesDir is the default policies directory used by the runtime security module
	DefaultRuntimePoliciesDir = "/etc/datadog-agent/runtime-security.d"

	// DefaultLogsSenderBackoffFactor is the default logs sender backoff randomness factor
	DefaultLogsSenderBackoffFactor = 2.0

	// DefaultLogsSenderBackoffBase is the default logs sender base backoff time, seconds
	DefaultLogsSenderBackoffBase = 1.0

	// DefaultLogsSenderBackoffMax is the default logs sender maximum backoff time, seconds
	DefaultLogsSenderBackoffMax = 120.0

	// DefaultLogsSenderBackoffRecoveryInterval is the default logs sender backoff recovery interval
	DefaultLogsSenderBackoffRecoveryInterval = 2
)

func DefaultConfig() *Config {
	cf := defaultConfig()

	cf.RootDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	if IsContainerized() {
		// In serverless-containerized environments (e.g Fargate)
		// it's impossible to mount host volumes.
		// Make sure the host paths exist before setting-up the default values.
		// Fallback to the container paths if host paths aren't mounted.
		if pathExists("/host/proc") {
			cf.ProcfsPath = "/host/proc"
			cf.Container.ProcRoot = "/host/proc"

			// Used by some librairies (like gopsutil)
			if v := os.Getenv("HOST_PROC"); v == "" {
				os.Setenv("HOST_PROC", "/host/proc")
			}
		} else {
			cf.ProcfsPath = "/proc"
			cf.Container.ProcRoot = "/proc"
		}
		if pathExists("/host/sys/fs/cgroup/") {
			cf.Container.CgroupRoot = "/host/sys/fs/cgroup/"
		} else {
			cf.Container.CgroupRoot = "/sys/fs/cgroup/"
		}
	} else {
		cf.Container.ProcRoot = "/proc"
		// for amazon linux the cgroup directory on host is /cgroup/
		// we pick memory.stat to make sure it exists and not empty
		if _, err := os.Stat("/cgroup/memory/memory.stat"); !os.IsNotExist(err) {
			cf.Container.CgroupRoot = "/cgroup/"
		} else {
			cf.Container.CgroupRoot = "/sys/fs/cgroup/"
		}
	}

	return cf
}

func defaultConfig() *Config {

	return &Config{
		m:                                       new(sync.RWMutex),
		CliQueryTimeout:                         api.NewDuration("5s"),
		DisablePage:                             false,
		PageSize:                                10,
		NoColor:                                 false,
		Ident:                                   "$ip",
		Alias:                                   "$hostname",
		Lang:                                    "zh",
		EnableN9eProvider:                       true,
		Endpoints:                               []string{"http://localhost:8000"},
		ContainerdNamespace:                     "k8s.io",
		VerboseReport:                           true,
		MaxProcs:                                "4",
		CoreDump:                                true,
		HealthPort:                              0,
		ECSMetadataTimeout:                      api.NewDuration("500ms"),
		MetadataEndpointsMaxHostnameSize:        255,
		GCEMetadataTimeout:                      api.NewDuration("1000ms"),
		CollectGCETags:                          true,
		EnableGohai:                             true,
		ChecksTagCardinality:                    "low",
		HistogramAggregates:                     []string{"max", "median", "avg", "count"},
		HistogramPercentiles:                    []string{"0.95"},
		AcLoadTimeout:                           api.NewDuration("3000ms"),
		AdConfigPollInterval:                    api.NewDuration("10s"),
		AggregatorBufferSize:                    100,
		AggregatorStopTimeout:                   api.NewDuration("2s"),
		AutoconfTemplateUrlTimeout:              5,
		CheckRunners:                            4,
		CacheSyncTimeout:                        api.NewDuration("2s"),
		CriConnectionTimeout:                    api.NewDuration("1s"),
		CriQueryTimeout:                         api.NewDuration("5s"),
		DockerQueryTimeout:                      api.NewDuration("5s"),
		EC2MetadataTimeout:                      api.NewDuration("300ms"),
		EC2MetadataTokenLifetime:                api.NewDuration("6h"),
		EcsAgentContainerName:                   "ecs-agent",
		EnableMetadataCollection:                true,
		ExcludeGCETags:                          []string{"kube-env", "kubelet-config", "containerd-configure-sh", "startup-script", "shutdown-script", "configure-sh", "sshKeys", "ssh-keys", "user-data", "cli-cert", "ipsec-cert", "ssl-cert", "google-container-manifest", "boshSettings", "windows-startup-script-ps1", "common-psm1", "k8s-node-setup-psm1", "serial-port-logging-enable", "enable-oslogin", "disable-address-manager", "disable-legacy-endpoints", "windows-keys", "kubeconfig"},
		ExternalMetricsAggregator:               "avg",
		HpaConfigmapName:                        "n9e-custom-metric",
		HpaWatcherGcPeriod:                      api.NewDuration("300s"),
		InventoriesEnabled:                      true,
		InventoriesMaxInterval:                  api.NewDuration("10m"),
		InventoriesMinInterval:                  api.NewDuration("5m"),
		KubeletCachePodsDuration:                api.NewDuration("5s"),
		KubeletListenerPollingInterval:          api.NewDuration("5s"),
		KubeletTlsVerify:                        true,
		KubernetesApiserverClientTimeout:        10,
		KubernetesCollectMetadataTags:           true,
		KubernetesHttpKubeletPort:               10225,
		KubernetesHttpsKubeletPort:              10250,
		KubernetesInformersResyncPeriod:         api.NewDuration("5m"),
		KubernetesMetadataTagUpdateFreq:         api.NewDuration("60s"),
		KubernetesPodExpirationDuration:         api.NewDuration("15m"),
		LeaderLeaseDuration:                     api.NewDuration("60s"),
		LoggingFrequency:                        500,
		LogFile:                                 "./logs/agent.log",
		LogLevel:                                "info",
		LogToConsole:                            true,
		LogFileMaxSize:                          10485760,
		LogFileMaxRolls:                         1,
		MetricsPort:                             5000,
		ProcRoot:                                "/proc",
		Python3LinterTimeout:                    api.NewDuration("120s"),
		PythonVersion:                           DefaultPython,
		EnableJsonStreamSharedCompressorBuffers: true,
		SerializerMaxPayloadSize:                int(2.5 * 1024 * 1024),
		SerializerMaxUncompressedPayloadSize:    4 * 1024 * 1024,
		EnableServiceChecksStreamPayloadSerialization: true,
		EnableEventsStreamPayloadSerialization:        true,
		EnableSketchStreamPayloadSerialization:        true,
		AzureHostnameStyle:                            "os",
		SecretBackendTimeout:                          30,
		SecretBackendOutputMaxSize:                    1024,
		UseV2Api: UseV2Api{
			Series: true,
		},
		ProcessConfig: ProcessConfig{
			ExpvarPort:                6062,
			ScrubArgs:                 true,
			QueueSize:                 256,
			QueueBytes:                resource.MustParse("60M"),
			MaxPerMessage:             100,
			MaxCtrProcessesPerMessage: 10000,
			GrpcConnectionTimeout:     api.NewDuration("60s"),
			Windows: ProcessWindows{
				ArgsRefreshInterval: 15,
				AddNewArgs:          true,
			},
		},
		EnablePayloads: EnablePayloads{
			Series:              true,
			Events:              false,
			ServiceChecks:       false,
			Sketches:            false,
			JsonToV1Intake:      false,
			Metadata:            false,
			HostMetadata:        false,
			AgentchecksMetadata: false,
		},
		ExternalMetricsProvider: ExternalMetricsProvider{
			BucketSize:           300,
			LocalCopyRefreshRate: api.NewDuration("30s"),
			MaxAge:               20,
			RefreshPeriod:        30,
			Rollup:               30,
		},
		AdminssionController: AdminssionController{
			CertificateExpirationThreshold: api.NewDuration("720h"),
			CertificateSecretName:          "webhook-certificate",
			CertificateValidityBound:       api.NewDuration("8760h"),
			InjectConfigEnabled:            true,
			InjectConfigEndpoint:           "/injectconfig",
			InjectTagsEnabled:              true,
			InjectTagsEndpoint:             "/injecttags",
			PodOwnersCacheValidity:         10,
			ServiceName:                    "admission-controller",
			TimeoutSeconds:                 api.NewDuration("30s"),
			WebhookName:                    "n9e-webhook",
		},
		Jmx: Jmx{
			CollectionTimeout:          api.NewDuration("60s"),
			MaxRestarts:                3,
			ReconnectionThreadPoolSize: 3,
			ReconnectionTimeout:        api.NewDuration("50s"),
			RestartInterval:            api.NewDuration("5s"),
			ThreadPoolSize:             3,
			CheckPeriod:                int(defaults.DefaultCheckInterval / time.Millisecond),
		},
		CloudFoundryGarden: CloudFoundryGarden{
			ListenNetwork: "unix",
			ListenAddress: "/var/vcap/data/garden/garden.sock",
		},
		Telemetry: Telemetry{
			Enabled: true,
			Port:    8011,
			Docs:    true,
			Metrics: true,
			Expvar:  true,
			Pprof:   true,
			Statsd: TelemetryStatsd{
				AggregatorChannelLatencyBuckets: []float64{100, 250, 500, 1000, 1000},
			},
		},
		ClusterChecks: ClusterChecks{
			ClcRunnersPort:        5005,
			ClusterTagName:        "cluisterName",
			NodeExpirationTimeout: api.NewDuration("30s"),
			WarmupDuration:        api.NewDuration("30s"),
		},
		ClusterAgent: ClusterAgent{
			CmdPort:               5005,
			KubernetesServiceName: "n9e-cluster-agent",
		},
		OrchestratorExplorer: OrchestratorExplorer{
			Enabled: true,
		},
		Logs: logs.Config{
			FileScanPeriod:          api.NewDuration("10s"),
			ValidatePodContainerId:  false,
			DevModeUseProto:         true,
			UseCompression:          true,
			CompressionLevel:        6,
			URL:                     "localhost:8080",
			BatchWait:               api.NewDuration("5s"),
			BatchMaxSize:            100,
			BatchMaxContentSize:     1000000,
			SenderBackoffFactor:     2,
			SenderBackoffBase:       1,
			SenderBackoffMax:        120,
			SenderRecoveryInterval:  2,
			AggregationTimeout:      api.NewDuration("1000ms"),
			CloseTimeout:            api.NewDuration("60s"),
			OpenFilesLimit:          100,
			DockerClientReadTimeout: api.NewDuration("30s"),
			FrameSize:               9000,
			StopGracePeriod:         api.NewDuration("30s"),
			AuditorTTL:              api.NewDuration(DefaultAuditorTTL),
		},
		Statsd: statsd.Config{
			Enabled:                  false,
			Port:                     8125,
			ContextExpirySeconds:     api.NewDuration("300s"),
			ExpirySeconds:            api.NewDuration("300s"),
			StatsEnable:              true,
			StatsBuffer:              10,
			MetricsStatsEnable:       false,
			BufferSize:               8192,
			MetricNamespaceBlacklist: StandardStatsdPrefixes,
			QueueSize:                1024,
			MapperCacheSize:          1000,
			StringInternerSize:       4096,
			PacketBufferSize:         32,
			PacketBufferFlushTimeout: api.NewDuration("100ms"),
			TagCardinality:           "low",
		},
		NetworkConfig: NetworkConfig{
			Enabled: true,
		},
		Autoconfig: Autoconfig{
			Enabled: true,
		},
		Forwarder: forwarder.Config{
			ApikeyValidationInterval:  api.NewDuration("60m"),
			BackoffBase:               2,
			BackoffFactor:             2,
			BackoffMax:                64,
			FlushToDiskMemRatio:       0.5,
			NumWorkers:                1,
			OutdatedFileInDays:        10,
			RecoveryInterval:          2,
			StopTimeout:               api.NewDuration("2s"),
			StorageMaxDiskRatio:       0.95,
			Timeout:                   api.NewDuration("20s"),
			HighPrioBufferSize:        100,
			LowPrioBufferSize:         100,
			RequeueBufferSize:         100,
			RetryQueueMaxSize:         0,
			RetryQueuePayloadsMaxSize: resource.MustParse("15Mi"),
		},
		InternalProfiling: internalprofiling.InternalProfiling{
			Enabled:     false,
			Period:      api.NewDuration("5m"),
			CpuDuration: api.NewDuration("5m"),
		},
		SystemProbe: systemprobe.Config{
			Enabled:             false,
			ExternalSystemProbe: false,
			MaxConnsPerMessage:  600,
			LogLevel:            "info",
			DebugPort:           0,
			StatsdHost:          "127.0.0.1",
			StatsdPort:          8125,
			DnsTimeout:          api.NewDuration("15s"),
		},
	}
}
