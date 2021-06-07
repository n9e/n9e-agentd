package api

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/info"
)

// makeInfoHandler returns a new handler for handling the discovery endpoint.
func (r *HTTPReceiver) makeInfoHandler() (hash string, handler http.HandlerFunc) {
	var all []string
	for _, e := range endpoints {
		if !e.Hidden {
			all = append(all, e.Pattern)
		}
	}
	type reducedObfuscationConfig struct {
		ElasticSearch        bool                         `json:"elasticSearch"`
		Mongo                bool                         `json:"mongo"`
		SQLExecPlan          bool                         `json:"sqlExecPlan"`
		SQLExecPlanNormalize bool                         `json:"sqlExecPlanNormalize"`
		HTTP                 config.HTTPObfuscationConfig `json:"http"`
		RemoveStackTraces    bool                         `json:"removeStackTraces"`
		Redis                bool                         `json:"redis"`
		Memcached            bool                         `json:"memcached"`
	}
	type reducedConfig struct {
		DefaultEnv             string                        `json:"defaultEnv"`
		TargetTPS              float64                       `json:"targetTps"`
		MaxEPS                 float64                       `json:"maxEps"`
		ReceiverPort           int                           `json:"receiverPort"`
		ReceiverSocket         string                        `json:"receiverSocket"`
		ConnectionLimit        int                           `json:"connectionLimit"`
		ReceiverTimeout        int                           `json:"receiverTimeout"`
		MaxRequestBytes        int64                         `json:"maxRequestBytes"`
		StatsdPort             int                           `json:"statsdPort"`
		MaxMemory              float64                       `json:"maxMemory"`
		MaxCPU                 float64                       `json:"maxCpu"`
		AnalyzedSpansByService map[string]map[string]float64 `json:"analyzedSpansByService"`
		Obfuscation            reducedObfuscationConfig      `json:"obfuscation"`
	}
	var oconf reducedObfuscationConfig
	if o := r.conf.Obfuscation; o != nil {
		oconf.ElasticSearch = o.ES.Enabled
		oconf.Mongo = o.Mongo.Enabled
		oconf.SQLExecPlan = o.SQLExecPlan.Enabled
		oconf.SQLExecPlanNormalize = o.SQLExecPlanNormalize.Enabled
		oconf.HTTP = o.HTTP
		oconf.RemoveStackTraces = o.RemoveStackTraces
		oconf.Redis = o.Redis.Enabled
		oconf.Memcached = o.Memcached.Enabled
	}
	txt, err := json.MarshalIndent(struct {
		Version       string        `json:"version"`
		GitCommit     string        `json:"gitCommit"`
		BuildDate     string        `json:"buildDate"`
		Endpoints     []string      `json:"endpoints"`
		FeatureFlags  []string      `json:"featureFlags,omitempty"`
		ClientDropP0s bool          `json:"clientDropP0s"`
		Config        reducedConfig `json:"config"`
	}{
		Version:       info.Version,
		GitCommit:     info.GitCommit,
		BuildDate:     info.BuildDate,
		Endpoints:     all,
		FeatureFlags:  config.Features(),
		ClientDropP0s: true,
		Config: reducedConfig{
			DefaultEnv:             r.conf.DefaultEnv,
			TargetTPS:              r.conf.TargetTPS,
			MaxEPS:                 r.conf.MaxEPS,
			ReceiverPort:           r.conf.ReceiverPort,
			ReceiverSocket:         r.conf.ReceiverSocket,
			ConnectionLimit:        r.conf.ConnectionLimit,
			ReceiverTimeout:        r.conf.ReceiverTimeout,
			MaxRequestBytes:        r.conf.MaxRequestBytes,
			StatsdPort:             r.conf.StatsdPort,
			MaxMemory:              r.conf.MaxMemory,
			MaxCPU:                 r.conf.MaxCPU,
			AnalyzedSpansByService: r.conf.AnalyzedSpansByService,
			Obfuscation:            oconf,
		},
	}, "", "\t")
	if err != nil {
		panic(fmt.Errorf("Error making /info handler: %v", err))
	}
	h := sha256.Sum256(txt)
	return fmt.Sprintf("%x", h), func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "%s", txt)
	}
}
