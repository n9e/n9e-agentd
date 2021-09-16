package mocker

import (
	"net/http"

	"github.com/n9e/agent-payload/gogen"
	"github.com/yubo/apiserver/pkg/rest"
)

func (p *mocker) installDatadogWs(http rest.GoRestfulContainer) {
	rest.SwaggerTagRegister("datadog", "Datadog Api - datadog api")

	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/api/v1",
		GoRestfulContainer: http,
		Produces:           []string{rest.MIME_JSON, rest.MIME_TXT},
		Consumes:           []string{rest.MIME_JSON, rest.MIME_PROTOBUF},
		Tags:               []string{"datadog"},
		Routes: []rest.WsRoute{{
			Method:  "POST",
			SubPath: "/series",
			Handle:  v1SeriesEndpoint,
		}, {
			Method:  "POST",
			SubPath: "/check_run",
			Handle:  non,
		}, {
			Method:  "POST",
			SubPath: "/sketches",
			Handle:  non,
		}, {
			Method:  "POST",
			SubPath: "/validate",
			Handle:  non,
		}, {
			Method:  "POST",
			SubPath: "/collector",
			Handle:  non,
		}, {
			Method:  "POST",
			SubPath: "/container",
			Handle:  non,
		}, {
			Method:  "POST",
			SubPath: "/orchestrator",
			Handle:  non,
		}},
	})

	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/api/v2",
		GoRestfulContainer: http,
		Tags:               []string{"datadog"},
		Routes: []rest.WsRoute{{
			Method:  "POST",
			SubPath: "/series",
			Handle:  v2SeriesEndpoint,
		}, {
			Method:  "POST",
			SubPath: "/events",
			Handle:  non,
		}, {
			Method:  "POST",
			SubPath: "/service_checks",
			Handle:  non,
		}, {
			Method:  "POST",
			SubPath: "/host_metadata",
			Handle:  non,
		}, {
			Method:  "POST",
			SubPath: "/metadata",
			Handle:  non,
		}},
	})

	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/api/beta",
		GoRestfulContainer: http,
		Tags:               []string{"datadog"},
		Routes: []rest.WsRoute{{
			Method:  "POST",
			SubPath: "/sketches",
			Handle:  non,
		}},
	})

}

func v1SeriesEndpoint(w http.ResponseWriter, req *http.Request, _ *rest.NoneParam, data *gogen.MetricsPayload) (string, error) {
	return "", nil
}

func v2SeriesEndpoint(w http.ResponseWriter, req *http.Request, _ *rest.NoneParam, data *gogen.MetricsPayload) (string, error) {
	return "", nil
}
