package config

import "sync"

const (
	S_statsd_metrics_stats_enable    = "statsd_metrics_stats_enable"    // statsd_metrics_stats_enable
	S_log_level                      = "log_level"                      // log_level
	S_runtime_mutex_profile_fraction = "runtime_mutex_profile_fraction" // internal_profiling.mutex_profile_fraction
	S_runtime_block_profile_rate     = "runtime_block_profile_rate"     // internal_profiling.block_profile_rate
	S_dogstatsd_stats                = "dogstatsd_stats"                // dogstatsd_metrics_stats_enable
	S_dogstatsd_capture_duration     = "dogstatsd_capture_duration"     // TODO
	S_internal_profiling_goroutines  = "internal_profiling_goroutines"  // internal_profiling.enable_goroutine_stacktraces
	S_internal_profiling             = "internal_profiling"
)

type settings struct {
	sync.RWMutex
	data map[string]interface{}
}

var (
	_settings = &settings{
		data: map[string]interface{}{},
	}
)

func Get(key string) interface{} {
	_settings.RLock()
	defer _settings.RUnlock()
	return _settings.data[key]
}

func Set(key string, value interface{}) {
	_settings.Lock()
	defer _settings.Unlock()
	_settings.data[key] = value
}
