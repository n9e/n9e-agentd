package forwarder

import (
	"net/http"
)

type N9eForwarder struct {
	*DefaultForwarder
}

func NewN9eForwarder(options *Options) *N9eForwarder {
	return &N9eForwarder{
		DefaultForwarder: NewDefaultForwarder(options),
	}
}

// SubmitV1Series will send timeserie to v1 endpoint (this will be remove once
// the backend handles v2 endpoints).
func (f *N9eForwarder) SubmitV1Series(payload Payloads, extra http.Header) error {
	transactions := f.createHTTPTransactions(n9eV1SeriesEndpoint, payload, true, extra)
	return f.sendHTTPTransactions(transactions)
}

// SubmitSeries will send a series type payload to Datadog backend.
func (f *N9eForwarder) SubmitSeries(payload Payloads, extra http.Header) error {
	transactions := f.createHTTPTransactions(n9eSeriesEndpoint, payload, false, extra)
	return f.sendHTTPTransactions(transactions)
}
