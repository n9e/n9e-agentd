module github.com/n9e/n9e-agentd

go 1.16

replace (
	github.com/DataDog/datadog-agent => ../datadog-agent
	github.com/DataDog/datadog-agent/pkg/util/log => ../datadog-agent/pkg/util/log
	github.com/DataDog/datadog-agent/pkg/util/winutil => ../datadog-agent/pkg/util/winutil
	github.com/n9e/agent-payload => ../agent-payload
	github.com/yubo/apiserver => ../apiserver
	github.com/yubo/golib => ../golib
)

// NOTE: Prefer using simple `require` directives instead of using `replace` if possible.
// See https://github.com/DataDog/datadog-agent/blob/main/docs/dev/gomodreplace.md
// for more details.

// Internal deps fix version
replace (
	github.com/cihub/seelog => github.com/cihub/seelog v0.0.0-20151216151435-d2c6e5aa9fbf // v2.6
	github.com/coreos/go-systemd => github.com/coreos/go-systemd v0.0.0-20180202092358-40e2722dffea
	github.com/docker/distribution => github.com/docker/distribution v2.7.1-0.20190104202606-0ac367fd6bee+incompatible
	github.com/florianl/go-conntrack => github.com/florianl/go-conntrack v0.2.0
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	github.com/iovisor/gobpf => github.com/DataDog/gobpf v0.0.0-20210322155958-9866ef4cd22c
	github.com/lxn/walk => github.com/lxn/walk v0.0.0-20180521183810-02935bac0ab8
	github.com/mholt/archiver => github.com/mholt/archiver v2.0.1-0.20171012052341-26cf5bb32d07+incompatible
	github.com/spf13/cast => github.com/DataDog/cast v1.3.1-0.20190301154711-1ee8c8bd14a3
	github.com/ugorji/go => github.com/ugorji/go v1.1.7
)

// pinned to grpc v1.27.0
replace (
	github.com/grpc-ecosystem/grpc-gateway => github.com/grpc-ecosystem/grpc-gateway v1.12.2
	google.golang.org/grpc => github.com/grpc/grpc-go v1.27.0
)

require (
	github.com/DataDog/datadog-agent v0.0.0-20210730134932-2365d4a4f838
	github.com/DataDog/datadog-agent/pkg/util/log v0.30.0-rc.7
	github.com/DataDog/datadog-agent/pkg/util/winutil v0.30.0-rc.7
	github.com/DataDog/datadog-go v4.8.0+incompatible
	github.com/DataDog/gohai v0.0.0-20210303102637-6b668acb50dd
	github.com/DataDog/gopsutil v0.0.0-20200624212600-1b53412ef321
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575
	github.com/coreos/etcd v3.3.25+incompatible // indirect
	github.com/fatih/color v1.12.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/gosnmp/gosnmp v1.32.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369
	github.com/n9e/agent-payload v4.71.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.23.0
	github.com/shirou/gopsutil v3.21.5+incompatible
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4
	github.com/spf13/afero v1.6.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635
	github.com/yubo/apiserver v0.0.0-00010101000000-000000000000
	github.com/yubo/golib v0.0.0-20210729083123-4040286093c6
	go.etcd.io/etcd v3.3.25+incompatible // indirect
	go.uber.org/automaxprocs v1.4.0
	golang.org/x/mobile v0.0.0-20201217150744-e6ae53a27f4f
	golang.org/x/sys v0.0.0-20210601080250-7ecdf8ef093b
	golang.org/x/text v0.3.6
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c
	google.golang.org/grpc v1.38.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/component-base v0.20.5
	k8s.io/klog/v2 v2.9.0
	k8s.io/kubectl v0.0.0
	sigs.k8s.io/yaml v1.2.0
)

// Pinned to kubernetes-v0.20.5
replace (
	k8s.io/api => k8s.io/api v0.20.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.5
	k8s.io/apiserver => k8s.io/apiserver v0.20.5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.5
	k8s.io/client-go => k8s.io/client-go v0.20.5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.5
	k8s.io/code-generator => k8s.io/code-generator v0.20.5
	k8s.io/component-base => k8s.io/component-base v0.20.5
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.5
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.5
	k8s.io/cri-api => k8s.io/cri-api v0.20.5
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.5
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.5
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.5
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.5
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.5
	k8s.io/kubectl => k8s.io/kubectl v0.20.5
	k8s.io/kubelet => k8s.io/kubelet v0.20.5
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.5
	k8s.io/metrics => k8s.io/metrics v0.20.5
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.3-rc.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.5
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.20.5
	k8s.io/sample-controller => k8s.io/sample-controller v0.20.5
)

replace gopkg.in/DataDog/dd-trace-go.v1 => gopkg.in/DataDog/dd-trace-go.v1 v1.30.0

// Remove once the PR kubernetes/kube-state-metrics#1516 is merged and released.
replace k8s.io/kube-state-metrics/v2 => github.com/ahmed-mez/kube-state-metrics/v2 v2.1.0-rc.0.0.20210629115837-e46f17606d22

replace github.com/aptly-dev/aptly => github.com/lebauce/aptly v0.7.2-0.20210723103859-345a32860f4d
