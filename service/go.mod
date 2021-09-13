module github.com/kubearmor/KVMService/service

go 1.16

replace github.com/coreos/bbolt v1.3.6 => go.etcd.io/bbolt v1.3.6

require (
	go.etcd.io/etcd/client/pkg/v3 v3.5.0
	go.etcd.io/etcd/client/v3 v3.5.0
	go.uber.org/zap v1.17.0
	google.golang.org/grpc v1.38.0
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
)
