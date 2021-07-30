package script

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/n9e/n9e-agentd/pkg/aggregator"
	"github.com/n9e/n9e-agentd/pkg/collector/check"
	core "github.com/n9e/n9e-agentd/pkg/collector/corechecks"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	checkName      = "script"
	defaultTimeout = 5
)

type InitConfig struct {
	Root    string            `json:"root"`
	Env     map[string]string `json:"env"`
	Timeout int               `json:"timeout"`
}

type InstanceConfig struct {
	FilePath string            `json:"file_path"`
	Root     string            `json:"root"`
	Params   string            `json:"params"`
	Env      map[string]string `json:"env"`
	Stdin    string            `json:"stdin"`
	Timeout  int               `json:"timeout"`
}

type checkConfig struct {
	filePath string
	env      []string
	params   []string
	stdin    string
	timeout  time.Duration
}

func (p checkConfig) String() string {
	return util.Prettify(p)
}

// Configure the Prom check
func (c *Check) Configure(rawInstance integration.Data, rawInitConfig integration.Data, source string) error {
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

func defaultInstanceConfig() InstanceConfig {
	return InstanceConfig{}
}

func buildConfig(rawInstance integration.Data, rawInitConfig integration.Data) (*checkConfig, error) {
	instance := defaultInstanceConfig()
	pwd, _ := os.Getwd()

	initConfig := InitConfig{
		Root:    filepath.Join(pwd, "script.d"),
		Timeout: defaultTimeout,
	}

	err := yaml.Unmarshal(rawInitConfig, &initConfig)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(rawInstance, &instance)
	if err != nil {
		return nil, err
	}

	if instance.Root == "" {
		instance.Root = initConfig.Root
	}

	if len(instance.Env) == 0 {
		instance.Env = initConfig.Env
	}

	if instance.Timeout <= 0 {
		instance.Timeout = initConfig.Timeout
	}

	if instance.Timeout <= 0 {
		instance.Timeout = defaultTimeout
	}

	config := &checkConfig{}

	if !filepath.IsAbs(instance.FilePath) && instance.Root != "" {
		if !isDir(instance.Root) {
			return nil, fmt.Errorf("root %s is not a valid dir", instance.Root)
		}
		config.filePath = filepath.Join(instance.Root, instance.FilePath)
	} else {
		config.filePath = instance.FilePath
	}

	if len(instance.Env) > 0 {
		for k, v := range instance.Env {
			config.env = append(config.env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	config.env = append(syscall.Environ(), config.env...)

	if len(instance.Params) > 0 {
		config.params = strings.Fields(instance.Params)
	}

	if len(instance.Stdin) > 0 {
		config.stdin = instance.Stdin
	}

	config.timeout = time.Duration(instance.Timeout) * time.Second

	return config, nil
}

// Check doesn't need additional fields
type Check struct {
	core.CheckBase
	config *checkConfig
}

// Run executes the check
func (c *Check) Run() error {
	sender, err := aggregator.GetSender(c.ID())
	if err != nil {
		return err
	}

	if err := c.collect(sender); err != nil {
		klog.Warningf("%s", err)
	}

	sender.Commit()
	return nil
}

func (c *Check) collect(sender aggregator.Sender) error {
	for _, file := range c.getFiles() {
		c._collect(sender, file)
	}

	sender.Commit()
	return nil
}

func (c *Check) _collect(sender aggregator.Sender, file string) {
	cf := c.config
	klog.V(4).Infof("file %s", file)

	ctx, cancel := context.WithTimeout(context.Background(), cf.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, file, cf.params...)

	stdout := bytes.NewBuffer([]byte{})
	stderr := bytes.NewBuffer([]byte{})

	cmd.Stdin = bytes.NewReader([]byte(cf.stdin))
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	cmd.Env = cf.env

	err := cmd.Run()
	if err != nil {
		klog.Warningf("%s run err %s", cmd.String(), err)
		if err := stderr.String(); err != "" {
			klog.Warningf("stderr: %s", err)
		}
		return
	}

	out := stdout.Bytes()
	if len(out) == 0 {
		klog.Infof("stdout of %s is blank", file)
		return
	}
	klog.V(6).Infof("%s stdout: %s", file, string(out))

	if err := send(sender, out); err != nil {
		klog.Warningf("send of %s err %s", file, err)
		return
	}
}

func (c *Check) getFiles() []string {
	files, _ := filepath.Glob(c.config.filePath)
	return files
}

func checkFactory() check.Check {
	return &Check{
		CheckBase: core.NewCheckBase(checkName),
	}
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

func init() {
	core.RegisterCheck(checkName, checkFactory)
}
