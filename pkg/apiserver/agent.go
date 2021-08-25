// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package agent implements the api endpoints for the `/agent` prefix.
// This group of endpoints is meant to provide high-level functionalities
// at the agent level.
package apiserver

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/DataDog/datadog-agent/pkg/autodiscovery"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/flare"
	"github.com/DataDog/datadog-agent/pkg/logs"
	"github.com/DataDog/datadog-agent/pkg/logs/diagnostic"
	"github.com/DataDog/datadog-agent/pkg/secrets"
	"github.com/DataDog/datadog-agent/pkg/status"
	"github.com/DataDog/datadog-agent/pkg/status/health"
	"github.com/DataDog/datadog-agent/pkg/tagger"
	"github.com/DataDog/datadog-agent/pkg/tagger/collectors"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/n9e/n9e-agentd/cmd/agent/common"
	"github.com/n9e/n9e-agentd/pkg/apiserver/response"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/config/settings"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/yubo/apiserver/pkg/request"
	"github.com/yubo/apiserver/pkg/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

func (p *module) installWs(c rest.GoRestfulContainer) {
	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/api/v1",
		GoRestfulContainer: c,
		Routes: []rest.WsRoute{
			{Method: "GET", SubPath: "/version", Handle: getVersion, Desc: "get version"},
			{Method: "GET", SubPath: "/hostname", Handle: getHostname, Desc: "get hostname"},
			{Method: "POST", SubPath: "/flare", Handle: makeFlare, Desc: "make flare"},
			{Method: "POST", SubPath: "/stop", Handle: stopAgent, Desc: "stop agent"},
			{Method: "GET", SubPath: "/status", Handle: getStatus, Desc: "get status"},
			{Method: "POST", SubPath: "/stream-logs", Handle: streamLogs, Desc: "post stream logs"},
			{Method: "GET", SubPath: "/statsd-stats", Handle: getDogstatsdStats, Desc: "get statsd stats"},
			{Method: "GET", SubPath: "/status/formatted", Handle: getFormattedStatus, Desc: "get formatted status"},
			{Method: "GET", SubPath: "/status/health", Handle: getHealth, Desc: "get health"},
			{Method: "GET", SubPath: "/py/status", Handle: getPythonStatus, Desc: "get python status"},
			{Method: "POST", SubPath: "/jmx/status", Handle: setJMXStatus, Desc: "set jmx status"},
			{Method: "GET", SubPath: "/jmx/configs", Handle: getJMXConfigs, Desc: "get jmx configs"},
			//{Method: "GET", SubPath: "/gui/csrf-token", Handle: nonHandle, Desc: "flare"},
			{Method: "GET", SubPath: "/config-check", Handle: getConfigCheck, Desc: "get config check"},
			{Method: "GET", SubPath: "/config", Handle: getFullRuntimeConfig, Desc: "get full runtime config"},
			{Method: "GET", SubPath: "/config/list-runtime", Handle: getRuntimeConfigurableSettings, Desc: "get runtime configure able settings"},
			{Method: "GET", SubPath: "/config/{setting}", Handle: getRuntimeConfig, Desc: "get runtime config"},
			{Method: "POST", SubPath: "/config/{setting}", Handle: setRuntimeConfig, Desc: "set runtime config"},
			{Method: "GET", SubPath: "/tagger-list", Handle: getTaggerList, Desc: "get tagger list"},
			{Method: "GET", SubPath: "/secrets", Handle: secretInfo, Desc: "get secrets info"},
		},
	})
}

func stopAgent(w http.ResponseWriter, r *http.Request) {
	util.Stop()
}

func getVersion(w http.ResponseWriter, r *http.Request) (string, error) {
	return options.Version, nil
}

func getHostname(w http.ResponseWriter, r *http.Request) (string, error) {
	return config.C.Hostname, nil
}

func nonHandle(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func makeFlare(w http.ResponseWriter, r *http.Request, _ *rest.NoneParam, profile flare.ProfileData) (string, error) {
	klog.Infof("Making a flare")
	return flare.CreateArchive(false, config.C.DistPath, config.C.PyChecksPath, []string{}, profile)
}

func getStatus(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	klog.Info("Got a request for the status. Making status.")
	return status.GetStatus()
}

func streamLogs(w http.ResponseWriter, r *http.Request, _ *rest.NoneParam, filters *diagnostic.Filters) error {
	klog.Info("Got a request for stream logs.")
	w.Header().Set("Transfer-Encoding", "chunked")

	logMessageReceiver := logs.GetMessageReceiver()

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("Expected a Flusher type, got: %v", w)
	}

	if logMessageReceiver == nil {
		flusher.Flush()
		klog.Info("Logs agent is not running - can't stream logs")
		return fmt.Errorf("The logs agent is not running")
	}

	if !logMessageReceiver.SetEnabled(true) {
		flusher.Flush()
		klog.Info("Logs are already streaming. Dropping connection.")
		return fmt.Errorf("Another client is already streaming logs.")
	}
	defer logMessageReceiver.SetEnabled(false)

	conn, _ := request.ConnFrom(r.Context())

	// Override the default server timeouts so the connection never times out
	_ = conn.SetDeadline(time.Time{})
	_ = conn.SetWriteDeadline(time.Time{})

	done := make(chan struct{})
	defer close(done)
	logChan := logMessageReceiver.Filter(filters, done)
	flushTimer := time.NewTicker(time.Second)
	for {
		// Handlers for detecting a closed connection (from either the server or client)
		select {
		case <-w.(http.CloseNotifier).CloseNotify():
			return nil
		case <-r.Context().Done():
			return nil
		case line := <-logChan:
			fmt.Fprint(w, line)
		case <-flushTimer.C:
			// The buffer will flush on its own most of the time, but when we run out of logs flush so the client is up to date.
			flusher.Flush()
		}
	}
}

func getDogstatsdStats(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	klog.Info("Got a request for the statsd stats.")

	if !config.C.Statsd.Enabled {
		return nil, fmt.Errorf("statsd not enabled in the Agent configuration")
	}

	if !config.C.Statsd.MetricsStatsEnable {
		return nil, fmt.Errorf("tatsd metrics stats not enabled in the Agent configuration")
	}

	// Weird state that should not happen: dogstatsd is enabled
	// but the server has not been successfully initialized.
	// Return no data.
	if common.DSD == nil {
		return []byte("{}"), nil
	}

	jsonStats, err := common.DSD.GetJSONDebugStats()
	if err != nil {
		return nil, fmt.Errorf("Error getting marshalled tatsd stats: %s", err)
	}

	return jsonStats, nil
}

func getFormattedStatus(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	klog.Info("Got a request for the formatted status. Making formatted status.")
	s, err := status.GetAndFormatStatus()
	if err != nil {
		return nil, fmt.Errorf("Error getting status. Error: %v, Status: %v", err, s)
	}

	return s, nil
}

func getHealth(w http.ResponseWriter, r *http.Request) (*health.Status, error) {
	h := health.GetReady()

	if len(h.Unhealthy) > 0 {
		klog.V(5).Infof("Healthcheck failed on: %v", h.Unhealthy)
	}

	return &h, nil
}

//func getCSRFToken(w http.ResponseWriter, r *http.Request) {
//	w.Write([]byte(gui.CsrfToken))
//}

func getConfigCheck(w http.ResponseWriter, r *http.Request) (*response.ConfigCheckResponse, error) {
	if common.AC == nil {
		return nil, fmt.Errorf("Trying to use /config-check before the agent has been initialized.")
	}

	configs := common.AC.GetLoadedConfigs()
	configSlice := make([]integration.Config, 0)
	for _, config := range configs {
		configSlice = append(configSlice, config)
	}
	sort.Slice(configSlice, func(i, j int) bool {
		return configSlice[i].Name < configSlice[j].Name
	})

	return &response.ConfigCheckResponse{
		Configs:         configSlice,
		ResolveWarnings: autodiscovery.GetResolveWarnings(),
		ConfigErrors:    autodiscovery.GetConfigErrors(),
		Unresolved:      common.AC.GetUnresolvedTemplates(),
	}, nil
}

func getFullRuntimeConfig(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	runtimeConfig, err := yaml.Marshal(config.C)
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal runtime config response: %s", err)
	}

	scrubbed, err := log.CredentialsCleanerBytes(runtimeConfig)
	if err != nil {
		return nil, fmt.Errorf("Unable to scrub sensitive data from runtime config: %s", err)
	}

	return scrubbed, nil
}

func getRuntimeConfigurableSettings(w http.ResponseWriter, r *http.Request) (map[string]settings.RuntimeSettingResponse, error) {
	configurableSettings := make(map[string]settings.RuntimeSettingResponse)
	for name, setting := range settings.RuntimeSettings() {
		configurableSettings[name] = settings.RuntimeSettingResponse{
			Description: setting.Description(),
			Hidden:      setting.Hidden(),
		}
	}
	return configurableSettings, nil
}

type SettingInput struct {
	Setting string `param:"path"`
	Value   string `param:"query"`
}

func getRuntimeConfig(w http.ResponseWriter, r *http.Request, in *SettingInput) (interface{}, error) {
	klog.Infof("Got a request to read a setting value: %s", in.Setting)
	return settings.GetRuntimeSetting(in.Setting)
}

func setRuntimeConfig(w http.ResponseWriter, r *http.Request, in *SettingInput) error {
	klog.Infof("Got a request to change a setting: %s", in.Setting)

	return settings.SetRuntimeSetting(in.Setting, in.Value)
}

func getTaggerList(w http.ResponseWriter, r *http.Request) (response.TaggerListResponse, error) {
	// query at the highest cardinality between checks and dogstatsd cardinalities
	cardinality := collectors.TagCardinality(max(int(tagger.ChecksCardinality), int(tagger.DogstatsdCardinality)))
	return tagger.List(cardinality), nil
}

func secretInfo(w http.ResponseWriter, r *http.Request) (*secrets.SecretInfo, error) {
	return secrets.GetDebugInfo()
}

// max returns the maximum value between a and b.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
