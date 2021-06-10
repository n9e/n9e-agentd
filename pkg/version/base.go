// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package version

import (
	"strings"

	"github.com/n9e/n9e-agentd/pkg/options"
)

// AgentVersion contains the version of the Agent
var AgentVersion string

// Commit is populated with the short commit hash from which the Agent was built
var Commit string

func init() {
	AgentVersion = strings.TrimPrefix(options.Version, "v")
	Commit = options.Revision
}
