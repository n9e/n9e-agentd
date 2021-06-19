package ebpf

import (
	"strings"

	aconfig "github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/process/util"
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
	cf := aconfig.C.SystemProbe

	return &Config{
		BPFDebug:                 cf.BPFDebug,
		BPFDir:                   cf.BPFDir,
		ExcludedBPFLinuxVersions: cf.ExcludedLinuxVersions,
		EnableTracepoints:        cf.EnableTracepoints,
		ProcRoot:                 util.GetProcRoot(),

		EnableRuntimeCompiler:    cf.EnableRuntimeCompiler,
		RuntimeCompilerOutputDir: cf.RuntimeCompilerOutputDir,
		KernelHeadersDirs:        cf.KernelHeaderDirs,
		AllowPrecompiledFallback: true,
	}
}
