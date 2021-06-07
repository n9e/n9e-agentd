package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/n9e/n9e-agentd/pkg/api"
	"sigs.k8s.io/yaml"
)

var (
	ruleData = []byte(`
initConfig:

instances:

    ## @param prometheusUrl - string - required
    ## The URL where your application metrics are exposed by Prometheus.
    #
  - prometheusUrl: http://localhost:10055/metrics

    ## @param namespace - string - required
    ## The namespace to be appended before all metrics namespace
    #
    namespace: <SERVICE_NAME>

    ## @param metrics - list of key:value elements - required
    ## List of <METRIC_TO_FETCH>: <NEW_METRIC_NAME> for metrics to be fetched from the prometheus endpoint.
    ## <NEW_METRIC_NAME> is optional. It transforms the name in Datadog if set.
    ## This list should contain at least one metric
    #
    metrics:
      - <METRIC_TO_FETCH>: <NEW_METRIC_NAME>

    ## @param prometheusMetricsPrefix - string - optional
    ## Prefix for exposed Prometheus metrics.
    #
    # prometheusMetricsPrefix: <PREFIX>_

    ## @param healthServiceCheck - boolean - optional - default: true
    ## Send a service check reporting about the health of the prometheus endpoint
    ## It will be named <NAMESPACE>.prometheus.health
    #
    # healthServiceCheck: true

    ## @param labelToHostname - string - optional
    ## Override the hostname with the value of one label.
    #
    # labelToHostname: <LABEL>

    ## @param labelJoins - object - optional
    ## The label join allows to target a metric and retrieve it's label via a 1:1 mapping
    #
    # labelJoins:
    #   targetMetric:
    #     labelToMatch: <MATCHED_LABEL>
    #     labelsToGet:
    #       - <LABEL_TO_GET>

    ## @param labelsMapper - list of key:value element - optional
    ## The label mapper allows you to rename some labels
    ## Format is <LABEL_TO_RENAME>: <NEW_LABEL_NAME>
    #
    # labelsMapper:
    #   flavor: origin

    ## @param typeOverrides - list of key:value element - optional
    ## Type override allows you to override a type in the prometheus payload
    ## or type an untyped metrics (they're ignored by default)
    ## Supported <METRIC_TYPE> are gauge, counter, histogram, summary
    #
    # typeOverrides:
    #   <METRIC_NAME>: <METRIC_TYPE>

    ## @param tags - list of key:value element - optional
    ## List of tags to attach to every metric, event and service check emitted by this integration.
    ##
    ## Learn more about tagging: https://docs.datadoghq.com/tagging/
    #
    # tags:
    #   - <KEY_1>:<VALUE_1>
    #   - <KEY_2>:<VALUE_2>

    ## @param sendHistogramsBuckets - boolean - optional - default: true
    ## Set sendHistogramsBuckets to true to send the histograms bucket.
    #
    # sendHistogramsBuckets: true

    ## @param sendMonotonicCounter - boolean - optional - default: true
    ## To send counters as monotonic counter
    ##
    ## see: https://github.com/DataDog/integrations-core/issues/1303
    #
    # sendMonotonicCounter: true

    ## @param excludeLabels - list of string - optional
    ## List of label to be excluded.
    #
    # excludeLabels:
    #   - timestamp

    ## @param sslCert - string - optional
    ## If your prometheus endpoint is secured, here are the settings to configure it
    ## Can either be only the path to the certificate and thus you should specify the private key
    ## or it can be the path to a file containing both the certificate & the private key
    #
    # sslCert: "<CERT_PATH>"

    ## @param sslPrivateKey - string - optional
    ## Needed if the certificate does not include the private key
    ## WARNING: The private key to your local certificate must be unencrypted.
    #
    # sslPrivateKey: "<KEY_PATH>"

    ## @param sslCaCert - string - optional
    ## The path to the trusted CA used for generating custom certificates. Set this to false to disable SSL certificate
    ## verification.
    #
    # sslCaCert: "<CA_CERT_PATH>"

    ## @param prometheusTimeout - integer - optional - default: 10
    ## Set a timeout in second for the prometheus query.
    #
    # prometheusTimeout: 10s

    ## @param maxReturnedMetrics - integer - optional - default: 2000
    ## The check limits itself to 2000 metrics by default, increase this limit if needed.
    #
    # maxReturnedMetrics: 2000
`)

	rules     Rules
	rulesIdx  int
	rulesData []byte
)

func installCollectRules() {
	http.HandleFunc(api.RoutePathGetCollectRules, getCollectRules)
	http.HandleFunc(api.RoutePathGetCollectRulesSummary, getCollectRulesSummary)
	http.HandleFunc("/api/collect-rules/add", addCollectRule)
	http.HandleFunc("/api/collect-rules/del", delCollectRule)
}

type Rules struct {
	rules           []api.CollectRule
	LatestUpdatedAt int64
}

func init() {
	ruleData, _ = yaml.YAMLToJSON(ruleData)
	_addCollectRule()
}

func getCollectRules(w http.ResponseWriter, _ *http.Request) {
	writeRawJSON(rules.rules, w)
}

func getCollectRulesSummary(w http.ResponseWriter, _ *http.Request) {
	writeRawJSON(api.CollectRulesSummary{
		LatestUpdatedAt: rules.LatestUpdatedAt,
		Total:           len(rules.rules),
	}, w)
}

func addCollectRule(w http.ResponseWriter, r *http.Request) {
	_addCollectRule()
	getCollectRulesSummary(w, r)
}

func _addCollectRule() {
	rulesIdx++
	rule := api.CollectRule{
		Name:      "test",
		Data:      string(ruleData),
		Type:      fmt.Sprintf("prometheus"),
		Interval:  15,
		Tags:      "a=b,b=c",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		Creator:   "creator",
		Updater:   "updater",
	}
	rules.rules = append(rules.rules, rule)
	rules.LatestUpdatedAt = time.Now().Unix()
}

func delCollectRule(w http.ResponseWriter, r *http.Request) {
	if len(rules.rules) > 0 {
		rules.rules = rules.rules[0 : len(rules.rules)-1]
	}
	getCollectRulesSummary(w, r)
}

func writeRawJSON(object interface{}, w http.ResponseWriter) {
	output, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}
