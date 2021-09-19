module github.com/kubearmor/KVMService/protobuf

go 1.15

replace (
	github.com/kubearmor/KVMService => ../
	github.com/kubearmor/KVMService/protobuf => ./
)

require (
	google.golang.org/grpc v1.34.0
	google.golang.org/protobuf v1.25.0
)
