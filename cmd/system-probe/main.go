// +build linux windows

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	_ "net/http/pprof"

	"github.com/n9e/n9e-agentd/cmd/system-probe/config"
	"github.com/n9e/n9e-agentd/cmd/system-probe/modules"
	"github.com/n9e/n9e-agentd/cmd/system-probe/utils"
	ddconfig "github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/pidfile"
	"github.com/n9e/n9e-agentd/pkg/process/net"
	"github.com/n9e/n9e-agentd/pkg/process/statsd"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/serializer"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util"
	"k8s.io/klog/v2"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/profiling"
	"github.com/n9e/n9e-agentd/pkg/version"
)

// Flag values
var opts struct {
	configPath  string
	pidFilePath string
	debug       bool
	version     bool
	console     bool // windows only; execute on console rather than via SCM
}

const loggerName = ddconfig.LoggerName("SYS-PROBE")

func runAgent(exit <-chan struct{}) {
	// --version
	if opts.version {
		fmt.Println(versionString("\n"))
		cleanupAndExit(0)
	}

	if err := util.SetupCoreDump(); err != nil {
		klog.Warningf("Can't setup core dumps: %v, core dumps might not be available after a crash", err)
	}

	// --pid
	if opts.pidFilePath != "" {
		if err := pidfile.WritePID(opts.pidFilePath); err != nil {
			klog.Errorf("Error while writing PID file, exiting: %v", err)
			cleanupAndExit(1)
		}

		klog.Infof("pid '%d' written to pid file '%s'", os.Getpid(), opts.pidFilePath)

		defer func() {
			os.Remove(opts.pidFilePath)
		}()
	}

	cfg, err := config.New(opts.configPath)
	if err != nil {
		klog.Warningf("Failed to create agent config: %s", err)
		cleanupAndExit(1)
	}

	err = ddconfig.SetupLogger(
		loggerName,
		cfg.LogLevel,
		cfg.LogFile,
		ddconfig.GetSyslogURI(),
		ddconfig.Datadog.GetBool("syslog_rfc"),
		ddconfig.Datadog.GetBool("log_to_console"),
		ddconfig.Datadog.GetBool("log_format_json"),
	)
	if err != nil {
		klog.Warningf("failed to setup configured logger: %s", err)
		cleanupAndExit(1)
	}

	// Exit if system probe is disabled
	if cfg.ExternalSystemProbe || !cfg.Enabled {
		klog.Info("system probe not enabled. exiting.")
		gracefulExit()
	}

	if cfg.ProfilingEnabled {
		if err := enableProfiling(cfg); err != nil {
			klog.Warningf("failed to enable profiling: %s", err)
		}
		defer profiling.Stop()
	}

	klog.Infof("running system-probe with version: %s", versionString(", "))

	// configure statsd
	if err := statsd.Configure(cfg.StatsdHost, cfg.StatsdPort); err != nil {
		klog.Warningf("Error configuring statsd: %s", err)
		cleanupAndExit(1)
	}

	conn, err := net.NewListener(cfg.SocketAddress)
	if err != nil {
		klog.Warningf("Error creating IPC socket: %s", err)
		cleanupAndExit(1)
	}

	// if a debug port is specified, we expose the default handler to that port
	if cfg.DebugPort > 0 {
		go func() {
			err := http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.DebugPort), http.DefaultServeMux)
			if err != nil && err != http.ErrServerClosed {
				klog.Warningf("Error creating debug HTTP server: %v", err)
				cleanupAndExit(1)
			}
		}()
	}

	loader := NewLoader()
	defer loader.Close()

	httpMux := http.NewServeMux()

	err = loader.Register(cfg, httpMux, modules.All)
	if err != nil {
		loader.Close()
		klog.Warningf("failed to create system probe: %s", err)
		cleanupAndExit(1)
	}

	// Register stats endpoint
	httpMux.HandleFunc("/debug/stats", func(w http.ResponseWriter, req *http.Request) {
		stats := loader.GetStats()
		utils.WriteAsJSON(w, stats)
	})

	go func() {
		err = http.Serve(conn.GetListener(), httpMux)
		if err != nil {
			klog.Warningf("Error creating HTTP server: %s", err)
			loader.Close()
			cleanupAndExit(1)
		}
	}()

	klog.Infof("system probe successfully started")
	<-exit
}

func enableProfiling(cfg *config.Config) error {
	var site string
	v, _ := version.Agent()

	// check if TRACE_AGENT_URL is set, in which case, forward the profiles to the trace agent
	if traceAgentURL := os.Getenv("TRACE_AGENT_URL"); len(traceAgentURL) > 0 {
		site = fmt.Sprintf(profiling.ProfilingLocalURLTemplate, traceAgentURL)
	} else {
		site = fmt.Sprintf(profiling.ProfileURLTemplate, cfg.ProfilingSite)
		if cfg.ProfilingURL != "" {
			site = cfg.ProfilingURL
		}
	}

	return profiling.Start(
		cfg.ProfilingAPIKey,
		site,
		cfg.ProfilingEnvironment,
		"system-probe",
		fmt.Sprintf("version:%v", v),
	)
}

func gracefulExit() {
	// A sleep is necessary to ensure that supervisor registers this process as "STARTED"
	// If the exit is "too quick", we enter a BACKOFF->FATAL loop even though this is an expected exit
	// http://supervisord.org/subprocess.html#process-states
	time.Sleep(5 * time.Second)
	cleanupAndExit(0)
}

// versionString returns the version information filled in at build time
func versionString(sep string) string {
	addString := func(buf *bytes.Buffer, format string, args ...interface{}) {
		if len(args) > 0 && args[0] != "" {
			_, _ = fmt.Fprintf(buf, format, args...)
		}
	}

	av, _ := version.Agent()

	var buf bytes.Buffer
	addString(&buf, "Version: %s %s%s", av.GetNumberAndPre(), av.Meta, sep)
	addString(&buf, "Commit: %s%s", av.Commit, sep)
	addString(&buf, "Serialization Version: %s%s", serializer.AgentPayloadVersion, sep)
	addString(&buf, "Go Version: %s%s", runtime.Version(), sep)
	return buf.String()
}

// cleanupAndExit cleans all resources allocated by system-probe before calling
// os.Exit
func cleanupAndExit(status int) {
	// remove pidfile if set
	if opts.pidFilePath != "" {
		if _, err := os.Stat(opts.pidFilePath); err == nil {
			os.Remove(opts.pidFilePath)
		}
	}

	os.Exit(status)
}
