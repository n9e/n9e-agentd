package apiserver

// apiserverConfig contains the config while running a generic api server.
type apiserverConfig struct {
	Enabled bool `json:"enabled" default:"false" flag:"apiserver-enable" description:"api server enable"`

	ExternalHost string `json:"externalHost" flag:"external-hostname" description:"The hostname to use when generating externalized URLs for this master (e.g. Swagger API Docs or OpenID Discovery)."`

	Host string/*net.IP*/ `json:"address" default:"127.0.0.1" flag:"bind-host" description:"The IP address on which to listen for the --bind-port port. The associated interface(s) must be reachable by the rest of the cluster, and by CLI/web clients. If blank or an unspecified address (127.0.0.1 or ::), all interfaces will be used."` // BindAddress

	Port int `json:"port" default:"8010" flag:"bind-port" description:"The port on which to serve HTTPS with authentication and authorization. It cannot be switched off with 0."` // BindPort is ignored when Listener is set, will serve https even with 0.

	Network string `json:"bindNetwork" flag:"cors-allowed-origins" description:"List of allowed origins for CORS, comma separated.  An allowed origin can be a regular expression to support subdomain matching. If this list is empty CORS will not be enabled."` // BindNetwork is the type of network to bind to - defaults to "tcp", accepts "tcp", "tcp4", and "tcp6".

	CorsAllowedOriginList []string `json:"corsAllowedOriginList"`

	HSTSDirectives []string `json:"hstsDirectives" flag:"strict-transport-security-directives" description:"List of directives for HSTS, comma separated. If this list is empty, then HSTS directives will not be added. Example: 'max-age=31536000,includeSubDomains,preload'"`

	MaxRequestsInFlight int `json:"maxRequestsInFlight" default:"400" flag:"max-requests-inflight" description:"The maximum number of non-mutating requests in flight at a given time. When the server exceeds this, it rejects requests. Zero for no limit."`

	MaxMutatingRequestsInFlight int `json:"maxMutatingRequestsInFlight" default:"200" flag:"max-mutating-requests-inflight" description:"The maximum number of mutating requests in flight at a given time. When the server exceeds this, it rejects requests. Zero for no limit."`

	RequestTimeout int `json:"requestTimeout" default:"60" flag:"request-timeout" description:"An optional field indicating the duration a handler must keep a request open before timing it out. This is the default request timeout for requests but may be overridden by flags such as --min-request-timeout for specific types of requests."`

	GoawayChance float64 `json:"goawayChance" flag:"goaway-chance" description:"To prevent HTTP/2 clients from getting stuck on a single apiserver, randomly close a connection (GOAWAY). The client's other in-flight requests won't be affected, and the client will reconnect, likely landing on a different apiserver after going through the load balancer again. This argument sets the fraction of requests that will be sent a GOAWAY. Clusters with single apiservers, or which don't use a load balancer, should NOT enable this. Min is 0 (off), Max is .02 (1/50 requests); .001 (1/1000) is a recommended starting point."`

	LivezGracePeriod  int `json:"livezGracePeriod" flag:"livez-grace-period" description:"This option represents the maximum amount of time it should take for apiserver to complete its startup sequence and become live. From apiserver's start time to when this amount of time has elapsed, /livez will assume that unfinished post-start hooks will complete successfully and therefore return true."`
	MinRequestTimeout int `json:"minRequestTimeout" default:"1800" flag:"min-request-timeout" description:"An optional field indicating the minimum number of seconds a handler must keep a request open before timing it out. Currently only honored by the watch request handler, which picks a randomized value above this number as the connection timeout, to spread out load."`

	ShutdownTimeout       int `json:"shutdownTimeout" default:"60" description:"ShutdownTimeout is the timeout used for server shutdown. This specifies the timeout before server gracefully shutdown returns."`
	ShutdownDelayDuration int `json:"shutdownDelayDuration" flag:"shutdown-delay-duration" description:"Time to delay the termination. During that time the server keeps serving requests normally. The endpoints /healthz and /livez will return success, but /readyz immediately returns failure. Graceful termination starts after this delay has elapsed. This can be used to allow load balancer to stop sending traffic to this server."`

	// The limit on the request body size that would be accepted and
	// decoded in a write request. 0 means no limit.
	// We intentionally did not add a flag for this option. Users of the
	// apiserver library can wire it to a flag.
	MaxRequestBodyBytes int64 `json:"maxRequestBodyBytes" default:"3145728" flag:"max-resource-write-bytes" description:"The limit on the request body size that would be accepted and decoded in a write request."`

	EnablePriorityAndFairness bool `json:"enablePriorityAndFairness" default:"true" flag:"enable-priority-and-fairness" description:"If true and the APIPriorityAndFairness feature gate is enabled, replace the max-in-flight handler with an enhanced one that queues and dispatches with priority and fairness"`

	EnableIndex               bool `json:"enableIndex" default:"false"`
	EnableProfiling           bool `json:"enableProfiling" default:"false"`
	EnableContentionProfiling bool `json:"enableContentionProfiling" default:"false"`
	EnableMetrics             bool `json:"enableMetrics" default:"false"`
}

// authzConfig contains all build-in authorization options for API Server
type authzConfig struct {
	Modes []string `json:"modes" default:"Login" flag:"authorization-mode" description:"Ordered list of plug-ins to do authorization on secure port."`
}
