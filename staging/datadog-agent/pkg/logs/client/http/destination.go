package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	httputils "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/http"
	"k8s.io/klog/v2"

	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/client"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/metrics"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/types"
)

// ContentType options,
const (
	TextContentType = "text/plain"
	JSONContentType = "application/json"
)

// HTTP errors.
var (
	errClient = errors.New("client error")
	errServer = errors.New("server error")
)

// emptyPayload is an empty payload used to check HTTP connectivity without sending logs.
var emptyPayload []byte

// Destination sends a payload over HTTP.
type Destination struct {
	url                 string
	contentType         string
	contentEncoding     ContentEncoding
	client              *httputils.ResetClient
	destinationsContext *client.DestinationsContext
	once                sync.Once
	payloadChan         chan []byte
	climit              chan struct{} // semaphore for limiting concurrent background sends
}

// NewDestination returns a new Destination.
// If `maxConcurrentBackgroundSends` > 0, then at most that many background payloads will be sent concurrently, else
// there is no concurrency and the background sending pipeline will block while sending each payload.
// TODO: add support for SOCKS5
func NewDestination(endpoint types.Endpoint, contentType string, destinationsContext *client.DestinationsContext, maxConcurrentBackgroundSends int) *Destination {
	return newDestination(endpoint, contentType, destinationsContext, time.Second*10, maxConcurrentBackgroundSends)
}

func newDestination(endpoint types.Endpoint, contentType string, destinationsContext *client.DestinationsContext, timeout time.Duration, maxConcurrentBackgroundSends int) *Destination {
	if maxConcurrentBackgroundSends < 0 {
		maxConcurrentBackgroundSends = 0
	}
	return &Destination{
		url:                 buildURL(endpoint),
		contentType:         contentType,
		contentEncoding:     buildContentEncoding(endpoint),
		client:              httputils.NewResetClient(endpoint.ConnectionResetInterval, httpClientFactory(timeout)),
		destinationsContext: destinationsContext,
		climit:              make(chan struct{}, maxConcurrentBackgroundSends),
	}
}

// Send sends a payload over HTTP,
// the error returned can be retryable and it is the responsibility of the callee to retry.
func (d *Destination) Send(payload []byte) error {
	klog.V(10).Infof("-- send entering")
	defer klog.V(10).Infof("-- send leaving")
	ctx := d.destinationsContext.Context()

	encodedPayload, err := d.contentEncoding.encode(payload)
	if err != nil {
		return err
	}
	metrics.BytesSent.Add(int64(len(payload)))
	metrics.EncodedBytesSent.Add(int64(len(encodedPayload)))

	klog.V(5).Infof("POST %s", d.url)
	req, err := http.NewRequest("POST", d.url, bytes.NewReader(encodedPayload))
	if err != nil {
		klog.V(5).Infof("POST %s err %s", d.url, err)
		// the request could not be built,
		// this can happen when the method or the url are valid.
		return err
	}
	req.Header.Set("Content-Type", d.contentType)
	req.Header.Set("Content-Encoding", d.contentEncoding.name())
	req = req.WithContext(ctx)

	resp, err := d.client.Do(req)
	if err != nil {
		if ctx.Err() == context.Canceled {
			return ctx.Err()
		}
		// most likely a network or a connect error, the callee should retry.
		return client.NewRetryableError(err)
	}

	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// the read failed because the server closed or terminated the connection
		// *after* serving the request.
		return err
	}

	if resp.StatusCode >= 500 {
		// the server could not serve the request,
		// most likely because of an internal error
		return client.NewRetryableError(errServer)
	} else if resp.StatusCode >= 400 {
		// the logs-agent is likely to be misconfigured,
		// the URL or the API key may be wrong.
		return errClient
	} else {
		return nil
	}
}

// SendAsync sends a payload in background.
func (d *Destination) SendAsync(payload []byte) {
	klog.V(10).Infof("-- sendAsync entering")
	d.once.Do(func() {
		payloadChan := make(chan []byte, config.ChanSize)
		d.sendInBackground(payloadChan)
		d.payloadChan = payloadChan
	})
	d.payloadChan <- payload
}

// sendInBackground sends all payloads from payloadChan in background.
func (d *Destination) sendInBackground(payloadChan chan []byte) {
	ctx := d.destinationsContext.Context()
	go func() {
		for {
			select {
			case payload := <-payloadChan:
				// if the channel is non-buffered then there is no concurrency and we block on sending each payload
				if cap(d.climit) == 0 {
					d.Send(payload) //nolint:errcheck
					break
				}
				d.climit <- struct{}{}
				go func() {
					d.Send(payload) //nolint:errcheck
					<-d.climit
				}()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func httpClientFactory(timeout time.Duration) func() *http.Client {
	return func() *http.Client {
		return &http.Client{
			Timeout: timeout,
			// reusing core agent HTTP transport to benefit from proxy settings.
			Transport: httputils.CreateHTTPTransport(),
		}
	}
}

// buildURL buils a url from a config endpoint.
func buildURL(endpoint types.Endpoint) string {
	var scheme string
	if endpoint.UseSSL {
		scheme = "https"
	} else {
		scheme = "http"
	}
	var address string
	if endpoint.Port != 0 {
		address = fmt.Sprintf("%v:%v", endpoint.Host, endpoint.Port)
	} else {
		address = endpoint.Host
	}
	return fmt.Sprintf("%v://%v%s", scheme, address, api.RoutePathLogs)
}

func buildContentEncoding(endpoint types.Endpoint) ContentEncoding {
	if endpoint.UseCompression {
		return NewGzipContentEncoding(endpoint.CompressionLevel)
	}
	return IdentityContentType
}

// CheckConnectivity check if sending logs through HTTP works
func CheckConnectivity(endpoint types.Endpoint) config.HTTPConnectivity {
	klog.Info("Checking HTTP connectivity...")
	ctx := client.NewDestinationsContext()
	ctx.Start()
	defer ctx.Stop()
	// Lower the timeout to 5s because HTTP connectivity test is done synchronously during the agent bootstrap sequence
	destination := newDestination(endpoint, JSONContentType, ctx, time.Second*5, 0)
	klog.Infof("Sending HTTP connectivity request to %s...", destination.url)
	err := destination.Send(emptyPayload)
	if err != nil {
		klog.Warningf("HTTP connectivity failure: %v", err)
	} else {
		klog.Info("HTTP connectivity successful")
	}
	return err == nil
}
