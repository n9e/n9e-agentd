package apiserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"path/filepath"
	"time"

	"github.com/n9e/n9e-agentd/pkg/options"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/yubo/golib/proc"
	"k8s.io/klog/v2"
)

const (
	moduleName          = "apiserver"
	authTokenMinimalLen = 32
	authTokenName       = "auth_token"
)

type config struct {
	IpcAddress    string        `yaml:"ipcAddress"`
	CmdPort       int           `yaml:"cmdPort"`
	AuthTokenFile string        `yaml:"authTokenFile"`
	ServerTimeout time.Duration `yaml:"serverTimeout"`
}

func (p config) String() string {
	return util.Prettify(p)
}

func (p *config) Validate() (err error) {
	klog.V(10).Infof("%s config\n%s", moduleName, p)
	return nil
}

func (p *module) prepare(configFileUsed string) (err error) {
	cf := p.config

	if p.ipcAddress, err = cf.getIPCAddress(); err != nil {
		return err
	}

	// get the transport we're going to use under HTTP
	if p.listener, err = net.Listen("tcp", p.ipcAddress); err != nil {
		// we use the listener to handle commands for the Agent, there's
		// no way we can recover from this error
		return fmt.Errorf("Unable to create the api server: %v", err)
	}

	coreModule := options.CoreMustFrom(p.ctx)
	if cf.AuthTokenFile != "" {
		p.tokenFile = cf.AuthTokenFile
	} else {
		p.tokenFile = filepath.Join(filepath.Dir(coreModule.ConfigFile()), authTokenName)
	}

	if p.token, err = p.createOrFetchAuthToken(); err != nil {
		return err
	}

	p.hostname = coreModule.Hostname()

	return nil
}

type module struct {
	config *config
	name   string

	tlsKeyPair  *tls.Certificate
	tlsCertPool *x509.CertPool
	tlsAddr     string

	ipcAddress string
	tokenFile  string
	token      string
	hostname   string
	listener   net.Listener

	ctx    context.Context
	cancel context.CancelFunc
}

var (
	_module = &module{name: moduleName}
	hookOps = []proc.HookOps{{
		Hook:     _module.start,
		Owner:    moduleName,
		HookNum:  proc.ACTION_START,
		Priority: proc.PRI_MODULE,
	}, {
		Hook:     _module.stop,
		Owner:    moduleName,
		HookNum:  proc.ACTION_STOP,
		Priority: proc.PRI_MODULE,
	}}
)

func (p *module) start(ops *proc.HookOps) error {
	ctx, configer := ops.ContextAndConfiger()
	p.ctx, p.cancel = context.WithCancel(ctx)

	cf := &config{}
	if err := configer.ReadYaml(p.name, cf); err != nil {
		return err
	}

	p.config = cf

	if err := p.prepare(configer.ConfigFilePath()); err != nil {
		return err
	}

	if err := p.startServer(); err != nil {
		return err
	}

	return nil
}

func (p *module) stop(ops *proc.HookOps) error {
	if p.cancel != nil {
		p.cancel()
	}

	if p.listener != nil {
		p.listener.Close()
	}
	return nil
}
