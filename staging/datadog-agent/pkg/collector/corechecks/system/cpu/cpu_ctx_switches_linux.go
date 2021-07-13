// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.
// +build linux

package cpu

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"
)

func readCtxSwitches(procStatPath string) (ctxSwitches int64, err error) {
	file, err := os.Open(procStatPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan(); i++ {
		txt := scanner.Text()
		if strings.HasPrefix(txt, "ctxt") {
			elemts := strings.Split(txt, " ")
			ctxSwitches, err = strconv.ParseInt(elemts[1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("%s in '%s' at line %d", err, procStatPath, i)
			}
			return ctxSwitches, nil
		}
	}

	return 0, fmt.Errorf("could not find the context switches in stat file")
}

func (c *Check) collectCtxSwitches(sender aggregator.Sender) error {
	procfsPath := "/proc"
	if config.C.ProcfsPath != "" {
		procfsPath = config.C.ProcfsPath
	}
	ctxSwitches, err := readCtxSwitches(filepath.Join(procfsPath, "/stat"))
	if err != nil {
		return err
	}
	sender.MonotonicCount("system.cpu.switches", float64(ctxSwitches), "", nil)
	return nil
}
