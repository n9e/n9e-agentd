// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package mock

import (
	"context"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/message"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/pipeline"
)

// mockProvider mocks pipeline providing logic
type mockProvider struct {
	msgChan chan *message.Message
}

// NewMockProvider returns a new mockProvider
func NewMockProvider() pipeline.Provider {
	return &mockProvider{
		msgChan: make(chan *message.Message),
	}
}

// Start does nothing
func (p *mockProvider) Start() {}

// Stop does nothing
func (p *mockProvider) Stop() {}

// Flush does nothing
func (p *mockProvider) Flush(ctx context.Context) {}

// NextPipelineChan returns the next pipeline
func (p *mockProvider) NextPipelineChan() chan *message.Message {
	return p.msgChan
}
