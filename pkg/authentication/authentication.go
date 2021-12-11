package authentication

import (
	"context"

	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/yubo/apiserver/pkg/authentication"
	"github.com/yubo/apiserver/pkg/authentication/authenticator"
	"github.com/yubo/apiserver/pkg/authentication/user"
	"github.com/yubo/apiserver/pkg/options"
	"github.com/yubo/golib/configer"
	"github.com/yubo/golib/proc"
	"k8s.io/klog/v2"
)

const (
	moduleName = "authentication.tokenAuthFile"
	modulePath = "authentication"
	fakeToken  = "xxxxxx"
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

type Config struct {
	AuthTokenFile        string `json:"auth_token_file" flag:"auth-token-file" env:"N9E_TOKEN_FILE" description:"If set, the file that will be used to secure the secure port of the API server via token authentication."`
	ClusterAuthTokenFile string `json:"cluster_auth_token_file" flag:"cluster-auth-token-file" description:"If set, the file that will be used to secure the secure port of the API server via token authentication."`
	Fake                 bool   `json:"fake" flag:"fake-auth" default:"false" description:"If set, you can use auth token"`
	RootDir              string `json:"-"` // from agent.root_dir
	Token                string `json:"-"` // generate from auth_token_file
	ClusterToken         string `json:"-"` // generate from cluster_auth_token_file
}

func (p *Config) Validate() error {
	root := util.NewRootDir(p.RootDir)
	p.AuthTokenFile = root.Abs(p.AuthTokenFile, "etc", "auth_token")
	klog.V(1).InfoS("authentication", "auth_token_file", p.AuthTokenFile)

	var err error
	if p.Token, err = p.fetchAuthToken(true); err != nil {
		return err
	}

	return nil
}

type authModule struct {
	*Config
	name string
}

func newConfig() *Config {
	return &Config{}
}

func (p *authModule) init(ctx context.Context) error {
	c := configer.ConfigerMustFrom(ctx)
	cf := newConfig()
	cf.RootDir = c.GetString("agent.root_dir")

	if err := c.Read(modulePath, cf); err != nil {
		return err
	}
	p.Config = cf

	return authentication.RegisterTokenAuthn(func(_ context.Context) (authenticator.Token, error) {
		return p.newAuthenticator()
	})
}

type TokenAuthenticator struct {
	tokens map[string]*user.DefaultInfo
}

func (p *authModule) newAuthenticator() (*TokenAuthenticator, error) {
	tokens := map[string]*user.DefaultInfo{}
	tokens[p.Token] = &user.DefaultInfo{Name: "system:agentd", UID: "0"}
	klog.V(6).Infof("auth.token %s", p.Token)

	if p.Fake {
		tokens[fakeToken] = &user.DefaultInfo{Name: "system:fake", UID: "0"}
	}

	return &TokenAuthenticator{tokens: tokens}, nil
}

func (a *TokenAuthenticator) AuthenticateToken(ctx context.Context, value string) (*authenticator.Response, bool, error) {
	user, ok := a.tokens[value]
	if !ok {
		return nil, false, nil
	}
	return &authenticator.Response{
		User: user,
	}, true, nil
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
