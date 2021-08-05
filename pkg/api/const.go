package api

const (
	// log
	RoutePathLogs = "/api/v1/logs/input"

	// collect rule
	RoutePathGetCollectRules        = "/v1/n9e/collect-rules-belong-to-ident"
	RoutePathGetCollectRulesSummary = "/v1/n9e/collect-rules-summary"

	// metrics
	// aggr.sender.{Gauge,Rate,count,MonotonicCount,Histogram,Historate}
	V1SeriesEndpoint       = "/v1/n9e/series"            // "series_v1"
	V1CheckRunsEndpoint    = "/api/v1/check_run"         // "check_run_v1"
	V1IntakeEndpoint       = "/intake/"                  // "intake" serializer.SendHostMetadata, ...
	V1SketchSeriesEndpoint = "/api/v1/sketches"          // "sketches_v1" nolint unused for now
	V1ValidateEndpoint     = "/api/v1/validate"          // "validate_v1"
	SeriesEndpoint         = "/api/v2/series"            // "series_v2"
	EventsEndpoint         = "/api/v2/events"            // "events_v2" // aggr.sender.Event()
	ServiceChecksEndpoint  = "/api/v2/service_checks"    // "services_checks_v2" // aggr.sender.ServiceCheck()
	SketchSeriesEndpoint   = "/api/beta/sketches"        // "sketches_v2"
	HostMetadataEndpoint   = "/api/v2/host_metadata"     // "host_metadata_v2"
	MetadataEndpoint       = "/api/v2/metadata"          // "metadata_v2"
	ProcessesEndpoint      = "/api/v1/collector/process" // "process"
	RtProcessesEndpoint    = "/api/v1/collector/rt"      // "rtprocess"
	ContainerEndpoint      = "/api/v1/container"         // "container"
	RtContainerEndpoint    = "/api/v1/container/rt"      // "rtcontainer"
	ConnectionsEndpoint    = "/api/v1/collector/conn"    // "connections"
	OrchestratorEndpoint   = "/api/v1/orchestrator"      // "orchestrator"

	// pkg/forwarder/telemetry.go
	//v1SeriesEndpoint       = transaction.Endpoint{Route: "/api/v1/series", Name: "series_v1"}
	//v1CheckRunsEndpoint    = transaction.Endpoint{Route: "/api/v1/check_run", Name: "check_run_v1"}
	//v1IntakeEndpoint       = transaction.Endpoint{Route: "/intake/", Name: "intake"}
	//v1SketchSeriesEndpoint = transaction.Endpoint{Route: "/api/v1/sketches", Name: "sketches_v1"} // nolint unused for now
	//v1ValidateEndpoint     = transaction.Endpoint{Route: "/api/v1/validate", Name: "validate_v1"}

	//seriesEndpoint        = transaction.Endpoint{Route: "/api/v2/series", Name: "series_v2"}
	//eventsEndpoint        = transaction.Endpoint{Route: "/api/v2/events", Name: "events_v2"}
	//serviceChecksEndpoint = transaction.Endpoint{Route: "/api/v2/service_checks", Name: "services_checks_v2"}
	//sketchSeriesEndpoint  = transaction.Endpoint{Route: "/api/beta/sketches", Name: "sketches_v2"}
	//hostMetadataEndpoint  = transaction.Endpoint{Route: "/api/v2/host_metadata", Name: "host_metadata_v2"}
	//metadataEndpoint      = transaction.Endpoint{Route: "/api/v2/metadata", Name: "metadata_v2"}

	//processesEndpoint    = transaction.Endpoint{Route: "/api/v1/collector", Name: "process"}
	//rtProcessesEndpoint  = transaction.Endpoint{Route: "/api/v1/collector", Name: "rtprocess"}
	//containerEndpoint    = transaction.Endpoint{Route: "/api/v1/container", Name: "container"}
	//rtContainerEndpoint  = transaction.Endpoint{Route: "/api/v1/container", Name: "rtcontainer"}
	//connectionsEndpoint  = transaction.Endpoint{Route: "/api/v1/collector", Name: "connections"}
	//orchestratorEndpoint = transaction.Endpoint{Route: "/api/v1/orchestrator", Name: "orchestrator"}

)
