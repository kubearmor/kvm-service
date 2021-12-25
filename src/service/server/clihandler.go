package server

import (
	"context"

	ct "github.com/kubearmor/KVMService/src/constants"
	kg "github.com/kubearmor/KVMService/src/log"
	gs "github.com/kubearmor/KVMService/src/service/genscript"
	pb "github.com/kubearmor/KVMService/src/service/protobuf"
)

// Variables / Struct
type CLIServer struct {
	pb.HandleCliServer
}

func (c *CLIServer) HandleCliRequest(ctx context.Context, request *pb.CliRequest) (*pb.ResponseStatus, error) {
	kg.Printf("Received the request KVMName:%s\n", request.KvmName)
	kvPair, err := EtcdClient.EtcdGet(context.Background(), ct.KvmOprEWNameToIdentity+request.KvmName)
	if err != nil {
		kg.Err(err.Error())
		return &pb.ResponseStatus{ScriptData: "", StatusMsg: "Error: DB reading failed", Status: -1}, err
	}

	if len(kvPair[ct.KvmOprEWNameToIdentity+request.KvmName]) == 0 {
		return &pb.ResponseStatus{ScriptData: "", StatusMsg: "Error: KVM Name is not present in DB", Status: -1}, nil
	}

	kg.Printf("Handling the CLI request for Identity '%s'\n", kvPair[ct.KvmOprEWNameToIdentity+request.KvmName])

	scriptData := gs.GenerateEWInstallationScript(request.KvmName, kvPair[ct.KvmOprEWNameToIdentity+request.KvmName])
	return &pb.ResponseStatus{ScriptData: scriptData, StatusMsg: "Success", Status: 0}, nil
}
