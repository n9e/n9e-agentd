package apiserver

import "github.com/yubo/apiserver/pkg/rest"

var (
	apiV1Routes = []rest.WsRoute{{
		Method: "GET", Scope: "read",
		SubPath: "/version",
		Handle:  getVersion,
		Desc:    "get version",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/hostname",
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
		SubPath: "/status",
		Handle:  getStatus,
		Desc:    "get status",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/stream-logs",
		Handle:  streamLogs,
		Desc:    "post stream logs",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/statsd-stats",
		Handle:  getDogstatsdStats,
		Desc:    "get statsd stats",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/status/formatted",
		Handle:  getFormattedStatus,
		Desc:    "get formatted status",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/status/health",
		Handle:  getHealth,
		Desc:    "get health",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/py/status",
		Handle:  getPythonStatus,
		Desc:    "get python status",
	}, {
		Method: "POST", Scope: "write",
		SubPath: "/jmx/status",
		Handle:  setJMXStatus,
		Desc:    "set jmx status",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/jmx/configs",
		Handle:  getJMXConfigs,
		Desc:    "get jmx configs",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/config-check",
		Handle:  getConfigCheck,
		Desc:    "get config check",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/config",
		Handle:  getFullRuntimeConfig,
		Desc:    "get full runtime config",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/config/list-runtime",
		Handle:  getRuntimeConfigurableSettings,
		Desc:    "get runtime configure able settings",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/config/{setting}",
		Handle:  getRuntimeConfig,
		Desc:    "get runtime config",
	}, {
		Method: "POST", Scope: "write",
		SubPath: "/config/{setting}",
		Handle:  setRuntimeConfig,
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
		Desc:    "get checks list",
	}, {
		Method: "GET", Scope: "read",
		SubPath: "/checks/{name}",
		Handle:  unsupported,
		Desc:    "get check detail",
	}, {
		Method: "DELETE", Scope: "write",
		SubPath: "/checks/{name}",
		Handle:  unsupported,
		Desc:    "get check detail",
	}, {
		Method: "POST", Scope: "write",
		SubPath: "/checks/{name}/reload",
		Handle:  unsupported,
		Desc:    "reload check",
	}, {
		// from grpc
		Method: "POST", Scope: "write",
		SubPath: "/statsd/capture-trigger",
		Handle:  statsdCaptureTrigger,
		Desc:    "triggers a dogstatsd traffic capture for the duration specified in the request. If a capture is already in progress, an error response is sent back",
	}, {
		Method: "POST", Scope: "write",
		SubPath: "/statsd/set-tagger-status",
		Handle:  statsdSetTaggerStatus,
		Desc:    "DogstatsdSetTaggerState allows setting a captured tagger state in the Tagger facilities. This endpoint is used when traffic replays are in progress. An empty state or nil request will result in the Tagger capture state being reset to nil."},
	}
)
