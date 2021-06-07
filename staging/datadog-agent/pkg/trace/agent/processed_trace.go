// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package agent

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/pb"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/stats"
)

// ProcessedTrace represents a trace being processed in the agent.
type ProcessedTrace struct {
	Trace            pb.Trace
	WeightedTrace    stats.WeightedTrace
	Root             *pb.Span
	Env              string
	ClientDroppedP0s bool
}
