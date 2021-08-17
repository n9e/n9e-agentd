package agentctl

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/yubo/apiserver/pkg/cmdcli"
	"github.com/yubo/apiserver/pkg/rest"
	apierrors "github.com/yubo/golib/api/errors"
	"github.com/yubo/golib/configer"
	"github.com/yubo/golib/scheme"
	gutil "github.com/yubo/golib/util"
	"k8s.io/klog/v2"
)

func NewSettings() *EnvSettings {
	return &EnvSettings{
		In:     os.Stdin,
		Out:    os.Stdout,
		Errout: os.Stderr,
	}
}

type EnvSettingsOutput struct {
	Host string
	//Token         string
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
	Agent     agentConfig
	Apiserver apiserverConfig
	Auth      authConfig

	In     io.Reader
	Out    io.Writer
	Errout io.Writer

	client  *rest.RESTClient
	clients map[string]*rest.RESTClient // metrcs/cmd

	TopCmd   *cobra.Command
	configer *configer.Configer
	fs       *pflag.FlagSet

	//token    string
	//Req      *http.Request
	//Resp     *http.Response
}

func (p *EnvSettings) Init() error {
	configer, err := configer.New()
	if err != nil {
		return err
	}
	p.configer = configer

	if err := configer.Read("agent", &p.Agent); err != nil {
		return err
	}
	if err := configer.Read("apiserver", &p.Apiserver); err != nil {
		return err
	}
	if err := configer.Read("authentication", &p.Auth); err != nil {
		return err
	}

	// client
	p.Auth.AuthTokenFile = util.AbsPath(p.Auth.AuthTokenFile, p.Agent.RootDir, "etc/auth_token")
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

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	s.fs = fs
	configer.Setting.AddFlags(fs)
	configer.AddConfigs(fs, "agent", &agentConfig{})
	configer.AddConfigs(fs, "apiserver", &apiserverConfig{})
	configer.AddConfigs(fs, "authentication", &authConfig{})
}

func (p EnvSettings) Output(verbose bool) *EnvSettingsOutput {
	ret := &EnvSettingsOutput{
		Host:          p.Apiserver.Host,
		AuthTokenFile: p.Auth.AuthTokenFile,
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
	if p.Agent.Timeout > 0 {
		opts = append(opts, cmdcli.WithOutput(output))
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
