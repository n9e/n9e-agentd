// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package processor

import (
	"context"
	"sync"

	"k8s.io/klog/v2"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/diagnostic"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/message"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/metrics"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/types"
)

// A Processor updates messages from an inputChan and pushes
// in an outputChan.
type Processor struct {
	inputChan                 chan *message.Message
	outputChan                chan *message.Message
	processingRules           []*types.ProcessingRule
	encoder                   Encoder
	done                      chan struct{}
	diagnosticMessageReceiver diagnostic.MessageReceiver
	mu                        sync.Mutex
}

// New returns an initialized Processor.
func New(inputChan, outputChan chan *message.Message, processingRules []*types.ProcessingRule, encoder Encoder, diagnosticMessageReceiver diagnostic.MessageReceiver) *Processor {
	//klog.V(6).Infof("process inputChan %p outputChan %p", inputChan, outputChan)
	return &Processor{
		inputChan:                 inputChan,
		outputChan:                outputChan,
		processingRules:           processingRules,
		encoder:                   encoder,
		done:                      make(chan struct{}),
		diagnosticMessageReceiver: diagnosticMessageReceiver,
	}
}

// Start starts the Processor.
func (p *Processor) Start() {
	go p.run()
}

// Stop stops the Processor,
// this call blocks until inputChan is flushed
func (p *Processor) Stop() {
	close(p.inputChan)
	<-p.done
}

// Flush processes synchronously the messages that this processor has to process.
func (p *Processor) Flush(ctx context.Context) {
	klog.V(6).Infof("---- entering flush")
	p.mu.Lock()
	defer p.mu.Unlock()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if len(p.inputChan) == 0 {
				return
			}
			msg := <-p.inputChan
			//klog.V(11).Infof("processor.inputChan %p -> %s", p.inputChan, string(msg.Content))
			p.processMessage(msg)
		}
	}
}

// run starts the processing of the inputChan
func (p *Processor) run() {
	defer func() {
		p.done <- struct{}{}
	}()
	for msg := range p.inputChan {
		//klog.V(11).Infof("processor.inputChan %p -> %s", p.inputChan, string(msg.Content))
		p.processMessage(msg)
		p.mu.Lock() // block here if we're trying to flush synchronously
		p.mu.Unlock()
	}
}

func (p *Processor) processMessage(msg *message.Message) {
	klog.V(6).Infof("entering processMessage")
	metrics.LogsDecoded.Add(1)
	metrics.TlmLogsDecoded.Inc()
	if shouldProcess, redactedMsg := p.applyRedactingRules(msg); shouldProcess {
		metrics.LogsProcessed.Add(1)
		metrics.TlmLogsProcessed.Inc()

		p.diagnosticMessageReceiver.HandleMessage(*msg, redactedMsg)

		// Encode the message to its final format
		content, err := p.encoder.Encode(msg, redactedMsg)
		if err != nil {
			klog.Error("unable to encode msg ", err)
			return
		}
		msg.Content = content
		p.outputChan <- msg
		//klog.V(11).Infof("processor.outputChan %p <- %s", p.outputChan, string(msg.Content))
	}
}

// applyRedactingRules returns given a message if we should process it or not,
// and a copy of the message with some fields redacted, depending on config
func (p *Processor) applyRedactingRules(msg *message.Message) (bool, []byte) {
	content := msg.Content
	rules := append(p.processingRules, msg.Origin.LogSource.Config.ProcessingRules...)
	for _, rule := range rules {
		switch rule.Type {
		case types.ExcludeAtMatch:
			if rule.Regex.Match(content) {
				return false, nil
			}
		case types.IncludeAtMatch:
			if !rule.Regex.Match(content) {
				return false, nil
			}
		case types.MaskSequences:
			content = rule.Regex.ReplaceAll(content, rule.Placeholder)
		}
	}
	return true, content
}