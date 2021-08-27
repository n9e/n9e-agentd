package apiserver

import "github.com/yubo/apiserver/pkg/rest"

var (
	apiV1Routes = []rest.WsRoute{
		{Method: "GET", SubPath: "/version", Handle: getVersion, Desc: "get version"},
		{Method: "GET", SubPath: "/hostname", Handle: getHostname, Desc: "get hostname"},
		{Method: "POST", SubPath: "/flare", Handle: makeFlare, Desc: "make flare"},
		{Method: "POST", SubPath: "/stop", Handle: stopAgent, Desc: "stop agent"},
		{Method: "GET", SubPath: "/status", Handle: getStatus, Desc: "get status"},
		{Method: "GET", SubPath: "/stream-logs", Handle: streamLogs, Desc: "post stream logs"},
		{Method: "GET", SubPath: "/statsd-stats", Handle: getDogstatsdStats, Desc: "get statsd stats"},
		{Method: "GET", SubPath: "/status/formatted", Handle: getFormattedStatus, Desc: "get formatted status"},
		{Method: "GET", SubPath: "/status/health", Handle: getHealth, Desc: "get health"},
		{Method: "GET", SubPath: "/py/status", Handle: getPythonStatus, Desc: "get python status"},
		{Method: "POST", SubPath: "/jmx/status", Handle: setJMXStatus, Desc: "set jmx status"},
		{Method: "GET", SubPath: "/jmx/configs", Handle: getJMXConfigs, Desc: "get jmx configs"},
		//{Method: "GET", SubPath: "/gui/csrf-token", Handle: nonHandle, Desc: "flare"},
		{Method: "GET", SubPath: "/config-check", Handle: getConfigCheck, Desc: "get config check"},
		{Method: "GET", SubPath: "/config", Handle: getFullRuntimeConfig, Desc: "get full runtime config"},
		{Method: "GET", SubPath: "/config/list-runtime", Handle: getRuntimeConfigurableSettings, Desc: "get runtime configure able settings"},
		{Method: "GET", SubPath: "/config/{setting}", Handle: getRuntimeConfig, Desc: "get runtime config"},
		{Method: "POST", SubPath: "/config/{setting}", Handle: setRuntimeConfig, Desc: "set runtime config"},
		{Method: "GET", SubPath: "/tagger-list", Handle: getTaggerList, Desc: "get tagger list"},
		{Method: "GET", SubPath: "/secrets", Handle: secretInfo, Desc: "get secrets info"},

		// from check
		{Method: "GET", SubPath: "/checks", Handle: unsupported, Desc: "get checks list"},
		{Method: "GET", SubPath: "/checks/{name}", Handle: unsupported, Desc: "get check detail"},
		{Method: "DELETE", SubPath: "/checks/{name}", Handle: unsupported, Desc: "get check detail"},
		{Method: "POST", SubPath: "/checks/{name}/reload", Handle: unsupported, Desc: "reload check"},

		// from grpc
		{Method: "POST", SubPath: "/statsd/capture-trigger", Handle: statsdCaptureTrigger, Desc: "triggers a dogstatsd traffic capture for the duration specified in the request. If a capture is already in progress, an error response is sent back"},
		{Method: "POST", SubPath: "/statsd/set-tagger-status", Handle: statsdSetTaggerStatus, Desc: "DogstatsdSetTaggerState allows setting a captured tagger state in the Tagger facilities. This endpoint is used when traffic replays are in progress. An empty state or nil request will result in the Tagger capture state being reset to nil."},
	}
)
