package stats

import (
	"strconv"
	"strings"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/pb"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/traceutil"
	"k8s.io/klog/v2"
)

const (
	tagHostname   = "_dd.hostname"
	tagStatusCode = "http.status_code"
	tagVersion    = "version"
	tagOrigin     = "_dd.origin"
	tagSynthetics = "synthetics"
)

// Aggregation contains all the dimension on which we aggregate statistics
// when adding or removing fields to Aggregation the methods ToTagSet, KeyLen and
// WriteKey should always be updated accordingly
type Aggregation struct {
	Env        string
	Resource   string
	Service    string
	Type       string
	Hostname   string
	StatusCode uint32
	Version    string
	Synthetics bool
}

func getStatusCode(s *pb.Span) uint32 {
	strC := traceutil.GetMetaDefault(s, tagStatusCode, "")
	if strC == "" {
		return 0
	}
	c, err := strconv.Atoi(strC)
	if err != nil {
		klog.V(5).Infof("Invalid status code %s. Using 0.", strC)
		return 0
	}
	return uint32(c)
}

// NewAggregationFromSpan creates a new aggregation from the provided span and env
func NewAggregationFromSpan(s *pb.Span, env string, agentHostname string) Aggregation {
	synthetics := strings.HasPrefix(traceutil.GetMetaDefault(s, tagOrigin, ""), tagSynthetics)
	hostname := traceutil.GetMetaDefault(s, tagHostname, "")
	if hostname == "" {
		hostname = agentHostname
	}
	return Aggregation{
		Env:        env,
		Resource:   s.Resource,
		Service:    s.Service,
		Type:       s.Type,
		Hostname:   hostname,
		StatusCode: getStatusCode(s),
		Version:    traceutil.GetMetaDefault(s, tagVersion, ""),
		Synthetics: synthetics,
	}
}
