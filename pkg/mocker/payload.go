package mocker

// datadog-agent/pkg/logs/processor/json.go
type jsonPayload struct {
	Message   string `json:"message"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Hostname  string `json:"hostname"`
	Service   string `json:"service"`
	Source    string `json:"source"`
	Tags      string `json:"tags"`
	Ident     string `json:"ident"`
	Alias     string `json:"alias"`
}

type LogsPayload []jsonPayload

/*
var (
	routes = []struct {
		pattern string
		payload interface{}
	}{
		{api.N9eV1SeriesEndpoint, gogen.N9EMetricsPayload{}},
		{api.V1CheckRunsEndpoint, nil},
		{api.V1IntakeEndpoint, nil},
		{api.V1SketchSeriesEndpoint, nil},
		{api.V1ValidateEndpoint, nil},
		//{api.SeriesEndpoint, nil},
		{api.EventsEndpoint, gogen.EventsPayload{}},
		{api.ServiceChecksEndpoint, gogen.ServiceChecksPayload{}},
		{api.SketchSeriesEndpoint, gogen.SketchPayload{}},
		{api.HostMetadataEndpoint, nil},
		{api.MetadataEndpoint, nil},
		{api.ProcessesEndpoint, nil},
		{api.RtProcessesEndpoint, nil},
		{api.ContainerEndpoint, nil},
		{api.RtContainerEndpoint, nil},
		{api.ConnectionsEndpoint, nil},
		{api.OrchestratorEndpoint, nil},

		{api.RoutePathLogs, LogsPayload{}},
	}
)

func payloadHandle(payload interface{}) func(http.ResponseWriter, *http.Request) {
	return func(_ http.ResponseWriter, r *http.Request) {
		if payload != nil {
			readAll(r, reflect.New(reflect.TypeOf(payload)).Interface())
			return
		}

		readAll(r, nil)
	}
}

type decoder interface {
	Unmarshal(b []byte) error
}

func readAll(r *http.Request, payload interface{}) {
	reader := r.Body

	var err error
	if encoding := r.Header.Get("Content-Encoding"); encoding == "gzip" {
		if reader, err = gzip.NewReader(r.Body); err != nil {
			klog.Error(err)
			return
		}
		defer reader.Close()
	} else if encoding == "deflate" {
		if reader, err = zlib.NewReader(r.Body); err != nil {
			klog.Error(err)
			return
		}
		defer reader.Close()
	}

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		klog.Error(err)
		return
	}

	for k, v := range r.Header {
		klog.InfoS("header", "k", k, "v", v)
	}

	if payload == nil {
		klog.InfoS("recv", "method", r.Method, "url", r.URL, "payload.size", len(b))
		return
	}

	switch r.Header.Get("Content-Type") {
	case "application/x-protobuf":
		d, ok := payload.(decode)
		if !ok {
			klog.Infof("invalid payload")
			return
		}

		if err := d.Unmarshal(b); err != nil {
			klog.Errorf("%s %s Unmarshal err %s", r.Method, r.URL, err)
			return
		}
	case "application/json":
		fallthrough
	default:
		if err := json.Unmarshal(b, payload); err != nil {
			klog.Errorf("%s %s Unmarshal err %s body %s content-type %s", r.Method, r.URL, err, string(b), r.Header.Get("Content-Type"))
			return
		}
	}
	buf, _ := json.Marshal(payload)
	klog.InfoS("recf", "method", r.Method, "url", r.URL, "payload.size", len(b), "payload", string(buf))
}
*/
