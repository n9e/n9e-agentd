// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package config

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-agent/pkg/trace/osutil"
	"github.com/DataDog/datadog-agent/pkg/trace/traceutil"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/config/apm"
)

// apiEndpointPrefix is the URL prefix prepended to the default site value from YamlAgentConfig.
const apiEndpointPrefix = "https://trace.agent."

// OTLP holds the configuration for the OpenTelemetry receiver.
type OTLP struct {
	// BindHost specifies the host to bind the receiver to.
	BindHost string `mapstructure:"-"`

	// HTTPPort specifies the port to use for the plain HTTP receiver.
	// If unset (or 0), the receiver will be off.
	HTTPPort int `mapstructure:"http_port"`

	// GRPCPort specifies the port to use for the plain HTTP receiver.
	// If unset (or 0), the receiver will be off.
	GRPCPort int `mapstructure:"grpc_port"`

	// MaxRequestBytes specifies the maximum number of bytes that will be read
	// from an incoming HTTP request.
	MaxRequestBytes int64 `mapstructure:"-"`
}

func (c *AgentConfig) applyDatadogConfig() error {
	if len(c.Endpoints) == 0 {
		c.Endpoints = []*Endpoint{{}}
	}
	if config.C.ApiKey != "" {
		c.Endpoints[0].APIKey = config.SanitizeAPIKey(config.C.ApiKey)
	}
	if config.C.Hostname != "" {
		c.Hostname = config.C.Hostname
	}
	if config.C.LogLevel != "" {
		c.LogLevel = config.C.LogLevel
	}
	if config.C.Statsd.Port > 0 {
		c.StatsdPort = config.C.Statsd.Port
	}

	endpoints := config.C.Endpoints
	if len(endpoints) > 0 {
		c.Endpoints[0].Hosts = endpoints
	}
	if host := config.C.Apm.ApmUrl; host != "" {
		c.Endpoints[0].Hosts = []string{host}
		if len(endpoints) > 0 {
			log.Infof("'endpoints' and 'apm.apm_url' are both set, using endpoint: %q", host)
		}
	}

	if endpoints := config.C.Apm.AdditionalEndpoints; len(endpoints) > 0 {
		for url, keys := range endpoints {
			if len(keys) == 0 {
				log.Errorf("'additional_endpoints' entries must have at least one API key present")
				continue
			}
			for _, key := range keys {
				key = config.SanitizeAPIKey(key)
				c.Endpoints = append(c.Endpoints, &Endpoint{Hosts: []string{url}, APIKey: key})
			}
		}
	}

	if proxyList := config.C.Proxy.NoProxy; len(proxyList) > 0 {
		noProxy := make(map[string]bool, len(proxyList))
		for _, host := range proxyList {
			// map of hosts that need to be skipped by proxy
			noProxy[host] = true
		}
		for _, e := range c.Endpoints {
			for _, h := range e.Hosts {
				e.NoProxy = noProxy[h]
			}
		}
	}
	if addr := config.C.Proxy.HTTPS; addr != "" {
		url, err := url.Parse(addr)
		if err == nil {
			c.ProxyURL = url
		} else {
			log.Errorf("Failed to parse proxy URL from proxy.https configuration: %s", err)
		}
	}

	cf := config.C

	if cf.SkipSSLValidation {
		c.SkipSSLValidation = cf.SkipSSLValidation
	}
	if cf.Apm.Enabled {
		c.Enabled = cf.Apm.Enabled
	}
	if cf.Apm.LogFile != "" {
		c.LogFilePath = cf.Apm.LogFile
	}
	if cf.Apm.Env != "" {
		c.DefaultEnv = cf.Apm.Env
		log.Debugf("Setting DefaultEnv to %q (from apm_config.env)", c.DefaultEnv)
	} else if cf.Env != "" {
		c.DefaultEnv = cf.Env
		log.Debugf("Setting DefaultEnv to %q (from 'env' config option)", c.DefaultEnv)
	} else {
		for _, tag := range config.GetConfiguredTags(false) {
			if strings.HasPrefix(tag, "env:") {
				c.DefaultEnv = strings.TrimPrefix(tag, "env:")
				log.Debugf("Setting DefaultEnv to %q (from `env:` entry under the 'tags' config option: %q)", c.DefaultEnv, tag)
				break
			}
		}
	}
	prevEnv := c.DefaultEnv
	c.DefaultEnv = traceutil.NormalizeTag(c.DefaultEnv)
	if c.DefaultEnv != prevEnv {
		log.Debugf("Normalized DefaultEnv from %q to %q", prevEnv, c.DefaultEnv)
	}
	if cf.Apm.ReceiverPort > 0 {
		c.ReceiverPort = cf.Apm.ReceiverPort
	}
	if cf.Apm.ReceiverSocket != "" {
		c.ReceiverSocket = cf.Apm.ReceiverSocket
	}
	if cf.Apm.ConnectionLimit > 0 {
		c.ConnectionLimit = cf.Apm.ConnectionLimit
	}
	if cf.Apm.ExtraSampleRate > 0 {
		c.ExtraSampleRate = cf.Apm.ExtraSampleRate
	}
	if cf.Apm.MaxEventsPerSecond > 0 {
		c.MaxEPS = cf.Apm.MaxEventsPerSecond
	}
	if cf.Apm.MaxTracesPerSecond > 0 {
		c.TargetTPS = cf.Apm.MaxTracesPerSecond
	}
	if v := cf.Apm.IgnoreResources; len(v) > 0 {
		c.Ignore["resource"] = v
	}
	if v := cf.Apm.MaxPayloadSize; v > 0 {
		c.MaxRequestBytes = v
	}
	if v := cf.Apm.ReplaceTags; len(v) > 0 {
		c.ReplaceTags = v
	}

	if cf.BindHost != "" || cf.Apm.ApmNonLocalTraffic {
		if cf.BindHost != "" {
			host := cf.BindHost
			c.StatsdHost = host
			c.ReceiverHost = host
		}

		if cf.Apm.ApmNonLocalTraffic {
			c.ReceiverHost = "0.0.0.0"
		}
	} else if config.IsContainerized() {
		// Automatically activate non local traffic in containerized environment if no explicit config set
		log.Info("Activating non-local traffic automatically in containerized environment, trace-agent will listen on 0.0.0.0")
		c.ReceiverHost = "0.0.0.0"
	}

	c.OTLPReceiver = &OTLP{
		BindHost:        c.ReceiverHost,
		HTTPPort:        cf.Experimental.OTLP.HTTPPort,
		GRPCPort:        cf.Experimental.OTLP.GRPCPort,
		MaxRequestBytes: c.MaxRequestBytes,
	}

	if cf.IsSet("apm_config.obfuscation") {
		//var o ObfuscationConfig
		c.Obfuscation = &cf.Apm.Obfuscation
		//if err == nil {
		//	c.Obfuscation = &o
		//	if c.Obfuscation.RemoveStackTraces {
		//		c.addReplaceRule("error.stack", `(?s).*`, "?")
		//	}
		//}
	}

	if v := cf.Apm.FilterTagsRequire; len(v) > 0 {
		for _, tag := range v {
			c.RequireTags = append(c.RequireTags, splitTag(tag))
		}
	}
	if v := cf.Apm.FilterTagsReject; len(v) > 0 {
		for _, tag := range v {
			c.RejectTags = append(c.RejectTags, splitTag(tag))
		}
	}

	// undocumented
	if v := cf.Apm.MaxCpuPercent; v > 0 {
		c.MaxCPU = v / 100
	}
	if v := cf.Apm.MaxMemory; v > 0 {
		c.MaxMemory = v
	}

	// undocumented writers
	if cf.Apm.TraceWriter != nil {
		c.TraceWriter = cf.Apm.TraceWriter
	}
	if cf.Apm.StatsWriter != nil {
		c.StatsWriter = cf.Apm.StatsWriter
	}

	if v := cf.Apm.ConnectionResetInterval; v > 0 {
		c.ConnectionResetInterval = getDuration(v)
	}
	if v := cf.Apm.SyncFlushing; v {
		c.SynchronousFlushing = v
	}

	// undocumeted
	for key, rate := range cf.Apm.AnalyzedSpans {
		serviceName, operationName, err := parseServiceAndOp(key)
		if err != nil {
			log.Errorf("Error parsing names: %v", err)
			continue
		}
		if _, ok := c.AnalyzedSpansByService[serviceName]; !ok {
			c.AnalyzedSpansByService[serviceName] = make(map[string]float64)
		}
		c.AnalyzedSpansByService[serviceName][operationName] = rate

	}

	// undocumented
	if v := cf.Apm.DDAgentBin; len(v) > 0 {
		c.DDAgentBin = v
	}

	if err := c.loadDeprecatedValues(); err != nil {
		return err
	}

	if strings.ToLower(c.LogLevel) == "debug" && !cf.Apm.LogThrottling {
		// if we are in "debug mode" and log throttling behavior was not
		// set by the user, disable it
		c.LogThrottling = false
	}

	return nil
}

// loadDeprecatedValues loads a set of deprecated values which are kept for
// backwards compatibility with Agent 5. These should eventually be removed.
// TODO(x): remove them gradually or fully in a future release.
func (c *AgentConfig) loadDeprecatedValues() error {
	cfg := config.C.Apm
	if v := cfg.ApiKey; v != "" {
		c.Endpoints[0].APIKey = config.SanitizeAPIKey(v)
	}
	if v := cfg.LogLevel; len(v) > 0 {
		c.LogLevel = v
	}
	if v := cfg.LogThrottling; v {
		c.LogThrottling = v
	}
	if v := cfg.BucketSizeSeconds; v > 0 {
		c.BucketInterval = getDuration(v)
	}
	if v := cfg.ReceiverTimeout; v > 0 {
		c.ReceiverTimeout = v
	}
	if v := cfg.WatchdogCheckDelay; v > 0 {
		c.WatchdogInterval = getDuration(v)
	}
	return nil
}

// addReplaceRule adds the specified replace rule to the agent configuration. If the pattern fails
// to compile as valid regexp, it exits the application with status code 1.
func (c *AgentConfig) addReplaceRule(tag, pattern, repl string) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		osutil.Exitf("error adding replace rule: %s", err)
	}
	c.ReplaceTags = append(c.ReplaceTags, &apm.ReplaceRule{
		Name:    tag,
		Pattern: pattern,
		Re:      re,
		Repl:    repl,
	})
}

// compileReplaceRules compiles the regular expressions found in the replace rules.
// If it fails it returns the first error.
func compileReplaceRules(rules []*apm.ReplaceRule) error {
	for _, r := range rules {
		if r.Name == "" {
			return errors.New(`all rules must have a "name" property (use "*" to target all)`)
		}
		if r.Pattern == "" {
			return errors.New(`all rules must have a "pattern"`)
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return fmt.Errorf("key %q: %s", r.Name, err)
		}
		r.Re = re
	}
	return nil
}

// getDuration returns the duration of the provided value in seconds
func getDuration(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}

func parseServiceAndOp(name string) (string, string, error) {
	splits := strings.Split(name, "|")
	if len(splits) != 2 {
		return "", "", fmt.Errorf("Bad format for operation name and service name in: %s, it should have format: service_name|operation_name", name)
	}
	return splits[0], splits[1], nil
}

func splitString(s string, sep rune) ([]string, error) {
	r := csv.NewReader(strings.NewReader(s))
	r.TrimLeadingSpace = true
	r.LazyQuotes = true
	r.Comma = sep

	return r.Read()
}

func toFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	default:
		return 0, fmt.Errorf("%v can not be converted to float64", val)
	}
}

// splitTag splits a "k:v" formatted string and returns a Tag.
func splitTag(tag string) *Tag {
	parts := strings.SplitN(tag, ":", 2)
	kv := &Tag{
		K: strings.TrimSpace(parts[0]),
	}
	if len(parts) > 1 {
		if v := strings.TrimSpace(parts[1]); v != "" {
			kv.V = v
		}
	}
	return kv
}
