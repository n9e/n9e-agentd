// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package agent

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/pb"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/traceutil"
	"k8s.io/klog/v2"
)

// Truncate checks that the span resource, meta and metrics are within the max length
// and modifies them if they are not
func Truncate(s *pb.Span) {
	r, ok := traceutil.TruncateResource(s.Resource)
	if !ok {
		klog.V(5).Infof("span.truncate: truncated `Resource` (max %d chars): %s", traceutil.MaxResourceLen, s.Resource)
	}
	s.Resource = r

	// Error - Nothing to do
	// Optional data, Meta & Metrics can be nil
	// Soft fail on those
	for k, v := range s.Meta {
		modified := false

		if len(k) > traceutil.MaxMetaKeyLen {
			klog.V(5).Infof("span.truncate: truncating `Meta` key (max %d chars): %s", traceutil.MaxMetaKeyLen, k)
			delete(s.Meta, k)
			k = traceutil.TruncateUTF8(k, traceutil.MaxMetaKeyLen) + "..."
			modified = true
		}

		if len(v) > traceutil.MaxMetaValLen {
			v = traceutil.TruncateUTF8(v, traceutil.MaxMetaValLen) + "..."
			modified = true
		}

		if modified {
			s.Meta[k] = v
		}
	}
	for k, v := range s.Metrics {
		if len(k) > traceutil.MaxMetricsKeyLen {
			klog.V(5).Infof("span.truncate: truncating `Metrics` key (max %d chars): %s", traceutil.MaxMetricsKeyLen, k)
			delete(s.Metrics, k)
			k = traceutil.TruncateUTF8(k, traceutil.MaxMetricsKeyLen) + "..."

			s.Metrics[k] = v
		}
	}
}
