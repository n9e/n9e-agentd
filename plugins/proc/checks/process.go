package checks

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DataDog/datadog-agent/pkg/process/config"
	"github.com/DataDog/datadog-agent/pkg/process/net"
	"github.com/DataDog/datadog-agent/pkg/process/procutil"
	"github.com/DataDog/datadog-agent/pkg/process/util"
	agentutil "github.com/DataDog/datadog-agent/pkg/util"
	"github.com/DataDog/datadog-agent/pkg/util/containers"
	"github.com/DataDog/gopsutil/cpu"
	model "github.com/n9e/agent-payload/process"
	"k8s.io/klog/v2"
)

const emptyCtrID = ""

var (
	// Process is a singleton ProcessCheck.
	Process         = &ProcessCheck{probe: procutil.NewProcessProbe()}
	errEmptyCPUTime = errors.New("empty CPU time information returned")
)

type ProcessFilter struct {
	Target        string  `json:"target"`
	CollectMethod string  `json:"collect_method" description:"name or cmdline"`
	Pids          []int32 `json:"-"`
}

func (f *ProcessFilter) Match(p *procutil.Process) bool {
	switch f.CollectMethod {
	case "name":
		return p.Name == f.Target
	case "cmdline":
		return strings.Contains(strings.Join(p.Cmdline, " "), f.Target)
	}
	return false
}

// ProcessCheck collects full state, including cmdline args and related metadata,
// for live and running processes. The instance will store some state between
// checks that will be used for rates, cpu calculations, etc.
type ProcessCheck struct {
	sync.RWMutex

	probe                  *procutil.Probe
	filters                []*ProcessFilter
	sysInfo                *model.SystemInfo
	lastCPUTime            cpu.TimesStat
	lastProcs              map[int32]*procutil.Process
	lastRun                time.Time
	networkID              string
	notInitializedLogLimit *util.LogLimit
	lastPIDs               atomic.Value // will be reused by RTProcessCheck to get stats
	procsByPID             map[int32]*model.Process
}

func (p *ProcessCheck) Filters() []*ProcessFilter {
	p.RLock()
	defer p.RUnlock()
	return p.filters
}

func (p *ProcessCheck) AddFilter(filter *ProcessFilter) {
	p.Lock()
	defer p.Unlock()
	p.filters = append(p.filters, filter)
}

func (p *ProcessCheck) DelFilter(filter *ProcessFilter) {
	p.Lock()
	defer p.Unlock()
	for i, v := range p.filters {
		if v == filter {
			p.filters[i] = p.filters[len(p.filters)-1]
			p.filters = p.filters[:len(p.filters)-1]
		}
	}
}

func (p *ProcessCheck) GetProcs() map[int32]*model.Process {
	p.RLock()
	defer p.RUnlock()

	return p.procsByPID
}

// Init initializes the singleton ProcessCheck.
func (p *ProcessCheck) Init(_ *config.AgentConfig, info *model.SystemInfo) {
	p.sysInfo = info
	p.notInitializedLogLimit = util.NewLogLimit(1, time.Minute*10)

	networkID, err := agentutil.GetNetworkID(context.Background())
	if err != nil {
		klog.Infof("no network ID detected: %s", err)
	}
	p.networkID = networkID
}

// Name returns the name of the ProcessCheck.
func (p *ProcessCheck) Name() string { return config.ProcessCheckName }

// RealTime indicates if this check only runs in real-time mode.
func (p *ProcessCheck) RealTime() bool { return false }

// Run runs the ProcessCheck to collect a list of running processes and relevant
// stats for each. On most POSIX systems this will use a mix of procfs and other
// OS-specific APIs to collect this information. The bulk of this collection is
// abstracted into the `gopsutil` library.
// Processes are split up into a chunks of at most 100 processes per message to
// limit the message size on intake.
// See agent.proto for the schema of the message and models used.
func (p *ProcessCheck) Run(cfg *config.AgentConfig, groupID int32) ([]model.MessageBody, error) {
	p.Lock()
	defer p.Unlock()

	start := time.Now()
	cpuTimes, err := cpu.Times(false)
	if err != nil {
		return nil, err
	}
	if len(cpuTimes) == 0 {
		return nil, errEmptyCPUTime
	}

	var sysProbeUtil *net.RemoteSysProbeUtil
	procutil.WithPermission(true)(p.probe)

	ps, err := getAllProcesses(p.probe)
	if err != nil {
		return nil, err
	}
	procs := map[int32]*procutil.Process{}
	for _, filter := range p.filters {
		filter.Pids = filter.Pids[:0]
		for pid, p := range ps {
			if filter.Match(p) {
				procs[pid] = p
				filter.Pids = append(filter.Pids, pid)
			}
		}
	}

	// stores lastPIDs to be used by RTProcess
	lastPIDs := make([]int32, 0, len(procs))
	for pid := range procs {
		lastPIDs = append(lastPIDs, pid)
	}
	p.lastPIDs.Store(lastPIDs)

	if sysProbeUtil != nil {
		mergeProcWithSysprobeStats(lastPIDs, procs, sysProbeUtil)
	}

	//ctrList, _ := util.GetContainers()

	// Keep track of containers addresses
	//LocalResolver.LoadAddrs(ctrList)

	//ctrByProc := ctrIDForPID(ctrList)
	// End check early if this is our first run.
	if p.lastProcs == nil {
		p.lastProcs = procs
		p.lastCPUTime = cpuTimes[0]
		//p.lastCtrRates = util.ExtractContainerRateMetric(ctrList)
		//p.lastCtrIDForPID = ctrByProc
		p.lastRun = time.Now()
		return nil, nil
	}

	connsByPID := Connections.getLastConnectionsByPID()
	procsByPID := fmtProcesses(cfg, procs, p.lastProcs, nil, cpuTimes[0], p.lastCPUTime, p.lastRun, connsByPID)

	//ctrs := fmtContainers(ctrList, p.lastCtrRates, p.lastRun)

	// Store the last state for comparison on the next run.
	// Note: not storing the filtered in case there are new processes that haven't had a chance to show up twice.
	p.lastProcs = procs
	//p.lastCtrRates = util.ExtractContainerRateMetric(ctrList)
	p.lastCPUTime = cpuTimes[0]
	p.lastRun = time.Now()
	//p.lastCtrIDForPID = ctrByProc
	p.procsByPID = procsByPID

	//statsd.Client.Gauge("process.containers.host_count", float64(totalContainers), []string{}, 1) //nolint:errcheck
	//statsd.Client.Gauge("process.processes.host_count", float64(totalProcs), []string{}, 1)       //nolint:errcheck
	klog.V(5).Infof("collected processes in %s", time.Now().Sub(start))
	return nil, nil
}

// GetLastPIDs returns the lastPIDs as []int32 slice
func (p *ProcessCheck) GetLastPIDs() []int32 {
	if result := p.lastPIDs.Load(); result != nil {
		return result.([]int32)
	}
	return nil
}

// chunkProcesses split non-container processes into chunks and return a list of chunks
func chunkProcesses(procs []*model.Process, size int) [][]*model.Process {
	chunkCount := len(procs) / size
	if chunkCount*size < len(procs) {
		chunkCount++
	}
	chunks := make([][]*model.Process, 0, chunkCount)

	for i := 0; i < len(procs); i += size {
		end := i + size
		if end > len(procs) {
			end = len(procs)
		}
		chunks = append(chunks, procs[i:end])
	}

	return chunks
}

func ctrIDForPID(ctrList []*containers.Container) map[int32]string {
	ctrIDForPID := make(map[int32]string, len(ctrList))
	for _, c := range ctrList {
		for _, p := range c.Pids {
			ctrIDForPID[p] = c.ID
		}
	}
	return ctrIDForPID
}

// fmtProcesses goes through each process, converts them to process object and group them by containers
// non-container processes would be in a single group with key as empty string ""
func fmtProcesses(
	cfg *config.AgentConfig,
	procs, lastProcs map[int32]*procutil.Process,
	ctrByProc map[int32]string,
	syst2, syst1 cpu.TimesStat,
	lastRun time.Time,
	connsByPID map[int32][]*model.Connection,
) map[int32]*model.Process {
	procsByPID := make(map[int32]*model.Process)
	connCheckIntervalS := int(cfg.CheckIntervals[config.ConnectionsCheckName] / time.Second)

	for _, fp := range procs {
		if skipProcess(cfg, fp, lastProcs) {
			continue
		}

		// Hide blacklisted args if the Scrubber is enabled
		fp.Cmdline = cfg.Scrubber.ScrubProcessCommand(fp)

		proc := &model.Process{
			Pid:                    fp.Pid,
			NsPid:                  fp.NsPid,
			Command:                formatCommand(fp),
			User:                   formatUser(fp),
			Memory:                 formatMemory(fp.Stats),
			Cpu:                    formatCPU(fp.Stats, fp.Stats.CPUTime, lastProcs[fp.Pid].Stats.CPUTime, syst2, syst1),
			CreateTime:             fp.Stats.CreateTime,
			OpenFdCount:            fp.Stats.OpenFdCount,
			State:                  model.ProcessState(model.ProcessState_value[fp.Stats.Status]),
			IoStat:                 formatIO(fp.Stats, lastProcs[fp.Pid].Stats.IOStat, lastRun),
			VoluntaryCtxSwitches:   uint64(fp.Stats.CtxSwitches.Voluntary),
			InvoluntaryCtxSwitches: uint64(fp.Stats.CtxSwitches.Involuntary),
			ContainerId:            ctrByProc[fp.Pid],
			Networks:               formatNetworks(connsByPID[fp.Pid], connCheckIntervalS),
		}
		procsByPID[proc.Pid] = proc
	}

	cfg.Scrubber.IncrementCacheAge()

	return procsByPID
}

func formatCommand(fp *procutil.Process) *model.Command {
	return &model.Command{
		Args:   fp.Cmdline,
		Cwd:    fp.Cwd,
		Root:   "",    // TODO
		OnDisk: false, // TODO
		Ppid:   fp.Ppid,
		Exe:    fp.Exe,
	}
}

func formatIO(fp *procutil.Stats, lastIO *procutil.IOCountersStat, before time.Time) *model.IOStat {
	// This will be nil for Mac
	if fp.IOStat == nil {
		return &model.IOStat{}
	}

	diff := time.Now().Unix() - before.Unix()
	if before.IsZero() || diff <= 0 {
		return &model.IOStat{}
	}
	// Reading -1 as counter means the file could not be opened due to permissions.
	// In that case we set the rate as -1 to distinguish from a real 0 in rates.
	readRate := float32(-1)
	if fp.IOStat.ReadCount >= 0 {
		readRate = calculateRate(uint64(fp.IOStat.ReadCount), uint64(lastIO.ReadCount), before)
	}
	writeRate := float32(-1)
	if fp.IOStat.WriteCount >= 0 {
		writeRate = calculateRate(uint64(fp.IOStat.WriteCount), uint64(lastIO.WriteCount), before)
	}
	readBytesRate := float32(-1)
	if fp.IOStat.ReadBytes >= 0 {
		readBytesRate = calculateRate(uint64(fp.IOStat.ReadBytes), uint64(lastIO.ReadBytes), before)
	}
	writeBytesRate := float32(-1)
	if fp.IOStat.WriteBytes >= 0 {
		writeBytesRate = calculateRate(uint64(fp.IOStat.WriteBytes), uint64(lastIO.WriteBytes), before)
	}
	return &model.IOStat{
		ReadRate:       readRate,
		WriteRate:      writeRate,
		ReadBytesRate:  readBytesRate,
		WriteBytesRate: writeBytesRate,
	}
}

func formatMemory(fp *procutil.Stats) *model.MemoryStat {
	ms := &model.MemoryStat{
		Rss:  fp.MemInfo.RSS,
		Vms:  fp.MemInfo.VMS,
		Swap: fp.MemInfo.Swap,
	}

	if fp.MemInfoEx != nil {
		ms.Shared = fp.MemInfoEx.Shared
		ms.Text = fp.MemInfoEx.Text
		ms.Lib = fp.MemInfoEx.Lib
		ms.Data = fp.MemInfoEx.Data
		ms.Dirty = fp.MemInfoEx.Dirty
	}
	return ms
}

func formatNetworks(conns []*model.Connection, interval int) *model.ProcessNetworks {
	connRate := float32(len(conns)) / float32(interval)
	totalTraffic := uint64(0)
	for _, conn := range conns {
		totalTraffic += conn.LastBytesSent + conn.LastBytesReceived
	}
	bytesRate := float32(totalTraffic) / float32(interval)
	return &model.ProcessNetworks{ConnectionRate: connRate, BytesRate: bytesRate}
}

// skipProcess will skip a given process if it's blacklisted or hasn't existed
// for multiple collections.
func skipProcess(
	cfg *config.AgentConfig,
	fp *procutil.Process,
	lastProcs map[int32]*procutil.Process,
) bool {
	if len(fp.Cmdline) == 0 {
		return true
	}
	if config.IsBlacklisted(fp.Cmdline, cfg.Blacklist) {
		return true
	}
	if _, ok := lastProcs[fp.Pid]; !ok {
		// Skipping any processes that didn't exist in the previous run.
		// This means short-lived processes (<2s) will never be captured.
		return true
	}
	return false
}

func (p *ProcessCheck) createTimesforPIDs(pids []int32) map[int32]int64 {
	p.RLock()
	defer p.RUnlock()

	createTimeForPID := make(map[int32]int64)
	for _, pid := range pids {
		if p, ok := p.lastProcs[pid]; ok {
			createTimeForPID[pid] = p.Stats.CreateTime
		}
	}
	return createTimeForPID
}

// mergeProcWithSysprobeStats takes a process by PID map and fill the stats from system probe into the processes in the map
func mergeProcWithSysprobeStats(pids []int32, procs map[int32]*procutil.Process, pu *net.RemoteSysProbeUtil) {
	pStats, err := pu.GetProcStats(pids)
	if err == nil {
		for pid, proc := range procs {
			if s, ok := pStats.StatsByPID[pid]; ok {
				proc.Stats.OpenFdCount = s.OpenFDCount
				proc.Stats.IOStat.ReadCount = s.ReadCount
				proc.Stats.IOStat.WriteCount = s.WriteCount
				proc.Stats.IOStat.ReadBytes = s.ReadBytes
				proc.Stats.IOStat.WriteBytes = s.WriteBytes
			}
		}
	} else {
		klog.V(5).Infof("cannot do GetProcStats from system-probe for process check: %s", err)
	}
}
