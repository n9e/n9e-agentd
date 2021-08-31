package apiserver

import "github.com/yubo/apiserver/pkg/rest"

var (
	apiV1Routes = []rest.WsRoute{{
		Method: "GET", Scope: "read",
		SubPath: "/version1", // versino
		Handle:  getVersion,
		Desc:    "get version",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/hostname", // hostname
		Handle:  getHostname,
		Desc:    "get hostname",
	}, {
		Method: "POST", Scope: "write",
		SubPath: "/flare",
		Handle:  makeFlare,
		Desc:    "make flare",
	}, {
		Method: "POST", Scope: "write",
		SubPath: "/stop",
		Handle:  stopAgent,
		Desc:    "stop agent",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/stream-logs",
		Handle:  streamLogs,
		Desc:    "post stream logs",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/status",
		Handle:  getStatus,
		Tags:    []string{"status"},
		Desc:    "get status",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/status/formatted",
		Handle:  getFormattedStatus,
		Tags:    []string{"status"},
		Desc:    "get formatted status",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/status/health",
		Handle:  getHealth,
		Tags:    []string{"status"},
		Desc:    "get health",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/status/py", // py/status
		Handle:  getPythonStatus,
		Tags:    []string{"status"},
		Desc:    "get python status",
	}, {
		Method: "POST", Scope: "write",
		SubPath: "/status/jmx", // jmx/status
		Handle:  setJMXStatus,
		Tags:    []string{"status"},
		Desc:    "set jmx status",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/config/jmx", // jmx/configs
		Handle:  getJMXConfigs,
		Tags:    []string{"config"},
		Desc:    "get jmx configs",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/config/check", // config-check
		Handle:  getConfigCheck,
		Tags:    []string{"config"},
		Desc:    "get config check",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/config",
		Handle:  getFullRuntimeConfig,
		Tags:    []string{"config"},
		Desc:    "get full runtime config",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/config/setting", // /config/list-runtime
		Handle:  getRuntimeConfigurableSettings,
		Tags:    []string{"config"},
		Desc:    "get runtime configure able settings",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/config/setting/{setting}",
		Handle:  getRuntimeConfig,
		Tags:    []string{"config"},
		Desc:    "get runtime config",
	}, {
		Method: "POST", Scope: "write",
		SubPath: "/config/setting/{setting}",
		Handle:  setRuntimeConfig,
		Tags:    []string{"config"},
		Desc:    "set runtime config",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/tagger-list",
		Handle:  getTaggerList,
		Desc:    "get tagger list",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/secrets",
		Handle:  secretInfo,
		Desc:    "get secrets info",
	}, {
		// from check
		Method: "GET", Scope: "read",
		SubPath: "/checks",
		Handle:  unsupported,
		Tags:    []string{"checks"},
		Desc:    "get checks list",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/checks/{name}",
		Handle:  unsupported,
		Tags:    []string{"checks"},
		Desc:    "get check detail",
	}, {
		Method: "DELETE", Scope: "write",
		SubPath: "/checks/{name}",
		Handle:  unsupported,
		Tags:    []string{"checks"},
		Desc:    "get check detail",
	}, {
		Method: "POST", Scope: "write",
		SubPath: "/checks/{name}/reload",
		Handle:  unsupported,
		Tags:    []string{"checks"},
		Desc:    "reload check",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/statsd/stats", // statsd-stats
		Handle:  getDogstatsdStats,
		Tags:    []string{"statsd"},
		Desc:    "get statsd stats",
	}, {
		// from grpc
		Method: "POST", Scope: "write",
		SubPath: "/statsd/capture-trigger",
		Handle:  statsdCaptureTrigger,
		Tags:    []string{"statsd"},
		Desc:    "triggers a dogstatsd traffic capture for the duration specified in the request. If a capture is already in progress, an error response is sent back",
	}, {
		Method: "POST", Scope: "write",
		SubPath: "/statsd/set-tagger-status",
		Handle:  statsdSetTaggerStatus,
		Tags:    []string{"statsd"},
		Desc:    "DogstatsdSetTaggerState allows setting a captured tagger state in the Tagger facilities. This endpoint is used when traffic replays are in progress. An empty state or nil request will result in the Tagger capture state being reset to nil."},
	}
)

func (p *module) installWs(c rest.GoRestfulContainer) {
	rest.SwaggerTagRegister("status", "n9e agentd status Api")
	rest.SwaggerTagRegister("config", "n9e agentd config Api")
	rest.SwaggerTagRegister("checks", "n9e agentd checks Api")
	rest.SwaggerTagRegister("statsd", "n9e agentd statsd Api")
	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/api/v1",
		GoRestfulContainer: c,
		Routes:             apiV1Routes,
	})
}
