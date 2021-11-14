package mocker

import (
	"io/ioutil"
	"net/http"

	"github.com/DataDog/datadog-agent/pkg/metadata/inventories"
	v5 "github.com/DataDog/datadog-agent/pkg/metadata/v5"
	"github.com/DataDog/datadog-agent/pkg/metrics"
	"github.com/n9e/agent-payload/gogen"
	"github.com/n9e/agent-payload/process"
	model "github.com/n9e/agent-payload/process"
	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/yubo/apiserver/pkg/rest"
	"github.com/yubo/golib/util"
	"k8s.io/klog/v2"
)

func (p *mocker) routesDDLogs() []rest.WsRoute {
	return []rest.WsRoute{{
		Method:  "POST",
		SubPath: "/v1/input",
		Handle:  p.logsInput,
	}}
}

func (p *mocker) routesDD() []rest.WsRoute {
	return []rest.WsRoute{{
		Method:  "POST",
		SubPath: "/v1/series",
		Handle:  p.series,
	}, {
		Method:       "POST",
		SubPath:      "/v1/check_run",
		InputPayload: []metrics.ServiceCheck{},
		Handle:       p.serviceChecks,
	}, {
		Method:  "POST",
		SubPath: "/v1/intake",
		Handle:  p.intake, //payloads("/v1/intake"),
	}, {
		Method:  "POST",
		SubPath: "/v1/sketches",
		Handle:  p.sketches,
	}, {
		Method:  "GET",
		SubPath: "/v1/validate",
		Handle:  p.validate,
	}, {
		Method:       "POST",
		SubPath:      "/v1/collector",
		InputPayload: process.Message{},
		Handle:       p.collector, //payloads("/v1/collector"),
	}, {
		Method:       "POST",
		SubPath:      "/v1/container",
		InputPayload: process.Message{},
		Handle:       p.container, //payloads("/v1/container"),
	}, {
		Method:       "POST",
		SubPath:      "/v1/orchestrator",
		InputPayload: process.Message{},
		Handle:       p.orchestrator, //payloads("v1/orchestrator"),
	}, {
		Method:  "POST",
		SubPath: "/v2/series",
		Handle:  p.series,
	}, {
		Method:  "POST",
		SubPath: "/v2/events",
		Handle:  p.events, //payloads("/v2/events"),
	}, {
		Method:       "POST",
		SubPath:      "/v2/service_checks",
		InputPayload: []metrics.ServiceCheck{},
		Handle:       p.serviceChecks, //payloads("v2/service_checks"),
	}, {
		Method:  "POST",
		SubPath: "/v2/host_metadata",
		Handle:  p.intake, //payloads("/v2/host_metadata"),
	}, {
		Method:  "POST",
		SubPath: "/v2/metadata",
		Handle:  p.intake, //payloads("/v2/metadata"),
	}, {
		Method:  "POST",
		SubPath: "/beta/sketches",
		Handle:  p.sketches,
	}}
}

func (p *mocker) installDatadogWs(http rest.GoRestfulContainer) {
	rest.SwaggerTagRegister("datadog", "Datadog API - datadog api")

	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/apis/logs.datadoghq.com",
		GoRestfulContainer: http,
		Produces:           []string{rest.MIME_JSON, rest.MIME_TXT},
		Consumes:           []string{rest.MIME_JSON, rest.MIME_PROTOBUF},
		Tags:               []string{"api groups"},
		Routes:             p.routesDDLogs(),
	})

	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/apis/datadoghq.com",
		GoRestfulContainer: http,
		Produces:           []string{rest.MIME_JSON, rest.MIME_TXT},
		Consumes:           []string{rest.MIME_JSON, rest.MIME_PROTOBUF},
		Tags:               []string{"api groups"},
		Routes:             p.routesDD(),
	})

	rest.WsRouteBuild(&rest.WsOption{
		Path:               "/api",
		GoRestfulContainer: http,
		Produces:           []string{rest.MIME_JSON, rest.MIME_TXT},
		Consumes:           []string{rest.MIME_JSON, rest.MIME_PROTOBUF},
		Tags:               []string{"datadog"},
		Routes:             append(p.routesDD(), p.routesDDLogs()...),
	})

}

func (p *mocker) series(w http.ResponseWriter, req *http.Request, _ *rest.NonParam, data *gogen.MetricsPayload) (string, error) {
	sp, _ := opentracing.StartSpanFromContext(req.Context(), "dd.series")
	defer sp.Finish()
	sp.LogFields(log.Object("series", data))

	klog.InfoS("series", "samples.len", len(data.Samples))
	return "", nil
}

func (p *mocker) serviceChecks(w http.ResponseWriter, req *http.Request, _ *rest.NonParam, data *metrics.ServiceChecks) error {
	sp, _ := opentracing.StartSpanFromContext(req.Context(), "dd.service_checks")
	defer sp.Finish()
	sp.LogFields(log.Object("service_checks", data))

	klog.InfoS("service_checks", "serviceChecks.len", len(*data))
	return nil
}

type IntakePayload struct {
	// SendHostMetadata
	v5.Payload

	// SendProcessesMetadata already include v5.Payload.ResourcesPayload
	// Resources *resources.Payload `json:"resources"`

	// SendMetadata
	Hostname      string                     `json:"hostname"`
	Timestamp     int64                      `json:"timestamp"`
	CheckMetadata *inventories.CheckMetadata `json:"check_metadata"`
	AgentMetadata *inventories.AgentMetadata `json:"agent_metadata"`
}

// SendProcessesMetadata
//   map[string]interface{} ./pkg/metadata/resources.go
//   *v5.Payload github.com/DataDog/datadog-agent/pkg/metadata/v5.Payload  pkg/metadata/host.go
func (p *mocker) intake(w http.ResponseWriter, req *http.Request, _ *rest.NonParam, data *IntakePayload) error {
	sp, _ := opentracing.StartSpanFromContext(req.Context(), "dd.intake")
	defer sp.Finish()
	sp.LogFields(log.Object("intake", data))

	klog.InfoS("intake-host-metadata", "uuid", data.UUID)
	klog.InfoS("intake-metadata", "hostname", data.Hostname, "payload", util.JsonStr(data))
	return nil
}

func (p *mocker) events(w http.ResponseWriter, req *http.Request, _ *rest.NonParam, data *metrics.Events) (string, error) {
	sp, _ := opentracing.StartSpanFromContext(req.Context(), "dd.events")
	defer sp.Finish()
	sp.LogFields(log.Object("events", data))

	klog.InfoS("events", "samples.len", len(*data))
	return "", nil
}

func (p *mocker) sketches(w http.ResponseWriter, req *http.Request, _ *rest.NonParam, data *metrics.SketchSeriesList) error {
	sp, _ := opentracing.StartSpanFromContext(req.Context(), "dd.sketches")
	defer sp.Finish()
	sp.LogFields(log.Object("sketches", data))

	klog.InfoS("sketches", "sketches.len", len(*data))
	return nil
}

type validateInput struct {
	ApiKey string `param:"query" name:"api_key"`
}

func (p *mocker) validate(w http.ResponseWriter, req *http.Request, in *validateInput) error {
	sp, _ := opentracing.StartSpanFromContext(req.Context(), "dd.validate")
	defer sp.Finish()
	sp.LogFields(log.Object("validate_param", in))

	klog.InfoS("validate", "api_key", in.ApiKey)
	return nil
}

type collectorInput struct {
	// HostHeader contains the hostname of the payload
	HostHeader string `param:"header" name:"X-Dd-Hostname"`
	// ContainerCountHeader contains the container count in the payload
	ContainerCountHeader int `param:"header" name:"X-Dd-ContainerCount"`
	// ProcessVersionHeader holds the process agent version sending the payload
	ProcessVersionHeader string `param:"header" name:"X-Dd-Processagentversion"`
	// ClusterIDHeader contains the orchestrator cluster ID of this agent
	ClusterIDHeader string `param:"header" name:"X-Dd-Orchestrator-ClusterID"`
	// TimestampHeader contains the timestamp that the check data was created
	TimestampHeader int64 `param:"header" name:"X-DD-Agent-Timestamp"`
}

// processesEndpoint
// rtProcessesEndpoint
// connectionsEndpoint
func (p *mocker) collector(w http.ResponseWriter, req *http.Request, in *collectorInput) error {
	sp, _ := opentracing.StartSpanFromContext(req.Context(), "dd.collector")
	defer sp.Finish()

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		sp.LogFields(log.Error(err))
		return err
	}

	m, err := model.DecodeMessage(b)
	if err != nil {
		sp.LogFields(log.Error(err))
		return err
	}

	sp.LogFields(
		log.Object("collector.params", in),
		log.Object("collector.payload.header", m.Header),
		log.String("collector.payload.body", m.Body.String()),
	)

	klog.InfoS("collector", "type", m.Header.Type.String(), "body.size", m.Body.Size())
	return nil
}

// containerEndpoint
// rtContainerEndpoint
func (p *mocker) container(w http.ResponseWriter, req *http.Request, in *collectorInput) error {
	sp, ctx := opentracing.StartSpanFromContext(req.Context(), "dd.container")
	defer sp.Finish()

	return p.collector(w, req.WithContext(ctx), in)
}

// orchestratorEndpoint
// SendOrchestratorMetadata
func (p *mocker) orchestrator(w http.ResponseWriter, req *http.Request, in *collectorInput) error {
	sp, ctx := opentracing.StartSpanFromContext(req.Context(), "dd.orchestrator")
	defer sp.Finish()

	return p.collector(w, req.WithContext(ctx), in)
}

func (p *mocker) logsInput(w http.ResponseWriter, req *http.Request, _ *rest.NonParam, data *api.LogsPayload) error {
	sp, _ := opentracing.StartSpanFromContext(req.Context(), "dd.collector")
	defer sp.Finish()

	sp.LogFields(log.Object("logs", data))

	klog.InfoS("logs_input", "logs.len", len(*data))
	return nil
}
