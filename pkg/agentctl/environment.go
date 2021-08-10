package ctl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/yubo/apiserver/pkg/cmdcli"
	"github.com/yubo/apiserver/pkg/rest"
	"github.com/yubo/apiserver/pkg/rsh"
	apierrors "github.com/yubo/golib/api/errors"
	"github.com/yubo/golib/configer"
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
	Host        string
	Token       string
	DisablePage bool
	PageSize    int
	Version     string
	Branch      string
	Revision    string
	BuildDate   string
}

type Telemetry struct {
	Port int `json:"port" default:"8070"` // expvar_port
}

type agentConfig struct {
	RootDir       string    `json:"root_dir" flag:"root" env:"AGENTD_ROOT_DIR" description:"root dir path"` // e.g. /opt/n9e/agentd
	ConfdPath     string    `json:"confd_path"`                                                             // confd_path
	PythonVersion string    `json:"python_version" default:"3"`                                             // python_version
	Token         string    `json:"token" env:"AGENTD_TOKEN" description:"token"`
	Timeout       int       `json:"tiemout" env:"AGENTD_TIMEOUT" flag:"timeout" default:"5" description:"timeout(Second)"`
	DisablePage   bool      `json:"disable_page" flag:"disable-page" default:"false" env:"AGENTD_DISABLE_PAGE"`
	PageSize      int       `json:"page_size" flag:"page-size" default:"10" env:"AGENTD_PAGE_SIZE"`
	NoColor       bool      `json:"no_color" flag:"no-color,n" default:"false" env:"AGENTD_NO_COLOR" description:"disable color output"`
	Telemetry     Telemetry `json:"telemetry,inline"` // telemetry
	PythonHome    string    `json:"python_home"`
}

func (p *agentConfig) Validate() (err error) {
	if p.NoColor {
		color.NoColor = true
	}

	if p.RootDir, err = util.ResolveRootPath(p.RootDir); err != nil {
		return err
	}
	p.ConfdPath = util.AbsPath(p.ConfdPath, p.RootDir, "conf.d")
	p.PythonHome = util.AbsPath(p.PythonHome, p.RootDir, "embedded")
	return nil
}

type apiserverConfig struct {
	Host string `json:"address" default:"127.0.0.1" flag:"bind-host" description:"The IP address on which to listen for the --bind-port port. The associated interface(s) must be reachable by the rest of the cluster, and by CLI/web clients. If blank or an unspecified address (127.0.0.1 or ::), all interfaces will be used."` // BindAddress
	Port int    `json:"port" default:"8010" flag:"bind-port" description:"The port on which to serve HTTPS with authentication and authorization. It cannot be switched off with 0."`                                                                                                                                                // BindPort is ignored when Listener is set, will serve https even with 0.
}

func (p *apiserverConfig) Validate() error {
	p.Host = strings.TrimRight(p.Host, "/")
	if p.Host == "" {
		p.Host = "localhost"
	}
	return nil
}

type authConfig struct {
	AuthTokenFile string `json:"auth_token_file" flag:"auth-token-file" default:"./etc/auth_token" description:"If set, the file that will be used to secure the secure port of the API server via token authentication."`
}

func (p *authConfig) Validate() error {
	return nil
}

// envSettings describes all of the environment settings.
type EnvSettings struct {
	Agent     agentConfig
	Apiserver apiserverConfig
	Auth      authConfig

	In       io.Reader
	Out      io.Writer
	Errout   io.Writer
	token    string
	TopCmd   *cobra.Command
	Req      *http.Request
	Resp     *http.Response
	configer *configer.Configer
	fs       *pflag.FlagSet
}

func (p *EnvSettings) Init() error {
	configer, err := configer.New(
		configer.WithFlagOptions(true, false, 5),
		configer.WithFlag(p.fs),
	)
	if err != nil {
		return err
	}

	if err := configer.Read("agent", &p.Agent); err != nil {
		return err
	}
	if err := configer.Read("apiserver", &p.Apiserver); err != nil {
		return err
	}
	if err := configer.Read("authentication", &p.Auth); err != nil {
		return err
	}

	// token
	p.Auth.AuthTokenFile = util.AbsPath(p.Auth.AuthTokenFile, p.Agent.RootDir, "etc/auth_token")
	if file := p.Auth.AuthTokenFile; file != "" {
		token, err := ioutil.ReadFile(file)
		if err != nil {
			klog.Errorf("unable to open token file %s", file)
		} else {
			p.token = string(token)
		}
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
	configer.AddFlags(fs, "agent", &agentConfig{})
	configer.AddFlags(fs, "apiserver", &apiserverConfig{})
	configer.AddFlags(fs, "authentication", &authConfig{})
}

func (p EnvSettings) Output(verbose bool) *EnvSettingsOutput {
	ret := &EnvSettingsOutput{
		Host:      p.Apiserver.Host,
		Version:   options.Version,
		Branch:    options.Branch,
		Revision:  options.Revision,
		BuildDate: options.BuildDate,
	}

	if verbose {
		ret.Token = p.token
	} else {
		if token := p.token; token != "" {
			ret.Token = gutil.SubStr3(token, 3, -3)
		}
	}

	return ret
}

func (p *EnvSettings) Write(b []byte) (int, error) {
	return p.Out.Write(b)
}

// ("GET", "https://example.com/api/v{version}/{model}/{subject}?a=1&b=2", {"subject":"abc", "model": "instance", "version": 1}, nil)
func (p *EnvSettings) apiCall(method, url string, input, output interface{}, cb ...func(interface{})) (err error) {

	opt := &rest.RequestOptions{
		Output:     output,
		Url:        url,
		Method:     method,
		Bearer:     &p.token,
		Ctx:        context.Background(),
		InputParam: input,
		Client: http.Client{
			Timeout: time.Duration(p.Agent.Timeout) * time.Second,
		},
	}

	if _, _, err = rest.HttpRequest(opt); err != nil {
		klog.V(2).Info(err)
		return
	}

	for _, fn := range cb {
		if fn != nil {
			fn(output)
		}
	}

	if w, ok := output.(io.Writer); ok {
		fmt.Fprintf(w, "\n")
	}

	return
}

type Preparer interface {
	Prepare() error
}

func (p *EnvSettings) ApiCall(method, svc, uri string, input, output interface{}, cb ...func(interface{})) error {
	if p, ok := input.(Preparer); ok {
		if err := p.Prepare(); err != nil {
			return err
		}
	}

	return p.apiCall(method, p.url(svc, uri), input, output, cb...)
}

func (p *EnvSettings) ApiCallDone(method, svc, uri string, input, output interface{}, cb ...func(interface{})) (err error) {
	if err = p.ApiCall(method, svc, uri, input, output, cb...); err != nil {
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

func (p *EnvSettings) ApiPaging(svc, uri string, input, output interface{}, cb ...func(interface{})) error {

	rv := reflect.Indirect(reflect.ValueOf(input)).FieldByName("PageSize")
	if !rv.CanSet() {
		return errors.New("expected pageSize field with input struct can set")
	}

	if rv.Kind() == reflect.Ptr {
		gutil.PrepareValue(rv, rv.Type())
		rv = rv.Elem()
	}

	rv.SetInt(int64(p.Agent.PageSize))

	return cmdcli.TermPaging(p.Agent.PageSize,
		p.Agent.DisablePage, p, svc+"/"+uri,
		input, output, p.apiCall, cb...)
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

func (p *EnvSettings) RshRequest(svc, url string, input interface{}) error {
	opt := &rest.RequestOptions{
		Method:     "GET",
		Bearer:     &p.token,
		InputParam: input,
		Url:        p.url(svc, url),
		Client:     http.Client{},
	}

	return rsh.RshRequest(opt)
}
