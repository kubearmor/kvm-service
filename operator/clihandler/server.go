package clihandler

import (
	"context"
	"net"

	etcd "github.com/kubearmor/KVMService/operator/etcd"
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

func NewServerInit(portVal string, Etcd *etcd.EtcdClient) *Server {
	kg.Printf("Initiliazing the CLIHandler => Port:%v", portVal)
	return &Server{port: portVal, EtcdClient: Etcd}
}

func (s *Server) HandleCliRequest(ctx context.Context, request *pb.CliRequest) (*pb.Status, error) {
	kg.Printf("Recieved the request KVMName:%s request:%d\n", request.KvmName, request.RequestType)
	return &pb.Status{Status: 0}, nil
}

/*
func (s *Server) RegisterAgentIdentity(ctx context.Context, in *pb.AgentIdentity) (*pb.Status, error) {
	kg.Print("Recieved the connection from the identity")
	var identity uint16
	// TODO : Which function for identity register with etcd
	if IsIdentityServing(in.Identity) == 0 {
		kg.Print("Connection refused due to already busy or invalid identity")
		return &pb.Status{Status: -1}, nil
	}

	value, _ := strconv.Atoi(in.Identity)
	identity = uint16(value)
	kg.Printf("New connection recieved RegisterAgentIdentity: %v podIp: %v", identity, podIp)

	EtcdClient.EtcdPutWithTTL(context.Background(), ct.KvmSvcIdentitiToPodIps+in.Identity, podIp)

	return &pb.Status{Status: 0}, nil
}*/

func (s *Server) InitServer() error {
	// TCP connection - Listen on port specified in input
	//PolicyChan = make(chan tp.K8sKubeArmorHostPolicyEventWithIdentity)
	tcpConn, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		kg.Printf("Error in CLIHandler listening on port %s", s.port)
	} else {
		kg.Printf("Successfully CLIHandler Listening on port %s", s.port)
	}

	// Start gRPC server and register for protobuf
	gRPCServer := grpc.NewServer()
	if gRPCServer == nil {
		kg.Err("Failed to create CLIHandler")
	}
	reflection.Register(gRPCServer)
	pb.RegisterHandleCliServer(gRPCServer, &Server{})

	err = gRPCServer.Serve(tcpConn)
	if err != nil {
		kg.Err("Failed to start CLIHandler")
	}

	return err
}
