// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package serializer

import (
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/forwarder"
	"github.com/n9e/n9e-agentd/pkg/process/util/api/headers"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/serializer/marshaler"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/serializer/split"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/compression"
	"k8s.io/klog/v2"
)

const (
	protobufContentType                         = "application/x-protobuf"
	jsonContentType                             = "application/json"
	payloadVersionHTTPHeader                    = "Agentd-Payload"
	maxItemCountForCreateMarshalersBySourceType = 100
)

var (
	// AgentPayloadVersion is the versions of the agent-payload repository
	// used to serialize to protobuf
	AgentPayloadVersion string

	jsonExtraHeaders                    http.Header
	protobufExtraHeaders                http.Header
	jsonExtraHeadersWithCompression     http.Header
	protobufExtraHeadersWithCompression http.Header

	expvars                                 = expvar.NewMap("serializer")
	expvarsSendEventsErrItemTooBigs         = expvar.Int{}
	expvarsSendEventsErrItemTooBigsFallback = expvar.Int{}
)

func init() {
	expvars.Set("SendEventsErrItemTooBigs", &expvarsSendEventsErrItemTooBigs)
	expvars.Set("SendEventsErrItemTooBigsFallback", &expvarsSendEventsErrItemTooBigsFallback)
	initExtraHeaders()
}

// initExtraHeaders initializes the global extraHeaders variables.
// Not part of the `init` function body to ease testing
func initExtraHeaders() {
	jsonExtraHeaders = make(http.Header)
	jsonExtraHeaders.Set("Content-Type", jsonContentType)

	jsonExtraHeadersWithCompression = make(http.Header)
	for k := range jsonExtraHeaders {
		jsonExtraHeadersWithCompression.Set(k, jsonExtraHeaders.Get(k))
	}

	protobufExtraHeaders = make(http.Header)
	protobufExtraHeaders.Set("Content-Type", protobufContentType)

	protobufExtraHeaders.Set(payloadVersionHTTPHeader, AgentPayloadVersion)

	protobufExtraHeadersWithCompression = make(http.Header)
	for k := range protobufExtraHeaders {
		protobufExtraHeadersWithCompression.Set(k, protobufExtraHeaders.Get(k))
	}

	if compression.ContentEncoding != "" {
		jsonExtraHeadersWithCompression.Set("Content-Encoding", compression.ContentEncoding)
		protobufExtraHeadersWithCompression.Set("Content-Encoding", compression.ContentEncoding)
	}
}

// EventsMarshaler handles two serialization logics.
type EventsMarshaler interface {
	marshaler.Marshaler

	// Create a single marshaler.
	//CreateSingleMarshaler() marshaler.StreamJSONMarshaler

	// If the single marshaler cannot serialize, use smaller marshalers.
	//CreateMarshalersBySourceType() []marshaler.StreamJSONMarshaler
}

// MetricSerializer represents the interface of method needed by the aggregator to serialize its data
type MetricSerializer interface {
	SendEvents(e EventsMarshaler) error
	SendServiceChecks(sc marshaler.Marshaler) error
	SendSeries(series marshaler.Marshaler) error
	SendSketch(sketches marshaler.Marshaler) error
	SendMetadata(m marshaler.Marshaler) error
	SendHostMetadata(m marshaler.Marshaler) error
	SendJSONToV1Intake(data interface{}) error
	SendOrchestratorMetadata(msgs []ProcessMessageBody, hostName, clusterID, payloadType string) error
}

// Serializer serializes metrics to the correct format and routes the payloads to the correct endpoint in the Forwarder
type Serializer struct {
	Forwarder             forwarder.Forwarder
	orchestratorForwarder forwarder.Forwarder

	//seriesJSONPayloadBuilder *stream.JSONPayloadBuilder

	// Those variables allow users to blacklist any kind of payload
	// from being sent by the agent. This was introduced for
	// environment where, for example, events or serviceChecks
	// might collect data considered too sensitive (database IP and
	// such). By default every kind of payload is enabled since
	// almost every user won't fall into this use case.
	enableEvents              bool
	enableSeries              bool
	enableServiceChecks       bool
	enableSketches            bool
	enableJSONToV1Intake      bool
	enableMetadata            bool
	enableHostMetadata        bool
	enableAgentchecksMetadata bool
	//enableJSONStream              bool
	//enableServiceChecksJSONStream bool
	//enableEventsJSONStream        bool
	//enableSketchProtobufStream    bool
}

// NewSerializer returns a new Serializer initialized
func NewSerializer(forwarder forwarder.Forwarder, orchestratorForwarder forwarder.Forwarder) *Serializer {
	cf := config.C
	s := &Serializer{
		Forwarder:             forwarder,
		orchestratorForwarder: orchestratorForwarder,
		//seriesJSONPayloadBuilder: stream.NewJSONPayloadBuilder(cf.EnableJsonStreamSharedCompressorBuffers),
		enableEvents:              cf.EnablePayloads.Events,
		enableSeries:              cf.EnablePayloads.Series,
		enableServiceChecks:       cf.EnablePayloads.ServiceChecks,
		enableSketches:            cf.EnablePayloads.Sketches,
		enableJSONToV1Intake:      cf.EnablePayloads.JsonToV1Intake,
		enableMetadata:            cf.EnablePayloads.Metadata,
		enableHostMetadata:        cf.EnablePayloads.HostMetadata,
		enableAgentchecksMetadata: cf.EnablePayloads.AgentchecksMetadata,
		//enableJSONStream:              stream.Available && cf.EnableStreamPayloadSerialization,
		//enableServiceChecksJSONStream: stream.Available && cf.EnableServiceChecksStreamPayloadSerialization,
		//enableEventsJSONStream:        stream.Available && cf.EnableEventsStreamPayloadSerialization,
		//enableSketchProtobufStream:    stream.Available && cf.EnableSketchStreamPayloadSerialization,
	}

	if !s.enableEvents {
		klog.Warning("event payloads are disabled: all events will be dropped")
	}
	if !s.enableSeries {
		klog.Warning("series payloads are disabled: all series will be dropped")
	}
	if !s.enableServiceChecks {
		klog.Warning("service_checks payloads are disabled: all service_checks will be dropped")
	}
	if !s.enableSketches {
		klog.Warning("sketches payloads are disabled: all sketches will be dropped")
	}
	if !s.enableJSONToV1Intake {
		klog.Warning("JSON to V1 intake is disabled: all payloads to that endpoint will be dropped")
	}
	if !s.enableMetadata {
		klog.Warning("metadata payloads are disabled: all metadata will be dropped")
	}
	if !s.enableHostMetadata {
		klog.Warning("host metadata payloads are disabled: all host metadata will be dropped")
	}
	if !s.enableAgentchecksMetadata {
		klog.Warning("agentchecks metadata payloads are disabled: all agentchecks metadata will be dropped")
	}

	return s
}

func (s Serializer) serializePayload(payload marshaler.Marshaler, compress bool, useV1API bool) (forwarder.Payloads, http.Header, error) {
	var marshalType split.MarshalType
	var extraHeaders http.Header

	if useV1API {
		marshalType = split.MarshalJSON
		if compress {
			extraHeaders = jsonExtraHeadersWithCompression
		} else {
			extraHeaders = jsonExtraHeaders
		}
	} else {
		marshalType = split.Marshal
		if compress {
			extraHeaders = protobufExtraHeadersWithCompression
		} else {
			extraHeaders = protobufExtraHeaders
		}
	}

	payloads, err := split.Payloads(payload, compress, marshalType)

	if err != nil {
		return nil, nil, fmt.Errorf("could not split payload into small enough chunks: %s", err)
	}

	return payloads, extraHeaders, nil
}

// SendEvents serializes a list of event and sends the payload to the forwarder
func (s *Serializer) SendEvents(e EventsMarshaler) error {
	if !s.enableEvents {
		klog.V(5).Info("events payloads are disabled: dropping it")
		return nil
	}

	var eventPayloads forwarder.Payloads
	var extraHeaders http.Header
	var err error

	eventPayloads, extraHeaders, err = s.serializePayload(e, true, false)
	if err != nil {
		return fmt.Errorf("dropping event payload: %s", err)
	}

	return s.Forwarder.SubmitEvents(eventPayloads, extraHeaders)
}

// SendServiceChecks serializes a list of serviceChecks and sends the payload to the forwarder
func (s *Serializer) SendServiceChecks(sc marshaler.Marshaler) error {
	if !s.enableServiceChecks {
		klog.V(5).Info("service_checks payloads are disabled: dropping it")
		return nil
	}

	var serviceCheckPayloads forwarder.Payloads
	var extraHeaders http.Header
	var err error

	serviceCheckPayloads, extraHeaders, err = s.serializePayload(sc, true, false)
	if err != nil {
		return fmt.Errorf("dropping service check payload: %s", err)
	}

	return s.Forwarder.SubmitServiceChecks(serviceCheckPayloads, extraHeaders)
}

// SendSeries serializes a list of serviceChecks and sends the payload to the forwarder
func (s *Serializer) SendSeries(series marshaler.Marshaler) error {
	if !s.enableSeries {
		klog.V(5).Info("series payloads are disabled: dropping it")
		return nil
	}

	var seriesPayloads forwarder.Payloads
	var extraHeaders http.Header
	var err error

	seriesPayloads, extraHeaders, err = s.serializePayload(series, true, false)

	if err != nil {
		return fmt.Errorf("dropping series payload: %s", err)
	}

	return s.Forwarder.SubmitSeries(seriesPayloads, extraHeaders)
}

// SendSummary serializes a list of SketSeriesList and sends the payload to the forwarder
func (s *Serializer) SendSummary(sketches marshaler.Marshaler) error {
	if !s.enableSketches {
		klog.V(5).Info("sketches payloads are disabled: dropping it")
		return nil
	}

	//if s.enableSketchProtobufStream {
	//	payloads, err := sketches.MarshalSplitCompress(marshaler.DefaultBufferContext())
	//	if err == nil {
	//		return s.Forwarder.SubmitSketchSeries(payloads, protobufExtraHeadersWithCompression)
	//	}
	//	klog.Warningf("Error: %v trying to stream compress SketchSeriesList - falling back to split/compress method", err)
	//}

	compress := true
	splitSketches, extraHeaders, err := s.serializePayload(sketches, compress, false)
	if err != nil {
		return fmt.Errorf("dropping sketch payload: %s", err)
	}

	return s.Forwarder.SubmitSketchSeries(splitSketches, extraHeaders)
}

// SendSketch serializes a list of SketSeriesList and sends the payload to the forwarder
func (s *Serializer) SendSketch(sketches marshaler.Marshaler) error {
	if !s.enableSketches {
		klog.V(5).Infof("sketches payloads are disabled: dropping it")
		return nil
	}

	//if s.enableSketchProtobufStream {
	//	payloads, err := sketches.MarshalSplitCompress(marshaler.DefaultBufferContext())
	//	if err == nil {
	//		return s.Forwarder.SubmitSketchSeries(payloads, protobufExtraHeadersWithCompression)
	//	}
	//	klog.Warningf("Error: %v trying to stream compress SketchSeriesList - falling back to split/compress method", err)
	//}

	compress := true
	splitSketches, extraHeaders, err := s.serializePayload(sketches, compress, false)
	if err != nil {
		return fmt.Errorf("dropping sketch payload: %s", err)
	}

	return s.Forwarder.SubmitSketchSeries(splitSketches, extraHeaders)
}

// SendMetadata serializes a metadata payload and sends it to the forwarder
func (s *Serializer) SendMetadata(m marshaler.Marshaler) error {
	if !s.enableMetadata {
		klog.V(5).Info("metadata payloads are disabled: dropping it")
		return nil
	}
	return s.sendMetadata(m, s.Forwarder.SubmitMetadata)
}

// SendHostMetadata serializes a metadata payload and sends it to the forwarder
func (s *Serializer) SendHostMetadata(m marshaler.Marshaler) error {
	if !s.enableHostMetadata {
		klog.V(5).Info("host metadata payloads are disabled: dropping it")
		return nil
	}
	return s.sendMetadata(m, s.Forwarder.SubmitHostMetadata)
}

// SendAgentchecksMetadata serializes a metadata payload and sends it to the forwarder
func (s *Serializer) SendAgentchecksMetadata(m marshaler.Marshaler) error {
	if !s.enableAgentchecksMetadata {
		klog.V(5).Info("agent metadata payloads are disabled: dropping it")
		return nil
	}
	return s.sendMetadata(m, s.Forwarder.SubmitAgentChecksMetadata)
}

func (s *Serializer) sendMetadata(m marshaler.Marshaler, submit func(payload forwarder.Payloads, extra http.Header) error) error {
	mustSplit, compressedPayload, payload, err := split.CheckSizeAndSerialize(m, true, split.MarshalJSON)
	if err != nil {
		return fmt.Errorf("could not determine size of metadata payload: %s", err)
	}

	klog.V(5).Infof("Sending metadata payload, content: %v", string(payload))

	if mustSplit {
		return fmt.Errorf("metadata payload was too big to send (%d bytes compressed, %d bytes uncompressed), metadata payloads cannot be split", len(compressedPayload), len(payload))
	}

	if err := submit(forwarder.Payloads{&compressedPayload}, jsonExtraHeadersWithCompression); err != nil {
		return err
	}
	klog.V(5).Infof("Sent metadata payload, size (raw/compressed): %d/%d bytes.", len(payload), len(compressedPayload))
	return nil
}

// SendJSONToV1Intake serializes a payload and sends it to the forwarder. Some code sends
// arbitrary payload the v1 API.
func (s *Serializer) SendJSONToV1Intake(data interface{}) error {
	if !s.enableJSONToV1Intake {
		klog.V(5).Info("JSON to V1 intake endpoint payloads are disabled: dropping it")
		return nil
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("could not serialize processes metadata payload: %s", err)
	}
	compressedPayload, err := compression.Compress(nil, payload)
	if err != nil {
		return fmt.Errorf("could not compress processes metadata payload: %s", err)
	}
	if err := s.Forwarder.SubmitV1Intake(forwarder.Payloads{&compressedPayload}, jsonExtraHeadersWithCompression); err != nil {
		return err
	}

	klog.Infof("Sent processes metadata payload, size: %d bytes.", len(payload))
	klog.V(5).Infof("Sent processes metadata payload, content: %v", string(payload))
	return nil
}

// SendOrchestratorMetadata serializes & send orchestrator metadata payloads
func (s *Serializer) SendOrchestratorMetadata(msgs []ProcessMessageBody, hostName, clusterID, payloadType string) error {
	if s.orchestratorForwarder == nil {
		return errors.New("orchestrator forwarder is not setup")
	}
	for _, m := range msgs {
		extraHeaders := make(http.Header)
		extraHeaders.Set(headers.HostHeader, hostName)
		extraHeaders.Set(headers.ClusterIDHeader, clusterID)
		extraHeaders.Set(headers.TimestampHeader, strconv.Itoa(int(time.Now().Unix())))

		body, err := processPayloadEncoder(m)
		if err != nil {
			return fmt.Errorf("Unable to encode message: %s", err)
		}

		payloads := forwarder.Payloads{&body}
		responses, err := s.orchestratorForwarder.SubmitOrchestratorChecks(payloads, extraHeaders, payloadType)
		if err != nil {
			return fmt.Errorf("Unable to submit payload: %s", err)
		}

		// Consume the responses so that writers to the channel do not become blocked
		// we don't need the bodies here though
		for range responses {

		}
	}
	return nil
}
