// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package logs

import (
	"context"
	"time"

	coreConfig "github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/status/health"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util"
	"k8s.io/klog/v2"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/auditor"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/client"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/diagnostic"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/input/channel"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/input/container"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/input/file"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/input/journald"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/input/listener"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/input/traps"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/input/windowsevent"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/pipeline"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/restart"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/service"
	logstypes "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/types"
)

// Agent represents the data pipeline that collects, decodes,
// processes and sends logs to the backend
// + ------------------------------------------------------ +
// |                                                        |
// | Collector -> Decoder -> Processor -> Sender -> Auditor |
// |                                                        |
// + ------------------------------------------------------ +
type Agent struct {
	auditor                   auditor.Auditor
	destinationsCtx           *client.DestinationsContext
	pipelineProvider          pipeline.Provider
	inputs                    []restart.Restartable
	health                    *health.Handle
	diagnosticMessageReceiver *diagnostic.BufferedMessageReceiver
}

// NewAgent returns a new Logs Agent
func NewAgent(sources *config.LogSources, services *service.Services, processingRules []*logstypes.ProcessingRule, endpoints *logstypes.Endpoints) *Agent {
	health := health.RegisterLiveness("logs-agent")

	// setup the auditor
	// We pass the health handle to the auditor because it's the end of the pipeline and the most
	// critical part. Arguably it could also be plugged to the destination.
	auditorTTL := coreConfig.C.LogsConfig.AuditorTTL
	auditor := auditor.New(coreConfig.C.LogsConfig.RunPath, auditor.DefaultRegistryFilename, auditorTTL, health)
	destinationsCtx := client.NewDestinationsContext()
	diagnosticMessageReceiver := diagnostic.NewBufferedMessageReceiver()

	// setup the pipeline provider that provides pairs of processor and sender
	pipelineProvider := pipeline.NewProvider(config.NumberOfPipelines, auditor, diagnosticMessageReceiver, processingRules, endpoints, destinationsCtx)

	// setup the inputs
	inputs := []restart.Restartable{
		file.NewScanner(sources, coreConfig.C.LogsConfig.OpenFilesLimit, pipelineProvider, auditor, file.DefaultSleepDuration),
		container.NewLauncher(
			coreConfig.C.LogsConfig.ContainerCollectAll,
			coreConfig.C.LogsConfig.K8SContainerUseFile,
			coreConfig.C.LogsConfig.DockerContainerUseFile,
			coreConfig.C.LogsConfig.DockerContainerForceUseFile,
			coreConfig.C.LogsConfig.DockerClientReadTimeout,
			sources, services, pipelineProvider, auditor),
		listener.NewLauncher(sources, coreConfig.C.LogsConfig.FrameSize, pipelineProvider),
		journald.NewLauncher(sources, pipelineProvider, auditor),
		windowsevent.NewLauncher(sources, pipelineProvider),
		traps.NewLauncher(sources, pipelineProvider),
	}

	return &Agent{
		auditor:                   auditor,
		destinationsCtx:           destinationsCtx,
		pipelineProvider:          pipelineProvider,
		inputs:                    inputs,
		health:                    health,
		diagnosticMessageReceiver: diagnosticMessageReceiver,
	}
}

// NewServerless returns a Logs Agent instance to run in a serverless environment.
// The Serverless Logs Agent has only one input being the channel to receive the logs to process.
// It is using a NullAuditor because we've nothing to do after having sent the logs to the intake.
func NewServerless(sources *config.LogSources, services *service.Services, processingRules []*logstypes.ProcessingRule, endpoints *logstypes.Endpoints) *Agent {
	health := health.RegisterLiveness("logs-agent")

	diagnosticMessageReceiver := diagnostic.NewBufferedMessageReceiver()

	// setup the a null auditor, not tracking data in any registry
	auditor := auditor.NewNullAuditor()
	destinationsCtx := client.NewDestinationsContext()

	// setup the pipeline provider that provides pairs of processor and sender
	pipelineProvider := pipeline.NewServerlessProvider(config.NumberOfPipelines, auditor, processingRules, endpoints, destinationsCtx)

	// setup the inputs
	inputs := []restart.Restartable{
		channel.NewLauncher(sources, pipelineProvider),
	}

	return &Agent{
		auditor:                   auditor,
		destinationsCtx:           destinationsCtx,
		pipelineProvider:          pipelineProvider,
		inputs:                    inputs,
		health:                    health,
		diagnosticMessageReceiver: diagnosticMessageReceiver,
	}
}

// Start starts all the elements of the data pipeline
// in the right order to prevent data loss
func (a *Agent) Start() {
	starter := restart.NewStarter(a.destinationsCtx, a.auditor, a.pipelineProvider, a.diagnosticMessageReceiver)
	for _, input := range a.inputs {
		starter.Add(input)
	}
	starter.Start()
}

// Flush flushes synchronously the pipelines managed by the Logs Agent.
func (a *Agent) Flush(ctx context.Context) {
	a.pipelineProvider.Flush(ctx)
}

// Stop stops all the elements of the data pipeline
// in the right order to prevent data loss
func (a *Agent) Stop() {
	inputs := restart.NewParallelStopper()
	for _, input := range a.inputs {
		inputs.Add(input)
	}
	stopper := restart.NewSerialStopper(
		inputs,
		a.pipelineProvider,
		a.auditor,
		a.destinationsCtx,
		a.diagnosticMessageReceiver,
	)

	// This will try to stop everything in order, including the potentially blocking
	// parts like the sender. After StopTimeout it will just stop the last part of the
	// pipeline, disconnecting it from the auditor, to make sure that the pipeline is
	// flushed before stopping.
	// TODO: Add this feature in the stopper.
	c := make(chan struct{})
	go func() {
		stopper.Stop()
		close(c)
	}()
	timeout := coreConfig.C.LogsConfig.StopGracePeriod
	select {
	case <-c:
	case <-time.After(timeout):
		klog.Info("Timed out when stopping logs-agent, forcing it to stop now")
		// We force all destinations to read/flush all the messages they get without
		// trying to write to the network.
		a.destinationsCtx.Stop()
		// Wait again for the stopper to complete.
		// In some situation, the stopper unfortunately never succeed to complete,
		// we've already reached the grace period, give it some more seconds and
		// then force quit.
		timeout := time.NewTimer(5 * time.Second)
		select {
		case <-c:
		case <-timeout.C:
			klog.Warning("Force close of the Logs Agent, dumping the Go routines.")
			if stack, err := util.GetGoRoutinesDump(); err != nil {
				klog.Warningf("can't get the Go routines dump: %s\n", err)
			} else {
				klog.Warning(stack)
			}
		}
	}
}
