package forwarder

import (
	"fmt"
	"net/url"

	"github.com/yubo/golib/api"
	"github.com/yubo/golib/api/resource"
)

func NewConfig() Config {
	// defualt config
	return Config{
		ApikeyValidationInterval:  api.NewDuration("60m"),
		BackoffBase:               2,
		BackoffFactor:             2,
		BackoffMax:                64,
		FlushToDiskMemRatio:       0.5,
		NumWorkers:                1,
		OutdatedFileInDays:        10,
		RecoveryInterval:          2,
		StopTimeout:               api.NewDuration("2s"),
		StorageMaxDiskRatio:       0.95,
		Timeout:                   api.NewDuration("20s"),
		HighPrioBufferSize:        100,
		LowPrioBufferSize:         100,
		RequeueBufferSize:         100,
		RetryQueueMaxSize:         0,
		RetryQueuePayloadsMaxSize: resource.MustParse("15Mi"),
	}
}

type Config struct {
	AdditionalEndpoints      []AdditionalEndpoint `json:"additional_endpoints"`                                                                                          // additional_endpoints
	ApikeyValidationInterval api.Duration         `json:"apikey_validation_interval" flag:"forwarder-apikey-validation-interval" description:"apikeyValidationInterval"` // forwarder_apikey_validation_interval
	BackoffBase              float64              `json:"backoff_base"`                                                                                                  // forwarder_backoff_base
	BackoffFactor            float64              `json:"backoff_factor"`                                                                                                // forwarder_backoff_factor
	BackoffMax               float64              `json:"backoff_max"`                                                                                                   // forwarder_backoff_max
	ConnectionResetInterval  api.Duration         `json:"connection_reset_interval" flag:"forwarder-connection-reset-interval" description:"connectionResetInterval"`    // forwarder_connection_reset_interval
	FlushToDiskMemRatio      float64              `json:"flush_to_disk_mem_ratio""`                                                                                      // forwarder_flush_to_disk_mem_ratio
	NumWorkers               int                  `json:"num_workers"`                                                                                                   // forwarder_num_workers
	OutdatedFileInDays       int                  `json:"outdated_file_in_days"`                                                                                         // forwarder_outdated_file_in_days
	RecoveryInterval         int                  `json:"recovery_interval"`                                                                                             // forwarder_recovery_interval
	RecoveryReset            bool                 `json:"recovery_reset"`                                                                                                // forwarder_recovery_reset
	StopTimeout              api.Duration         `json:"stop_timeout" flag:"forwarder-stop-timeout" description:"stopTimeout"`                                          // forwarder_stop_timeout
	StorageMaxDiskRatio      float64              `json:"storage_max_disk_ratio"`                                                                                        // forwarder_storage_max_disk_ratio
	StorageMaxSizeInBytes    resource.Quantity    `json:"storage_max_size_in_bytes"`                                                                                     // forwarder_storage_max_size_in_bytes
	StoragePath              string               `json:"storage_path" description:"default {root}/transactions_to_retry"`                                               // forwarder_storage_path
	Timeout                  api.Duration         `json:"timeout" flag:"forwarder-timeout"description:"timeout"`                                                         // forwarder_timeout

	HighPrioBufferSize        int               `json:"high_prio_buffer_size"`                            //forwarder_high_prio_buffer_size
	LowPrioBufferSize         int               `json:"low_prio_buffer_size"`                             //forwarder_low_prio_buffer_size
	RequeueBufferSize         int               `json:"requeue_buffer_size"`                              //forwarder_requeue_buffer_size
	RetryQueueMaxSize         int               `json:"retry_queue_max_size"`                             // forwarder_retry_queue_max_size
	RetryQueuePayloadsMaxSize resource.Quantity `json:"retry_queue_payloads_max_size" description:"15Mi"` // forwarder_retry_queue_payloads_max_size

}

func (p *Config) Validate() error {
	for i, addtion := range p.AdditionalEndpoints {
		for j, endpoint := range addtion.Endpoints {
			if _, err := url.Parse(endpoint); err != nil {
				return fmt.Errorf("could not parse agent.forwarder.addtionEndpoints[%d][%d] %s %s", i, j, endpoint, err)
			}
		}
	}

	if p.RecoveryInterval <= 0 {
		return fmt.Errorf("Configured forwarder.recovery_interval (%v) is not positive", p.RecoveryInterval)
	}
	return nil
}

type AdditionalEndpoint struct {
	Endpoints []string `json:"endpoints"`
	ApiKeys   []string `json:"api_keys"`
}
