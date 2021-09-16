package mocker

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"k8s.io/klog/v2"
)

// https://github.com/DataDog/datadog-go
// https://pkg.go.dev/github.com/DataDog/datadog-go/statsd#Option
// https://docs.datadoghq.com/developers/dogstatsd/?tab=hostagent#client-instantiation-parameters
// https://docs.datadoghq.com/developers/metrics/dogstatsd_metrics_submission/?tab=go#count

func (p *mocker) installStatsdSender() error {
	if !p.config.SendStatsd {
		return nil
	}

	client, err := statsd.New("127.0.0.1:8125")
	if err != nil {
		return err
	}

	go func() {
		var i float64
		t1 := time.NewTicker(10 * time.Second)
		t2 := time.NewTicker(2 * time.Second)

		defer t1.Stop()
		defer t2.Stop()
		defer client.Close()

		for {
			select {
			case <-t1.C:
				// count
				if err := client.Incr("example_metric.increment", []string{"environment:dev"}, 1); err != nil {
					klog.Error(err)
				}

				client.Decr("example_metric.decrement", []string{"environment:dev"}, 1)
				client.Count("example_metric.count", 2, []string{"environment:dev"}, 1)

				// guage
				i += 1
				client.Gauge("example_metric.gauge", i, []string{"environment:dev"}, 1)

				// set
				client.Set("example_metric.set", fmt.Sprintf("%f", i), []string{"environment:dev"}, 1)

				// evnet
				client.SimpleEvent("An error occurred", "Error message")

				// service check
				client.SimpleServiceCheck("application.service_check", 0)

			case <-t2.C:
				// histogram
				if err := client.Histogram("example_metric.histogram", float64(rand.Intn(20)), []string{"environment:dev"}, 1); err != nil {
					klog.Error(err)
				}

				// distribution
				client.Distribution("example_metric.distribution", float64(rand.Intn(20)), []string{"environment:dev"}, 1)

			}
		}
	}()
	return nil
}
