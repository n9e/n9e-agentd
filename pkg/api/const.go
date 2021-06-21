package api

const (
	RoutePathSeries        = "/v1/n9e/series"
	RoutePathCheckRuns     = "/api/v1/check_run"
	RoutePathIntake        = "/intake/" // hostMetadata
	RoutePathValidate      = "/api/v1/validate"
	RoutePathEvents        = "/api/v2/events"
	RoutePathServiceChecks = "/api/v2/service_checks"
	RoutePathSketchSeries  = "/api/beta/sketches"
	RoutePathMetadata      = "/api/v2/metadata"
	RoutePathCollector     = "/api/v1/collector"
	RoutePathContainer     = "/api/v1/container"
	RoutePathOrchestrator  = "/api/v1/orchestrator"
	RoutePathLogs          = "/v1/logs/input"
	//RoutePathSeries        = "/api/v2/series"
	//RoutePathHostMetadata  = "/api/v2/host_metadata"

	//RoutePathGetCollectRules        = "/api/collect-rules"
	RoutePathGetCollectRules        = "/v1/n9e/collect-rules-belong-to-ident"
	RoutePathGetCollectRulesSummary = "/v1/n9e/collect-rules-summary"
)
