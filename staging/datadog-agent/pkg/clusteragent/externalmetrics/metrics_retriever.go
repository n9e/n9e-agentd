// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build kubeapiserver

package externalmetrics

import (
	"fmt"
	"time"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/clusteragent/externalmetrics/model"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/kubernetes/autoscalers"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2"
)

const (
	invalidMetricBackendErrorMessage  string = "Invalid metric (from backend), query: %s"
	invalidMetricOutdatedErrorMessage string = "Outdated result from backend, query: %s"
	invalidMetricNoDataErrorMessage   string = "No data from backend, query: %s"
	invalidMetricGlobalErrorMessage   string = "Global error (all queries) from backend"
	metricRetrieverStoreID            string = "mr"
)

type MetricsRetriever struct {
	refreshPeriod int64
	metricsMaxAge int64
	processor     autoscalers.ProcessorInterface
	store         *DatadogMetricsInternalStore
	isLeader      func() bool
}

func NewMetricsRetriever(refreshPeriod, metricsMaxAge int64, processor autoscalers.ProcessorInterface, isLeader func() bool, store *DatadogMetricsInternalStore) (*MetricsRetriever, error) {
	return &MetricsRetriever{
		refreshPeriod: refreshPeriod,
		metricsMaxAge: metricsMaxAge,
		processor:     processor,
		store:         store,
		isLeader:      isLeader,
	}, nil
}

func (mr *MetricsRetriever) Run(stopCh <-chan struct{}) {
	klog.Infof("Starting MetricsRetriever")
	tickerRefreshProcess := time.NewTicker(time.Duration(mr.refreshPeriod) * time.Second)
	for {
		select {
		case <-tickerRefreshProcess.C:
			if mr.isLeader() {
				mr.retrieveMetricsValues()
			}
		case <-stopCh:
			klog.Infof("Stopping MetricsRetriever")
			return
		}
	}
}

func (mr *MetricsRetriever) retrieveMetricsValues() {
	// We only update active DatadogMetrics
	datadogMetrics := mr.store.GetFiltered(func(datadogMetric model.DatadogMetricInternal) bool { return datadogMetric.Active })
	if len(datadogMetrics) == 0 {
		klog.V(5).Infof("No active DatadogMetric, nothing to refresh")
		return
	}

	queries := getUniqueQueries(datadogMetrics)
	klog.V(5).Infof("Starting refreshing external metrics with: %d queries", len(queries))

	results, err := mr.processor.QueryExternalMetric(queries)
	globalError := false
	// Check for global failure
	if len(results) == 0 && err != nil {
		globalError = true
		klog.Errorf("Unable to fetch external metrics: %v", err)
	}

	// Update store with current results
	currentTime := time.Now().UTC()
	for _, datadogMetric := range datadogMetrics {
		datadogMetricFromStore := mr.store.LockRead(datadogMetric.ID, false)
		if datadogMetricFromStore == nil {
			// This metric is not in the store anymore, discard it
			klog.Infof("Discarding results for DatadogMetric: %s as not present in store anymore", datadogMetric.ID)
			continue
		}

		query := datadogMetric.Query()
		if queryResult, found := results[query]; found {
			klog.V(5).Infof("QueryResult from DD for %q: %v", query, queryResult)

			if queryResult.Valid {
				datadogMetricFromStore.Value = queryResult.Value

				// If we get a valid but old metric, flag it as invalid
				if currentTime.Unix()-queryResult.Timestamp <= mr.metricsMaxAge {
					datadogMetricFromStore.Valid = true
					datadogMetricFromStore.Error = nil
					datadogMetricFromStore.UpdateTime = time.Unix(queryResult.Timestamp, 0).UTC()
				} else {
					datadogMetricFromStore.Valid = false
					datadogMetricFromStore.Error = fmt.Errorf(invalidMetricOutdatedErrorMessage, query)
					datadogMetricFromStore.UpdateTime = currentTime
				}
			} else {
				datadogMetricFromStore.Valid = false
				datadogMetricFromStore.Error = fmt.Errorf(invalidMetricBackendErrorMessage, query)
				datadogMetricFromStore.UpdateTime = currentTime
			}
		} else {
			datadogMetricFromStore.Valid = false
			if globalError {
				datadogMetricFromStore.Error = fmt.Errorf(invalidMetricGlobalErrorMessage)
			} else {
				datadogMetricFromStore.Error = fmt.Errorf(invalidMetricNoDataErrorMessage, query)
			}
			datadogMetricFromStore.UpdateTime = currentTime
		}

		mr.store.UnlockSet(datadogMetric.ID, *datadogMetricFromStore, metricRetrieverStoreID)
	}
}

func getUniqueQueries(datadogMetrics []model.DatadogMetricInternal) []string {
	queries := make([]string, 0, len(datadogMetrics))
	unique := make(map[string]struct{}, len(queries))
	for _, datadogMetric := range datadogMetrics {
		query := datadogMetric.Query()
		if _, found := unique[query]; !found {
			unique[query] = struct{}{}
			queries = append(queries, query)
		}
	}

	return queries
}