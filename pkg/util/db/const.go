package db

import "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metrics"

var (
	SECOND      = 1
	MILLISECOND = 1000
	MICROSECOND = 1000000
	NANOSECOND  = 1000000000

	TIME_UNITS = map[string]int{
		"microsecond": MICROSECOND,
		"millisecond": MILLISECOND,
		"nanosecond":  NANOSECOND,
		"second":      SECOND,
	}

	serviceCheckStatus = map[string]metrics.ServiceCheckStatus{
		"OK":       metrics.ServiceCheckOK,
		"WARNING":  metrics.ServiceCheckWarning,
		"CRITICAL": metrics.ServiceCheckCritical,
		"UNKNOWN":  metrics.ServiceCheckUnknown,
	}
)
