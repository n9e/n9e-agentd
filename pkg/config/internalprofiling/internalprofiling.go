package internalprofiling

import (
	"github.com/yubo/golib/api"
)

type InternalProfiling struct { // internal_profiling
	Enabled                    bool         `json:"enabled" env:"N9E_INTERNAL_PROFILING_ENABLED"` // enabled
	Site                       string       `json:"site" env:"N9E_INTERNAL_PROFILING_SITE"`       // site
	Url                        string       `json:"url" env:"N9E_INTERNAL_PROFILING_URL"`         // profile_dd_url
	ApiKey                     string       `json:"api_key" env:"N9E_API_KEY"`                    //
	Env                        string       `json:"env" env:"N9E_INTERNAL_PROFILING_ENV"`         // env
	Period                     api.Duration `json:"period" description:"period time"`             // period
	CpuDuration                api.Duration `json:"cpu_duration" description:"cpu duration"`      // cpu_duration
	MutexProfileFraction       int          `json:"mutex_profile_fraction"`                       // mutex_profile_fraction
	BlockProfileRate           int          `json:"block_profile_rate"`                           // block_profile_rate
	EnableGoroutineStacktraces bool         `json:"enable_goroutine_stacktraces"`                 // enable_goroutine_stacktraces
}

func (p *InternalProfiling) Validate() error {
	return nil
}
