package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/autodiscovery/providers"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

var (
	rules         CollectRules
	flushInterval = time.Second * 5
)

type CollectRules struct {
	sync.RWMutex
	rules           []api.CollectRule
	latestUpdatedAt int64
}

func (c *CollectRules) Set(configs []integration.Config) {
	var rs []api.CollectRule
	for _, config := range configs {
		r, err := configToRule(config)
		if err != nil {
			klog.Errorf("config to rule err %s", err)
		}
		rs = append(rs, r)
	}

	c.Lock()
	defer c.Unlock()

	c.rules = rs
	c.latestUpdatedAt = time.Now().Unix()
	klog.Infof("rules %d", len(rs))
}

type RulesPayload struct {
	Data []api.CollectRule `json:"dat"`
	Err  string            `json:"err"`
}

func (c *CollectRules) GetRules() api.CollectRuleWrap {
	c.RLock()
	defer c.RUnlock()

	return api.CollectRuleWrap{Data: c.rules}
}

func (c *CollectRules) GetSummary() api.CollectRulesSummaryWrap {
	c.RLock()
	defer c.RUnlock()

	return api.CollectRulesSummaryWrap{
		Data: api.CollectRulesSummary{
			LatestUpdatedAt: c.latestUpdatedAt,
			Total:           len(c.rules),
		},
	}
}

func installCollectRules(confd string) {
	http.HandleFunc(api.RoutePathGetCollectRules, getCollectRules)
	http.HandleFunc(api.RoutePathGetCollectRulesSummary, getCollectRulesSummary)

	go startCollectFileConfig(confd)
}

func startCollectFileConfig(path string) {
	p := providers.NewFileConfigProvider([]string{path})
	t := time.NewTicker(flushInterval)

	for {
		<-t.C

		configs, err := p.Collect()
		if err != nil {
			klog.Errorf("collect err %s", err)
			continue
		}
		for _, c := range configs {
			for _, i := range c.Instances {
				klog.V(6).Infof("%s", string(i))
			}
		}
		rules.Set(configs)
	}
}

func getCollectRules(w http.ResponseWriter, _ *http.Request) {
	writeRawJSON(rules.GetRules(), w)
}

func getCollectRulesSummary(w http.ResponseWriter, _ *http.Request) {
	writeRawJSON(rules.GetSummary(), w)
}

func writeRawJSON(object interface{}, w http.ResponseWriter) {
	output, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

func configToRule(config integration.Config) (rule api.CollectRule, err error) {
	rule = api.CollectRule{
		Name: config.Name + "-name",
		Type: config.Name,
	}

	c := api.ConfigFormat{
		ADIdentifiers:           config.ADIdentifiers,
		ClusterCheck:            config.ClusterCheck,
		IgnoreAutodiscoveryTags: config.IgnoreAutodiscoveryTags,
	}

	yaml.Unmarshal(config.InitConfig, &c.InitConfig) //nolint:errcheck

	var instances []interface{}
	for _, i := range config.Instances {
		var instance interface{}
		yaml.Unmarshal(i, &instance) //nolint:errcheck
		instances = append(instances, instance)
	}
	c.Instances = instances

	yaml.Unmarshal(config.LogsConfig, &c.LogsConfig) //nolint:errcheck

	buf, err := yaml.Marshal(c)
	if err != nil {
		klog.Error(err)
		return rule, err

	}
	buf, err = yaml.YAMLToJSON(buf)
	if err != nil {
		klog.Error(err)
		return rule, err
	}
	{
		klog.Infof("%s", string(buf))
		printInstances(c)
		var c api.ConfigFormat
		if err := json.Unmarshal(buf, &c); err != nil {
			klog.Errorf("ummarshal err %s", err)
		}
		printInstances(c)
	}

	rule.Data = string(buf)

	return rule, nil
}

func printInstances(config api.ConfigFormat) {
	for i, c := range config.Instances {
		klog.Infof("%d %v", i, c)
	}
}
