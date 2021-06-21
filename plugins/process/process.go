package process

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/pkg/process/config"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/n9e/n9e-agentd/plugins/process/checks"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/check"
	core "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	checkName = "process"
)

var (
	collector *Collector
)

type Collector struct {
	groupID int32

	rtIntervalCh chan time.Duration
	cfg          *config.AgentConfig

	// counters for each type of check
	runCounters   sync.Map
	enabledChecks []checks.Check

	// Controls the real-time interval, can change live.
	realTimeInterval time.Duration

	ctx    context.Context
	cancel context.CancelFunc

	process   *checks.ProcessCheck
	rtProcess *checks.RTProcessCheck
}

func initCollector() (*Collector, error) {
	c, err := newCollector(config.NewDefaultAgentConfig(false))
	if err != nil {
		return nil, err
	}
	c.run()
	return c, nil
}

// newCollector creates a new Collector
func newCollector(cfg *config.AgentConfig) (*Collector, error) {
	sysInfo, err := checks.CollectSystemInfo(cfg)
	if err != nil {
		return nil, err
	}

	for _, c := range checks.All {
		c.Init(cfg, sysInfo)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Collector{
		rtIntervalCh:  make(chan time.Duration),
		cfg:           cfg,
		groupID:       rand.Int31(),
		enabledChecks: checks.All,

		// Defaults for real-time on start
		realTimeInterval: 2 * time.Second,
		ctx:              ctx,
		cancel:           cancel,
		process:          checks.Process,
		rtProcess:        checks.RTProcess,
	}, nil
}

func (p *Collector) run() {
	for _, c := range p.enabledChecks {
		go func(c checks.Check) {
			klog.Infof("process run %s", c.Name())
			if !c.RealTime() {
				p.runCheck(c)
			}

			ticker := time.NewTicker(p.cfg.CheckInterval(c.Name()))
			for {
				select {
				case <-ticker.C:
					p.runCheck(c)
				case <-p.ctx.Done():
					return
				}
			}

		}(c)
	}
}

func (p *Collector) runCheck(c checks.Check) {
	klog.V(11).Infof("process run check %s", c.Name())
	messages, err := c.Run(p.cfg, atomic.AddInt32(&p.groupID, 1))
	if err != nil {
		klog.Errorf("unable to run check %s %s", c.Name(), err)
		return
	}

	for _, s := range messages {
		klog.V(6).Infof("process run check %s %s", c.Name(), s.String())
	}
}

type InstanceConfig struct {
	checks.ProcessFilter `json:",inline"`
}

type checkConfig struct {
	InstanceConfig
}

func (p checkConfig) String() string {
	return util.Prettify(p)
}

func defaultInstanceConfig() InstanceConfig {
	return InstanceConfig{}
}

func buildConfig(rawInstance integration.Data, _ integration.Data) (*checkConfig, error) {
	instance := defaultInstanceConfig()

	if err := yaml.Unmarshal(rawInstance, &instance); err != nil {
		return nil, err
	}

	return &checkConfig{
		InstanceConfig: instance,
	}, nil
}

// Check doesn't need additional fields
type Check struct {
	core.CheckBase
	*Collector
	filter *checks.ProcessFilter
}

// Run executes the check
func (c *Check) Run() error {
	klog.V(6).Infof("entering Run()")
	sender, err := aggregator.GetSender(c.ID())
	if err != nil {
		return err
	}

	//checks.Process.GetProcs()
	checks.RTProcess.GetProcs()
	sender.Gauge("proc.process", 1, "", nil)
	sender.Commit()
	return nil
}

func (c *Check) Cancel() {
	defer c.CheckBase.Cancel()
	c.process.DelFilter(c.filter)
}

// Configure the Prom check
func (c *Check) Configure(rawInstance integration.Data, rawInitConfig integration.Data, source string) (err error) {
	if collector == nil {
		if collector, err = initCollector(); err != nil {
			return err
		}
	}
	c.Collector = collector

	// Must be called before c.CommonConfigure
	c.BuildID(rawInstance, rawInitConfig)

	if err = c.CommonConfigure(rawInstance, source); err != nil {
		return fmt.Errorf("common configure failed: %s", err)
	}

	config, err := buildConfig(rawInstance, rawInitConfig)
	if err != nil {
		return fmt.Errorf("build config failed: %s", err)
	}
	c.filter = &config.ProcessFilter

	c.process.AddFilter(c.filter)

	return nil
}

func checkFactory() check.Check {
	return &Check{
		CheckBase: core.NewCheckBase(checkName),
	}
}

func init() {
	core.RegisterCheck(checkName, checkFactory)
}
