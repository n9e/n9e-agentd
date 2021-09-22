package log

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/pkg/aggregator"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	core "github.com/DataDog/datadog-agent/pkg/collector/corechecks"
	"github.com/DataDog/datadog-agent/pkg/logs/auditor"
	"github.com/DataDog/datadog-agent/pkg/logs/config"
	"github.com/DataDog/datadog-agent/pkg/logs/input/file"
	"github.com/DataDog/datadog-agent/pkg/logs/message"
	"github.com/DataDog/datadog-agent/pkg/logs/pipeline"
	"github.com/DataDog/datadog-agent/pkg/status/health"
	coreConfig "github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/util"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	checkName               = "log"
	DefaultRegistryFilename = "registry-plugins-log.json"
	numWorkers              = 4
)

var (
	commitInterval = 5 * time.Second
	agent          *Agent
)

type InstanceConfig struct {
	MetricName   string            `json:"metric_name"`  //
	FilePath     string            `json:"file_path"`    //
	Pattern      string            `json:"pattern"`      //
	TagsPattern  map[string]string `json:"tags_pattern"` //
	Func         string            `json:"func"`         // count(c), histogram(h)
	Encoding     string            `json:"encoding"`     //
	ExcludePaths []string          `json:"exclude_path"` //
	TailingMode  string            `json:"tailing_mode"` //
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

// Check doesn't need additional fields
type Check struct {
	sync.RWMutex
	core.CheckBase
	config *checkConfig
	agent  *Agent
	source *config.LogSource

	tagRegs    map[string]*regexp.Regexp
	patternReg *regexp.Regexp
}

// Run executes the check
func (c *Check) Run() error {
	klog.V(6).Infof("entering Run id %s", string(c.ID()))
	c.agent.addCheck(c)
	return nil
}

func (c *Check) process(msg *message.Message) (err error) {
	klog.V(6).Infof("entering process")
	line := string(msg.Content)
	cf := c.config

	var value float64
	if m := c.patternReg.FindStringSubmatch(line); len(m) == 0 || (cf.Func == "histogram" && len(m) < 2) {
		return nil
	} else if len(m) >= 2 {
		if value, err = strconv.ParseFloat(m[1], 64); err != nil {
			klog.V(6).Infof("parsefloat err %s", err)
			return err
		}
	}
	//处理tag 正则
	tags := []string{}
	for k, v := range c.config.TagsPattern {
		var regTag *regexp.Regexp
		regTag, ok := c.tagRegs[k]
		if !ok {
			klog.V(6).Infof("leaving process")
			return fmt.Errorf("get tag reg error %s:%s", k, v)
		}
		if m := regTag.FindStringSubmatch(line); len(m) < 2 {
			klog.V(6).Infof("leaving process")
			return nil
		} else {
			tags = append(tags, k+":"+m[1])
		}
	}
	sort.Strings(tags)

	sender, err := aggregator.GetSender(c.ID())
	if err != nil {
		klog.V(6).Infof("leaving process")
		return err
	}

	switch cf.Func {
	case "count", "c", "cnt":
		klog.V(6).Infof(`sender.Count(%s, 1, "", %v)`, cf.MetricName, tags)
		sender.Count(cf.MetricName, 1, "", tags)
	case "histogram", "h":
		klog.V(6).Infof(`sender.Histogram(%s, %f, "", %v)`, cf.MetricName, value, tags)
		sender.Histogram(cf.MetricName, value, "", tags)
	}
	klog.V(6).Infof("leaving process")
	return nil
}

func (c *Check) Cancel() {
	defer c.CheckBase.Cancel()
	c.agent.removeCheck(c)
}

// Configure the Prom check
func (c *Check) Configure(rawInstance integration.Data, rawInitConfig integration.Data, source string) error {
	// init & start logs-agent
	if agent == nil {
		if a, err := NewAgent(); err != nil {
			return err
		} else {
			agent = a
		}
		agent.start()
	}

	// Must be called before c.CommonConfigure
	c.BuildID(rawInstance, rawInitConfig)

	err := c.CommonConfigure(rawInstance, source)
	if err != nil {
		return fmt.Errorf("common configure failed: %s", err)
	}

	cf, err := buildConfig(rawInstance, rawInitConfig)
	if err != nil {
		return fmt.Errorf("build config failed: %s", err)
	}

	c.config = cf
	c.agent = agent

	if len(cf.Pattern) == 0 {
		return fmt.Errorf("pattern and exclude are all empty")
	}

	if c.patternReg, err = regexp.Compile(cf.Pattern); err != nil {
		return fmt.Errorf("compile pattern regexp failed %s %v", cf.Pattern, err)
	}

	for k, v := range cf.TagsPattern {
		reg, err := regexp.Compile(v)
		if err != nil {
			return fmt.Errorf("compile tag failed %s %v", v, err)
		}
		c.tagRegs[k] = reg
	}

	return nil
}

func buildConfig(rawInstance integration.Data, rawInitConfig integration.Data) (*checkConfig, error) {
	instance := defaultInstanceConfig()

	if err := yaml.Unmarshal(rawInstance, &instance); err != nil {
		return nil, err
	}

	return &checkConfig{
		InstanceConfig: instance,
	}, nil
}

type Agent struct {
	sync.RWMutex
	msgCh   chan *message.Message
	scanner *file.Scanner
	sources *config.LogSources
	health  *health.Handle
	auditor auditor.Auditor
	ctx     context.Context
	cancel  context.CancelFunc
	workers int
	checks  map[string]*Check
	files   map[string]map[string]*Check

	fn string
}

func (a *Agent) start() {
	a.scanner.Start()

	for i := 0; i < numWorkers; i++ {
		a.addWork()
	}
	go a.flush()
}

func (a *Agent) flush() {
	t := time.NewTicker(commitInterval)

	for {
		select {
		case <-t.C:
			a.RLock()
			for _, c := range a.checks {
				sender, err := aggregator.GetSender(c.ID())
				if err != nil {
					continue
				}
				sender.Commit()
			}
			a.RUnlock()
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *Agent) stop() {
	a.scanner.Stop()
	a.cancel()
}

func (a *Agent) addWork() {
	go a.work(a.workers)
	a.workers++
}

func (a *Agent) addCheck(c *Check) {
	a.Lock()
	defer a.Unlock()

	cf := c.config
	id := string(c.ID())
	if _, ok := a.checks[id]; ok {
		return
	}

	if _, ok := a.files[cf.FilePath]; !ok {
		a.files[cf.FilePath] = make(map[string]*Check)
	}
	a.files[cf.FilePath][id] = c
	a.checks[id] = c

	c.source = config.NewLogSource(id, &config.LogsConfig{
		Type:         config.FileType,
		Source:       "log",
		Path:         cf.FilePath,
		Encoding:     cf.Encoding,
		ExcludePaths: cf.ExcludePaths,
		TailingMode:  cf.TailingMode,
	})
	a.sources.AddSource(c.source)
}

func (a *Agent) removeCheck(c *Check) {
	a.Lock()
	defer a.Unlock()

	a.sources.RemoveSource(c.source)
	delete(a.checks, string(c.ID()))
	delete(a.files[c.config.FilePath], string(c.ID()))
}

func (a *Agent) getCheckByID(id string) *Check {
	a.Lock()
	defer a.Unlock()
	return a.checks[id]
}

func (a *Agent) getChecksByFile(file string) (checks map[string]*Check) {
	a.Lock()
	defer a.Unlock()

	return a.files[file]
}

func (a *Agent) process(msg *message.Message) {
	checks := a.getChecksByFile(msg.Origin.LogSource.Config.Path)

	for _, c := range checks {
		c.process(msg)
	}
}

func (a *Agent) work(id int) {
	klog.V(6).Infof("worker %d is running", id)
	for {
		select {
		case msg := <-a.msgCh:
			a.process(msg)

		case <-a.ctx.Done():
			klog.V(6).Infof("worker %d is exiting", id)
			return
		}
	}
}

func NewAgent() (*Agent, error) {
	health := health.RegisterLiveness("plugins-logs")
	auditorTTL := coreConfig.C.Logs.AuditorTTL.Duration
	auditor := auditor.New(coreConfig.C.Logs.RunPath, DefaultRegistryFilename, auditorTTL, health)
	sources := config.NewLogSources()
	msgCh := make(chan *message.Message, 10)
	pipelineProvider := NewChProvider(msgCh)
	scanner := file.NewScanner(sources, coreConfig.C.Logs.OpenFilesLimit, pipelineProvider, auditor, file.DefaultSleepDuration,
		coreConfig.C.Logs.ValidatePodContainerId, coreConfig.C.Logs.FileScanPeriod.Duration)
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		msgCh:   msgCh,
		scanner: scanner,
		sources: sources,
		auditor: auditor,
		health:  health,
		ctx:     ctx,
		cancel:  cancel,
		checks:  make(map[string]*Check),
		files:   make(map[string]map[string]*Check),
	}, nil
}

// chProvider mocks pipeline providing logic
type chProvider struct {
	msgChan chan *message.Message
}

// NewChProvider returns a new chProvider
func NewChProvider(msgCh chan *message.Message) pipeline.Provider {
	return &chProvider{msgChan: msgCh}
}

// Start does nothing
func (p *chProvider) Start() {}

// Stop does nothing
func (p *chProvider) Stop() {}

// Flush does nothing
func (p *chProvider) Flush(ctx context.Context) {}

// NextPipelineChan returns the next pipeline
func (p *chProvider) NextPipelineChan() chan *message.Message {
	return p.msgChan
}

func checkFactory() check.Check {
	return &Check{
		CheckBase: core.NewCheckBaseWithInterval(checkName, 0),
		tagRegs:   make(map[string]*regexp.Regexp),
	}
}

func init() {
	core.RegisterCheck(checkName, checkFactory)
}
