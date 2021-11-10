package mocker

import (
	"context"
	"fmt"

	"github.com/yubo/apiserver/pkg/options"
	"github.com/yubo/apiserver/pkg/rest"
	"github.com/yubo/apiserver/pkg/rest/protobuf"
	"github.com/yubo/golib/proc"

	_ "github.com/yubo/apiserver/pkg/rest/swagger/register"
	_ "github.com/yubo/apiserver/pkg/server/register"
	_ "github.com/yubo/apiserver/pkg/tracing/register"
)

const (
	moduleName = "mocker"
)

var (
	_mocker = &mocker{name: moduleName}
	hookOps = []proc.HookOps{{
		Hook:     _mocker.start,
		Owner:    moduleName,
		HookNum:  proc.ACTION_START,
		Priority: proc.PRI_MODULE,
	}}
)

type Config struct {
	Port        int    `flag:"port" default:"8000" env:"N9E_MOCKER_PORT" description:"listen port"`
	CollectRule bool   `flag:"collect-rule" description:"enable send statsd sample data"`
	SendStatsd  bool   `flag:"send-statsd" description:"enable collect rule provider"`
	Confd       string `flag:"confd" default:"./etc/mocker.d" description:"config dir"`
}
type mocker struct {
	config *Config
	name   string

	ctx   context.Context
	rules CollectRules
}

func (p *mocker) start(ctx context.Context) error {
	c := proc.ConfigerMustFrom(ctx)

	cf := &Config{}
	if err := c.Read(moduleName, cf); err != nil {
		return err
	}
	p.config = cf
	p.ctx = ctx

	if err := p.installCollectRules(); err != nil {
		return err
	}
	if err := p.installStatsdSender(); err != nil {
		return err
	}
	if err := p.installWs(); err != nil {
		return err
	}
	return nil
}

func (p *mocker) installWs() error {
	http, ok := options.APIServerFrom(p.ctx)
	if !ok {
		return fmt.Errorf("unable to get http server from the context")
	}

	rest.SwaggerTagRegister("api groups", "api groups")
	p.installDatadogWs(http)
	p.installN9eWs(http)

	return nil
}

func init() {
	proc.RegisterHooks(hookOps)
	proc.RegisterFlags("mocker", "mocker config", &Config{})
	protobuf.RegisterEntityAccessor()
}
