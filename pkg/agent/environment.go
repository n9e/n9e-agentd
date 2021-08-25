package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/n9e/n9e-agentd/pkg/apiserver"
	"github.com/n9e/n9e-agentd/pkg/authentication"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/options"
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

	client  *rest.RESTClient
	clients map[string]*rest.RESTClient // metrcs/cmd

	TopCmd    *cobra.Command
	ServerCmd *cobra.Command
	configer  *configer.Configer
	fs        *pflag.FlagSet
}

func (p *EnvSettings) Init() error {
	opts, _ := proc.ConfigOptsFrom(p.ctx)
	configer, err := configer.New(opts...)
	if err != nil {
		return err
	}
	p.configer = configer

	p.Agent = config.NewConfig(configer)
	if err := configer.Read("agent", p.Agent); err != nil {
		return err
	}
	if err := configer.Read("apiserver", &p.Apiserver); err != nil {
		return err
	}
	if err := configer.Read("authentication", &p.Auth); err != nil {
		return err
	}

	// client
	p.Auth.ValidatePath(p.Agent.RootDir)

	p.client, err = rest.RESTClientFor(&rest.Config{
		Host:            fmt.Sprintf("%s:%d", p.Apiserver.Host, p.Apiserver.Port),
		ContentConfig:   rest.ContentConfig{NegotiatedSerializer: scheme.Codecs},
		BearerTokenFile: p.Auth.AuthTokenFile,
	})
	if err != nil {
		klog.Infof("unable to create client, err %s", err)
	}

	klog.V(5).Infof("config %s", p)

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

func (p *EnvSettings) ApiCall(method, uri string, input, output interface{}, opts ...cmdcli.RequestOption) error {
	req, err := p.Request(method, uri, input, output, opts...)
	if err != nil {
		return err
	}

	return req.Do(context.Background())
}

func (p *EnvSettings) Request(method, uri string, input, output interface{}, opts ...cmdcli.RequestOption) (*cmdcli.Request, error) {
	if p, ok := input.(Preparer); ok {
		if err := p.Prepare(); err != nil {
			return nil, err
		}
	}
	if uri != "" {
		opts = append(opts, cmdcli.WithPrifix(uri))
	}
	if input != nil {
		opts = append(opts, cmdcli.WithInput(input))
	}
	if output != nil {
		opts = append(opts, cmdcli.WithOutput(output))
	}
	if p.Agent.CliQueryTimeout > 0 {
		opts = append(opts, cmdcli.WithTimeout(p.Agent.CliQueryTimeout))
	}

	return cmdcli.NewRequestWithClient(p.client, opts...), nil
}

func (p *EnvSettings) ApiCallDone(method, uri string, input, output interface{}, opts ...cmdcli.RequestOption) (err error) {
	if err = p.ApiCall(method, uri, input, output, opts...); err != nil {
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

func (p *EnvSettings) ApiPaging(uri string, input, output interface{}, opts ...cmdcli.RequestOption) error {
	req, err := p.Request("GET", uri, input, output, opts...)
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
