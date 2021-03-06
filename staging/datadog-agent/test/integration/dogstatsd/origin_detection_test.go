// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package dogstatsd

import (
	"testing"

	"github.com/n9e/n9e-agentd/pkg/config"
)

func TestUDSOriginDetection(t *testing.T) {
	config.SetupLogger(
		config.LoggerName("test"),
		"debug",
		"",
		"",
		false,
		true,
		false,
	)

	testUDSOriginDetection(t)
}
