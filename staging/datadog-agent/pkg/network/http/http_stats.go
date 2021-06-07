package http

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/process/util"
	"k8s.io/klog/v2"
	"github.com/DataDog/sketches-go/ddsketch"
)

// Key is an identifier for a group of HTTP transactions
type Key struct {
	SrcIPHigh uint64
	SrcIPLow  uint64
	SrcPort   uint16

	DstIPHigh uint64
	DstIPLow  uint64
	DstPort   uint16

	Path string
}

// NewKey generates a new Key
func NewKey(saddr, daddr util.Address, sport, dport uint16, path string) Key {
	saddrl, saddrh := util.ToLowHigh(saddr)
	daddrl, daddrh := util.ToLowHigh(daddr)
	return Key{
		SrcIPHigh: saddrh,
		SrcIPLow:  saddrl,
		SrcPort:   sport,
		DstIPHigh: daddrh,
		DstIPLow:  daddrl,
		DstPort:   dport,
		Path:      path,
	}
}

// RelativeAccuracy defines the acceptable error in quantile values calculated by DDSketch.
// For example, if the actual value at p50 is 100, with a relative accuracy of 0.01 the value calculated
// will be between 99 and 101
const RelativeAccuracy = 0.01

// NumStatusClasses represents the number of HTTP status classes (1XX, 2XX, 3XX, 4XX, 5XX)
const NumStatusClasses = 5

// RequestStats stores stats for HTTP requests to a particular path, organized by the class
// of the response code (1XX, 2XX, 3XX, 4XX, 5XX)
type RequestStats [NumStatusClasses]struct {
	// Note: every time we add a latency value to the DDSketch below, it's possible for the sketch to discard that value
	// (ie if it is outside the range that is tracked by the sketch). For that reason, in order to keep an accurate count
	// the number of http transactions processed, we have our own count field (rather than relying on DDSketch.GetCount())
	Count     int
	Latencies *ddsketch.DDSketch

	// This field holds the value (in milliseconds) of the first HTTP request
	// in this bucket. We do this as optimization to avoid creating sketches with
	// a single value. This is quite common in the context of HTTP requests without
	// keep-alives where a short-lived TCP connection is used for a single request.
	FirstLatencySample float64
}

// CombineWith merges the data in 2 RequestStats objects
// newStats is kept as it is, while the method receiver gets mutated
func (r *RequestStats) CombineWith(newStats RequestStats) {
	for i := 0; i < len(r); i++ {
		statusClass := 100 * (i + 1)

		if newStats[i].Count == 0 {
			// Nothing to do in this case
			continue
		}

		if newStats[i].Count == 1 {
			// The other bucket has a single latency sample, so we "manually" add it
			r.AddRequest(statusClass, newStats[i].FirstLatencySample)
			continue
		}

		// The other bucket (newStats) has multiple samples and therefore a DDSketch object
		// We first ensure that the bucket we're merging to has a DDSketch object
		if r[i].Latencies == nil {
			// TODO: Consider calling Copy() on the other sketch instead
			if err := r.initSketch(i); err != nil {
				continue
			}

			// If we have a latency sample in this bucket we now add it to the DDSketch
			if r[i].Count == 1 {
				err := r[i].Latencies.Add(r[i].FirstLatencySample)
				if err != nil {
					klog.V(5).Infof("could not add request latency to ddsketch: %v", err)
				}
			}
		}

		// Finally merge both sketches
		r[i].Count += newStats[i].Count
		err := r[i].Latencies.MergeWith(newStats[i].Latencies)
		if err != nil {
			klog.V(5).Infof("error merging http transactions: %v", err)
		}
	}
}

// AddRequest takes information about a HTTP transaction and adds it to the request stats
func (r *RequestStats) AddRequest(statusClass int, latency float64) {
	i := statusClass/100 - 1
	if i >= len(r) {
		return
	}

	r[i].Count++
	if r[i].Count == 1 {
		// We postpone the creation of histograms when we have only one latency sample
		r[i].FirstLatencySample = latency
		return
	}

	if r[i].Latencies == nil {
		if err := r.initSketch(i); err != nil {
			return
		}

		// Add the defered latency sample
		err := r[i].Latencies.Add(r[i].FirstLatencySample)
		if err != nil {
			klog.V(5).Infof("could not add request latency to ddsketch: %v", err)
		}
	}

	err := r[i].Latencies.Add(latency)
	if err != nil {
		klog.V(5).Infof("could not add request latency to ddsketch: %v", err)
	}
}

func (r *RequestStats) initSketch(i int) (err error) {
	r[i].Latencies, err = ddsketch.NewDefaultDDSketch(RelativeAccuracy)
	if err != nil {
		klog.V(5).Infof("error recording http transaction latency: could not create new ddsketch: %v", err)
	}
	return
}
