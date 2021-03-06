package agent

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/n9e/n9e-agentd/cmd/agent/common"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/yubo/apiserver/pkg/cmdcli"
	"github.com/yubo/apiserver/pkg/rest"
	apierrors "github.com/yubo/golib/api/errors"
	"github.com/yubo/golib/configer"
	"github.com/yubo/golib/proc"
	"github.com/yubo/golib/scheme"
	gutil "github.com/yubo/golib/util"
	"k8s.io/klog/v2"
)

func NewSettings(ctx context.Context) *EnvSettings {
	return &EnvSettings{
		In:     os.Stdin,
		Out:    os.Stdout,
		Errout: os.Stderr,
		ctx:    ctx,
	}
}

type EnvSettingsOutput struct {
	Endpoint      string
	AuthTokenFile string
	DisablePage   bool
	PageSize      int
	Version       string
	Branch        string
	Revision      string
	BuildDate     string
}

// envSettings describes all of the environment settings.
type EnvSettings struct {
	Agent *config.Config
	//Apiserver apiserver.Config
	//Auth      authentication.Config
	ctx context.Context

	In     io.Reader
	Out    io.Writer
	Errout io.Writer

	Client  *rest.RESTClient
	Clients map[string]*rest.RESTClient // metrcs/cmd

	TopCmd   *cobra.Command
	configer configer.ParsedConfiger
	fs       *pflag.FlagSet
}

const (
	defaultConfig = `apiserver:
  secureServing:
    enabled: false
  insecureServing:
    enabled: true
    bindAddress: 127.0.0.1
    bindPort: 8010

authorization:
  modes:
  - AlwaysAllow
  alwaysAllowPaths:
  - /healthz
  - /readyz
  - /livez
  - /swagger*
  - /apidocs.json
`
)

// Init: init proc
func (p *EnvSettings) Init(cmd *cobra.Command) error {
	p.TopCmd = cmd

	opts := []configer.ConfigerOption{
		configer.WithEnv(true, false),
		configer.WithMaxDepth(5),
		configer.WithDefaultYaml("", defaultConfig),
	}

	if len(configer.ValueFiles()) == 0 {
		if f := util.DefaultConfigfile(); util.IsFile(f) {
			klog.Infof("use default config file %s", f)
			opts = append(opts, configer.WithValueFile(f))
		}
	}

	// proc
	proc.Init(cmd,
		proc.WithContext(p.ctx),
		proc.WithConfigOptions(opts...),
	)

	return nil
}

func (p *EnvSettings) Parse(fs *pflag.FlagSet, override map[string]string) error {
	klog.V(10).Infof("entering setting parse")
	defer klog.V(10).Infof("leaving setting parse")
	opts := []configer.ConfigerOption{}
	for k, v := range override {
		opts = append(opts, configer.WithOverrideYaml(k, v))
	}

	c, err := proc.Parse(fs, opts...)
	if err != nil {
		return err
	}

	p.configer = c

	agentConfig, err := config.NewConfig(c)
	if err != nil {
		return err
	}

	if err := c.Read("agent", agentConfig); err != nil {
		return err
	}

	if err := config.ResolveSecrets(agentConfig); err != nil {
		return err
	}

	p.Agent = agentConfig
	config.C = p.Agent

	// client
	if err := p.initClients(); err != nil {
		return err
	}

	common.Client = p

	//klog.V(10).Infof("config %s", p)

	return nil
}

func (p *EnvSettings) GetClient(name string) *rest.RESTClient {
	return p.Clients[name]
}

func (p *EnvSettings) newClient(host string, port int) (*rest.RESTClient, error) {
	return rest.RESTClientFor(&rest.Config{
		Host:            fmt.Sprintf("%s:%d", host, port),
		ContentConfig:   rest.ContentConfig{NegotiatedSerializer: scheme.Codecs},
		BearerTokenFile: p.Agent.AuthTokenFile,
	})

}

func (p *EnvSettings) initClients() error {
	p.Clients = map[string]*rest.RESTClient{}

	// /vars, /metrics, /debug/pprof
	if port := p.Agent.Telemetry.Port; port > 0 {
		client, err := p.newClient("127.0.0.1", port)
		if err != nil {
			return err
		}
		p.Clients["telemetry"] = client
		p.Clients["metrics"] = client
		p.Clients["pprof"] = client
	}

	// apm
	if port := p.Agent.Apm.ReceiverPort; port > 0 {
		client, err := p.newClient("127.0.0.1", port)
		if err != nil {
			return err
		}
		p.Clients["apm"] = client
	}

	// apiserver
	if port := p.Agent.BindPort; port > 0 {
		client, err := p.newClient(p.Agent.BindHost, port)
		if err != nil {
			klog.Infof("unable to create client, err %s", err)
		}
		p.Client = client
	}

	return nil
}

func (p EnvSettings) String() string {
	return gutil.Prettify(p)
}

func (p EnvSettings) Output() *EnvSettingsOutput {
	ret := &EnvSettingsOutput{
		Endpoint:      fmt.Sprintf("%s:%d", p.Agent.BindHost, p.Agent.BindPort),
		AuthTokenFile: p.Agent.AuthTokenFile,
		DisablePage:   p.Agent.DisablePage,
		PageSize:      p.Agent.PageSize,
		Version:       options.Version,
		Branch:        options.Branch,
		Revision:      options.Revision,
		BuildDate:     options.BuildDate,
	}
	return ret
}

func (p *EnvSettings) Write(b []byte) (int, error) {
	return p.Out.Write(b)
}

type Preparer interface {
	Prepare() error
}

type Validator interface {
	Validate() error
}

func (p *EnvSettings) ApiCall(method, uri string, param, body, output interface{}, opts ...cmdcli.RequestOption) error {
	req, err := p.Request(method, uri, param, body, output, opts...)
	if err != nil {
		return err
	}

	return req.Do(context.Background())
}

func (p *EnvSettings) Request(method, uri string, param, body, output interface{}, opts ...cmdcli.RequestOption) (*cmdcli.Request, error) {
	if p.Client == nil {
		return nil, fmt.Errorf("unable to get client")
	}
	if p, ok := param.(Preparer); ok {
		if err := p.Prepare(); err != nil {
			return nil, err
		}
	}
	if p, ok := param.(Validator); ok {
		if err := p.Validate(); err != nil {
			return nil, err
		}
	}
	if p, ok := body.(Validator); ok {
		if err := p.Validate(); err != nil {
			return nil, err
		}
	}

	if method != "" {
		opts = append(opts, cmdcli.WithMethod(method))
	}

	if uri != "" {
		opts = append(opts, cmdcli.WithPrefix(uri))
	}
	if param != nil {
		opts = append(opts, cmdcli.WithParam(param))
	}
	if body != nil {
		opts = append(opts, cmdcli.WithBody(body))
	}
	if output != nil {
		opts = append(opts, cmdcli.WithOutput(output))
	}
	if p.Agent.CliQueryTimeout.Duration > 0 {
		opts = append(opts, cmdcli.WithTimeout(p.Agent.CliQueryTimeout.Duration))
	}

	return cmdcli.NewRequestWithClient(p.Client, opts...), nil
}

func (p *EnvSettings) ApiCallDone(method, uri string, param, body, output interface{}, opts ...cmdcli.RequestOption) (err error) {
	if err = p.ApiCall(method, uri, param, body, output, opts...); err != nil {
		if apierrors.IsNotFound(err) {
			p.Write([]byte("No Data\n"))
			return nil
		}
		return
	}

	if output == nil {
		p.Write([]byte("succeeded\n"))
		return
	}

	_, err = p.Write(cmdcli.Table(output))
	return
}

func (p *EnvSettings) ApiPaging(uri string, param, body, output interface{}, opts ...cmdcli.RequestOption) error {
	req, err := p.Request("GET", uri, param, body, output, opts...)
	if err != nil {
		return err
	}

	return req.Pager(p.Out, p.Agent.DisablePage).Do(context.Background())
}

func SetFlagFromEnv(name, envName string, fs *pflag.FlagSet) {
	flag := fs.Lookup(name)
	if flag == nil {
		return
	}

	if flag.Changed {
		klog.V(5).Infof("getFlagFromEnv name %s is changed", name)
		return
	}

	if v, ok := os.LookupEnv(envName); ok {
		klog.V(5).Infof("getFlagFromEnv set %s = %s from env(%s)",
			name, v, envName)
		if err := fs.Set(name, v); err != nil {
			klog.Error(err)
		}
	}
}
