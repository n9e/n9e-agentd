package all

import (
	// plugin checks
	_ "github.com/n9e/n9e-agentd/plugins/log"
	_ "github.com/n9e/n9e-agentd/plugins/mysql"
	_ "github.com/n9e/n9e-agentd/plugins/port"
	_ "github.com/n9e/n9e-agentd/plugins/process"
	_ "github.com/n9e/n9e-agentd/plugins/prometheus"
	_ "github.com/n9e/n9e-agentd/plugins/script"

	// register core checks
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/cluster/ksm"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/cluster/kubernetesapiserver"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/cluster/orchestrator"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/containers/containerd"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/containers/cri"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/containers/docker"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/ebpf"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/embed"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/net"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/nvidia/jetson"
	_ "github.com/n9e/n9e-agentd/plugins/snmp"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/system"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/system/cpu"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/system/disk"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/system/filehandles"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/system/memory"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/system/uptime"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/system/winproc"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks/systemd"

	// register metadata providers
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/metadata"
	_ "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metadata"
)
