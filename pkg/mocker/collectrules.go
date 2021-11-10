package mocker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/providers"
	"github.com/n9e/n9e-agentd/pkg/api"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

var (
	flushInterval = time.Second * 5
)

type CollectRules struct {
	sync.RWMutex
	rules           []api.CollectRule
	latestUpdatedAt int64
}

func (c *CollectRules) String() string {
	if c == nil {
		return "null"
	}

	return fmt.Sprintf("len: %d, lasted %s", len(c.rules), time.Unix(c.latestUpdatedAt, 0))
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

func (c *CollectRules) GetRules() []api.CollectRule {
	c.RLock()
	defer c.RUnlock()

	return c.rules
}

func (c *CollectRules) GetSummary() *api.CollectRulesSummary {
	c.RLock()
	defer c.RUnlock()

	return &api.CollectRulesSummary{
		LatestUpdatedAt: c.latestUpdatedAt,
		Total:           len(c.rules),
	}
}

func (p *mocker) installCollectRules() error {
	if !p.config.CollectRule {
		return nil
	}

	fp := providers.NewFileConfigProvider([]string{p.config.Confd})
	t := time.NewTicker(flushInterval)

	go func() {
		for {
			select {
			case <-p.ctx.Done():
				return
			case <-t.C:
				configs, err := fp.Collect(context.Background())
				if err != nil {
					klog.Errorf("collect err %s", err)
					continue
				}
				for _, c := range configs {
					for _, i := range c.Instances {
						klog.V(6).Infof("%s", string(i))
					}
				}
				p.rules.Set(configs)
			}
		}
	}()

	return nil
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
