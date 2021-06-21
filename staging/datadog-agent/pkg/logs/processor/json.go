// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package processor

import (
	"encoding/json"
	"time"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/message"
)

const nanoToMillis = 1000000

// JSONEncoder is a shared json encoder.
var JSONEncoder Encoder = &jsonEncoder{}

// jsonEncoder transforms a message into a JSON byte array.
type jsonEncoder struct{}

// JSON representation of a message.
type jsonPayload struct {
	Message   string `json:"message"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Hostname  string `json:"hostname"`
	Service   string `json:"service"`
	Source    string `json:"source"`
	Tags      string `json:"tags"`
	Ident     string `json:"ident"`
	Alias     string `json:"alias"`
}

// Encode encodes a message into a JSON byte array.
func (j *jsonEncoder) Encode(msg *message.Message, redactedMsg []byte) ([]byte, error) {
	ts := time.Now().UTC()
	if !msg.Timestamp.IsZero() {
		ts = msg.Timestamp
	}
	return json.Marshal(jsonPayload{
		Message:   toValidUtf8(redactedMsg),
		Status:    msg.GetStatus(),
		Timestamp: ts.UnixNano() / nanoToMillis,
		Hostname:  getHostname(),
		Service:   msg.Origin.Service(),
		Source:    msg.Origin.Source(),
		Tags:      msg.Origin.TagsToString(),
		Ident:     config.C.Ident,
		Alias:     config.C.Alias,
	})
}
