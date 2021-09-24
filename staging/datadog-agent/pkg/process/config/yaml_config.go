package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/DataDog/datadog-agent/pkg/process/util"
	apicfg "github.com/DataDog/datadog-agent/pkg/process/util/api/config"
	httputils "github.com/DataDog/datadog-agent/pkg/util/http"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/pkg/errors"
)

const (
	ns = "process_config"
)

func key(pieces ...string) string {
	return strings.Join(pieces, ".")
}

// LoadProcessYamlConfig load Process-specific configuration
func (a *AgentConfig) LoadProcessYamlConfig() error {
	//loadEnvVariables()

	// Resolve any secrets
	//if err := config.ResolveSecrets(config.Datadog, filepath.Base(path)); err != nil {
	//	return err
	//}

	URL, err := url.Parse("http://" + config.GetMainEndpoint())
	if err != nil {
		return fmt.Errorf("error parsing process_dd_url: %s", err)
	}
	a.APIEndpoints[0].Endpoint = URL
	a.APIEndpoints[0].APIKey = config.C.ApiKey
	a.HostName = config.C.Hostname

	pc := &config.C.ProcessConfig
	// Note: The enabled environment flag operates differently than that of our YAML configuration
	a.Enabled = pc.Enabled
	if pc.ProcessCheck {
		a.EnabledChecks = processChecks
	}
	// Whether or not the process-agent should output logs to console
	a.LogToConsole = config.C.LogToConsole
	// The full path to the file where process-agent logs will be written.
	if logFile := pc.LogFile; logFile != "" {
		a.LogFile = logFile
	}

	// The interval, in seconds, at which we will run each check. If you want consistent
	// behavior between real-time you may set the Container/ProcessRT intervals to 10.
	// Defaults to 10s for normal checks and 2s for others.
	{
		c := pc.Intervals
		a.setCheckInterval(ns, ContainerCheckName, c.Container.Duration)
		a.setCheckInterval(ns, RTContainerCheckName, c.RTContainer.Duration)
		a.setCheckInterval(ns, ProcessCheckName, c.Process.Duration)
		a.setCheckInterval(ns, RTProcessCheckName, c.RTProcess.Duration)
		a.setCheckInterval(ns, ConnectionsCheckName, c.Connections.Duration)
	}

	// A list of regex patterns that will exclude a process if matched.
	//if k := key(ns, "blacklist_patterns"); config.Datadog.IsSet(k) {
	for _, b := range pc.BlacklistPatterns {
		r, err := regexp.Compile(b)
		if err != nil {
			log.Warnf("Ignoring invalid blacklist pattern: %s", b)
			continue
		}
		a.Blacklist = append(a.Blacklist, r)
	}

	{
		port := pc.ExpvarPort
		if port <= 0 {
			return errors.Errorf("invalid process_config.expvar_port -- %d", port)
		}
		a.ProcessExpVarPort = port
	}

	// Enable/Disable the DataScrubber to obfuscate process args
	a.Scrubber.Enabled = pc.ScrubArgs

	// A custom word list to enhance the default one used by the DataScrubber
	if len(pc.CustomSensitiveWords) > 0 {
		a.Scrubber.AddCustomSensitiveWords(pc.CustomSensitiveWords)
	}

	// Strips all process arguments
	a.Scrubber.StripAllArguments = pc.StripProcArguments

	// How many check results to buffer in memory when POST fails. The default is usually fine.
	if queueSize := pc.QueueSize; queueSize > 0 {
		a.QueueSize = queueSize
	}

	if queueBytes := int(pc.QueueBytes.Value()); queueBytes > 0 {
		a.ProcessQueueBytes = queueBytes
	}

	// The maximum number of processes, or containers per message. Note: Only change if the defaults are causing issues.
	if maxPerMessage := pc.MaxPerMessage; maxPerMessage <= 0 {
		log.Warn("Invalid item count per message (<= 0), ignoring...")
	} else if maxPerMessage <= maxMessageBatch {
		a.MaxPerMessage = maxPerMessage
	} else if maxPerMessage > 0 {
		log.Warn("Overriding the configured item count per message limit because it exceeds maximum")
	}

	// The maximum number of processes belonging to a container per message. Note: Only change if the defaults are causing issues.
	if maxCtrProcessesPerMessage := pc.MaxCtrProcessesPerMessage; maxCtrProcessesPerMessage <= 0 {
		log.Warnf("Invalid max container processes count per message (<= 0), using default value of %d", defaultMaxCtrProcsMessageBatch)
	} else if maxCtrProcessesPerMessage <= maxCtrProcsMessageBatch {
		a.MaxCtrProcessesPerMessage = maxCtrProcessesPerMessage
	} else {
		log.Warnf("Overriding the configured max container processes count per message limit because it exceeds maximum limit of %d", maxCtrProcsMessageBatch)
	}

	// Overrides the path to the Agent bin used for getting the hostname. The default is usually fine.
	//a.DDAgentBin = defaultDDAgentBin
	//if k := key(ns, "dd_agent_bin"); config.Datadog.IsSet(k) {
	//	if agentBin := config.Datadog.GetString(k); agentBin != "" {
	//		a.DDAgentBin = agentBin
	//	}
	//}

	// Overrides the grpc connection timeout setting to the main agent.
	if timeout := pc.GrpcConnectionTimeout.Duration; timeout > 0 {
		a.grpcConnectionTimeout = timeout
	}

	// Windows: Sets windows process table refresh rate (in number of check runs)
	a.Windows.ArgsRefreshInterval = pc.Windows.ArgsRefreshInterval

	// Windows: Controls getting process arguments immediately when a new process is discovered
	a.Windows.AddNewArgs = pc.Windows.AddNewArgs

	// Optional additional pairs of endpoint_url => []apiKeys to submit to other locations.
	for endpointURL, apiKeys := range pc.AdditionalEndpoints {
		u, err := URL.Parse(endpointURL)
		if err != nil {
			return fmt.Errorf("invalid additional endpoint url '%s': %s", endpointURL, err)
		}
		for _, k := range apiKeys {
			a.APIEndpoints = append(a.APIEndpoints, apicfg.Endpoint{
				APIKey:   config.SanitizeAPIKey(k),
				Endpoint: u,
			})
		}
	}

	// use `internal_profiling.enabled` field in `process_config` section to enable/disable profiling for process-agent,
	// but use the configuration from main agent to fill the settings
	{
		c := &pc.InternalProfiling
		if c.Enabled {
			a.ProfilingEnabled = c.Enabled
			a.ProfilingSite = c.Site
			a.ProfilingURL = c.Url
			a.ProfilingEnvironment = c.Env
			a.ProfilingPeriod = c.Period.Duration
			a.ProfilingCPUDuration = c.CpuDuration.Duration
			a.ProfilingMutexFraction = c.MutexProfileFraction
			a.ProfilingBlockRate = c.BlockProfileRate
			a.ProfilingWithGoroutines = c.EnableGoroutineStacktraces
		}
	}

	// Used to override container source auto-detection
	// and to enable multiple collector sources if needed.
	// "docker", "ecs_fargate", "kubelet", "kubelet docker", etc.
	// container_source can be nil since we're not forcing default values in the main config file
	// make sure we don't pass nil value to GetStringSlice to avoid spammy warnings
	if len(pc.ContainerSource) > 0 {
		util.SetContainerSources(pc.ContainerSource)
	}

	// Pull additional parameters from the global config file.
	a.LogLevel = config.C.LogLevel

	a.StatsdPort = config.C.Statsd.Port

	if bindHost := config.C.GetBindHost(); bindHost != "" {
		a.StatsdHost = bindHost
	}

	// Build transport (w/ proxy if needed)
	a.Transport = httputils.CreateHTTPTransport()

	return nil
}

func (a *AgentConfig) setCheckInterval(ns, checkKey string, interval time.Duration) {
	if interval > 0 {
		log.Infof("Overriding container check interval to %ds", interval.Seconds())
		a.CheckIntervals[checkKey] = interval
	}
}
