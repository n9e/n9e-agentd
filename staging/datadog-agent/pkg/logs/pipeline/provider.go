// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package pipeline

import (
	"context"
	"sync/atomic"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/diagnostic"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/auditor"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/client"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/message"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/restart"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/types"
)

// Provider provides message channels
type Provider interface {
	Start()
	Stop()
	NextPipelineChan() chan *message.Message
	// Flush flushes all pipeline contained in this Provider
	Flush(ctx context.Context)
}

// provider implements providing logic
type provider struct {
	numberOfPipelines         int
	auditor                   auditor.Auditor
	diagnosticMessageReceiver diagnostic.MessageReceiver
	outputChan                chan *message.Message
	processingRules           []*types.ProcessingRule
	endpoints                 *types.Endpoints

	pipelines            []*Pipeline
	currentPipelineIndex int32
	destinationsContext  *client.DestinationsContext

	serverless bool
}

// NewProvider returns a new Provider
func NewProvider(numberOfPipelines int, auditor auditor.Auditor, diagnosticMessageReceiver diagnostic.MessageReceiver, processingRules []*types.ProcessingRule, endpoints *types.Endpoints, destinationsContext *client.DestinationsContext) Provider {
	return newProvider(numberOfPipelines, auditor, diagnosticMessageReceiver, processingRules, endpoints, destinationsContext, false)
}

// NewServerlessProvider returns a new Provider in serverless mode
func NewServerlessProvider(numberOfPipelines int, auditor auditor.Auditor, processingRules []*types.ProcessingRule, endpoints *types.Endpoints, destinationsContext *client.DestinationsContext) Provider {
	return newProvider(numberOfPipelines, auditor, &diagnostic.NoopMessageReceiver{}, processingRules, endpoints, destinationsContext, true)
}

func newProvider(numberOfPipelines int, auditor auditor.Auditor, diagnosticMessageReceiver diagnostic.MessageReceiver, processingRules []*types.ProcessingRule, endpoints *types.Endpoints, destinationsContext *client.DestinationsContext, serverless bool) Provider {
	return &provider{
		numberOfPipelines:         numberOfPipelines,
		auditor:                   auditor,
		diagnosticMessageReceiver: diagnosticMessageReceiver,
		processingRules:           processingRules,
		endpoints:                 endpoints,
		pipelines:                 []*Pipeline{},
		destinationsContext:       destinationsContext,
		serverless:                serverless,
	}
}

// Start initializes the pipelines
func (p *provider) Start() {
	// This requires the auditor to be started before.
	p.outputChan = p.auditor.Channel()

	for i := 0; i < p.numberOfPipelines; i++ {
		pipeline := NewPipeline(p.outputChan, p.processingRules, p.endpoints,
			p.destinationsContext, p.diagnosticMessageReceiver, p.serverless)
		pipeline.Start()
		p.pipelines = append(p.pipelines, pipeline)
	}
}

// Stop stops all pipelines in parallel,
// this call blocks until all pipelines are stopped
func (p *provider) Stop() {
	stopper := restart.NewParallelStopper()
	for _, pipeline := range p.pipelines {
		stopper.Add(pipeline)
	}
	stopper.Stop()
	p.pipelines = p.pipelines[:0]
	p.outputChan = nil
}

// NextPipelineChan returns the next pipeline input channel
func (p *provider) NextPipelineChan() chan *message.Message {
	pipelinesLen := len(p.pipelines)
	if pipelinesLen == 0 {
		return nil
	}
	index := int(p.currentPipelineIndex+1) % pipelinesLen
	defer atomic.StoreInt32(&p.currentPipelineIndex, int32(index))
	nextPipeline := p.pipelines[index]
	return nextPipeline.InputChan
}

// Flush flushes synchronously all the contained pipeline of this provider.
func (p *provider) Flush(ctx context.Context) {
	for _, p := range p.pipelines {
		select {
		case <-ctx.Done():
			return
		default:
			p.Flush(ctx)
		}
	}
}
