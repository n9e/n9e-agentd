package authentication

import (
	"context"

	"github.com/n9e/n9e-agentd/pkg/util"
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
}

func (p *Config) ValidatePath(rootDir string) error {
	root := util.NewRootDir(rootDir)
	p.AuthTokenFile = root.Abs(p.AuthTokenFile, "etc", "auth_token")
	klog.V(1).InfoS("authentication", "auth_token_file", p.AuthTokenFile)

	return nil
}

func (p *Config) Validate() error {
	return nil
}

type authModule struct {
	name   string
	token  string
	config *Config
}

func newConfig() *Config {
	return &Config{}
}

func (p *authModule) init(ctx context.Context) error {
	c := proc.ConfigerMustFrom(ctx)

	cf := newConfig()
	if err := c.Read(modulePath, cf); err != nil {
		return err
	}
	p.config = cf

	p.config.ValidatePath("")

	auth, err := p.newAuthToken()
	if err != nil {
		return err
	}

	return authentication.RegisterTokenAuthn(auth)
}

type TokenAuthenticator struct {
	tokens map[string]*user.DefaultInfo
}

func (p *authModule) newAuthToken() (*TokenAuthenticator, error) {
	authToken, err := CreateOrFetchToken()
	if err != nil {
		return nil, err
	}

	tokens := map[string]*user.DefaultInfo{}
	tokens[authToken] = &user.DefaultInfo{Name: "system:agentd", UID: "0"}

	if p.config.Fake {
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
