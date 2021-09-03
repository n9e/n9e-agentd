package apiserver

import (
	"context"
	"fmt"

	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/yubo/apiserver/pkg/apiserver"
	"github.com/yubo/apiserver/pkg/authorization" // authz
	apioptions "github.com/yubo/apiserver/pkg/options"
	"github.com/yubo/apiserver/pkg/rest"
	"github.com/yubo/golib/proc"

	_ "github.com/n9e/n9e-agentd/pkg/authentication"
	_ "github.com/yubo/apiserver/pkg/authentication/register" // authn
	_ "github.com/yubo/apiserver/pkg/rest/swagger/register"

	// authz.login
	"github.com/yubo/apiserver/pkg/authorization/abac/api"
	"github.com/yubo/apiserver/pkg/authorization/abac/register" // authz.login
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

	// httpserver
	apiserver.RegisterHooks()

	// authz
	authorization.RegisterHooks()

	// override apiserver config's flags & envs
	proc.RegisterFlags("apiserver", "apiserver", &Config{})

	// override authorization config's flags & envs
	proc.RegisterFlags("authorization", "authorization", &authzConfig{})

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
