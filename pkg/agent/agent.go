package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/n9e/n9e-agentd/cmd/agentd/common"
	"github.com/n9e/n9e-agentd/cmd/agentd/common/misconfig"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/forwarder"
	"github.com/n9e/n9e-agentd/pkg/forwarder/transaction"
	"github.com/n9e/n9e-agentd/pkg/i18n"
	"github.com/n9e/n9e-agentd/pkg/options"
	registrymetrics "github.com/n9e/n9e-agentd/pkg/registry/metrics"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/n9e/n9e-agentd/pkg/version"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/api/healthprobe"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/embed/jmx"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/dogstatsd"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metadata"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metadata/host"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/pidfile"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/serializer"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/snmp/traps"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/status/health"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/telemetry"
	ddutil "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util"
	"github.com/yubo/golib/proc"
	"k8s.io/klog/v2"
)

const (
	moduleName = "agent"
)

type module struct {
	config *config.Config
	name   string

	hostname string
	//confdPath string

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

func (p *module) start(ops *proc.HookOps) error {
	ctx, configer := ops.ContextAndConfiger()
	p.ctx, p.cancel = context.WithCancel(ctx)

	cf := config.NewDefaultConfig()
	if err := configer.ReadYaml(p.name, cf); err != nil {
		return err
	}

	if err := cf.Prepare(configer.ConfigFilePath()); err != nil {
		return err
	}
	p.config = cf
	config.C = cf

	setMaxProcs(cf.MaxProcs)

	if cf.CoreDump {
		if err := util.SetupCoreDump(); err != nil {
			klog.Warningf("Can't setup core dumps: %v, core dumps might not be available after a crash", err)
		}
	}

	// exporter
	//if err := p.startExporter(); err != nil {
	//	return err
	//}
	// outputs
	// serializer
	// aggreator
	// inputs

	// i18n
	i18n.SetDefaultPrinter(config.C.Lang, "zh")

	// Setup expvar server
	var port = config.C.ExpvarPort
	if config.C.EnableDocs {
		http.Handle("/docs/metrics", registrymetrics.Handler())
	}
	if config.C.Telemetry.Enabled {
		http.Handle("/metrics", telemetry.Handler())
	}
	go func() {
		err := http.ListenAndServe("127.0.0.1:"+strconv.Itoa(port), http.DefaultServeMux)
		if err != nil && err != http.ErrServerClosed {
			klog.Errorf("Error creating expvar server on port %d: %s", port, err)
		}
	}()

	// Setup healthcheck port
	if cf.HealthPort > 0 {
		if err := healthprobe.Serve(ctx, cf.HealthPort); err != nil {
			return fmt.Errorf("Error starting health port, exiting: %v", err)
		}
		klog.V(5).Infof("Health check listening on port %d", cf.HealthPort)
	}

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

	// start the cmd HTTP server
	//if runtime.GOOS != "android" {
	//	if err = api.StartServer(); err != nil {
	//		return klog.Errorf("Error while starting api server, exiting: %v", err)
	//	}
	//}

	// setup the forwarder
	keysPerDomain, err := config.GetMultipleEndpoints()
	if err != nil {
		klog.Error("Misconfiguration of agent endpoints: ", err)
	}

	// Enable core agent specific features like persistence-to-disk
	options := forwarder.NewOptions(keysPerDomain)
	options.EnabledFeatures = forwarder.SetFeature(options.EnabledFeatures, forwarder.CoreFeatures)
	options.CompletionHandler = completionHandler
	common.Forwarder = forwarder.NewDefaultForwarder(options)
	klog.V(5).Infof("Starting forwarder")
	common.Forwarder.Start() //nolint:errcheck
	klog.V(5).Infof("Forwarder started")

	// start dogstatsd
	//agg := &aggregator.BufferedAggregator{}
	s := serializer.NewSerializer(common.Forwarder, nil)
	agg := aggregator.InitAggregator(s, p.hostname)
	agg.AddAgentStartupTelemetry(version.AgentVersion)

	if cf.Statsd.Enabled {
		var err error
		common.DSD, err = dogstatsd.NewServer(agg, nil)
		if err != nil {
			return fmt.Errorf("Could not start statsd: %s", err)
		}
	}
	klog.V(5).Infof("statsd started")

	// Start SNMP trap server
	if config.C.SnmpTraps.Enabled {
		if config.C.LogsConfig.Enabled {
			err = traps.StartServer(&cf.SnmpTraps)
			if err != nil {
				klog.Errorf("Failed to start snmp-traps server: %s", err)
			}
		} else {
			klog.Warning("snmp-traps server did not start, as log collection is disabled. " +
				"Please enable log collection to collect and forward traps.",
			)
		}
	}

	// start logs-agent
	if config.C.LogsConfig.Enabled {
		if err := logs.Start(func() *autodiscovery.AutoConfig { return common.AC }); err != nil {
			klog.Error("Could not start logs-agent: ", err)
		}
	} else {
		klog.Info("logs-agent disabled")
	}

	//if err = common.SetupSystemProbeConfig(sysProbeConfFilePath); err != nil {
	//	klog.Infof("System probe config not found, disabling pulling system probe info in the status page: %v", err)
	//}

	// Detect Cloud Provider
	go ddutil.DetectCloudProvider()

	// Append version and timestamp to version history log file if this Agent is different than the last run version
	ddutil.LogVersionHistory()

	// create and setup the Autoconfig instance
	common.LoadComponents(p.config)
	// start the autoconfig, this will immediately run any configured check
	common.StartAutoConfig()

	// check for common misconfigurations and report them to log
	misconfig.ToLog()

	// setup the metadata collector
	common.MetadataScheduler = metadata.NewScheduler(s)
	if err := metadata.SetupMetadataCollection(common.MetadataScheduler, metadata.AllDefaultCollectors); err != nil {
		return err
	}

	if err := metadata.SetupInventories(common.MetadataScheduler, common.AC, common.Coll); err != nil {
		return err
	}

	return nil
}

func (p *module) stop(ops *proc.HookOps) error {
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
	//api.StopServer()
	jmx.StopJmxfetch()
	aggregator.StopDefaultAggregator()
	if common.Forwarder != nil {
		common.Forwarder.Stop()
	}
	//if orchestratorForwarder != nil {
	//	orchestratorForwarder.Stop()
	//}

	logs.Stop()
	//gui.StopGUIServer()
	//profiler.Stop()

	os.Remove(p.config.PidfilePath)
	klog.Info("See ya!")

	return nil
}

func completionHandler(transaction *transaction.HTTPTransaction, statusCode int, body []byte, err error) {
	if len(body) > 0 {
		klog.Info("err %v code %d body %s", err, statusCode, string(body))
	}
}

func init() {
	proc.RegisterHooks(hookOps)
}
