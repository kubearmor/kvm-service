package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/kubearmor/KVMService/protobuf"
	"google.golang.org/grpc"
)

// Variables / Struct
type server struct {
	pb.UnimplementedKVMServer
}

var sendPolicyToHost bool
var policy pb.PolicyData

func testSendPolicy() {
	fmt.Println("SEND POLICY TEST")
	time.Sleep(time.Second * 30) // Sleep for 30 seconds
	fmt.Println("TEST CASE #1")
	data := []byte("Test_case_001")
	policyName := "Test1"
	sendHostPolicy(data, policyName)
	time.Sleep(time.Second * 10)
	fmt.Println("TEST CASE #2")
	data = []byte("Test_case_002")
	policyName = "Test2"
	sendHostPolicy(data, policyName)
	time.Sleep(time.Second * 10)
	fmt.Println("TEST CASE #3")
	data = []byte("Test_case_003")
	policyName = "Test3"
	sendHostPolicy(data, policyName)
}

func sendHostPolicy(data []byte, name string) {
	if !sendPolicyToHost {
		policy.PolicyData = data
		policy.PolicyName = name
		sendPolicyToHost = true
	}
}

func (s *server) SendPolicy(stream pb.KVM_SendPolicyServer) error {
	err := *new(error)
	err = nil
	sendPolicyToHost = false

	if err == nil {
		log.Printf("Started Policy Streamer\n")
	}

	for {
		if sendPolicyToHost {
			err := stream.Send(&policy)
			if err == io.EOF {
				sendPolicyToHost = false
				policy.Reset()
			}
			if err != nil {
				log.Fatal("Failed to send")
				return err
			}
			sendPolicyToHost = false
			policy.Reset()
		} else {
			continue
		}
	}
}

func (s *server) RegisterAgentIdentity(ctx context.Context, in *pb.AgentIdentity) (*pb.Status, error) {
	log.Println("Agent Identity is ", in.Identity)
	os.Setenv("AGENT_IDENTITY", in.Identity)
	return &pb.Status{Status: 200, ErrorMessage: "nil"}, nil
}

func main() {

	// Get input flag arguments
	gRPCPtr := flag.String("gRPC", "4000", "gRPC port number")
	flag.Parse()

	go testSendPolicy()

	// TCP connection - Listen on port specified in input
	tcpConn, err := net.Listen("tcp", ":"+*gRPCPtr)
	if err != nil {
		log.Fatal("Error listening on port ", *gRPCPtr)
	} else {
		fmt.Printf("Listening on port %s\n", *gRPCPtr)
	}

	// Start gRPC server and register for protobuf
	gRPCServer := grpc.NewServer()
	pb.RegisterKVMServer(gRPCServer, &server{})

	err = gRPCServer.Serve(tcpConn)
	if err != nil {
		log.Fatal("Failed to init gRPC Server ")
	}
}
