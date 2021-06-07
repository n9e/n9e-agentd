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
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/n9e/n9e-agentd/pkg/util/tls"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/check"
	core "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks"
	"github.com/yubo/golib/staging/util/yaml"
)

const checkName = "prometheus"

type PromInitConfig struct {
}

type PromInstanceConfig struct {
	PrometheusUrl           string            `yaml:"prometheusUrl"`           // prometheus_url
	Namespace               string            `yaml:"namespace"`               // namespace
	Metrics                 []string          `yaml:"metrics"`                 // metrics
	PrometheusMetricsPrefix string            `yaml:"prometheusMetricsPrefix"` // prometheus_metrics_prefix
	HealthServiceCheck      bool              `yaml:"healthServiceCheck"`      // health_service_check
	LabelToHostname         string            `yaml:"labelToHostname"`         // label_to_hostname
	LabelJoins              interface{}       `yaml:"labelJoins"`              // label_joins
	LabelsMapper            map[string]string `yaml:"labelsMapper"`            // labels_mapper
	TypeOverrides           map[string]string `yaml:"typeOverrides"`           // type_overrides
	Tags                    []string          `yaml:"tags"`                    // tags
	SendHistogramsBuckets   bool              `yaml:"sendHistogramsBuckets"`   // send_histograms_buckets
	SendMonotonicCounter    bool              `yaml:"sendMonotonicCounter"`    // send_monotonic_counter
	ExcludeLabels           []string          `yaml:"excludeLabels"`           // exclude_labels
	SslCert                 string            `yaml:"sslCert"`                 // ssl_cert
	SslPrivateKey           string            `yaml:"sslPrivateKey"`           // ssl_private_key
	SslCaCert               string            `yaml:"sslCaCert"`               // ssl_ca_cert
	PrometheusTimeout       time.Duration     `yaml:"prometheusTimeout"`       // prometheus_timeout
	MaxReturnedMetrics      int               `yaml:"maxReturnedMetrics"`      // max_returned_metrics
	tls.ClientConfig        `yaml:"-"`
	PromInitConfig          `yaml:"-"`
}

type promConfig struct {
	PromInstanceConfig
	PromInitConfig
}

func (p promConfig) String() string {
	return util.Prettify(p)
}

func defaultInstanceConfig() PromInstanceConfig {
	return PromInstanceConfig{
		SendHistogramsBuckets: true,
		SendMonotonicCounter:  true,
		PrometheusTimeout:     10 * time.Second,
		MaxReturnedMetrics:    2000,
	}
}

func buildConfig(rawInstance integration.Data, rawInitConfig integration.Data) (promConfig, error) {
	instance := defaultInstanceConfig()
	initConfig := PromInitConfig{}

	err := yaml.Unmarshal(rawInitConfig, &initConfig)
	if err != nil {
		return promConfig{}, err
	}

	err = yaml.Unmarshal(rawInstance, &instance)
	if err != nil {
		return promConfig{}, err
	}

	return promConfig{
		PromInitConfig:     initConfig,
		PromInstanceConfig: instance,
	}, nil

}

// Check doesn't need additional fields
type Check struct {
	core.CheckBase
	config promConfig
	client *http.Client
}

// Run executes the check
func (c *Check) Run() error {
	sender, err := aggregator.GetSender(c.ID())
	if err != nil {
		return err
	}

	cf := c.config

	req, err := http.NewRequest("GET", cf.PrometheusUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "*/*")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("error making HTTP request to %s: %s", cf.PrometheusUrl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", cf.PrometheusUrl, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %s", err)
	}

	mfs, err := readBody(body, resp.Header)
	if err != nil {
		return fmt.Errorf("error reading metrics for %s: %s", cf.PrometheusUrl, err)
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
	tlsCfg, err := c.config.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   tlsCfg,
			DisableKeepAlives: true,
		},
		Timeout: c.config.PrometheusTimeout,
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

	c.config = config

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
