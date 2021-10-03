package server

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"strconv"

	kg "github.com/kubearmor/KVMService/service/log"
	pb "github.com/kubearmor/KVMService/service/protobuf"
	tp "github.com/kubearmor/KVMService/service/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	PolicyChan chan tp.K8sKubeArmorHostPolicyEventWithIdentity
)

// Variables / Struct
type server struct {
	pb.UnimplementedKVMServer
}

func GetIdentityFromContext(ctx context.Context) uint16 {
	var values []string
	var token string

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values = md.Get("identity")
		if len(values) > 0 {
			token = values[0]
		}
	}
	identity, _ := strconv.Atoi(token)

	return uint16(identity)
}

func (s *server) SendPolicy(stream pb.KVM_SendPolicyServer) error {
	var policy pb.PolicyData
	var loop bool
	loop = true

	kg.Print("Started Policy Streamer\n")

	go func() {
		for loop {
			// select {
			// case <-stream.Context().Done():
			<-stream.Context().Done()
			closeEvent := tp.K8sKubeArmorHostPolicyEventWithIdentity{}
			closeEvent.Identity = GetIdentityFromContext(stream.Context())
			closeEvent.CloseConnection = true
			kg.Errf("Closing client connections for identity %d\n", closeEvent.Identity)
			loop = false
			PolicyChan <- closeEvent
			//}
		}
	}()

	for {
		// select {
		// case event := <-PolicyChan:
		event := <-PolicyChan
		if event.Identity == GetIdentityFromContext(stream.Context()) {
			if !event.CloseConnection {
				policyBytes, err := json.Marshal(&event.Event)
				if err != nil {
					kg.Err("Failed to marshall data")
				} else {
					policy.PolicyData = policyBytes
					err := stream.Send(&policy)
					if err == io.EOF {
						kg.Err("client disconnected")
					}
					if err != nil {
						kg.Err("Failed to send")
					}
					response, err := stream.Recv()
					kg.Printf("Policy Enforcement status in host : %d err=%v", response.Status, err)
				}
			} else {
				kg.Print("Closing connection\n")
				break
			}
			break
		}
		// }
	}
	return nil
}

func (s *server) RegisterAgentIdentity(ctx context.Context, in *pb.AgentIdentity) (*pb.Status, error) {
	value, _ := strconv.Atoi(in.Identity)
	identity := uint16(value)
	kg.Printf("Agent Identity is %v\n", identity)
	return &pb.Status{Status: 0}, nil
}

func InitServer(grpcPort string) {

	// TCP connection - Listen on port specified in input
	tcpConn, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		kg.Errf("Error listening on port %s", grpcPort)
	}

	// Start gRPC server and register for protobuf
	gRPCServer := grpc.NewServer()
	pb.RegisterKVMServer(gRPCServer, &server{})

	// Create a channel for posting policy messages
	PolicyChan = make(chan tp.K8sKubeArmorHostPolicyEventWithIdentity)

	err = gRPCServer.Serve(tcpConn)
	if err != nil {
		kg.Err("Failed to serve")
	}
}
