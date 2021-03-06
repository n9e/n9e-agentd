// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package logs

import (
	"context"
	"time"

	"github.com/DataDog/datadog-agent/pkg/status/health"
	"github.com/DataDog/datadog-agent/pkg/util"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	coreConfig "github.com/n9e/n9e-agentd/pkg/config"

	"github.com/DataDog/datadog-agent/pkg/logs/auditor"
	"github.com/DataDog/datadog-agent/pkg/logs/client"
	"github.com/DataDog/datadog-agent/pkg/logs/config"
	"github.com/DataDog/datadog-agent/pkg/logs/diagnostic"
	"github.com/DataDog/datadog-agent/pkg/logs/input/channel"
	"github.com/DataDog/datadog-agent/pkg/logs/input/container"
	"github.com/DataDog/datadog-agent/pkg/logs/input/docker"
	"github.com/DataDog/datadog-agent/pkg/logs/input/file"
	"github.com/DataDog/datadog-agent/pkg/logs/input/journald"
	"github.com/DataDog/datadog-agent/pkg/logs/input/kubernetes"
	"github.com/DataDog/datadog-agent/pkg/logs/input/listener"
	"github.com/DataDog/datadog-agent/pkg/logs/input/traps"
	"github.com/DataDog/datadog-agent/pkg/logs/input/windowsevent"
	"github.com/DataDog/datadog-agent/pkg/logs/pipeline"
	"github.com/DataDog/datadog-agent/pkg/logs/restart"
	"github.com/DataDog/datadog-agent/pkg/logs/service"
	"github.com/n9e/n9e-agentd/pkg/config/logs"
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
func NewAgent(sources *config.LogSources, services *service.Services, processingRules []*logs.ProcessingRule, endpoints *logs.Endpoints) *Agent {
	health := health.RegisterLiveness("logs-agent")

	cf := coreConfig.C.Logs
	// setup the auditor
	// We pass the health handle to the auditor because it's the end of the pipeline and the most
	// critical part. Arguably it could also be plugged to the destination.
	auditor := auditor.New(cf.RunPath, auditor.DefaultRegistryFilename, cf.AuditorTTL.Duration, health)
	destinationsCtx := client.NewDestinationsContext()
	diagnosticMessageReceiver := diagnostic.NewBufferedMessageReceiver()

	// setup the pipeline provider that provides pairs of processor and sender
	pipelineProvider := pipeline.NewProvider(config.NumberOfPipelines, auditor, diagnosticMessageReceiver, processingRules, endpoints, destinationsCtx)

	containerLaunchables := []container.Launchable{
		{
			IsAvailable: docker.IsAvailable,
			Launcher: func() restart.Restartable {
				return docker.NewLauncher(
					cf.DockerClientReadTimeout.Duration,
					sources,
					services,
					pipelineProvider,
					auditor,
					cf.DockerContainerUseFile,
					cf.DockerContainerForceUseFile,
				)
			},
		},
		{
			IsAvailable: kubernetes.IsAvailable,
			Launcher: func() restart.Restartable {
				return kubernetes.NewLauncher(sources, services, cf.ContainerCollectAll)
			},
		},
	}

	// when k8s_container_use_file is true, always attempt to use the kubernetes launcher first
	if cf.K8SContainerUseFile {
		containerLaunchables[0], containerLaunchables[1] = containerLaunchables[1], containerLaunchables[0]
	}

	validatePodContainerID := cf.ValidatePodContainerId

	// setup the inputs
	inputs := []restart.Restartable{
		file.NewScanner(sources, cf.OpenFilesLimit, pipelineProvider, auditor,
			file.DefaultSleepDuration, validatePodContainerID, cf.FileScanPeriod.Duration),
		listener.NewLauncher(sources, cf.FrameSize, pipelineProvider),
		journald.NewLauncher(sources, pipelineProvider, auditor),
		windowsevent.NewLauncher(sources, pipelineProvider),
		traps.NewLauncher(sources, pipelineProvider),
	}

	// Only try to start the container launchers if Docker or Kubernetes is available
	if coreConfig.IsFeaturePresent(coreConfig.Docker) || coreConfig.IsFeaturePresent(coreConfig.Kubernetes) {
		inputs = append(inputs, container.NewLauncher(containerLaunchables))
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
func NewServerless(sources *config.LogSources, services *service.Services, processingRules []*logs.ProcessingRule, endpoints *logs.Endpoints) *Agent {
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
	timeout := coreConfig.C.Logs.StopGracePeriod.Duration
	select {
	case <-c:
	case <-time.After(timeout):
		log.Info("Timed out when stopping logs-agent, forcing it to stop now")
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
			log.Warn("Force close of the Logs Agent, dumping the Go routines.")
			if stack, err := util.GetGoRoutinesDump(); err != nil {
				log.Warnf("can't get the Go routines dump: %s\n", err)
			} else {
				log.Warn(stack)
			}
		}
	}
}
