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
	"github.com/yubo/apiserver/pkg/rest"
	"k8s.io/klog/v2"
)

func (p *mocker) routesDDLogs() []rest.WsRoute {
	return []rest.WsRoute{{
		Method:  "POST",
		SubPath: "/v1/input",
		Handle:  p.series,
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

func (p *mocker) series(w http.ResponseWriter, req *http.Request, _ *rest.NoneParam, data *gogen.MetricsPayload) (string, error) {
	klog.InfoS("series", "samples.len", len(data.Samples))
	return "", nil
}

func (p *mocker) serviceChecks(w http.ResponseWriter, req *http.Request, _ *rest.NoneParam, data *metrics.ServiceChecks) error {
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
func (p *mocker) intake(w http.ResponseWriter, req *http.Request, _ *rest.NoneParam, data *IntakePayload) error {
	klog.InfoS("intake-host-metadata", "uuid", data.UUID)
	klog.InfoS("intake-metadata", "hostname", data.Hostname)
	return nil
}

func (p *mocker) events(w http.ResponseWriter, req *http.Request, _ *rest.NoneParam, data *metrics.Events) (string, error) {
	klog.InfoS("series", "samples.len", len(*data))
	return "", nil
}

func (p *mocker) sketches(w http.ResponseWriter, req *http.Request, _ *rest.NoneParam, data *metrics.SketchSeriesList) error {
	klog.InfoS("sketches", "sketches.len", len(*data))
	return nil
}

type validateInput struct {
	ApiKey string `param:"query" name:"api_key"`
}

func (p *mocker) validate(w http.ResponseWriter, req *http.Request, in *validateInput) error {
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
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	m, err := model.DecodeMessage(b)
	if err != nil {
		return err
	}

	klog.InfoS("collector", "type", m.Header.Type.String(), "body.size", m.Body.Size())
	return nil
}

// containerEndpoint
// rtContainerEndpoint
func (p *mocker) container(w http.ResponseWriter, req *http.Request, in *collectorInput) error {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	m, err := model.DecodeMessage(b)
	if err != nil {
		return err
	}

	klog.InfoS("collector", "type", m.Header.Type.String(), "body.size", m.Body.Size())
	return nil
}

// orchestratorEndpoint
// SendOrchestratorMetadata
func (p *mocker) orchestrator(w http.ResponseWriter, req *http.Request, in *collectorInput) error {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	m, err := model.DecodeMessage(b)
	if err != nil {
		return err
	}

	klog.InfoS("orchestrator", "type", m.Header.Type.String(), "body.size", m.Body.Size())
	return nil
}
