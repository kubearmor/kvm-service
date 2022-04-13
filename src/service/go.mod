module github.com/kubearmor/KVMService/src/service

go 1.17

replace (
	github.com/kubearmor/KVMService/src/common => ../common
	github.com/kubearmor/KVMService/src/constants => ../constants
	github.com/kubearmor/KVMService/src/etcd => ../etcd
	github.com/kubearmor/KVMService/src/log => ../log
	github.com/kubearmor/KVMService/src/service/cilium => ./cilium
	github.com/kubearmor/KVMService/src/service/config => ./config
	github.com/kubearmor/KVMService/src/service/core => ./core
	github.com/kubearmor/KVMService/src/service/genscript => ./genscript
	github.com/kubearmor/KVMService/src/service/protobuf => ./protobuf
	github.com/kubearmor/KVMService/src/service/server => ./server
	github.com/kubearmor/KVMService/src/types => ../types
)

require (
	github.com/kubearmor/KVMService/src/log v0.0.0-20220228115540-2211247620dd
	github.com/kubearmor/KVMService/src/service/config v0.0.0-00010101000000-000000000000
	github.com/kubearmor/KVMService/src/service/core v0.0.0-00010101000000-000000000000
)

require (
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kubearmor/KVMService/src/common v0.0.0-00010101000000-000000000000 // indirect
	github.com/kubearmor/KVMService/src/constants v0.0.0-00010101000000-000000000000 // indirect
	github.com/kubearmor/KVMService/src/etcd v0.0.0-00010101000000-000000000000 // indirect
	github.com/kubearmor/KVMService/src/service/cilium v0.0.0-00010101000000-000000000000 // indirect
	github.com/kubearmor/KVMService/src/service/genscript v0.0.0-00010101000000-000000000000 // indirect
	github.com/kubearmor/KVMService/src/service/protobuf v0.0.0-00010101000000-000000000000 // indirect
	github.com/kubearmor/KVMService/src/service/server v0.0.0-00010101000000-000000000000 // indirect
	github.com/kubearmor/KVMService/src/types v0.0.0-00010101000000-000000000000 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.10.1 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.1 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.1 // indirect
	go.etcd.io/etcd/client/v3 v3.5.1 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/net v0.0.0-20211209124913-491a49abca63 // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sys v0.0.0-20211210111614-af8b64212486 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa // indirect
	google.golang.org/grpc v1.43.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.24.0-alpha.0 // indirect
	k8s.io/apimachinery v0.24.0-alpha.0 // indirect
	k8s.io/client-go v0.24.0-alpha.0 // indirect
	k8s.io/klog/v2 v2.30.0 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b // indirect
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.0 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
