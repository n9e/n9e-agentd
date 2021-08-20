// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package tag

import (
	"context"
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/pkg/logs/config"
	"github.com/DataDog/datadog-agent/pkg/metadata/host"
	coreConfig "github.com/n9e/n9e-agentd/pkg/config"
)

type localProvider struct {
	tags                 []string
	expectedTags         []string
	expectedTagsDeadline time.Time
	sync.RWMutex
}

// NewLocalProvider returns a new local Provider.
func NewLocalProvider(t []string) Provider {
	p := &localProvider{
		tags:         t,
		expectedTags: t,
	}

	if config.IsExpectedTagsSet() {
		p.expectedTags = append(p.tags, host.GetHostTags(context.TODO(), false).System...)
		p.expectedTagsDeadline = coreConfig.StartTime.Add(coreConfig.C.Logs.ExpectedTagsDuration)

		// reset submitExpectedTags after deadline elapsed
		go func() {
			<-time.After(time.Until(p.expectedTagsDeadline))

			p.Lock()
			defer p.Unlock()
			p.expectedTags = p.tags
		}()
	}

	return p
}

// GetTags returns an empty list of tags.
func (p *localProvider) GetTags() []string {

	p.RLock()
	defer p.RUnlock()

	return p.expectedTags
}
