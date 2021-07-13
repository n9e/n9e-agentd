// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package forwarder

import (
	"expvar"

	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/n9e/n9e-agentd/pkg/forwarder/transaction"
	"github.com/n9e/n9e-agentd/pkg/telemetry"
)

var (
	transactionsIntakePod        = expvar.Int{}
	transactionsIntakeDeployment = expvar.Int{}
	transactionsIntakeReplicaSet = expvar.Int{}
	transactionsIntakeService    = expvar.Int{}
	transactionsIntakeNode       = expvar.Int{}

	v1CheckRunsEndpoint = transaction.Endpoint{Route: api.RoutePathCheckRuns, Name: "check_run_v1"}
	v1IntakeEndpoint    = transaction.Endpoint{Route: api.RoutePathIntake, Name: "intake"}
	v1ValidateEndpoint  = transaction.Endpoint{Route: api.RoutePathValidate, Name: "validate_v1"}

	seriesEndpoint        = transaction.Endpoint{Route: api.RoutePathSeries, Name: "series_v2"}
	eventsEndpoint        = transaction.Endpoint{Route: api.RoutePathEvents, Name: "events_v2"}
	serviceChecksEndpoint = transaction.Endpoint{Route: api.RoutePathServiceChecks, Name: "services_checks_v2"}
	sketchSeriesEndpoint  = transaction.Endpoint{Route: api.RoutePathSketchSeries, Name: "sketches_v2"}
	//hostMetadataEndpoint  = transaction.Endpoint{Route: api.RoutePathHostMetadata, Name: "host_metadata_v2"}
	metadataEndpoint = transaction.Endpoint{Route: api.RoutePathMetadata, Name: "metadata_v2"}

	processesEndpoint    = transaction.Endpoint{Route: api.RoutePathCollector, Name: "process"}
	rtProcessesEndpoint  = transaction.Endpoint{Route: api.RoutePathCollector, Name: "rtprocess"}
	containerEndpoint    = transaction.Endpoint{Route: api.RoutePathContainer, Name: "container"}
	rtContainerEndpoint  = transaction.Endpoint{Route: api.RoutePathContainer, Name: "rtcontainer"}
	connectionsEndpoint  = transaction.Endpoint{Route: api.RoutePathCollector, Name: "connections"}
	orchestratorEndpoint = transaction.Endpoint{Route: api.RoutePathOrchestrator, Name: "orchestrator"}

	transactionsDroppedOnInput       = expvar.Int{}
	transactionsInputBytesByEndpoint = expvar.Map{}
	transactionsInputCountByEndpoint = expvar.Map{}
	transactionsRequeued             = expvar.Int{}
	transactionsRequeuedByEndpoint   = expvar.Map{}
	transactionsRetried              = expvar.Int{}
	transactionsRetriedByEndpoint    = expvar.Map{}
	transactionsRetryQueueSize       = expvar.Int{}

	tlmTxInputBytes = telemetry.NewCounter("transactions", "input_bytes",
		[]string{"domain", "endpoint"}, "Incoming transaction sizes in bytes")
	tlmTxInputCount = telemetry.NewCounter("transactions", "input_count",
		[]string{"domain", "endpoint"}, "Incoming transaction count")
	tlmTxDroppedOnInput = telemetry.NewCounter("transactions", "dropped_on_input",
		[]string{"domain", "endpoint"}, "Count of transactions dropped on input")
	tlmTxRequeued = telemetry.NewCounter("transactions", "requeued",
		[]string{"domain", "endpoint"}, "Transaction requeue count")
	tlmTxRetried = telemetry.NewCounter("transactions", "retries",
		[]string{"domain", "endpoint"}, "Transaction retry count")
	tlmTxRetryQueueSize = telemetry.NewGauge("transactions", "retry_queue_size",
		[]string{"domain"}, "Retry queue size")
)

func init() {
	initOrchestratorExpVars()
	initTransactionsExpvars()
	initForwarderHealthExpvars()
	initEndpointExpvars()
}

func initEndpointExpvars() {
	endpoints := []transaction.Endpoint{
		connectionsEndpoint,
		containerEndpoint,
		eventsEndpoint,
		//hostMetadataEndpoint,
		metadataEndpoint,
		orchestratorEndpoint,
		processesEndpoint,
		rtContainerEndpoint,
		rtProcessesEndpoint,
		seriesEndpoint,
		serviceChecksEndpoint,
		sketchSeriesEndpoint,
		v1CheckRunsEndpoint,
		v1IntakeEndpoint,
		v1ValidateEndpoint,
	}

	for _, endpoint := range endpoints {
		transaction.TransactionsSuccessByEndpoint.Set(endpoint.Name, expvar.NewInt(endpoint.Name))
	}
}

func initOrchestratorExpVars() {
	transaction.TransactionsExpvars.Set("Pods", &transactionsIntakePod)
	transaction.TransactionsExpvars.Set("Deployments", &transactionsIntakeDeployment)
	transaction.TransactionsExpvars.Set("ReplicaSets", &transactionsIntakeReplicaSet)
	transaction.TransactionsExpvars.Set("Services", &transactionsIntakeService)
	transaction.TransactionsExpvars.Set("Nodes", &transactionsIntakeNode)
}

func initTransactionsExpvars() {
	transactionsInputBytesByEndpoint.Init()
	transactionsInputCountByEndpoint.Init()
	transactionsRequeuedByEndpoint.Init()
	transactionsRetriedByEndpoint.Init()
	transaction.TransactionsExpvars.Set("InputCountByEndpoint", &transactionsInputCountByEndpoint)
	transaction.TransactionsExpvars.Set("InputBytesByEndpoint", &transactionsInputBytesByEndpoint)
	transaction.TransactionsExpvars.Set("DroppedOnInput", &transactionsDroppedOnInput)
	transaction.TransactionsExpvars.Set("Requeued", &transactionsRequeued)
	transaction.TransactionsExpvars.Set("RequeuedByEndpoint", &transactionsRequeuedByEndpoint)
	transaction.TransactionsExpvars.Set("Retried", &transactionsRetried)
	transaction.TransactionsExpvars.Set("RetriedByEndpoint", &transactionsRetriedByEndpoint)
	transaction.TransactionsExpvars.Set("RetryQueueSize", &transactionsRetryQueueSize)
}
