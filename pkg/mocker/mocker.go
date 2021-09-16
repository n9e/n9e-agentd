package mocker

import (
	"context"
	"fmt"

	"github.com/yubo/apiserver/pkg/options"
	"github.com/yubo/apiserver/pkg/rest/protobuf"
	"github.com/yubo/golib/proc"
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
	http, ok := options.ApiServerFrom(p.ctx)
	if !ok {
		return fmt.Errorf("unable to get http server from the context")
	}

	p.installDatadogWs(http)
	p.installN9eWs(http)

	return nil
}

func init() {
	proc.RegisterHooks(hookOps)
	proc.RegisterFlags("mocker", "mocker config", &Config{})
	protobuf.RegisterEntityAccessor()
}
