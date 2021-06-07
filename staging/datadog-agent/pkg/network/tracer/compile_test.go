// +build linux_bpf

package tracer

import (
	"testing"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/ebpf/bytecode/runtime"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/network/config"
	"github.com/stretchr/testify/require"
)

func TestTracerCompile(t *testing.T) {
	cfg := config.New()
	cfg.BPFDebug = true
	cflags := getCFlags(cfg)
	_, err := runtime.Tracer.Compile(&cfg.Config, cflags)
	require.NoError(t, err)
}

func TestConntrackCompile(t *testing.T) {
	cfg := config.New()
	cfg.BPFDebug = true
	cflags := getCFlags(cfg)
	_, err := runtime.Conntrack.Compile(&cfg.Config, cflags)
	require.NoError(t, err)
}
