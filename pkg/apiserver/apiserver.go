package apiserver

import (
	"context"
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/options"
	apioptions "github.com/yubo/apiserver/pkg/options"
	"github.com/yubo/apiserver/pkg/rest"
	"github.com/yubo/apiserver/pkg/server"
	"github.com/yubo/apiserver/plugin/authorizer/abac/api"
	"github.com/yubo/apiserver/plugin/authorizer/abac/register"
	"github.com/yubo/golib/proc"
)

const (
	moduleName = "apiserver"
)

type module struct {
	name   string
	server server.APIServer

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
	server, ok := apioptions.APIServerFrom(ctx)
	if !ok {
		return fmt.Errorf("unable to get server server from the context")
	}

	p.installWs(server)

	rest.ScopeRegister("write", "write resource")
	rest.ScopeRegister("read", "read resource")

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

	// register server module
	//servermodule.RegisterHooks()
	//proc.RegisterFlags("apiserver", "apiserver", &Config{})

	// register authz
	//authzmodule.RegisterHooks()
	//proc.RegisterFlags("authorization", "authorization", &authzConfig{})

	register.PolicyList = []*api.Policy{
		{Spec: api.PolicySpec{
			Group:           "system:unauthenticated",
			Readonly:        true,
			NonResourcePath: "/swagger*",
		}}, {Spec: api.PolicySpec{
			Group:           "system:unauthenticated",
			Readonly:        true,
			NonResourcePath: "/apidocs.json",
		}}, {Spec: api.PolicySpec{
			Group:           "system:authenticated",
			NonResourcePath: "*",
			Resource:        "*",
		}},
	}
}
