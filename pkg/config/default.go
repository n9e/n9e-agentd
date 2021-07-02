package config

import (
	"os"
	"time"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/check/defaults"
	"gopkg.in/yaml.v2"
)

func NewDefaultConfig() *Config {
	cf := &Config{}
	if err := yaml.Unmarshal([]byte(defaultConfig), cf); err != nil {
		panic(err)
	}

	if IsContainerized() {
		// In serverless-containerized environments (e.g Fargate)
		// it's impossible to mount host volumes.
		// Make sure the host paths exist before setting-up the default values.
		// Fallback to the container paths if host paths aren't mounted.
		if pathExists("/host/proc") {
			cf.ProcfsPath = "/host/proc"
			cf.ContainerProcRoot = "/host/proc"

			// Used by some librairies (like gopsutil)
			if v := os.Getenv("HOST_PROC"); v == "" {
				os.Setenv("HOST_PROC", "/host/proc")
			}
		} else {
			cf.ProcfsPath = "/proc"
			cf.ContainerProcRoot = "/proc"
		}
		if pathExists("/host/sys/fs/cgroup/") {
			cf.ContainerCgroupRoot = "/host/sys/fs/cgroup/"
		} else {
			cf.ContainerCgroupRoot = "/sys/fs/cgroup/"
		}
	} else {
		cf.ContainerProcRoot = "/proc"
		// for amazon linux the cgroup directory on host is /cgroup/
		// we pick memory.stat to make sure it exists and not empty
		if _, err := os.Stat("/cgroup/memory/memory.stat"); !os.IsNotExist(err) {
			cf.ContainerCgroupRoot = "/cgroup/"
		} else {
			cf.ContainerCgroupRoot = "/sys/fs/cgroup/"
		}

	}

	cf.Statsd.MetricNamespaceBlacklist = StandardStatsdPrefixes
	cf.Jmx.CheckPeriod = int(defaults.DefaultCheckInterval / time.Millisecond)
	cf.LogsConfig.AuditorTTL = DefaultAuditorTTL
	cf.PythonVersion = DefaultPython

	return cf
}

const (
	defaultConfig = `# agent default config
endpoints:
  - http://localhost:8080
n9eSeriesFormat: true
verboseReport: true
hostname: test
lang: zh
enableDocs: true
enableN9eProvider: true
maxProcs: 4
coreDump: true
exporterPort: 8070
pprof: true
expvar: true
metrics: true
authTokenFile: # dir(configfile)/auth_token

#cloudProviderMetadata: ["aws", "gcp", "azure", "alibaba"]
confPath: "."
loggingFrequency: 500
cmdHost: localhost
cmdPort: 5001
enableMetadataCollection: true
enableGohai: true
checkRunners: 4
ipcAddress: localhost
healthPort: 0

python3LinterTimeout: 120s

serverTimeout: 15s

procRoot: /proc
histogramAggregates: ["max", "median", "avg", "count"]
histogramPercentiles: ["0.95"]
aggregatorStopTimeout: 2s
aggregatorBufferSize: 100
basicTelemetryAddContainerTags: false

serializerMaxPayloadSize: 2621440             # 2.5 * 1024 * 1024
serializerMaxUncompressedPayloadSize: 4194304 #   4 * 1024 * 1024

enablePayloads:
  series: true
  events: false
  serviceChecks: false
  sketches: false
  jsonToV1Intake: false
  metadata: false
  hostMetadata: false
  agentchecksMetadata: false

statsd:
  enabled: true
  port: 8125
  bufferSize: 8192
  packetBufferSize: 32
  packetBufferFlushTimeout: 100ms
  queueSize: 1024
  statsBuffer: 10
  expirySeconds: 300s
  contextExpirySeconds: 300s
  mapperCacheSize: 1000
  stringInternerSize: 4096
  statsEnable: true
  metricsStatsEnable: false


acLoadTimeout: 30000ms
adConfigPollInterval: 10s

autoconfig:
  enabled: true

dockerQueryTimeout: 5s
criConnectionTimeout: 1s
criQueryTimeout: 5s

containerdNamespace: k8s.io
kubernetesHttpKubeletPort: 10255
kubernetesHttpsKubeletPort: 10250

kubeletTlsVerify: true
kubernetesPodExpirationDuration: 15m

kubeletCachePodsDuration: 5s
kubeletListenerPollingInterval: 5s
kubernetesCollectMetadataTags: true
kubernetesMetadataTagUpdateFreq: 60s
kubernetesApiserverClientTimeout: 10s

snmpTraps:
  port: 162
  bindHost: localhost
  stopTimeout: 5s

# Kube ApiServer
leaderLeaseDuration: 60s
cacheSyncTimeout: 2s

# cluster agent
clusterAgent:
  enabled: false
  cmdPort: 5005
  authToken:
  kubernetesServiceName: n9e-cluster-agent
metricsPort: 5000

# EC2
metadataEndpointsMaxHostnameSize: 255
ec2MetadataTimeout: 300ms
ec2MetadataTokenLifetime: 21600s

# ECS
ecsAgentContainerName: ecs-agent
ecsMetadataTimeout: 500ms

# GCE
collectGCETags: true
excludeGCETags: ["kube-env", "kubelet-config", "containerd-configure-sh", "startup-script", "shutdown-script", "configure-sh", "sshKeys", "ssh-keys", "user-data", "cli-cert", "ipsec-cert", "ssl-cert", "google-container-manifest", "boshSettings", "windows-startup-script-ps1", "common-psm1", "k8s-node-setup-psm1", "serial-port-logging-enable", "enable-oslogin", "disable-address-manager", "disable-legacy-endpoints", "windows-keys", "kubeconfig"]
gceMetadataTimeout: 1000ms

# Cloud Foundry
cloudFoundry: false
boshID:
cfOSHostnameAliasing: false

cloudFoundryGarden:
  listenNetwork: unix
  listenAddress: /var/vcap/data/garden/garden.sock

# JMXFetch
jmx:
  maxRestarts: 3
  restartInterval: 5s
  threadPoolSize: 3
  reconnectionThreadPoolSize: 3
  collectionTimeout: 60s
  reconnectionTimeout: 50s

expvarPort: 5000
profilingEnabled: false

logsConfig:
  enabled: false
  url: "localhost:8080"
  batchWait: 5s
  frameSize: 9000
  openFilesLimit: 100
  dockerClientReadTimeout: 30
  useSSL: false
  useCompression: true
  compressionLevel: 6
  devModeUseProto: true
  # url443: agent-443-intake.logs.datadoghq.com
  stopGracePeriod: 30s
  closeTimeout: 60s
  aggregationTimeout: 1000ms


checksTagCardinality: low
statsdTagCardinality: low

hpaWatcherPollingFreq: 10s
hpaWatcherGcPeriod: 5m
hpaConfigmapName: datadog-custom-metrics

externalMetricsProvider:
  enabled: false
  port: 443
  refreshPeriod: 30
  batchWindow: 10
  maxAge: 120
  bucketSize: 300
  rollup: 30
  localCopyRefreshRate: 30s

externalMetricsAggregator: avg

kubernetesEventCollectionTimeout: 100s
kubernetesInformersResyncPeriod: 5m

clusterChecks:
  nodeExpirationTimeout: 30s
  warmupDuration: 30s
  clusterTagName: clusterName
  clcRunnersPort: 5005

#clcRunnerPort: 5005
#clcRunnerServerWriteTimeout: 15s
#clcRunnerServerReadheaderTimeout: 10s

admissionController:
  port: 8000
  timeoutSeconds: 30s
  serviceName: datadog-admission-controller
  certificateValidityBound: 8760h		# 365*24h
  certificateExpirationThreshold: 720h		# 30*24h
  certificateSecretName: webhook-certificate
  webhookName: datadog-webhook
  injectConfigEnabled: ture
  injectConfigEndpoint: /injectconfig
  injectTagsEnabled: true
  injectTagsEndpoint: /injecttags
  podOwnersCacheValidity: 10485760		# 10m

telemetry:
  enabled: true

orchestratorExplorer:
  containerScrubbingEnabled: true

# inventories
inventoriesEnabled: true
inventoriesMaxInterval: 10m
inventoriesMinInterval: 5m


complianceConfigEnabled: false
complianceConfigDir: /etc/datadog-agent/compliance.d
#complianceConfigCheckInterval:  20m

# forwarder
forwarder:
  retryQueueMaxSize:
  retryQueuePayloadsMaxSize:  15728640		# 15 * 1024 * 1024
  apikeyValidationInterval: 60m
  timeout: 20s
  numWorkers: 1
  stopTimeout: 2s
  backoffFactor: 2
  backoffBase: 2 # 2s
  backoffMax: 64 # 64s
  recoveryInterval: 2
  recoveryReset: false
  outdatedFileInDays: 10
  flushToDiskMemRatio: 0.5
  storageMaxSizeInBytes: 0
  storageMaxDiskRatio: 0.95
`
)
