package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/n9e/n9e-agentd/cmd/agent/common"
	"github.com/n9e/n9e-agentd/pkg/apiserver"
	"github.com/n9e/n9e-agentd/pkg/authentication"
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
	Agent     *config.Config
	Apiserver apiserver.Config
	Auth      authentication.Config
	ctx       context.Context

	In     io.Reader
	Out    io.Writer
	Errout io.Writer

	Client  *rest.RESTClient
	Clients map[string]*rest.RESTClient // metrcs/cmd

	TopCmd    *cobra.Command
	ServerCmd *cobra.Command
	configer  *configer.Configer
	fs        *pflag.FlagSet
}

func (p *EnvSettings) Init(override map[string]string) error {
	opts, _ := proc.ConfigOptsFrom(p.ctx)

	if len(configer.GlobalOptions.ValueFiles()) == 0 {
		if f := util.DefaultConfigfile(); util.IsFile(f) {
			klog.V(1).Infof("use default config file %s", f)
			opts = append(opts, configer.WithValueFile(f))
		}
	}

	for k, v := range override {
		opts = append(opts, configer.WithOverrideYaml(k, v))
	}

	c, err := configer.New(opts...)
	if err != nil {
		return err
	}
	p.configer = c

	p.Agent = config.NewConfig(c)
	if err := c.Read("agent", p.Agent); err != nil {
		return err
	}
	config.C = p.Agent

	if err := c.Read("apiserver", &p.Apiserver); err != nil {
		return err
	}
	if err := c.Read("authentication", &p.Auth); err != nil {
		return err
	}

	// validate
	p.Auth.ValidatePath(p.Agent.RootDir)

	// client
	if err := p.initClients(); err != nil {
		return err
	}

	proc.WithConfiger(p.ctx, p.configer)
	common.Client = p

	klog.V(10).Infof("config %s", p)

	return nil
}

func (p *EnvSettings) newClient(host string, port int) (*rest.RESTClient, error) {
	return rest.RESTClientFor(&rest.Config{
		Host:            fmt.Sprintf("%s:%d", host, port),
		ContentConfig:   rest.ContentConfig{NegotiatedSerializer: scheme.Codecs},
		BearerTokenFile: p.Auth.AuthTokenFile,
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
		p.Clients["exporter"] = client
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
	if port := p.Apiserver.Port; port > 0 {
		client, err := p.newClient(p.Apiserver.Host, port)
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
		Endpoint:      fmt.Sprintf("%s:%d", p.Apiserver.Host, p.Apiserver.Port),
		AuthTokenFile: p.Auth.AuthTokenFile,
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
		opts = append(opts, cmdcli.WithPrifix(uri))
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
	if p.Agent.CliQueryTimeout > 0 {
		opts = append(opts, cmdcli.WithTimeout(p.Agent.CliQueryTimeout))
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

// svc: cmd
// uri: /health
// return: http://localhost:5001/health
func (p *EnvSettings) url(svc, uri string) string {
	uri = strings.Trim(uri, "/")
	switch strings.ToLower(svc) {
	case "cmd":
		return fmt.Sprintf("http://%s:%d/%v", p.Apiserver.Host, p.Apiserver.Port, uri)
	default:
		return fmt.Sprintf("http://%s:%d/%v", p.Apiserver.Host, p.Agent.Telemetry.Port, uri) // metrics
	}
}

// SetupCLI
func (p *EnvSettings) SetupCLI() error {
	cf := config.NewConfig(p.configer)
	cf.IsCliRunner = true
	if err := p.configer.Read("agent", cf); err != nil {
		return err
	}
	config.C = cf
	return nil
}
