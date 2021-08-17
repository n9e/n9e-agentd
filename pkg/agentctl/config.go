package agentctl

import (
	"strings"

	"github.com/fatih/color"
	"github.com/n9e/n9e-agentd/pkg/util"
)

type Telemetry struct {
	Port int `json:"port" default:"8070"` // expvar_port
}

type agentConfig struct {
	RootDir       string    `json:"root_dir" flag:"root" env:"N9E_ROOT_DIR" description:"root dir path"` // e.g. /opt/n9e/agentd
	ConfdPath     string    `json:"confd_path"`                                                          // confd_path
	PythonVersion string    `json:"python_version" default:"3"`                                          // python_version
	Token         string    `json:"token" env:"AGENTD_TOKEN" description:"token"`
	Timeout       int       `json:"tiemout" env:"AGENTD_TIMEOUT" flag:"timeout" default:"5" description:"timeout(Second)"`
	DisablePage   bool      `json:"disable_page" flag:"disable-page" default:"false" env:"AGENTD_DISABLE_PAGE"`
	PageSize      int       `json:"page_size" flag:"page-size" default:"10" env:"AGENTD_PAGE_SIZE"`
	NoColor       bool      `json:"no_color" flag:"no-color,n" default:"false" env:"AGENTD_NO_COLOR" description:"disable color output"`
	Telemetry     Telemetry `json:"telemetry,inline"` // telemetry
	PythonHome    string    `json:"python_home"`
}

func (p *agentConfig) Validate() (err error) {
	if p.NoColor {
		color.NoColor = true
	}

	if p.RootDir, err = util.ResolveRootPath(p.RootDir); err != nil {
		return err
	}
	p.ConfdPath = util.AbsPath(p.ConfdPath, p.RootDir, "conf.d")
	p.PythonHome = util.AbsPath(p.PythonHome, p.RootDir, "embedded")
	return nil
}

type apiserverConfig struct {
	Host string `json:"address" default:"127.0.0.1" flag:"bind-host" env:"N9E_API_HOST" description:"The IP address on which to listen for the --bind-port port. The associated interface(s) must be reachable by the rest of the cluster, and by CLI/web clients. If blank or an unspecified address (127.0.0.1 or ::), all interfaces will be used."` // BindAddress
	Port int    `json:"port" default:"15001" flag:"bind-port" env:"N9E_API_PORT" description:"The port on which to serve HTTPS with authentication and authorization. It cannot be switched off with 0."`                                                                                                                                               // BindPort is ignored when Listener is set, will serve https even with 0.
}

func (p *apiserverConfig) Validate() error {
	p.Host = strings.TrimRight(p.Host, "/")
	if p.Host == "" {
		p.Host = "localhost"
	}
	return nil
}

type authConfig struct {
	AuthTokenFile string `json:"auth_token_file" flag:"auth-token-file" env:"N9E_TOKEN_FILE" default:"./etc/auth_token" description:"If set, the file that will be used to secure the secure port of the API server via token authentication."`
}

func (p *authConfig) Validate() error {
	return nil
}
