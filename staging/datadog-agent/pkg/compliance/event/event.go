// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package event

const (
	// Passed is used to report successful result of a rule check (condition passed)
	Passed = "passed"
	// Failed is used to report unsuccessful result of a rule check (condition failed)
	Failed = "failed"
	// Error is used to report result of a rule check that resulted in an error (unable to evaluate condition)
	Error = "error"
)

// Data defines a key value map for storing attributes of a reported rule event
type Data map[string]interface{}

// Event describes a log event sent for an evaluated compliance/security rule.
type Event struct {
	AgentRuleID      string      `json:"agentRuleId,omitempty"`
	AgentRuleVersion int         `json:"agentRuleVersion,omitempty"`
	Result           string      `json:"result,omitempty"`
	ResourceType     string      `json:"resourceType,omitempty"`
	ResourceID       string      `json:"resourceId,omitempty"`
	Tags             []string    `json:"tags"`
	Data             interface{} `json:"data,omitempty"`
}
