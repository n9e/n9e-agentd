package testsender

import (
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	"github.com/DataDog/datadog-agent/pkg/metrics"
	"github.com/DataDog/datadog-agent/pkg/serializer"
)

//Rate adds a rate type to the mock calls.
func (m *TestSender) Rate(metric string, value float64, hostname string, tags []string) {
	m.Logf("rate(%s, %f, %s, %v)", metric, value, hostname, tags)
	m.Called(metric, value, hostname, tags)
}

//Count adds a count type to the mock calls.
func (m *TestSender) Count(metric string, value float64, hostname string, tags []string) {
	m.Logf("count(%s, %f, %s, %v)", metric, value, hostname, tags)
	m.Called(metric, value, hostname, tags)
}

//MonotonicCount adds a monotonic count type to the mock calls.
func (m *TestSender) MonotonicCount(metric string, value float64, hostname string, tags []string) {
	m.Logf("monotonicCount(%s, %f, %s, %v)", metric, value, hostname, tags)
	m.Called(metric, value, hostname, tags)
}

//MonotonicCountWithFlushFirstValue adds a monotonic count type to the mock calls with flushFirstValue parameter
func (m *TestSender) MonotonicCountWithFlushFirstValue(metric string, value float64, hostname string, tags []string, flushFirstValue bool) {
	m.Logf("MonotonicCountWithFlushFirstValue(%s, %f, %s, %v, %v)", metric, value, hostname, tags, flushFirstValue)
	m.Called(metric, value, hostname, tags, flushFirstValue)
}

//Counter adds a counter type to the mock calls.
func (m *TestSender) Counter(metric string, value float64, hostname string, tags []string) {
	m.Logf("counter(%s, %f, %s, %v)", metric, value, hostname, tags)
	m.Called(metric, value, hostname, tags)
}

//Histogram adds a histogram type to the mock calls.
func (m *TestSender) Histogram(metric string, value float64, hostname string, tags []string) {
	m.Logf("histogram(%s, %f, %s, %v)", metric, value, hostname, tags)
	m.Called(metric, value, hostname, tags)
}

//Historate adds a historate type to the mock calls.
func (m *TestSender) Historate(metric string, value float64, hostname string, tags []string) {
	m.Logf("Historate(%s, %f, %s, %v)", metric, value, hostname, tags)
	m.Called(metric, value, hostname, tags)
}

//Gauge adds a gauge type to the mock calls.
func (m *TestSender) Gauge(metric string, value float64, hostname string, tags []string) {
	m.Logf("gauge(%s, %f, %s, %v)", metric, value, hostname, tags)
	m.Called(metric, value, hostname, tags)
}

//ServiceCheck enables the service check mock call.
func (m *TestSender) ServiceCheck(checkName string, status metrics.ServiceCheckStatus, hostname string, tags []string, message string) {
	m.Logf("serviceCheck(%s, %s, %s, %v, %s)", checkName, status, hostname, tags, message)
	m.Called(checkName, status, hostname, tags, message)
}

//DisableDefaultHostname enables the hostname mock call.
func (m *TestSender) DisableDefaultHostname(d bool) {
	m.Logf("DisableDefaultHostname(%v)", d)
	m.Called(d)
}

//Event enables the event mock call.
func (m *TestSender) Event(e metrics.Event) {
	m.Logf("Event(%#v)", e)
	m.Called(e)
}

//EventPlatformEvent enables the event platform event mock call.
func (m *TestSender) EventPlatformEvent(rawEvent string, eventType string) {
	m.Called(rawEvent, eventType)
}

//HistogramBucket enables the histogram bucket mock call.
func (m *TestSender) HistogramBucket(metric string, value int64, lowerBound, upperBound float64, monotonic bool, hostname string, tags []string, flushFirstValue bool) {
	m.Logf("HistogramBucket(%s, %d, %f, %f, %v, %s, %v)", metric, value, lowerBound, upperBound, monotonic, hostname, tags)
	m.Called(metric, value, lowerBound, upperBound, monotonic, hostname, tags, flushFirstValue)
}

//Commit enables the commit mock call.
func (m *TestSender) Commit() {
	m.Logf("Commit()")
	m.Called()
}

//SetCheckCustomTags enables the set of check custom tags mock call.
func (m *TestSender) SetCheckCustomTags(tags []string) {
	m.Logf("SetCheckCustomTags(%v)", tags)
	m.Called(tags)
}

//SetCheckService enables the setting of check service mock call.
func (m *TestSender) SetCheckService(service string) {
	m.Logf("SetCheckService(%s)", service)
	m.Called(service)
}

//FinalizeCheckServiceTag enables the sending of check service tag mock call.
func (m *TestSender) FinalizeCheckServiceTag() {
	m.Logf("FinalizeCheckServiceTag()")
	m.Called()
}

//GetSenderStats enables the get metric stats mock call.
func (m *TestSender) GetSenderStats() check.SenderStats {
	m.Called()
	return check.NewSenderStats()
}

// OrchestratorMetadata submit orchestrator metadata messages
func (m *TestSender) OrchestratorMetadata(msgs []serializer.ProcessMessageBody, clusterID, payloadType string) {
	m.Logf("OrchestratorMetadata(%v, %s, %s)", msgs, clusterID, payloadType)
	m.Called(msgs, clusterID, payloadType)
}
