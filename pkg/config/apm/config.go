package apm

import (
	"errors"
	"fmt"
	"regexp"
)

type Config struct {
	AdditionalEndpoints           map[string][]string `json:"additional_endpoints"`             // apm_config.additional_endpoints
	AnalyzedSpans                 map[string]float64  `json:"analyzed_spans"`                   // apm_config.analyzed_spans
	ApiKey                        string              `json:"api_key"`                          // apm_config.api_key
	ApmUrl                        string              `json:"apm_url"`                          // apm_config.apm_dd_url
	ApmNonLocalTraffic            bool                `json:"apm_non_local_traffic"`            // apm_config.apm_non_local_traffic
	ConnectionLimit               int                 `json:"connection_limit"`                 // apm_config.connection_limit
	ConnectionResetInterval       int                 `json:"connection_reset_interval"`        // apm_config.connection_reset_interval
	DDAgentBin                    string              `json:"dd_agent_bin"`                     // apm_config.dd_agent_bin
	Enabled                       bool                `json:"enabled"`                          // apm_config.enabled
	Env                           string              `json:"env"`                              // apm_config.env
	ExtraSampleRate               float64             `json:"extra_sample_rate"`                // apm_config.extra_sample_rate
	FilterTagsReject              []string            `json:"filter_tags_reject"`               // apm_config.filter_tags.reject
	FilterTagsRequire             []string            `json:"filter_tags_require"`              // apm_config.filter_tags.require
	IgnoreResources               []string            `json:"ignore_resources"`                 // apm_config.ignore_resources
	LogFile                       string              `json:"log_file"`                         // apm_config.log_file
	LogLevel                      string              `json:"log_level"`                        // apm_config.log_level
	LogThrottling                 bool                `json:"log_throttling"`                   // apm_config.log_throttling
	MaxCpuPercent                 float64             `json:"max_cpu_percent"`                  // apm_config.max_cpu_percent
	MaxEventsPerSecond            float64             `json:"max_events_per_second"`            // apm_config.max_events_per_second
	MaxMemory                     float64             `json:"max_memory"`                       // apm_config.max_memory
	MaxTracesPerSecond            float64             `json:"max_traces_per_second"`            // apm_config.max_traces_per_second
	Obfuscation                   *ObfuscationConfig  `json:"obfuscation"`                      // apm_config.obfuscation
	ProfilingAdditionalEndpoints  bool                `json:"profiling_additional_endpoints"`   // apm_config.profiling_additional_endpoints
	ProfilingDdUrl                bool                `json:"profiling_dd_url"`                 // apm_config.profiling_dd_url
	ReceiverPort                  int                 `json:"receiver_port"`                    // apm_config.receiver_port
	ReceiverSocket                string              `json:"receiver_socket"`                  // apm_config.receiver_socket
	ReceiverTimeout               int                 `json:"receiver_timeout"`                 // apm_config.receiver_timeout
	RemoteTagger                  bool                `json:"remote_tagger"`                    // apm_config.remote_tagger
	SyncFlushing                  bool                `json:"sync_flushing"`                    // apm_config.sync_flushing
	WindowsPipeBufferSize         bool                `json:"windows_pipe_buffer_size"`         // apm_config.windows_pipe_buffer_size
	WindowsPipeName               bool                `json:"windows_pipe_name"`                // apm_config.windows_pipe_name
	WindowsPipeSecurityDescriptor bool                `json:"windows_pipe_security_descriptor"` // apm_config.windows_pipe_security_descriptor
	MaxPayloadSize                int64               `json:"max_payload_size"`
	ReplaceTags                   []*ReplaceRule      `json:"replace_tags"`
	TraceWriter                   *WriterConfig       `json:"trace_writer"`
	StatsWriter                   *WriterConfig       `json:"stats_writer"`
	BucketSizeSeconds             int                 `json:"bucket_size_seconds"`
	WatchdogCheckDelay            int                 `json:"watchdog_check_delay"`
}

// WriterConfig specifies configuration for an API writer.
type WriterConfig struct {
	// ConnectionLimit specifies the maximum number of concurrent outgoing
	// connections allowed for the sender.
	ConnectionLimit int `mapstructure:"connection_limit"`

	// QueueSize specifies the maximum number or payloads allowed to be queued
	// in the sender.
	QueueSize int `mapstructure:"queue_size"`

	// FlushPeriodSeconds specifies the frequency at which the writer's buffer
	// will be flushed to the sender, in seconds. Fractions are permitted.
	FlushPeriodSeconds float64 `mapstructure:"flush_period_seconds"`
}

// compileReplaceRules compiles the regular expressions found in the replace rules.
// If it fails it returns the first error.
func compileReplaceRules(rules []*ReplaceRule) error {
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

// ReplaceRule specifies a replace rule.
type ReplaceRule struct {
	// Name specifies the name of the tag that the replace rule addresses. However,
	// some exceptions apply such as:
	// • "resource.name" will target the resource
	// • "*" will target all tags and the resource
	Name string `mapstructure:"name"`

	// Pattern specifies the regexp pattern to be used when replacing. It must compile.
	Pattern string `mapstructure:"pattern"`

	// Re holds the compiled Pattern and is only used internally.
	Re *regexp.Regexp `mapstructure:"-"`

	// Repl specifies the replacement string to be used when Pattern matches.
	Repl string `mapstructure:"repl"`
}

func (p *Config) Validate() error {
	if err := compileReplaceRules(p.ReplaceTags); err != nil {
		return err
	}
	return nil
}

// ObfuscationConfig holds the configuration for obfuscating sensitive data
// for various span types.
type ObfuscationConfig struct {
	// ES holds the obfuscation configuration for ElasticSearch bodies.
	ES JSONObfuscationConfig `mapstructure:"elasticsearch"`

	// Mongo holds the obfuscation configuration for MongoDB queries.
	Mongo JSONObfuscationConfig `mapstructure:"mongodb"`

	// SQLExecPlan holds the obfuscation configuration for SQL Exec Plans. This is strictly for safety related obfuscation,
	// not normalization. Normalization of exec plans is configured in SQLExecPlanNormalize.
	SQLExecPlan JSONObfuscationConfig `mapstructure:"sql_exec_plan"`

	// SQLExecPlanNormalize holds the normalization configuration for SQL Exec Plans.
	SQLExecPlanNormalize JSONObfuscationConfig `mapstructure:"sql_exec_plan_normalize"`

	// HTTP holds the obfuscation settings for HTTP URLs.
	HTTP HTTPObfuscationConfig `mapstructure:"http"`

	// RemoveStackTraces specifies whether stack traces should be removed.
	// More specifically "error.stack" tag values will be cleared.
	RemoveStackTraces bool `mapstructure:"remove_stack_traces"`

	// Redis holds the configuration for obfuscating the "redis.raw_command" tag
	// for spans of type "redis".
	Redis Enablable `mapstructure:"redis"`

	// Memcached holds the configuration for obfuscating the "memcached.command" tag
	// for spans of type "memcached".
	Memcached Enablable `mapstructure:"memcached"`
}

// JSONObfuscationConfig holds the obfuscation configuration for sensitive
// data found in JSON objects.
type JSONObfuscationConfig struct {
	// Enabled will specify whether obfuscation should be enabled.
	Enabled bool `mapstructure:"enabled"`

	// KeepValues will specify a set of keys for which their values will
	// not be obfuscated.
	KeepValues []string `mapstructure:"keep_values"`

	// ObfuscateSQLValues will specify a set of keys for which their values
	// will be passed through SQL obfuscation
	ObfuscateSQLValues []string `mapstructure:"obfuscate_sql_values"`
}

// Enablable can represent any option that has an "enabled" boolean sub-field.
type Enablable struct {
	Enabled bool `mapstructure:"enabled"`
}

// HTTPObfuscationConfig holds the configuration settings for HTTP obfuscation.
type HTTPObfuscationConfig struct {
	// RemoveQueryStrings determines query strings to be removed from HTTP URLs.
	RemoveQueryString bool `mapstructure:"remove_query_string" json:"remove_query_string"`

	// RemovePathDigits determines digits in path segments to be obfuscated.
	RemovePathDigits bool `mapstructure:"remove_paths_with_digits" json:"remove_path_digits"`
}
