// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package forwarder

import (
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/pkg/util/backoff"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/n9e/n9e-agentd/pkg/config"
)

type block struct {
	nbError int
	until   time.Time
}

type blockedEndpoints struct {
	errorPerEndpoint map[string]*block
	backoffPolicy    backoff.Policy
	m                sync.RWMutex
}

func newBlockedEndpoints() *blockedEndpoints {
	backoffFactor := config.C.Forwarder.BackoffFactor
	if backoffFactor < 2 {
		log.Warnf("Configured forwarder_backoff_factor (%v) is less than 2; 2 will be used", backoffFactor)
		backoffFactor = 2
	}

	backoffBase := config.C.Forwarder.BackoffBase
	if backoffBase <= 0 {
		log.Warnf("Configured forwarder_backoff_base (%v) is not positive; 2 will be used", backoffBase)
		backoffBase = 2
	}

	backoffMax := config.C.Forwarder.BackoffMax
	if backoffMax <= 0 {
		log.Warnf("Configured forwarder_backoff_max (%v) is not positive; 64 seconds will be used", backoffMax)
		backoffMax = 64
	}

	recInterval := config.C.Forwarder.RecoveryInterval
	if recInterval <= 0 {
		log.Warnf("Configured forwarder_recovery_interval (%v) is not positive; %v will be used", recInterval, config.DefaultForwarderRecoveryInterval)
		recInterval = config.DefaultForwarderRecoveryInterval
	}

	recoveryReset := config.C.Forwarder.RecoveryReset

	return &blockedEndpoints{
		errorPerEndpoint: make(map[string]*block),
		backoffPolicy:    backoff.NewPolicy(backoffFactor, backoffBase, backoffMax, recInterval, recoveryReset),
	}
}

func (e *blockedEndpoints) close(endpoint string) {
	e.m.Lock()
	defer e.m.Unlock()

	var b *block
	if knownBlock, ok := e.errorPerEndpoint[endpoint]; ok {
		b = knownBlock
	} else {
		b = &block{}
	}

	b.nbError = e.backoffPolicy.IncError(b.nbError)
	b.until = time.Now().Add(e.getBackoffDuration(b.nbError))

	e.errorPerEndpoint[endpoint] = b
}

func (e *blockedEndpoints) recover(endpoint string) {
	e.m.Lock()
	defer e.m.Unlock()

	var b *block
	if knownBlock, ok := e.errorPerEndpoint[endpoint]; ok {
		b = knownBlock
	} else {
		b = &block{}
	}

	b.nbError = e.backoffPolicy.DecError(b.nbError)
	b.until = time.Now().Add(e.getBackoffDuration(b.nbError))

	e.errorPerEndpoint[endpoint] = b
}

func (e *blockedEndpoints) isBlock(endpoint string) bool {
	e.m.RLock()
	defer e.m.RUnlock()

	if b, ok := e.errorPerEndpoint[endpoint]; ok && time.Now().Before(b.until) {
		return true
	}
	return false
}

func (e *blockedEndpoints) getBackoffDuration(numErrors int) time.Duration {
	return e.backoffPolicy.GetBackoffDuration(numErrors)
}
