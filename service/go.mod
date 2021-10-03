module github.com/kubearmor/KVMService/service

go 1.15

replace (
	github.com/coreos/bbolt v1.3.6 => go.etcd.io/bbolt v1.3.6
	github.com/kubearmor/KVMService/service/protobuf => ./protobuf
	github.com/kubearmor/KVMService/service/server => ./server
)

require (
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/kubearmor/KVMService/service/protobuf v0.0.0-00010101000000-000000000000
	go.etcd.io/etcd/client/pkg/v3 v3.5.0
	go.etcd.io/etcd/client/v3 v3.5.0
	go.uber.org/zap v1.18.1
	golang.org/x/net v0.0.0-20210726213435-c6fcb2dbf985 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/tools v0.1.5 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/grpc v1.40.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
)
