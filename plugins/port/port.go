package port

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/check"
	core "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const checkName = "port"

type InitConfig struct {
	Timeout int `json:"timeout"`
}

type InstanceConfig struct {
	Protocol   string        `json:"protocol" description:"udp or tcp"`
	Port       int           `json:"port"`
	addrs      []string      `json:"-"`
	timeout    time.Duration `json:"-"`
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
	return InstanceConfig{
		Protocol: "tcp",
	}
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

	ifAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return checkConfig{}, err
	}
	for _, addr := range ifAddrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			klog.Warningf("parse cidr %s %s", addr.String(), err)
		}
		instance.addrs = append(instance.addrs, fmt.Sprintf("%s:%d", ip.String(), instance.Port))
	}

	if initConfig.Timeout <= 0 {
		instance.timeout = time.Second * 3
	} else {
		instance.timeout = time.Second * time.Duration(initConfig.Timeout)
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
	tags   []string
}

// Run executes the check
func (c *Check) Run() error {
	sender, err := aggregator.GetSender(c.ID())
	if err != nil {
		return err
	}

	value := 0
	if ok := c.check(); ok {
		value = 1
	}
	sender.Gauge("proc.port.listen", float64(value), "", c.tags)

	sender.Commit()
	return nil
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
	c.tags = []string{
		fmt.Sprintf("port:%d", config.Port),
		fmt.Sprintf("protocol:%s", config.Protocol),
	}
	return nil
}

func (c *Check) SampleConfig() string {
	return ""
}

func (c *Check) check() bool {
	cf := c.config

	klog.V(6).Infof("port check address %v", cf.addrs)

	if cf.Protocol == "udp" {
		for i, addr := range cf.addrs {
			fmt.Printf(" udp addr  %v \n", addr)
			udpAddr, err := net.ResolveUDPAddr("udp", addr)
			if err != nil {
				continue
			}
			l, err := net.ListenUDP("udp", udpAddr)
			if err != nil && strings.Contains(err.Error(), "address already in use") {
				if i > 0 {
					cf.addrs[0], cf.addrs[i] = cf.addrs[i], cf.addrs[0]
				}
				return true
			}
			l.Close()
		}
		return false
	}

	for i, addr := range cf.addrs {
		conn, err := net.DialTimeout("tcp", addr, cf.timeout)
		if err != nil {
			continue
		}
		conn.Close()

		if i > 0 {
			cf.addrs[0], cf.addrs[i] = cf.addrs[i], cf.addrs[0]
		}

		return true
	}

	return false
}

func checkFactory() check.Check {
	return &Check{
		CheckBase: core.NewCheckBase(checkName),
	}
}

func init() {
	core.RegisterCheck(checkName, checkFactory)
}
