package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/providers"
	"k8s.io/klog/v2"
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
	c.rules = rs
	c.latestUpdatedAt = time.Now().Unix()
	klog.Infof("rules %d", len(rs))
	c.Unlock()
}

func (c *CollectRules) GetRules() []api.CollectRule {
	c.RLock()
	defer c.RUnlock()

	return c.rules
}

func (c *CollectRules) GetSummary() api.CollectRulesSummary {
	c.RLock()
	defer c.RUnlock()

	return api.CollectRulesSummary{
		LatestUpdatedAt: c.latestUpdatedAt,
		Total:           len(c.rules),
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

	buf, err := json.Marshal(config)
	if err != nil {
		return rule, err
	}

	rule.Data = string(buf)

	return rule, nil
}
