// Copyright 2018,2019 freewheel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package agentdctl

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/yubo/apiserver/pkg/cmdcli"
	"github.com/yubo/apiserver/pkg/rest"
	"github.com/yubo/apiserver/pkg/rsh"
	apierrors "github.com/yubo/golib/api/errors"
	"github.com/yubo/golib/configer"
	"github.com/yubo/golib/util"
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

const (
	moduleName = "agentdctl"
)

// envSettings describes all of the environment settings.
type EnvSettings struct {
	Host        string `json:"host" flag:"host" default:"localhost" env:"AGENTD_HOST" description:"agentd endpoint"`
	CmdPort     int    `json:"cmd_port" flag:"cmd-port" default:"5001" env:"AGENTD_CMD_PORT" description:"cmd port"` // cmd_port
	AgentdRoot  string `json:"agentd_root" flag:"root" default:"/opt/n9e/agentd" env:"AGENTD_ROOT" description:"agentd server work dir"`
	Token       string `json:"token" env:"AGENTD_TOKEN" description:"token"`
	Timeout     int    `json:"tiemout" env:"AGENTD_TIMEOUT" flag:"timeout" default:"5" description:"timeout(Second)"`
	DisablePage bool   `json:"disable_page" flag:"disable-page" default:"false" env:"AGENTD_DISABLE_PAGE"`
	PageSize    int    `json:"page_size" flag:"page-size" default:"10" env:"AGENTD_PAGE_SIZE"`
	NoColor     bool   `json:"no_color" flag:"no-color" default:"false" env:"AGENTD_NO_COLOR"`

	In       io.Reader      `json:"-"`
	Out      io.Writer      `json:"-"`
	Errout   io.Writer      `json:"-"`
	TopCmd   *cobra.Command `json:"-"`
	Req      *http.Request  `json:"-"`
	Resp     *http.Response `json:"-"`
	configer *configer.Configer
	fs       *pflag.FlagSet
}

func (p *EnvSettings) Validate() error {
	p.Host = strings.TrimRight(p.Host, "/")
	if p.Host == "" {
		p.Host = "localhost"
	}
	return nil
}

func (p *EnvSettings) Init() error {
	configer, err := configer.New(
		configer.WithFlagOptions(true, false, 5),
		configer.WithFlag(p.fs),
	)
	if err != nil {
		return err
	}

	if err := configer.Read(moduleName, p); err != nil {
		return err
	}

	if p.NoColor {
		color.NoColor = true
	}
	klog.V(5).Infof("config %s", p)

	return nil
}

func (p EnvSettings) String() string {
	return util.Prettify(p)
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	s.fs = fs
	configer.AddFlags(fs, moduleName, &EnvSettings{})
}

func (p EnvSettings) Output(verbose bool) *EnvSettingsOutput {
	ret := &EnvSettingsOutput{
		Host:        p.Host,
		DisablePage: p.DisablePage,
		Version:     options.Version,
		Branch:      options.Branch,
		Revision:    options.Revision,
		BuildDate:   options.BuildDate,
	}

	if verbose {
		ret.Token = p.Token
	} else {
		if token := p.Token; token != "" {
			ret.Token = util.SubStr3(token, 3, -3)
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
		Bearer:     &p.Token,
		InputParam: input,
		Client: http.Client{
			Timeout: time.Duration(p.Timeout) * time.Second,
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
		util.PrepareValue(rv, rv.Type())
		rv = rv.Elem()
	}

	rv.SetInt(int64(p.PageSize))

	return cmdcli.TermPaging(p.PageSize,
		p.DisablePage, p, svc+"/"+uri,
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
	switch strings.ToLower(svc) {
	case "cmd":
		return fmt.Sprintf("http://%s:%d/%v", p.Host, p.CmdPort, uri)
	default:
		return fmt.Sprintf("http://%s:%d/%v", p.Host, p.CmdPort, uri)
	}
}

func (p *EnvSettings) RshRequest(svc, url string, input interface{}) error {
	opt := &rest.RequestOptions{
		Method:     "GET",
		Bearer:     &p.Token,
		InputParam: input,
		Url:        p.url(svc, url),
		Client:     http.Client{},
	}

	return rsh.RshRequest(opt)
}

func newEnvCmd(env *EnvSettings) *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "env",
		Short: "show grab env information",
		RunE: func(cmd *cobra.Command, args []string) error {
			env.Out.Write(cmdcli.Table(env.Output(verbose)))
			return nil
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&verbose, "verbvose", "V", false, "show secret infomation")
	return cmd
}
