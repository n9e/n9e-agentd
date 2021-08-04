package forwarder

import (
	"fmt"
	"net/url"
	"time"
)

type Config struct {
	AdditionalEndpoints       []AdditionalEndpoint `json:"additional_endpoints"` // additional_endpoints
	ApikeyValidationInterval  time.Duration        `json:"-"`
	ApikeyValidationInterval_ int                  `json:"apikey_validation_interval" flag:"forwarder-apikey-validation-interval" default:"3600" description:"apikeyValidationInterval(Second)"` // forwarder_apikey_validation_interval
	BackoffBase               float64              `json:"backoff_base" default:"2"`                                                                                                             // forwarder_backoff_base
	BackoffFactor             float64              `json:"backoff_factor" default:"2"`                                                                                                           // forwarder_backoff_factor
	BackoffMax                float64              `json:"backoff_max" default:"64"`                                                                                                             // forwarder_backoff_max
	ConnectionResetInterval   time.Duration        `json:"-"`
	ConnectionResetInterval_  int                  `json:"connection_reset_interval" flag:"forwarder-connection-reset-interval" description:"connectionResetInterval(Second)"` // forwarder_connection_reset_interval
	FlushToDiskMemRatio       float64              `json:"flush_to_disk_mem_ratio" default:"0.5"`                                                                              // forwarder_flush_to_disk_mem_ratio
	NumWorkers                int                  `json:"num_workers" default:"1"`                                                                                            // forwarder_num_workers
	OutdatedFileInDays        int                  `json:"outdated_file_in_days" default:"10"`                                                                                 // forwarder_outdated_file_in_days
	RecoveryInterval          int                  `json:"recovery_interval" default:"2"`                                                                                      // forwarder_recovery_interval
	RecoveryReset             bool                 `json:"recovery_reset"`                                                                                                     // forwarder_recovery_reset
	StopTimeout               time.Duration        `json:"-"`
	StopTimeout_              int                  `json:"stop_timeout" flag:"forwarder-stop-timeout" default:"2" description:"stopTimeout(Second)"` // forwarder_stop_timeout
	StorageMaxDiskRatio       float64              `json:"storage_max_disk_ratio" default:"0.95"`                                                    // forwarder_storage_max_disk_ratio
	StorageMaxSizeInBytes     int64                `json:"storage_max_size_in_bytes"`                                                                // forwarder_storage_max_size_in_bytes
	StoragePath               string               `json:"storage_path"`                                                                             // forwarder_storage_path
	Timeout                   time.Duration        `json:"-"`
	Timeout_                  int                  `json:"timeout" flag:"forwarder-timeout" default:"20" description:"timeout(Second)"` // forwarder_timeout

	HighPrioBufferSize        int `json:"high_prio_buffer_size"`                                              //forwarder_high_prio_buffer_size
	LowPrioBufferSize         int `json:"low_prio_buffer_size"`                                               //forwarder_low_prio_buffer_size
	RequeueBufferSize         int `json:"requeue_buffer_size"`                                                //forwarder_requeue_buffer_size
	RetryQueueMaxSize         int `json:"retry_queue_max_size"`                                               // forwarder_retry_queue_max_size
	RetryQueuePayloadsMaxSize int `json:"retry_queue_payloads_max_size" default:"15728640" description:"15m"` // forwarder_retry_queue_payloads_max_size

}

func (p *Config) Validate() error {
	p.ApikeyValidationInterval = time.Second * time.Duration(p.ApikeyValidationInterval_)
	p.ConnectionResetInterval = time.Second * time.Duration(p.ConnectionResetInterval_)
	p.StopTimeout = time.Second * time.Duration(p.StopTimeout_)
	p.Timeout = time.Second * time.Duration(p.Timeout_)
	for i, addtion := range p.AdditionalEndpoints {
		for j, endpoint := range addtion.Endpoints {
			if _, err := url.Parse(endpoint); err != nil {
				return fmt.Errorf("could not parse agent.forwarder.addtionEndpoints[%d][%d] %s %s", i, j, endpoint, err)
			}
		}
	}

	if p.RecoveryInterval <= 0 {
		return fmt.Errorf("Configured forwarder.recoveryInterval (%v) is not positive", p.RecoveryInterval)
	}
	return nil
}

type AdditionalEndpoint struct {
	Endpoints []string `json:"endpoints"`
	ApiKeys   []string `json:"api_keys"`
}
