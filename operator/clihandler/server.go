package clihandler

import (
	"context"
	"log"
	"net"

	ct "github.com/kubearmor/KVMService/operator/constants"
	etcd "github.com/kubearmor/KVMService/operator/etcd"
	gs "github.com/kubearmor/KVMService/operator/genscript"
	kg "github.com/kubearmor/KVMService/operator/log"
	pb "github.com/kubearmor/KVMService/operator/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Variables / Struct
type Server struct {
	pb.UnimplementedHandleCliServer
	EtcdClient *etcd.EtcdClient
	port       string
}

type CLIServer struct {
	pb.HandleCliServer
}

var cs Server

func NewServerInit(portVal string, Etcd *etcd.EtcdClient) *Server {
	kg.Printf("Initiliazing the CLIHandler => Port:%v", portVal)
	cs.EtcdClient = Etcd
	return &Server{port: portVal, EtcdClient: Etcd}
}

func (c *CLIServer) HandleCliRequest(ctx context.Context, request *pb.CliRequest) (*pb.ResponseStatus, error) {
	kg.Printf("Recieved the request KVMName:%s\n", request.KvmName)
	kvPair, err := cs.EtcdClient.EtcdGet(context.Background(), ct.KvmOprEWNameToIdentity+request.KvmName)
	if err != nil {
		log.Fatal(err)
		return &pb.ResponseStatus{ScriptData: "", StatusMsg: "Error: DB reading failed", Status: -1}, err
	}

	if len(kvPair[ct.KvmOprEWNameToIdentity+request.KvmName]) == 0 {
		return &pb.ResponseStatus{ScriptData: "", StatusMsg: "Error: KVM Name is not present in DB", Status: -1}, nil
	}

	kg.Printf("Handling the CLI request for Identity '%s'\n", kvPair[ct.KvmOprEWNameToIdentity+request.KvmName])

	scriptData := gs.GenerateEWInstallationScript(request.KvmName, kvPair[ct.KvmOprEWNameToIdentity+request.KvmName])
	return &pb.ResponseStatus{ScriptData: scriptData, StatusMsg: "Success", Status: 0}, nil
}

func (s *Server) InitServer() error {
	tcpConn, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		kg.Printf("Error in CLIHandler listening on port %s", s.port)
	} else {
		kg.Printf("Successfully CLIHandler Listening on port %s", s.port)
	}

	gRPCServer := grpc.NewServer()
	if gRPCServer == nil {
		kg.Err("Failed to create CLIHandler")
	}
	reflection.Register(gRPCServer)
	pb.RegisterHandleCliServer(gRPCServer, &CLIServer{})

	err = gRPCServer.Serve(tcpConn)
	if err != nil {
		kg.Err("Failed to start CLIHandler")
	}

	return err
}
