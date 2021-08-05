// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build !windows

package file

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/decoder"
	"k8s.io/klog/v2"
)

// setup sets up the file tailer
func (t *Tailer) setup(offset int64, whence int) error {
	fullpath, err := filepath.Abs(t.file.Path)
	if err != nil {
		return err
	}
	t.fullpath = fullpath

	// adds metadata to enable users to filter logs by filename
	t.tags = t.buildTailerTags()

	klog.Infof("Opening %s for tailer key %s", t.file.Path, t.file.GetScanKey())
	f, err := openFile(fullpath)
	if err != nil {
		return err
	}

	t.osFile = f
	ret, _ := f.Seek(offset, whence)
	t.readOffset = ret
	t.decodedOffset = ret

	return nil
}

// read lets the tailer tail the content of a file
// until it is closed or the tailer is stopped.
func (t *Tailer) read() (int, error) {
	// keep reading data from file
	inBuf := make([]byte, 4096)
	n, err := t.osFile.Read(inBuf)
	if err != nil && err != io.EOF {
		// an unexpected error occurred, stop the tailor
		t.file.Source.Status.Error(err)
		return 0, fmt.Errorf("Unexpected error occurred while reading file: %s", err)
	}
	if n == 0 {
		return 0, nil
	}
	t.decoder.InputChan <- decoder.NewInput(inBuf[:n])
	t.incrementReadOffset(n)
	return n, nil
}