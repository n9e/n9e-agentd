package ebpf

import (
	"strings"
)

const (
	spNS = "system_probe_config"
)

// Config stores all common flags used by system-probe
type Config struct {
	// BPFDebug enables bpf debug logs
	BPFDebug bool

	// BPFDir is the directory to load the eBPF program from
	BPFDir string

	// ExcludedBPFLinuxVersions lists Linux kernel versions that should not use BPF features
	ExcludedBPFLinuxVersions []string

	// ProcRoot is the root path to the proc filesystem
	ProcRoot string

	// EnableTracepoints enables use of tracepoints instead of kprobes for probing syscalls (if available on system)
	EnableTracepoints bool

	// EnableRuntimeCompiler enables the use of the embedded compiler to build eBPF programs on-host
	EnableRuntimeCompiler bool

	// KernelHeadersDir is the directories of the kernel headers to use for runtime compilation
	KernelHeadersDirs []string

	// KernelHeadersDownloadDir is the directory where the system-probe will attempt to download kernel headers, if necessary
	KernelHeadersDownloadDir string

	// RuntimeCompilerOutputDir is the directory where the runtime compiler will store compiled programs
	RuntimeCompilerOutputDir string

	// AllowPrecompiledFallback indicates whether we are allowed to fallback to the prebuilt probes if runtime compilation fails.
	AllowPrecompiledFallback bool
}

func key(pieces ...string) string {
	return strings.Join(pieces, ".")
}

// NewConfig creates a config with ebpf-related settings
func NewConfig() *Config {
	return nil
	//cfg := aconfig.C
	//aconfig.InitSystemProbeConfig(cfg)

	//return &Config{
	//	BPFDebug:                 cfg.BPFDebug,
	//	BPFDir:                   cfg.BPFDir,
	//	ExcludedBPFLinuxVersions: cfg.SystemProbe.ExcludedLinuxVersions,
	//	EnableTracepoints:        cfg.SystemProbe.EnableTracepoints,
	//	ProcRoot:                 util.GetProcRoot(),

	//	EnableRuntimeCompiler:    cfg.SystemProbe.EnableRuntimeCompiler,
	//	RuntimeCompilerOutputDir: cfg.SystemProbe.RuntimeCompilerOutputDir,
	//	KernelHeadersDirs:        cfg.SystemProbe.KernelHeaderDirs,
	//	KernelHeadersDownloadDir: cfg.SystemProb.KernelHeadersDownloadDir,
	//	AllowPrecompiledFallback: true,
	//}
}
