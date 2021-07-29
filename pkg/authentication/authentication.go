package authentication

import (
	"context"

	"github.com/yubo/apiserver/pkg/authentication"
	"github.com/yubo/apiserver/pkg/authentication/authenticator"
	"github.com/yubo/apiserver/pkg/authentication/user"
	"github.com/yubo/apiserver/pkg/options"
	"github.com/yubo/golib/proc"
	"k8s.io/klog/v2"
)

const (
	moduleName = "authentication.tokenAuthFile"
	modulePath = "authentication"
)

var (
	_auth   = &authModule{name: moduleName}
	hookOps = []proc.HookOps{{
		Hook:        _auth.init,
		Owner:       moduleName,
		HookNum:     proc.ACTION_START,
		Priority:    proc.PRI_SYS_INIT,
		SubPriority: options.PRI_M_AUTHN - 1,
	}}
)

type config struct {
	AuthTokenFile string `json:"authTokenFile" flag:"auth-token-file" default:"./etc/auth_token" description:"If set, the file that will be used to secure the secure port of the API server via token authentication."`
}

func (p *config) Validate() error {
	return nil
}

type authModule struct {
	name   string
	config *config
}

func newConfig() *config {
	return &config{}
}

func (p *authModule) init(ctx context.Context) error {
	c := proc.ConfigerFrom(ctx)

	cf := newConfig()
	if err := c.Read(modulePath, cf); err != nil {
		return err
	}
	p.config = cf
	AuthTokenFile = cf.AuthTokenFile

	if len(cf.AuthTokenFile) == 0 {
		klog.InfoS("skip authModule", "name", p.name, "reason", "tokenfile not set")
		return nil
	}
	klog.V(5).InfoS("authmodule init", "name", p.name, "file", cf.AuthTokenFile)

	auth, err := NewAuthToken()
	if err != nil {
		return err
	}

	return authentication.RegisterTokenAuthn(auth)
}

type TokenAuthenticator struct {
	tokens map[string]*user.DefaultInfo
}

func NewAuthToken() (*TokenAuthenticator, error) {
	authToken, err := CreateOrFetchToken()
	if err != nil {
		return nil, err
	}

	return &TokenAuthenticator{
		tokens: map[string]*user.DefaultInfo{
			authToken: &user.DefaultInfo{Name: "system:agentd", UID: "0"},
		},
	}, nil
}

func (a *TokenAuthenticator) AuthenticateToken(ctx context.Context, value string) (*authenticator.Response, bool, error) {
	user, ok := a.tokens[value]
	if !ok {
		return nil, false, nil
	}
	return &authenticator.Response{User: user}, true, nil
}

func (a *TokenAuthenticator) Name() string {
	return "token file token authenticator"
}

func (a *TokenAuthenticator) Priority() int {
	return authenticator.PRI_TOKEN_FILE
}

func (a *TokenAuthenticator) Available() bool {
	return true
}
func init() {
	proc.RegisterHooks(hookOps)
	proc.RegisterFlags(modulePath, "authentication", newConfig())
}
