package apiserver

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/DataDog/datadog-agent/pkg/autodiscovery"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	dsdReplay "github.com/DataDog/datadog-agent/pkg/dogstatsd/replay"
	"github.com/DataDog/datadog-agent/pkg/flare"
	"github.com/DataDog/datadog-agent/pkg/logs/diagnostic"
	pb "github.com/DataDog/datadog-agent/pkg/proto/pbgo"
	"github.com/DataDog/datadog-agent/pkg/secrets"
	"github.com/DataDog/datadog-agent/pkg/status"
	"github.com/DataDog/datadog-agent/pkg/status/health"
	"github.com/DataDog/datadog-agent/pkg/tagger"
	"github.com/DataDog/datadog-agent/pkg/tagger/collectors"
	"github.com/DataDog/datadog-agent/pkg/tagger/replay"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/n9e/n9e-agentd/cmd/agent/common"
	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/n9e/n9e-agentd/pkg/apiserver/response"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/config/settings"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/yubo/apiserver/pkg/handlers"
	"github.com/yubo/apiserver/pkg/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

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

func makeFlare(w http.ResponseWriter, r *http.Request, _ *rest.NonParam, profile *flare.ProfileData) (string, error) {
	klog.Infof("Making a flare")
	return flare.CreateArchive(false, config.C.DistPath, config.C.PyChecksPath, []string{}, *profile)
}

func getStatus(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	klog.Info("Got a request for the status. Making status.")
	return status.GetStatus()
}

func streamLogs(w http.ResponseWriter, r *http.Request, filters *diagnostic.Filters) error {
	watcher, err := newLogsWatch(filters)
	if err != nil {
		return err
	}
	watcher.Start()

	return handlers.ServeWatch(watcher, r, w, 0)
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

func getConfigCheck(w http.ResponseWriter, r *http.Request, in *api.QueryInput) (*response.ConfigCheckResponse, error) {
	if common.AC == nil {
		return nil, fmt.Errorf("Trying to use /config-check before the agent has been initialized.")
	}

	configs := common.AC.GetLoadedConfigs()
	configSlice := make([]integration.Config, 0)
	if in.Query == "" {
		for _, config := range configs {
			configSlice = append(configSlice, config)
		}
	} else {
		for _, config := range configs {
			if strings.Contains(config.Name, in.Query) {
				configSlice = append(configSlice, config)
			}
		}
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

func getRuntimeConfig(w http.ResponseWriter, r *http.Request, in *api.SettingInput) (interface{}, error) {
	klog.Infof("Got a request to read a setting value: %s", in.Setting)
	return settings.GetRuntimeSetting(in.Setting)
}

func setRuntimeConfig(w http.ResponseWriter, r *http.Request, in *api.SettingInput) error {
	klog.Infof("Got a request to change a setting: %s", in.Setting)

	return settings.SetRuntimeSetting(in.Setting, in.Value)
}

func getTaggerList(w http.ResponseWriter, r *http.Request) (*response.TaggerListResponse, error) {
	// query at the highest cardinality between checks and dogstatsd cardinalities
	cardinality := collectors.TagCardinality(max(int(tagger.ChecksCardinality), int(tagger.DogstatsdCardinality)))
	resp := tagger.List(cardinality)
	return &resp, nil
}

func secretInfo(w http.ResponseWriter, r *http.Request) (*secrets.SecretInfo, error) {
	return secrets.GetDebugInfo()
}

func statsdCaptureTrigger(w http.ResponseWriter, r *http.Request, _ *rest.NonParam, input *api.StatsdCaptureTriggerInput) (*string, error) {
	err := common.DSD.Capture(time.Second*time.Duration(input.Duration), input.Compressed)
	if err != nil {
		return nil, err
	}

	// wait for the capture to start
	for !common.DSD.TCapture.IsOngoing() {
		time.Sleep(500 * time.Millisecond)
	}

	path, err := common.DSD.TCapture.Path()
	if err != nil {
		return nil, err
	}

	return &path, nil
}

func statsdSetTaggerStatus(w http.ResponseWriter, r *http.Request, _ *rest.NonParam, req *pb.TaggerState) (*pb.TaggerStateResponse, error) {
	// Reset and return if no state pushed
	if req == nil || req.State == nil {
		log.Debugf("API: empty request or state")
		tagger.ResetCaptureTagger()
		dsdReplay.SetPidMap(nil)
		return &pb.TaggerStateResponse{Loaded: false}, nil
	}

	// FiXME: we should perhaps lock the capture processing while doing this...
	t := replay.NewTagger()
	if t == nil {
		return nil, fmt.Errorf("unable to instantiate state")
	}
	t.LoadState(req.State)

	log.Debugf("API: setting capture state tagger")
	tagger.SetCaptureTagger(t)
	dsdReplay.SetPidMap(req.PidMap)

	log.Debugf("API: loaded state successfully")

	return &pb.TaggerStateResponse{Loaded: true}, nil
}

func unsupported(w http.ResponseWriter, r *http.Request) error {
	return fmt.Errorf("unsupported")
}
