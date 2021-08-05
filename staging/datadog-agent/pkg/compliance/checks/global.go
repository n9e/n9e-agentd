// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package checks

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/compliance/eval"
)

var (
	// TODO: Add global functions for string manipulation
	globalFunctions = eval.FunctionMap{}

	globalVars = eval.VarMap{}

	globalInstance = &eval.Instance{
		Vars:      globalVars,
		Functions: globalFunctions,
	}
)