// +build linux

package checks

import (
	"time"

	"github.com/n9e/n9e-agentd/pkg/process/procutil"
)

func getAllProcesses(probe *procutil.Probe) (map[int32]*procutil.Process, error) {
	return probe.ProcessesByPID(time.Now())
}

func getAllProcStats(probe *procutil.Probe, pids []int32) (map[int32]*procutil.Stats, error) {
	return probe.StatsForPIDs(pids, time.Now())
}
