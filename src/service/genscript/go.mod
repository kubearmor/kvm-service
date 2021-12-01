module github.com/kubearmor/KVMService/src/service/genscript

go 1.17

replace github.com/kubearmor/KVMService/src/log => ../../log

require github.com/kubearmor/KVMService/src/log v0.0.0-00010101000000-000000000000

require (
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.19.1 // indirect
)
