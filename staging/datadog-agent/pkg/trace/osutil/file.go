// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package osutil

import (
	"fmt"
	"os"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/flags"
	"k8s.io/klog/v2"
)

// Exists reports whether the given path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Exit prints the message and exits the program with status code 1.
func Exit(msg string) {
	if flags.Info || flags.Version {
		fmt.Println(msg)
	} else {
		klog.Error(msg)
		log.Flush()
	}
	os.Exit(1)
}

// Exitf prints the formatted text and exits the program with status code 1.
func Exitf(format string, args ...interface{}) {
	if flags.Info || flags.Version {
		fmt.Printf(format, args...)
		fmt.Println("")
	} else {
		klog.Warningf(format, args...)
		log.Flush()
	}
	os.Exit(1)
}
