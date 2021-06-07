// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package common provides a set of common symbols needed by different packages,
// to avoid circular dependencies.
package common

// TransformationFunc type represents transformation applicable to byte slices
type TransformationFunc func(rawData []byte) ([]byte, error)

// ImportConfig imports the agent5 configuration into the agent6 yaml config
//func ImportConfig(oldConfigDir string, newConfigDir string, force bool) error {
//}

// Copy the src file to dst. File attributes won't be copied. Apply all TransformationFunc while copying.
//func copyFile(src, dst string, overwrite bool, transformations []TransformationFunc) error {
//}

// configTraceAgent extracts trace-agent specific info and dump to its own config file
//func configTraceAgent(datadogConfPath, traceAgentConfPath string, overwrite bool) (bool, error) {
//}

//func relocateMinCollectionInterval(rawData []byte) ([]byte, error) {
//}

//func insertMinCollectionInterval(rawData map[interface{}]interface{}, interval int) {
//}
