package demo

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/check"
	core "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks"
	"k8s.io/klog"
	"sigs.k8s.io/yaml"
)

// y = cos(2*pi*x/3600)

const checkName = "demo"

type InitConfig struct {
}

type InstanceConfig struct {
	Period int `json:"period"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

type checkConfig struct {
	InstanceConfig
	InitConfig
}

func defaultInstanceConfig() InstanceConfig {
	return InstanceConfig{
		Period: 3600,
		Count:  8,
	}
}

func buildConfig(rawInstance integration.Data, rawInitConfig integration.Data) (*checkConfig, error) {
	//initConfig := InitConfig{}
	instance := defaultInstanceConfig()

	//err := yaml.Unmarshal(rawInitConfig, &initConfig)
	//if err != nil {
	//	return nil, err
	//}

	if err := yaml.Unmarshal(rawInstance, &instance); err != nil {
		return nil, err
	}

	return &checkConfig{
		//InitConfig:     initConfig,
		InstanceConfig: instance,
	}, nil
}

type Check struct {
	core.CheckBase

	count int
	cos   *cos
}

// Run executes the check
func (c *Check) Run() error {
	klog.V(6).Infof("entering Run()")
	sender, err := aggregator.GetSender(c.ID())
	if err != nil {
		return err
	}

	for i := 0; i < c.count; i++ {
		sender.Gauge("demo", c.cos.value(i), "", []string{"n:" + strconv.Itoa(i)})
	}

	sender.Commit()
	return nil
}

func (c *Check) Cancel() {
	defer c.CheckBase.Cancel()
	klog.V(6).Infof("see ya!!")
}

// Configure the Prom check
func (c *Check) Configure(rawInstance integration.Data, rawInitConfig integration.Data, source string) (err error) {
	// Must be called before c.CommonConfigure
	c.BuildID(rawInstance, rawInitConfig)

	if err = c.CommonConfigure(rawInstance, source); err != nil {
		return fmt.Errorf("common configure failed: %s", err)
	}

	config, err := buildConfig(rawInstance, rawInitConfig)
	if err != nil {
		return fmt.Errorf("build config failed: %s", err)
	}

	c.count = config.InstanceConfig.Count
	period := float64(config.InstanceConfig.Period)
	offset := float64(config.InstanceConfig.Offset)
	if c.count <= 0 {
		c.count = 8
	}

	c.cos = &cos{
		period: period,
		unit:   period / float64(c.count),
		offset: offset,
	}

	return nil
}

type cos struct {
	period float64
	unit   float64
	offset float64
}

func (c *cos) value(i int) float64 {
	return math.Cos(2 * math.Pi * (float64(time.Now().Unix()) + c.unit*float64(i) + c.offset) / c.period)
}

func checkFactory() check.Check {
	return &Check{
		CheckBase: core.NewCheckBase(checkName),
	}
}
func init() {
	core.RegisterCheck(checkName, checkFactory)
}
