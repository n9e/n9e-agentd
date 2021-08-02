package process

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	model "github.com/n9e/agent-payload/process"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/process/config"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/n9e/n9e-agentd/plugins/process/checks"
	"github.com/DataDog/datadog-agent/pkg/aggregator"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	core "github.com/DataDog/datadog-agent/pkg/collector/corechecks"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const checkName = "process"

var collector *Collector

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

	procs := c.process.GetProcs()
	stats := c.rtProcess.GetStats()

	for _, pid := range c.filter.Pids {
		proc, ok := procs[pid]
		if !ok {
			continue
		}
		stat, ok := stats[pid]
		if !ok {
			continue
		}
		collectProc(sender, proc, stat, []string{"target:" + c.filter.Target})
	}

	sender.Commit()
	return nil
}

func collectProc(sender aggregator.Sender, proc *model.Process, stat *model.ProcessStat, tags []string) {
	sender.Count("proc.num", 1, "", tags)

	// uptime
	sender.Gauge("proc.uptime", float64(time.Now().Unix()-stat.CreateTime/1000), "", tags)
	sender.Gauge("proc.createtime", float64(stat.CreateTime)/1000, "", tags)

	// fd
	sender.Count("proc.open_fd_count", float64(stat.OpenFdCount), "", tags)

	if mem := stat.Memory; mem != nil {
		sender.Count("proc.mem.rss", float64(mem.Rss), "", tags)
		sender.Count("proc.mem.vms", float64(mem.Vms), "", tags)
		sender.Count("proc.mem.swap", float64(mem.Swap), "", tags)
		sender.Count("proc.mem.shared", float64(mem.Shared), "", tags)
		sender.Count("proc.mem.text", float64(mem.Text), "", tags)
		sender.Count("proc.mem.lib", float64(mem.Lib), "", tags)
		sender.Count("proc.mem.data", float64(mem.Data), "", tags)
		sender.Count("proc.mem.dirty", float64(mem.Dirty), "", tags)
	}

	if cpu := stat.Cpu; cpu != nil {
		sender.Count("proc.cpu.total", float64(cpu.TotalPct), "", tags)
		sender.Count("proc.cpu.user", float64(cpu.UserPct), "", tags)
		sender.Count("proc.cpu.sys", float64(cpu.SystemPct), "", tags)
		sender.Count("proc.cpu.threads", float64(cpu.NumThreads), "", tags)
	}

	if io := stat.IoStat; io != nil {
		sender.Count("proc.io.read_rate", float64(io.ReadRate), "", tags)
		sender.Count("proc.io.write_rate", float64(io.WriteRate), "", tags)
		sender.Count("proc.io.readbytes_rate", float64(io.ReadBytesRate), "", tags)
		sender.Count("proc.io.writebytes_rate", float64(io.WriteBytesRate), "", tags)
	}
	if net := stat.Networks; net != nil {
		sender.Count("proc.net.conn_rate", float64(net.ConnectionRate), "", tags)
		sender.Count("proc.net.bytes_rate", float64(net.BytesRate), "", tags)
	}
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

func init() {
	core.RegisterCheck(checkName, checkFactory)
}
