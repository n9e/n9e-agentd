package apiserver

import (
	"context"
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/yubo/apiserver/pkg/apiserver"
	apioptions "github.com/yubo/apiserver/pkg/options"
	"github.com/yubo/golib/proc"

	// authn
	_ "github.com/yubo/apiserver/pkg/apiserver"
	_ "github.com/yubo/apiserver/pkg/options"
)

const (
	moduleName = "apiserver"
)

type module struct {
	name string
	http apioptions.ApiServer

	ctx    context.Context
	cancel context.CancelFunc
}

var (
	_module = &module{name: moduleName}
	hookOps = []proc.HookOps{{
		Hook:        _module.start,
		Owner:       moduleName,
		HookNum:     proc.ACTION_START,
		Priority:    proc.PRI_MODULE,
		SubPriority: options.PRI_M_APISERVER,
	}, {
		Hook:        _module.stop,
		Owner:       moduleName,
		HookNum:     proc.ACTION_STOP,
		Priority:    proc.PRI_MODULE,
		SubPriority: options.PRI_M_APISERVER,
	}}
)

func (p *module) start(ctx context.Context) error {
	server, ok := apioptions.ApiServerFrom(ctx)
	if !ok {
		return fmt.Errorf("unable to get http server from the context")
	}

	p.installWs(server)

	return nil
}

func (p *module) stop(ctx context.Context) error {
	if p.cancel != nil {
		p.cancel()
	}

	return nil
}

func init() {
	proc.RegisterHooks(hookOps)
	apiserver.RegisterHooks()
	proc.RegisterFlags(moduleName, "apiserver", &Config{})
}
