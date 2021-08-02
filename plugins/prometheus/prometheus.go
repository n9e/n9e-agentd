package prometheus

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"time"

	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/pkg/util/tls"
	"github.com/DataDog/datadog-agent/pkg/aggregator"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	core "github.com/DataDog/datadog-agent/pkg/collector/corechecks"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"sigs.k8s.io/yaml"
)

const checkName = "prometheus"

type InitConfig struct {
}

type InstanceConfig struct {
	PrometheusUrl           string            `json:"prometheus_url"`            // prometheus_url
	Namespace               string            `json:"namespace"`                 // namespace
	Metrics                 []string          `json:"metrics"`                   // metrics
	PrometheusMetricsPrefix string            `json:"prometheus_metrics_prefix"` // prometheus_metrics_prefix
	HealthServiceCheck      bool              `json:"health_service_check"`      // health_service_check
	LabelToHostname         string            `json:"label_to_hostname"`         // label_to_hostname
	LabelJoins              interface{}       `json:"label_joins"`               // label_joins
	LabelsMapper            map[string]string `json:"labels_mapper"`             // labels_mapper
	TypeOverrides           map[string]string `json:"type_overrides"`            // type_overrides
	Tags                    []string          `json:"tags"`                      // tags
	SendHistogramsBuckets   bool              `json:"send_histograms_buckets"`   // send_histograms_buckets
	SendMonotonicCounter    bool              `json:"send_monotonic_counter"`    // send_monotonic_counter
	ExcludeLabels           []string          `json:"exclude_labels"`            // exclude_labels
	SslCert                 string            `json:"ssl_cert"`                  // ssl_cert
	SslPrivateKey           string            `json:"ssl_private_key"`           // ssl_private_key
	SslCaCert               string            `json:"ssl_ca_cert"`               // ssl_ca_cert
	PrometheusTimeout       int               `json:"prometheus_timeout"`        // prometheus_timeout
	MaxReturnedMetrics      int               `json:"max_returned_metrics"`      // max_returned_metrics

	timeout          time.Duration
	tls.ClientConfig `json:"-"`
	InitConfig       `json:"-"`
}

type promConfig struct {
	InstanceConfig
	InitConfig
}

func (p *promConfig) Validate() error {
	p.timeout = time.Second * time.Duration(p.PrometheusTimeout)
	return nil
}

func defaultInstanceConfig() InstanceConfig {
	return InstanceConfig{
		SendHistogramsBuckets: true,
		SendMonotonicCounter:  true,
		PrometheusTimeout:     10,
		MaxReturnedMetrics:    2000,
	}
}

func buildConfig(rawInstance integration.Data, rawInitConfig integration.Data) (*promConfig, error) {
	instance := defaultInstanceConfig()
	initConfig := InitConfig{}

	err := yaml.Unmarshal(rawInitConfig, &initConfig)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(rawInstance, &instance)
	if err != nil {
		return nil, err
	}

	c := &promConfig{
		InitConfig:     initConfig,
		InstanceConfig: instance,
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// Check doesn't need additional fields
type Check struct {
	core.CheckBase
	*promConfig
	client *http.Client
}

// Run executes the check
func (c *Check) Run() error {
	sender, err := aggregator.GetSender(c.ID())
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", c.PrometheusUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "*/*")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", c.PrometheusUrl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", c.PrometheusUrl, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %s", err)
	}

	mfs, err := readBody(body, resp.Header)
	if err != nil {
		return fmt.Errorf("error reading metrics for %s: %s", c.PrometheusUrl, err)
	}

	collectMetrics(sender, mfs)

	sender.Commit()
	return nil
}

func collectMetrics(sender aggregator.Sender, mfs map[string]*dto.MetricFamily) error {
	for metricName, mf := range mfs {
		for _, m := range mf.Metric {
			switch mf.GetType() {
			case dto.MetricType_SUMMARY:
				sendSummary(metricName, m, sender)
			case dto.MetricType_HISTOGRAM:
				sendHistogram(metricName, m, sender)
			case dto.MetricType_COUNTER:
				sendCounter(metricName, m, sender)
			case dto.MetricType_GAUGE:
				sendGauge(metricName, m, sender)
			case dto.MetricType_UNTYPED:
				sendUntyped(metricName, m, sender)
			}
		}
	}
	return nil
}

func readBody(buf []byte, header http.Header) (map[string]*dto.MetricFamily, error) {
	var parser expfmt.TextParser
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	buffer := bytes.NewBuffer(buf)
	reader := bufio.NewReader(buffer)

	mediatype, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	// Prepare output
	metricFamilies := make(map[string]*dto.MetricFamily)

	if err == nil && mediatype == "application/vnd.google.protobuf" &&
		params["encoding"] == "delimited" &&
		params["proto"] == "io.prometheus.client.MetricFamily" {
		for {
			mf := &dto.MetricFamily{}
			if _, ierr := pbutil.ReadDelimited(reader, mf); ierr != nil {
				if ierr == io.EOF {
					break
				}
				return nil, fmt.Errorf("reading metric family protocol buffer failed: %s", ierr)
			}
			metricFamilies[mf.GetName()] = mf
		}
		return metricFamilies, nil
	}

	return parser.TextToMetricFamilies(reader)
}

func (c *Check) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := c.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   tlsCfg,
			DisableKeepAlives: true,
		},
		Timeout: c.timeout,
	}

	return client, nil
}

// Configure the Prom check
func (c *Check) Configure(rawInstance integration.Data, rawInitConfig integration.Data, source string) error {
	// Must be called before c.CommonConfigure
	c.BuildID(rawInstance, rawInitConfig)

	err := c.CommonConfigure(rawInstance, source)
	if err != nil {
		return fmt.Errorf("common configure failed: %s", err)
	}

	config, err := buildConfig(rawInstance, rawInitConfig)
	if err != nil {
		return fmt.Errorf("build config failed: %s", err)
	}

	c.promConfig = config

	if c.client, err = c.createHTTPClient(); err != nil {
		return nil
	}

	return nil
}

func promFactory() check.Check {
	return &Check{
		CheckBase: core.NewCheckBase(checkName),
	}
}

func init() {
	core.RegisterCheck(checkName, promFactory)
}
