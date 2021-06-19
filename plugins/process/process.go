package process

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/n9e/n9e-agentd/plugins/process/checks"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/check"
	core "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks"
	"github.com/n9e/n9e-agentd/pkg/process/config"
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
	// Set to 1 if enabled 0 is not. We're using an integer
	// so we can use the sync/atomic for thread-safe access.
	realTimeEnabled int32

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
}

// NewCollector creates a new Collector
func NewCollector(cfg *config.AgentConfig) (*Collector, error) {
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
		realTimeEnabled:  0,
		ctx:              ctx,
		cancel:           cancel,
	}, nil
}

func (p *Collector) init() {
	//TODO: check global pro
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

type InitConfig struct {
}

type InstanceConfig struct {
	Name          string `json:"name"`          // no used
	CollectMethod string `json:"collectMethod"` // enum: cmdline, cmd
	Target        string `json:"target"`        //
	Comment       string `json:"comment"`       // no used
	//CollectType   string         `json:"collect_type"`
	//Id            int64          `json:"id"`
	//Nid           int64          `json:"nid"`
	//Tags          string         `json:"tags"`
	//Step          int            `json:"step"`
	//Creator       string         `json:"creator"`
	//Created       time.Time      `json:"created"`
	//LastUpdator   string         `json:"last_updator"`
	//LastUpdated   time.Time      `json:"last_updated"`
	//ProcJiffy     map[int]uint64 `json:"-"`
	//Jiffy         uint64         `json:"-"`
	//RBytes        map[int]uint64 `json:"-"`
	//WBytes        map[int]uint64 `json:"-"`

	InitConfig `json:"-"`
}

type checkConfig struct {
	InstanceConfig
	InitConfig
}

func (p checkConfig) String() string {
	return util.Prettify(p)
}

func defaultInstanceConfig() InstanceConfig {
	return InstanceConfig{}
}

func buildConfig(rawInstance integration.Data, rawInitConfig integration.Data) (checkConfig, error) {
	instance := defaultInstanceConfig()
	initConfig := InitConfig{}

	err := yaml.Unmarshal(rawInitConfig, &initConfig)
	if err != nil {
		return checkConfig{}, err
	}

	err = yaml.Unmarshal(rawInstance, &instance)
	if err != nil {
		return checkConfig{}, err
	}

	return checkConfig{
		InitConfig:     initConfig,
		InstanceConfig: instance,
	}, nil
}

// Check doesn't need additional fields
type Check struct {
	core.CheckBase
	config checkConfig
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

// Configure the Prom check
func (c *Check) Configure(rawInstance integration.Data, rawInitConfig integration.Data, source string) error {
	if collector == nil {
		var err error
		if collector, err = NewCollector(config.NewDefaultAgentConfig(false)); err != nil {
			return err
		}
		collector.init()
	}

	// Must be called before c.CommonConfigure
	c.BuildID(rawInstance, rawInitConfig)

	err := c.CommonConfigure(rawInstance, source)
	if err != nil {
		return fmt.Errorf("common configure failed: %s", err)
	}

	config, err := buildConfig(rawInstance, rawInitConfig)
	if err != nil {
		return fmt.Errorf("build config failed: %s", err)
	}

	c.config = config
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
