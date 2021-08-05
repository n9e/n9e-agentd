// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build kubeapiserver

package kubernetesapiserver

import (
	"errors"
	"fmt"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/check"
	core "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/cluster"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metrics"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/kubernetes/clustername"
	"k8s.io/klog/v2"
)

// Covers the Control Plane service check and the in memory pod metadata.
const (
	KubeControlPaneCheck          = "kube_apiserver_controlplane.up"
	kubernetesAPIServerCheckName  = "kubernetes_apiserver"
	eventTokenKey                 = "event"
	maxEventCardinality           = 300
	defaultResyncPeriodInSecond   = 300
	defaultTimeoutEventCollection = 2000

	defaultCacheExpire = 2 * time.Minute
	defaultCachePurge  = 10 * time.Minute
)

// KubeASConfig is the config of the API server.
type KubeASConfig struct {
	CollectEvent             bool     `yaml:"collect_events"`
	CollectOShiftQuotas      bool     `yaml:"collect_openshift_clusterquotas"`
	FilteredEventTypes       []string `yaml:"filtered_event_types"`
	EventCollectionTimeoutMs int      `yaml:"kubernetes_event_read_timeout_ms"`
	MaxEventCollection       int      `yaml:"max_events_per_run"`
	LeaderSkip               bool     `yaml:"skip_leader_election"`
	ResyncPeriodEvents       int      `yaml:"kubernetes_event_resync_period_s"`
}

// EventC holds the information pertaining to which event we collected last and when we last re-synced.
type EventC struct {
	LastResVer string
	LastTime   time.Time
}

// KubeASCheck grabs metrics and events from the API server.
type KubeASCheck struct {
	core.CheckBase
	instance        *KubeASConfig
	eventCollection EventC
	ignoredEvents   string
	ac              *apiserver.APIClient
	oshiftAPILevel  apiserver.OpenShiftAPILevel
	providerIDCache *cache.Cache
}

func (c *KubeASConfig) parse(data []byte) error {
	// default values
	c.CollectEvent = config.Datadog.GetBool("collect_kubernetes_events")
	c.CollectOShiftQuotas = true
	c.ResyncPeriodEvents = defaultResyncPeriodInSecond

	return yaml.Unmarshal(data, c)
}

func NewKubeASCheck(base core.CheckBase, instance *KubeASConfig) *KubeASCheck {
	return &KubeASCheck{
		CheckBase:       base,
		instance:        instance,
		providerIDCache: cache.New(defaultCacheExpire, defaultCachePurge),
	}
}

// KubernetesASFactory is exported for integration testing.
func KubernetesASFactory() check.Check {
	return NewKubeASCheck(core.NewCheckBase(kubernetesAPIServerCheckName), &KubeASConfig{})
}

// Configure parses the check configuration and init the check.
func (k *KubeASCheck) Configure(config, initConfig integration.Data, source string) error {
	err := k.CommonConfigure(config, source)
	if err != nil {
		return err
	}

	// Check connectivity to the APIServer
	err = k.instance.parse(config)
	if err != nil {
		klog.Error("could not parse the config for the API server")
		return err
	}
	if k.instance.EventCollectionTimeoutMs == 0 {
		k.instance.EventCollectionTimeoutMs = defaultTimeoutEventCollection
	}

	if k.instance.MaxEventCollection == 0 {
		k.instance.MaxEventCollection = maxEventCardinality
	}
	k.ignoredEvents = convertFilter(k.instance.FilteredEventTypes)

	return nil
}

func convertFilter(conf []string) string {
	var formatedFilters []string
	for _, filter := range conf {
		f := strings.Split(filter, "=")
		if len(f) == 1 {
			formatedFilters = append(formatedFilters, fmt.Sprintf("reason!=%s", f[0]))
			continue
		}
		formatedFilters = append(formatedFilters, filter)
	}
	return strings.Join(formatedFilters, ",")
}

// Run executes the check.
func (k *KubeASCheck) Run() error {
	sender, err := aggregator.GetSender(k.ID())
	if err != nil {
		return err
	}
	defer sender.Commit()

	if config.Datadog.GetBool("cluster_agent.enabled") {
		klog.V(5).Info("Cluster agent is enabled. Not running Kubernetes API Server check or collecting Kubernetes Events.")
		return nil
	}
	// If the check is configured as a cluster check, the cluster check worker needs to skip the leader election section.
	// The Cluster Agent will passed in the `skip_leader_election` bool.
	if !k.instance.LeaderSkip {
		// Only run if Leader Election is enabled.
		if !config.Datadog.GetBool("leader_election") {
			return klog.Error("Leader Election not enabled. Not running Kubernetes API Server check or collecting Kubernetes Events.")
		}
		leader, errLeader := cluster.RunLeaderElection()
		if errLeader != nil {
			if errLeader == apiserver.ErrNotLeader {
				// Only the leader can instantiate the apiserver client.
				klog.V(5).Infof("Not leader (leader is %q). Skipping the Kubernetes API Server check", leader)
				return nil
			}

			_ = k.Warn("Leader Election error. Not running the Kubernetes API Server check.")
			return err
		}

		klog.V(6).Infof("Current leader: %q, running the Kubernetes API Server check", leader)
	}
	// API Server client initialisation on first run
	if k.ac == nil {
		// Using GetAPIClient (no wait) as check we'll naturally retry with each check run
		k.ac, err = apiserver.GetAPIClient()
		if err != nil {
			k.Warnf("Could not connect to apiserver: %s", err) //nolint:errcheck
			return err
		}

		// We detect OpenShift presence for quota collection
		if k.instance.CollectOShiftQuotas {
			k.oshiftAPILevel = k.ac.DetectOpenShiftAPILevel()
		}
	}

	// Running the Control Plane status check.
	componentsStatus, err := k.ac.ComponentStatuses()
	if err != nil {
		k.Warnf("Could not retrieve the status from the control plane's components %s", err.Error()) //nolint:errcheck
	} else {
		err = k.parseComponentStatus(sender, componentsStatus)
		if err != nil {
			k.Warnf("Could not collect API Server component status: %s", err.Error()) //nolint:errcheck
		}
	}

	// Running OpenShift ClusterResourceQuota collection if available
	if k.instance.CollectOShiftQuotas && k.oshiftAPILevel != apiserver.NotOpenShift {
		quotas, err := k.retrieveOShiftClusterQuotas()
		if err != nil {
			k.Warnf("Could not collect OpenShift cluster quotas: %s", err.Error()) //nolint:errcheck
		} else {
			k.reportClusterQuotas(quotas, sender)
		}
	}

	// Running the event collection.
	if !k.instance.CollectEvent {
		return nil
	}

	// Get the events from the API server
	events, err := k.eventCollectionCheck()
	if err != nil {
		return err
	}

	// Process the events to have a Datadog format.
	err = k.processEvents(sender, events)
	if err != nil {
		k.Warnf("Could not submit new event %s", err.Error()) //nolint:errcheck
	}
	return nil
}

func (k *KubeASCheck) eventCollectionCheck() (newEvents []*v1.Event, err error) {
	resVer, lastTime, err := k.ac.GetTokenFromConfigmap(eventTokenKey)
	if err != nil {
		return nil, err
	}

	// This is to avoid getting in a situation where we list all the events for multiple runs in a row.
	if resVer == "" && k.eventCollection.LastResVer != "" {
		klog.Errorf("Resource Version stored in the ConfigMap is incorrect. Will resume collecting from %s", k.eventCollection.LastResVer)
		resVer = k.eventCollection.LastResVer
	}

	timeout := int64(k.instance.EventCollectionTimeoutMs / 1000)
	limit := int64(k.instance.MaxEventCollection)
	resync := int64(k.instance.ResyncPeriodEvents)
	newEvents, k.eventCollection.LastResVer, k.eventCollection.LastTime, err = k.ac.RunEventCollection(resVer, lastTime, timeout, limit, resync, k.ignoredEvents)

	if err != nil {
		k.Warnf("Could not collect events from the api server: %s", err.Error()) //nolint:errcheck
		return nil, err
	}

	configMapErr := k.ac.UpdateTokenInConfigmap(eventTokenKey, k.eventCollection.LastResVer, k.eventCollection.LastTime)
	if configMapErr != nil {
		k.Warnf("Could not store the LastEventToken in the ConfigMap: %s", configMapErr.Error()) //nolint:errcheck
	}
	return newEvents, nil
}

func (k *KubeASCheck) parseComponentStatus(sender aggregator.Sender, componentsStatus *v1.ComponentStatusList) error {
	for _, component := range componentsStatus.Items {

		if component.ObjectMeta.Name == "" {
			return errors.New("metadata structure has changed. Not collecting API Server's Components status")
		}
		if component.Conditions == nil || component.Name == "" {
			klog.V(5).Info("API Server component's structure is not expected")
			continue
		}
		tagComp := []string{fmt.Sprintf("component:%s", component.Name)}
		for _, condition := range component.Conditions {
			statusCheck := metrics.ServiceCheckUnknown
			message := ""

			// We only expect the Healthy condition. May change in the future. https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#typical-status-properties
			if condition.Type != "Healthy" {
				klog.V(5).Infof("Condition %q not supported", condition.Type)
				continue
			}
			// We only expect True, False and Unknown (default).
			switch condition.Status {
			case "True":
				statusCheck = metrics.ServiceCheckOK
				message = condition.Message
			case "False":
				statusCheck = metrics.ServiceCheckCritical
				message = condition.Error
			}
			sender.ServiceCheck(KubeControlPaneCheck, statusCheck, "", tagComp, message)
		}
	}
	return nil
}

// processEvents:
// - iterates over the Kubernetes Events
// - extracts some attributes and builds a structure ready to be submitted as a Datadog event (bundle)
// - formats the bundle and submit the Datadog event
func (k *KubeASCheck) processEvents(sender aggregator.Sender, events []*v1.Event) error {
	eventsByObject := make(map[string]*kubernetesEventBundle)

	for _, event := range events {
		id := bundleID(event)
		bundle, found := eventsByObject[id]
		if found == false {
			bundle = newKubernetesEventBundler(event)
			eventsByObject[id] = bundle
		}
		err := bundle.addEvent(event)
		if err != nil {
			k.Warnf("Error while bundling events, %s.", err.Error()) //nolint:errcheck
		}
	}
	hostname, _ := util.GetHostname()
	clusterName := clustername.GetClusterName(hostname)
	for _, bundle := range eventsByObject {
		datadogEv, err := bundle.formatEvents(clusterName, k.providerIDCache)
		if err != nil {
			k.Warnf("Error while formatting bundled events, %s. Not submitting", err.Error()) //nolint:errcheck
			continue
		}
		sender.Event(datadogEv)
	}
	return nil
}

// bundleID generates a unique ID to separate k8s events
// based on their InvolvedObject UIDs and event Types
func bundleID(e *v1.Event) string {
	return fmt.Sprintf("%s/%s", e.InvolvedObject.UID, e.Type)
}

func init() {
	core.RegisterCheck(kubernetesAPIServerCheckName, KubernetesASFactory)
}