module github.com/n9e/n9e-agentd

go 1.16

require (
	code.cloudfoundry.org/bbs v0.0.0-20200403215808-d7bc971db0db
	code.cloudfoundry.org/cfhttp/v2 v2.0.0 // indirect
	code.cloudfoundry.org/clock v1.0.0 // indirect
	code.cloudfoundry.org/consuladapter v0.0.0-20200131002136-ac1daf48ba97 // indirect
	code.cloudfoundry.org/diego-logging-client v0.0.0-20200130234554-60ef08820a45 // indirect
	code.cloudfoundry.org/executor v0.0.0-20200218194701-024d0bdd52d4 // indirect
	code.cloudfoundry.org/garden v0.0.0-20210208153517-580cadd489d2
	code.cloudfoundry.org/go-diodes v0.0.0-20190809170250-f77fb823c7ee // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible // indirect
	code.cloudfoundry.org/lager v2.0.0+incompatible
	code.cloudfoundry.org/locket v0.0.0-20200131001124-67fd0a0fdf2d // indirect
	code.cloudfoundry.org/rep v0.0.0-20200325195957-1404b978e31e // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20180905210152-236a6d29298a // indirect
	code.cloudfoundry.org/tlsconfig v0.0.0-20200131000646-bbe0f8da39b3 // indirect
	github.com/DataDog/datadog-go v4.5.0+incompatible
	github.com/DataDog/datadog-operator v0.3.1
	github.com/DataDog/ebpf v0.0.0-20210407211251-e1ec0c7b627d
	github.com/DataDog/gohai v0.0.0-20210303102637-6b668acb50dd
	github.com/DataDog/gopsutil v0.0.0-20200624212600-1b53412ef321
	github.com/DataDog/sketches-go v1.0.0
	github.com/DataDog/watermarkpodautoscaler v0.2.1-0.20210209165213-28eb1cc35a8d
	github.com/DataDog/zstd_0 v0.0.0-20210310093942-586c1286621f
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Microsoft/go-winio v0.4.15
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/alecthomas/participle v0.4.4
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1
	github.com/andybalholm/brotli v1.0.1 // indirect
	github.com/avast/retry-go v2.7.0+incompatible
	github.com/aws/aws-sdk-go v1.38.45
	github.com/beevik/ntp v0.3.0
	github.com/benesch/cgosymbolizer v0.0.0-20190515212042-bec6fe6e597b
	github.com/bhmj/jsonslice v0.0.0-20200323023432-92c3edaad8e2
	github.com/blabber/go-freebsd-sysctl v0.0.0-20201130114544-503969f39d8f
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575
	github.com/clbanning/mxj v1.8.4
	github.com/cloudfoundry-community/go-cfclient v0.0.0-20201123235753-4f46d6348a05
	github.com/cobaugh/osrelease v0.0.0-20181218015638-a93a0a55a249
	github.com/containerd/cgroups v0.0.0-20190919134610-bf292b21730f
	github.com/containerd/containerd v1.3.10
	github.com/containerd/fifo v0.0.0-20191213151349-ff969a566b00 // indirect
	github.com/containerd/typeurl v1.0.0
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/davecgh/go-spew v1.1.1
	github.com/dgraph-io/ristretto v0.0.3
	github.com/docker/docker v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/dustin/go-humanize v1.0.0
	github.com/elastic/go-libaudit v0.4.0
	github.com/fatih/color v1.9.0
	github.com/fatih/structtag v1.2.0
	github.com/florianl/go-conntrack v0.1.1-0.20191002182014-06743d3a59db
	github.com/freddierice/go-losetup v0.0.0-20170407175016-fc9adea44124
	github.com/go-ole/go-ole v1.2.5
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-test/deep v1.0.5 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/gogo/googleapis v1.3.2 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/gopacket v1.1.17
	github.com/google/pprof v0.0.0-20201117184057-ae444373da19
	github.com/gorilla/mux v1.7.4
	github.com/gosnmp/gosnmp v1.29.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.5
	github.com/hashicorp/consul/api v1.4.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95
	github.com/ianlancetaylor/cgosymbolizer v0.0.0-20201204192058-7acc97e53614 // indirect
	github.com/iceber/iouring-go v0.0.0-20201110085921-73520a520aac
	github.com/iovisor/gobpf v0.0.0
	github.com/itchyny/gojq v0.10.2
	github.com/json-iterator/go v1.1.11
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/klauspost/compress v1.11.12 // indirect
	github.com/klauspost/pgzip v1.2.5 // indirect
	github.com/kubernetes-incubator/custom-metrics-apiserver v0.0.0-20190918110929-3d9be26a50eb // Pinned to kubernetes-1.16.2
	github.com/mailru/easyjson v0.7.6
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369
	github.com/mdlayher/netlink v1.4.0
	github.com/mholt/archiver/v3 v3.5.0
	github.com/miekg/dns v1.1.31
	github.com/moby/sys/mountinfo v0.4.0
	github.com/n9e/agent-payload v0.0.0-20210619031503-b72325474651
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852 // indirect
	github.com/opencontainers/runc v1.0.0-rc2.0.20190611121236-6cc515888830 // indirect
	github.com/opencontainers/runtime-spec v1.0.2
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.3 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.18.0
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
	github.com/shirou/gopsutil v3.21.2+incompatible
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4
	github.com/shuLhan/go-bindata v3.6.1+incompatible
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/soniah/gosnmp v1.26.0
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2
	github.com/tedsuo/ifrit v0.0.0-20191009134036-9a97d0632f00 // indirect
	github.com/tinylib/msgp v1.1.2
	github.com/tklauser/go-sysconf v0.3.4 // indirect
	github.com/ttacon/chalk v0.0.0-20160626202418-22c06c80ed31 // indirect
	github.com/twmb/murmur3 v1.1.3
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/vishvananda/netlink v1.1.1-0.20201206203632-88079d98e65d
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f
	github.com/vito/go-sse v1.0.0 // indirect
	github.com/vmihailenco/msgpack/v4 v4.3.11
	github.com/yubo/apiserver v0.0.0-20210729120429-544a148d5f22
	github.com/yubo/golib v0.0.0-20210729083123-4040286093c6
	github.com/zorkian/go-datadog-api v2.28.0+incompatible // indirect
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	go.uber.org/automaxprocs v1.2.0
	golang.org/x/crypto v0.0.0-20210503195802-e9a32991a82e
	golang.org/x/mobile v0.0.0-20201217150744-e6ae53a27f4f
	golang.org/x/mod v0.3.1-0.20200828183125-ce943fd02449 // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	golang.org/x/oauth2 v0.0.0-20210622215436-a8dc77f794b6 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22
	golang.org/x/text v0.3.6
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	golang.org/x/tools v0.1.0
	gomodules.xyz/jsonpatch/v3 v3.0.1
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210303154014-9728d6b83eeb
	google.golang.org/grpc v1.37.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.23.1
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gopkg.in/zorkian/go-datadog-api.v2 v2.29.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.21.3
	k8s.io/autoscaler/vertical-pod-autoscaler v0.9.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/component-base v0.18.3
	k8s.io/cri-api v0.0.0
	k8s.io/klog v1.0.1-0.20200310124935-4ad0115ba9e4 // indirect; Min version that includes fix for Windows Nano
	k8s.io/klog/v2 v2.9.0
	k8s.io/kube-state-metrics v1.8.1-0.20200108124505-369470d6ead8
	k8s.io/kubectl v0.18.2
	k8s.io/kubernetes v1.18.2
	k8s.io/metrics v0.18.3
	sigs.k8s.io/yaml v1.2.0
)

replace (
	#github.com/yubo/apiserver => ../../yubo/apiserver
	github.com/cihub/seelog => github.com/cihub/seelog v0.0.0-20151216151435-d2c6e5aa9fbf // v2.6
	github.com/containerd/cgroups => github.com/containerd/cgroups v0.0.0-20200327175542-b44481373989
	github.com/containerd/containerd => github.com/containerd/containerd v1.2.13
	github.com/coreos/go-systemd => github.com/coreos/go-systemd v0.0.0-20180202092358-40e2722dffea
	github.com/docker/distribution => github.com/docker/distribution v2.7.1-0.20190104202606-0ac367fd6bee+incompatible
	github.com/florianl/go-conntrack => github.com/florianl/go-conntrack v0.2.0
	github.com/grpc-ecosystem/grpc-gateway => github.com/grpc-ecosystem/grpc-gateway v1.12.2
	github.com/iovisor/gobpf => github.com/DataDog/gobpf v0.0.0-20210322155958-9866ef4cd22c
	github.com/kubernetes-incubator/custom-metrics-apiserver => github.com/kubernetes-incubator/custom-metrics-apiserver v0.0.0-20200618121405-54026617ec44
	github.com/lxn/walk => github.com/lxn/walk v0.0.0-20180521183810-02935bac0ab8
	github.com/mholt/archiver => github.com/mholt/archiver v2.0.1-0.20171012052341-26cf5bb32d07+incompatible
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	github.com/spf13/cast => github.com/DataDog/cast v1.3.1-0.20190301154711-1ee8c8bd14a3
	github.com/ugorji/go => github.com/ugorji/go v1.1.7
	github.com/vishvananda/netlink => github.com/DataDog/netlink v1.0.1-0.20210312173533-6d628a7fc6f3
	google.golang.org/grpc => github.com/grpc/grpc-go v1.26.0
	k8s.io/api => k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.6
	k8s.io/apiserver => k8s.io/apiserver v0.18.2
	k8s.io/autoscaler => k8s.io/autoscaler v0.18.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.2
	k8s.io/client-go => k8s.io/client-go v0.18.6
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191016115326-20453efc2458
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.2
	k8s.io/code-generator => k8s.io/code-generator v0.18.2
	k8s.io/component-base => k8s.io/component-base v0.18.2
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191016114939-2b2b218dc1df
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191016114407-2e83b6f20229
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191016114748-65049c67a58b
	k8s.io/kube-state-metrics => k8s.io/kube-state-metrics v1.9.8-0.20200729235434-0ff5482079c8
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191016114556-7841ed97f1b2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191016115753-cf0698c3a16b
	k8s.io/metrics => k8s.io/metrics v0.18.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191016112829-06bb3c9d77c9
)
