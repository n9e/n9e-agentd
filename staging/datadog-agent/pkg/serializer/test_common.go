// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build test

package serializer

import (
	"github.com/stretchr/testify/mock"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/serializer/marshaler"
)

// MockSerializer is a mock for the MetricSerializer
type MockSerializer struct {
	mock.Mock
}

// SendEvents serializes a list of event and sends the payload to the forwarder
func (s *MockSerializer) SendEvents(e EventsStreamJSONMarshaler) error {
	return s.Called(e).Error(0)
}

// SendServiceChecks serializes a list of serviceChecks and sends the payload to the forwarder
func (s *MockSerializer) SendServiceChecks(sc marshaler.StreamJSONMarshaler) error {
	return s.Called(sc).Error(0)
}

// SendSeries serializes a list of serviceChecks and sends the payload to the forwarder
func (s *MockSerializer) SendSeries(series marshaler.StreamJSONMarshaler) error {
	return s.Called(series).Error(0)
}

// SendSketch serializes a list of SketSeriesList and sends the payload to the forwarder
func (s *MockSerializer) SendSketch(sketches marshaler.Marshaler) error {
	return s.Called(sketches).Error(0)
}

// SendMetadata serializes a metadata payload and sends it to the forwarder
func (s *MockSerializer) SendMetadata(m marshaler.Marshaler) error {
	return s.Called(m).Error(0)
}

// SendHostMetadata serializes a host metadata payload and sends it to the forwarder
func (s *MockSerializer) SendHostMetadata(m marshaler.Marshaler) error {
	return s.Called(m).Error(0)
}

// SendJSONToV1Intake serializes a payload and sends it to the forwarder. Some code sends
// arbitrary payload the v1 API.
func (s *MockSerializer) SendJSONToV1Intake(data interface{}) error {
	return s.Called(data).Error(0)
}

// SendOrchestratorMetadata serializes & send orchestrator metadata payloads
func (s *MockSerializer) SendOrchestratorMetadata(msgs []ProcessMessageBody, hostName, clusterID, payloadType string) error {
	return s.Called(msgs, hostName, clusterID, payloadType).Error(0)
}
