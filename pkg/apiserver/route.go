package apiserver

import "github.com/yubo/apiserver/pkg/rest"

func (p *module) installWs(c rest.GoRestfulContainer) {
	p.installConfigWs(c)
	p.installStatusWs(c)
	p.installChecksWs(c)
	p.installStatsdWs(c)
	p.installGenericWs(c)
}

func (p *module) installConfigWs(c rest.GoRestfulContainer) {
	rest.SwaggerTagRegister("config", "n9e agentd config Api")
	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/api/v1/config",
		Tags:               []string{"config"},
		GoRestfulContainer: c,
		Routes: []rest.WsRoute{{
			Method: "GET", Scope: "read",
			SubPath: "/",
			Handle:  getFullRuntimeConfig,
			Desc:    "get full runtime config",
		}, {
			Method: "GET", Scope: "read",
			SubPath: "/jmx", // jmx/configs
			Handle:  getJMXConfigs,
			Desc:    "get jmx configs",
		}, {
			Method: "GET", Scope: "read",
			SubPath: "/check", // config-check
			Handle:  getConfigCheck,
			Desc:    "get config check",
		}, {
			Method: "GET", Scope: "read",
			SubPath: "/settings", // /config/list-runtime
			Handle:  getRuntimeConfigurableSettings,
			Desc:    "get runtime configure able settings",
		}, {
			Method: "GET", Scope: "read",
			SubPath: "/settings/{setting}",
			Handle:  getRuntimeConfig,
			Desc:    "get runtime config",
		}, {
			Method: "POST", Scope: "write",
			SubPath: "/settings/{setting}",
			Handle:  setRuntimeConfig,
			Desc:    "set runtime config",
		}},
	})
}

func (p *module) installStatusWs(c rest.GoRestfulContainer) {
	rest.SwaggerTagRegister("statsd", "n9e agentd statsd Api")
	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/api/v1/status",
		Tags:               []string{"status"},
		GoRestfulContainer: c,
		Routes: []rest.WsRoute{{
			Method: "GET", Scope: "read",
			SubPath: "/",
			Handle:  getStatus,
			Desc:    "get status",
		}, {
			Method: "GET", Scope: "read",
			SubPath: "/formatted",
			Handle:  getFormattedStatus,
			Desc:    "get formatted status",
		}, {
			Method: "GET", Scope: "read",
			SubPath: "/health",
			Handle:  getHealth,
			Desc:    "get health",
		}, {
			Method: "GET", Scope: "read", // py/status
			SubPath: "/py",
			Handle:  getPythonStatus,
			Desc:    "get python status",
		}, {
			Method: "POST", Scope: "write", // jmx/status
			SubPath: "/jmx",
			Handle:  setJMXStatus,
			Desc:    "set jmx status",
		}},
	})
}

func (p *module) installChecksWs(c rest.GoRestfulContainer) {
	rest.SwaggerTagRegister("checks", "n9e agentd checks Api")
	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/api/v1/checks",
		Tags:               []string{"checks"},
		GoRestfulContainer: c,
		Routes: []rest.WsRoute{{
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
		}},
	})
}

func (p *module) installStatsdWs(c rest.GoRestfulContainer) {
	rest.SwaggerTagRegister("status", "n9e agentd status Api")
	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/api/v1/statsd",
		Tags:               []string{"statsd"},
		GoRestfulContainer: c,
		Routes: []rest.WsRoute{{ // statsd-stats
			Method: "GET", Scope: "read",
			SubPath: "/stats",
			Handle:  getDogstatsdStats,
			Desc:    "get statsd stats",
		}, { // from grpc
			Method: "POST", Scope: "write",
			SubPath: "/capture-trigger",
			Handle:  statsdCaptureTrigger,
			Desc:    "triggers a dogstatsd traffic capture for the duration specified in the request. If a capture is already in progress, an error response is sent back",
		}, {
			Method: "POST", Scope: "write",
			SubPath: "/set-tagger-status",
			Handle:  statsdSetTaggerStatus,
			Desc:    "DogstatsdSetTaggerState allows setting a captured tagger state in the Tagger facilities. This endpoint is used when traffic replays are in progress. An empty state or nil request will result in the Tagger capture state being reset to nil.",
		}},
	})
}

func (p *module) installGenericWs(c rest.GoRestfulContainer) {
	rest.SwaggerTagRegister("generic", "n9e agentd generic Api")
	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/api/v1",
		GoRestfulContainer: c,
		Tags:               []string{"generic"},
		Routes: []rest.WsRoute{{
			Method: "GET", Scope: "read",
			SubPath: "/version", // versino
			Handle:  getVersion,
			Desc:    "get version info",
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
			SubPath: "/logs",
			Handle:  streamLogs,
			Desc:    "post stream logs",
		}, {
			Method: "GET", Scope: "read",
			SubPath: "/tagger",
			Handle:  getTaggerList,
			Desc:    "get tagger list",
		}, {
			Method: "GET", Scope: "read",
			SubPath: "/secrets",
			Handle:  secretInfo,
			Desc:    "get secrets info",
		}},
	})
}
