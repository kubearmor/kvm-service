module github.com/kubearmor/KVMService/src/service/cilium

go 1.17

replace (
	github.com/kubearmor/KVMService/src/common => ../../common
	github.com/kubearmor/KVMService/src/constants => ../../constants
	github.com/kubearmor/KVMService/src/etcd => ../../etcd
	github.com/kubearmor/KVMService/src/log => ../../log
	github.com/kubearmor/KVMService/src/service/cilium/kvstore => ./kvstore
	github.com/kubearmor/KVMService/src/service/cilium/labels => ./labels
	github.com/kubearmor/KVMService/src/service/cilium/types => ./types
	github.com/kubearmor/KVMService/src/types => ../../types
)

require (
	github.com/cilium/cilium v1.10.0
	github.com/google/uuid v1.3.0
	github.com/kubearmor/KVMService/src/etcd v0.0.0-00010101000000-000000000000
	github.com/kubearmor/KVMService/src/log v0.0.0-00010101000000-000000000000
	github.com/kubearmor/KVMService/src/types v0.0.0-00010101000000-000000000000
	k8s.io/apimachinery v0.23.1
)

require (
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.4.10-0.20200417215612-7f4cf4dd2b52 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-openapi/analysis v0.19.16 // indirect
	github.com/go-openapi/errors v0.19.9 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/loads v0.20.0 // indirect
	github.com/go-openapi/runtime v0.19.26 // indirect
	github.com/go-openapi/spec v0.20.3 // indirect
	github.com/go-openapi/strfmt v0.20.0 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/go-openapi/validate v0.20.1 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/kubearmor/KVMService/src/common v0.0.0-00010101000000-000000000000 // indirect
	github.com/kubearmor/KVMService/src/constants v0.0.0-00010101000000-000000000000 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/miekg/dns v1.1.26 // indirect
	github.com/mitchellh/mapstructure v1.4.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.1-0.20200623203004-60555c9708c7 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/sasha-s/go-deadlock v0.2.1-0.20190427202633-1595213edefa // indirect
	github.com/shirou/gopsutil/v3 v3.21.2 // indirect
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/spf13/afero v1.3.4 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.1 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.1 // indirect
	go.etcd.io/etcd/client/v3 v3.5.1 // indirect
	go.mongodb.org/mongo-driver v1.4.4 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/crypto v0.0.0-20210503195802-e9a32991a82e // indirect
	golang.org/x/net v0.0.0-20211209124913-491a49abca63 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210831042530-f4d43177bf5e // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	google.golang.org/grpc v1.38.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.22.4 // indirect
	k8s.io/client-go v0.22.4 // indirect
	k8s.io/klog/v2 v2.30.0 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b // indirect
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
