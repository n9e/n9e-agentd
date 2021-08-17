package core

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/DataDog/datadog-agent/pkg/aggregator"
	"github.com/DataDog/datadog-agent/pkg/api/healthprobe"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/embed/jmx"
	"github.com/DataDog/datadog-agent/pkg/dogstatsd"
	"github.com/DataDog/datadog-agent/pkg/epforwarder"
	"github.com/DataDog/datadog-agent/pkg/forwarder"
	"github.com/DataDog/datadog-agent/pkg/forwarder/transaction"
	"github.com/DataDog/datadog-agent/pkg/logs"
	"github.com/DataDog/datadog-agent/pkg/metadata"
	"github.com/DataDog/datadog-agent/pkg/metadata/host"
	"github.com/DataDog/datadog-agent/pkg/pidfile"
	"github.com/DataDog/datadog-agent/pkg/serializer"
	"github.com/DataDog/datadog-agent/pkg/snmp/traps"
	"github.com/DataDog/datadog-agent/pkg/status/health"
	ddutil "github.com/DataDog/datadog-agent/pkg/util"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/n9e/n9e-agentd/cmd/agent/common"
	"github.com/n9e/n9e-agentd/cmd/agent/common/misconfig"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/config/settings"
	commonsettings "github.com/n9e/n9e-agentd/pkg/config/settings"
	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/n9e/n9e-agentd/pkg/version"
	"github.com/yubo/golib/proc"
	"k8s.io/klog/v2"
)

const (
	moduleName                      = "agent"
	loggerName    config.LoggerName = "CORE"
	jmxLoggerName config.LoggerName = "JMXFETCH"
)

type module struct {
	config                 *config.Config
	name                   string
	serializer             *serializer.Serializer
	eventPlatformForwarder epforwarder.EventPlatformForwarder

	hostname string

	ctx    context.Context
	cancel context.CancelFunc
}

var (
	_module = &module{name: moduleName}
	hookOps = []proc.HookOps{{
		Hook:        _module.start,
		Owner:       moduleName,
		HookNum:     proc.ACTION_START,
		Priority:    proc.PRI_MODULE,
		SubPriority: options.PRI_M_CORE,
	}, {
		Hook:        _module.stop,
		Owner:       moduleName,
		HookNum:     proc.ACTION_STOP,
		Priority:    proc.PRI_MODULE,
		SubPriority: options.PRI_M_CORE,
	}}
)

func (p *module) start(ctx context.Context) error {
	p.ctx, p.cancel = context.WithCancel(ctx)

	if err := p.readConfig(); err != nil {
		return err
	}

	if err := p.setupLogger(); err != nil {
		return err
	}

	// coredump, hostMetadata, ...
	if err := p.setupGeneric(); err != nil {
		return nil
	}

	if err := p.startExporter(); err != nil {
		klog.Errorf("exporter start err %s, ignore", err)
	}

	if err := p.startHealthprobe(); err != nil {
		return err
	}

	if err := p.startForwarder(); err != nil {
		return err
	}

	// statsd & aggr
	if err := p.startAggStatsd(); err != nil {
		return err
	}

	if err := initRuntimeSettings(); err != nil {
		log.Warnf("Can't initiliaze the runtime settings: %v", err)
	}

	if err := p.startSnmpTrap(); err != nil {
		return err
	}

	if err := p.startLogsAgent(); err != nil {
		return err
	}

	if err := p.startSystemProbe(); err != nil {
		return err
	}

	if err := p.startDetectCloudProvider(); err != nil {
		return nil
	}

	if err := p.startLogVersionRecord(); err != nil {
		return err
	}

	if err := p.startAutoconfig(); err != nil {
		return nil
	}

	// check for common misconfigurations and report them to log
	misconfig.ToLog()

	if err := p.setupMedatadaCollector(); err != nil {
		return err
	}

	return nil
}

func (p *module) stop(ctx context.Context) error {
	health, err := health.GetReadyNonBlocking()
	if err != nil {
		klog.Warningf("Agent health unknown: %s", err)
	} else if len(health.Unhealthy) > 0 {
		klog.Warningf("Some components were unhealthy: %v", health.Unhealthy)
	}

	if common.DSD != nil {
		common.DSD.Stop()
	}
	if common.AC != nil {
		common.AC.Stop()
	}
	if common.MetadataScheduler != nil {
		common.MetadataScheduler.Stop()
	}
	traps.StopServer()

	jmx.StopJmxfetch()
	aggregator.StopDefaultAggregator()
	if common.Forwarder != nil {
		common.Forwarder.Stop()
	}
	//if orchestratorForwarder != nil {
	//	orchestratorForwarder.Stop()
	//}

	if p.eventPlatformForwarder != nil {
		p.eventPlatformForwarder.Stop()
	}

	logs.Stop()
	//gui.StopGUIServer()
	//profiler.Stop()

	os.Remove(p.config.PidfilePath)
	klog.Info("See ya!")

	return nil
}

func completionHandler(transaction *transaction.HTTPTransaction, statusCode int, body []byte, err error) {
	if len(body) > 0 {
		klog.InfoS("transaction completion", "err", err, "code", statusCode, "body", string(body))
	}
}

func init() {
	proc.RegisterHooks(hookOps)

	config.AddFlags()
}

func (p *module) readConfig() error {
	configer := proc.ConfigerFrom(p.ctx)
	if config.Configfile != "" {
		klog.Warningf("--config has been Deprecated, use -f(--values) instead of it")
		if _, err := os.Stat(config.Configfile); err != nil {
			// path/to/whatever exists
			return fmt.Errorf("openfile %s error %s", config.Configfile, err)
		}
	}

	cf := config.NewConfig(configer)
	if err := configer.Read(moduleName, cf); err != nil {
		return err
	}

	if config.TestConfig {
		configer.PrintFlags()
		fmt.Printf("agentd: configuration test is successful\n")
		os.Exit(0)
	}

	p.config = cf
	config.C = cf
	config.Context = p.ctx

	return nil
}

func (p *module) setupLogger() error {
	cf := p.config
	// Setup logger
	if runtime.GOOS != "android" {
		syslogURI := config.GetSyslogURI()
		logFile := cf.LogFile
		jmxLogFile := cf.Jmx.LogFile

		if config.C.DisableFileLogging {
			// this will prevent any logging on file
			logFile = ""
			jmxLogFile = ""
		}

		if err := config.SetupLogger(
			loggerName,
			cf.LogLevel,
			logFile,
			syslogURI,
			cf.SyslogRfc,
			cf.LogToConsole,
			cf.LogFormatJson,
		); err != nil {
			return err
		}

		// Setup JMX logger
		return config.SetupJMXLogger(
			jmxLoggerName,
			cf.LogLevel,
			jmxLogFile,
			syslogURI,
			cf.SyslogRfc,
			cf.LogToConsole,
			cf.LogFormatJson,
		)

	}
	if err := config.SetupLogger(
		loggerName,
		cf.LogLevel,
		"", // no log file on android
		"", // no syslog on android,
		false,
		true,  // always log to console
		false, // not in json
	); err != nil {
		return err
	}

	// Setup JMX logger
	return config.SetupJMXLogger(
		jmxLoggerName,
		cf.LogLevel,
		"", // no log file on android
		"", // no syslog on android,
		false,
		true,  // always log to console
		false, // not in json
	)
}

func (p *module) setupGeneric() error {
	cf := p.config

	setMaxProcs(cf.MaxProcs)

	if cf.CoreDump {
		if err := util.SetupCoreDump(); err != nil {
			klog.Warningf("Can't setup core dumps: %v, core dumps might not be available after a crash", err)
		}
	}

	// i18n
	i18n.SetDefaultPrinter(config.C.Lang, "zh")

	// pidfile
	if cf.PidfilePath != "" {
		if err := pidfile.WritePID(cf.PidfilePath); err != nil {
			return fmt.Errorf("Error while writing PID file, exiting: %v", err)
		}
		klog.Infof("pid '%d' written to pid file '%s'", os.Getpid(), cf.PidfilePath)
	}

	var err error
	if p.hostname, err = util.GetHostname(cf.Hostname); err != nil {
		return fmt.Errorf("Error while getting hostname, exiting: %s", err)
	}
	klog.Infof("Hostname is: %s", p.hostname)

	// HACK: init host metadata module (CPU) early to avoid any
	//       COM threading model conflict with the python checks
	if err := host.InitHostMetadata(); err != nil {
		klog.Errorf("Unable to initialize host metadata: %v", err)
	}

	return nil
}

// Setup healthcheck port
func (p *module) startHealthprobe() error {
	port := p.config.HealthPort
	if port == 0 {
		return nil
	}
	if err := healthprobe.Serve(p.ctx, port); err != nil {
		return fmt.Errorf("Error starting health port, exiting: %v", err)
	}
	klog.Infof("Health check listening on port %d", port)

	return nil
}

func (p *module) startForwarder() error {
	// setup the forwarder
	keysPerDomain, err := config.GetMultipleEndpoints()
	if err != nil {
		klog.Error("Misconfiguration of agent endpoints: ", err)
	}

	// Enable core agent specific features like persistence-to-disk
	options := forwarder.NewOptions(keysPerDomain)
	options.EnabledFeatures = forwarder.SetFeature(options.EnabledFeatures, forwarder.CoreFeatures)
	options.CompletionHandler = completionHandler

	f := forwarder.NewDefaultForwarder(options)
	klog.V(5).Infof("Starting forwarder")
	f.Start() //nolint:errcheck
	klog.V(5).Infof("Forwarder started")

	p.serializer = serializer.NewSerializer(f, nil)
	common.Forwarder = f

	p.eventPlatformForwarder = epforwarder.NewEventPlatformForwarder()
	p.eventPlatformForwarder.Start()

	return nil
}

// start Agg and dogstatsd
func (p *module) startAggStatsd() (err error) {
	agg := aggregator.InitAggregator(p.serializer, p.eventPlatformForwarder, p.hostname)
	agg.AddAgentStartupTelemetry(version.AgentVersion)

	if !p.config.Statsd.Enabled {
		return nil
	}

	common.DSD, err = dogstatsd.NewServer(agg, nil)
	if err != nil {
		return fmt.Errorf("Could not start statsd: %s", err)
	}
	klog.V(5).Infof("statsd started")

	return nil
}

func (p *module) startSnmpTrap() error {
	if !p.config.SnmpTraps.Enabled {
		return nil
	}

	if !p.config.Logs.Enabled {
		klog.Warning("snmp-traps server did not start, as log collection is disabled. " +
			"Please enable log collection to collect and forward traps.",
		)
		return nil
	}

	if err := traps.StartServer(); err != nil {
		klog.Errorf("Failed to start snmp-traps server: %s", err)
	}

	return nil
}

// start logs-agent
func (p *module) startLogsAgent() error {
	if !p.config.Logs.Enabled {
		klog.Info("logs-agent disabled")
		return nil
	}

	if err := logs.Start(func() *autodiscovery.AutoConfig { return common.AC }); err != nil {
		klog.Error("Could not start logs-agent: ", err)
	}

	return nil
}

func (p *module) startSystemProbe() error {
	//if err = common.SetupSystemProbeConfig(sysProbeConfFilePath); err != nil {
	//	klog.Infof("System probe config not found, disabling pulling system probe info in the status page: %v", err)
	//}

	return nil
}

// Detect Cloud Provider
func (p *module) startDetectCloudProvider() error {
	go ddutil.DetectCloudProvider(context.Background())
	return nil
}

func (p *module) startLogVersionRecord() error {
	// Append version and timestamp to version history log file if this Agent is different than the last run version
	ddutil.LogVersionHistory()

	return nil
}

func (p *module) startAutoconfig() error {
	// create and setup the Autoconfig instance
	common.LoadComponents()
	// start the autoconfig, this will immediately run any configured check
	common.StartAutoConfig()

	return nil
}

// setup the metadata collector
func (p *module) setupMedatadaCollector() error {
	common.MetadataScheduler = metadata.NewScheduler(p.serializer)
	if err := metadata.SetupMetadataCollection(common.MetadataScheduler, metadata.AllDefaultCollectors); err != nil {
		return err
	}

	if err := metadata.SetupInventories(common.MetadataScheduler, common.AC, common.Coll); err != nil {
		return err
	}

	return nil
}

// initRuntimeSettings builds the map of runtime settings configurable at runtime.
func initRuntimeSettings() error {
	// Runtime-editable settings must be registered here to dynamically populate command-line information
	if err := commonsettings.RegisterRuntimeSetting(
		commonsettings.LogLevelRuntimeSetting(config.S_log_level), config.C.LogLevel); err != nil {
		return err
	}
	if err := commonsettings.RegisterRuntimeSetting(
		commonsettings.RuntimeMutexProfileFraction(config.S_runtime_mutex_profile_fraction), config.C.InternalProfiling.MutexProfileFraction); err != nil {
		return err
	}
	if err := commonsettings.RegisterRuntimeSetting(
		commonsettings.RuntimeBlockProfileRate(config.S_runtime_block_profile_rate), config.C.InternalProfiling.BlockProfileRate); err != nil {
		return err
	}
	if err := commonsettings.RegisterRuntimeSetting(
		settings.DsdStatsRuntimeSetting(config.S_dogstatsd_stats), config.C.Statsd.MetricsStatsEnable); err != nil {
		return err
	}
	if err := commonsettings.RegisterRuntimeSetting(
		settings.DsdCaptureDurationRuntimeSetting(config.S_dogstatsd_capture_duration), 0); err != nil {
		return err
	}
	if err := commonsettings.RegisterRuntimeSetting(
		commonsettings.ProfilingGoroutines(config.S_internal_profiling_goroutines), config.C.InternalProfiling.EnableGoroutineStacktraces); err != nil {
		return err
	}
	return commonsettings.RegisterRuntimeSetting(
		commonsettings.ProfilingRuntimeSetting(config.S_internal_profiling), config.C.InternalProfiling.Enabled)
}
