// +build linux,!linux_bpf ebpf_bindata

package probe

import (
	"fmt"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/ebpf/bytecode"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/security/config"
)

func getRuntimeCompiledProbe(config *config.Config, useSyscallWrapper bool) (bytecode.AssetReader, error) {
	return nil, fmt.Errorf("runtime compilation unsupported")
}
