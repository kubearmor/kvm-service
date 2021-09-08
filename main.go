package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"

	pb "github.com/kubearmor/KVMService/protobuf"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedKVMServer
}

func (s *server) RegisterAgentIdentity(ctx context.Context, in *pb.AgentIdentity) (*pb.Status, error) {
	log.Println("Agent Identity is ", in.Identity)
	os.Setenv("AGENT_IDENTITY", in.Identity)
	return &pb.Status{Status: 200, ErrorMessage: "nil"}, nil
}

func main() {

	// Get input flag arguments
	gRPCPtr := flag.String("gRPC", "32767", "gRPC port number")
	flag.Parse()

	// TCP connection - Listen on port specified in input
	tcpConn, err := net.Listen("tcp", ":"+*gRPCPtr)
	if err != nil {
		log.Fatal("Error listening on port ", *gRPCPtr)
	}

	// Start gRPC server and register for protobuf
	gRPCServer := grpc.NewServer()
	pb.RegisterKVMServer(gRPCServer, &server{})

	err = gRPCServer.Serve(tcpConn)
	if err != nil {
		log.Fatal("Failed to init gRPC Server ")
	}
}
