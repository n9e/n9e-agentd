// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package agent implements the api endpoints for the `/agent` prefix.
// This group of endpoints is meant to provide high-level functionalities
// at the agent level.

// +build !jmx

package apiserver

import (
	"errors"
	"net/http"
)

var noJMXErrorString = errors.New("jmx is not compiled in this agent")

func getJMXConfigs(w http.ResponseWriter, r *http.Request) error {
	return noJMXErrorString
}

func setJMXStatus(w http.ResponseWriter, r *http.Request) error {
	return noJMXErrorString
}
