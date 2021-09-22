// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package logs

import (
	"time"

	"github.com/n9e/n9e-agentd/pkg/api"
)

// EPIntakeVersion is the events platform intake API version
type EPIntakeVersion uint8

// IntakeTrackType indicates the type of an endpoint intake.
type IntakeTrackType string

// IntakeProtocol indicates the protocol to use for an endpoint intake.
type IntakeProtocol string

// IntakeSource indicates the log source to use for an endpoint intake.
type IntakeSource string

const (
	_ EPIntakeVersion = iota
	// EPIntakeVersion1 is version 1 of the envets platform intake API
	EPIntakeVersion1
	// EPIntakeVersion2 is version 2 of the envets platform intake API
	EPIntakeVersion2
)

// Endpoint holds all the organization and network parameters to send logs to Datadog.
type Endpoint struct {
	APIKey                  string       `json:"api_key"`
	Host                    string       `json:"host"`
	Port                    int          `json:"port"`
	UseSSL                  bool         `json:"use_ssl"`
	UseCompression          bool         `json:"use_compression"`
	CompressionLevel        int          `json:"compression_level"`
	ProxyAddress            string       `json:"proxy_address"`
	ConnectionResetInterval api.Duration `json:"connection_reset_interval"`

	BackoffFactor    float64
	BackoffBase      float64
	BackoffMax       float64
	RecoveryInterval int
	RecoveryReset    bool

	Version   EPIntakeVersion
	TrackType IntakeTrackType
	Protocol  IntakeProtocol
	Source    IntakeSource
}

func (p *Endpoint) Validate() error {
	return nil
}

// Endpoints holds the main endpoint and additional ones to dualship logs.
type Endpoints struct {
	Main                   Endpoint
	Additionals            []Endpoint
	UseProto               bool
	UseHTTP                bool
	BatchWait              time.Duration
	BatchMaxConcurrentSend int
	BatchMaxSize           int
	BatchMaxContentSize    int
}

// NewEndpoints returns a new endpoints composite with default batching settings
func NewEndpoints(main Endpoint, additionals []Endpoint, useProto bool, useHTTP bool) *Endpoints {
	return &Endpoints{
		Main:                   main,
		Additionals:            additionals,
		UseProto:               useProto,
		UseHTTP:                useHTTP,
		BatchWait:              DefaultBatchWait,
		BatchMaxConcurrentSend: DefaultBatchMaxConcurrentSend,
		BatchMaxSize:           DefaultBatchMaxSize,
		BatchMaxContentSize:    DefaultBatchMaxContentSize,
	}
}

// NewEndpointsWithBatchSettings returns a new endpoints composite with non-default batching settings specified
func NewEndpointsWithBatchSettings(main Endpoint, additionals []Endpoint, useProto bool, useHTTP bool, batchWait time.Duration, batchMaxConcurrentSend int, batchMaxSize int, batchMaxContentSize int) *Endpoints {
	return &Endpoints{
		Main:                   main,
		Additionals:            additionals,
		UseProto:               useProto,
		UseHTTP:                useHTTP,
		BatchWait:              batchWait,
		BatchMaxConcurrentSend: batchMaxConcurrentSend,
		BatchMaxSize:           batchMaxSize,
		BatchMaxContentSize:    batchMaxContentSize,
	}
}
