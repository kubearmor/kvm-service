package server

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strconv"

	pb "github.com/kubearmor/KVMService/protobuf"
	kg "github.com/kubearmor/KVMService/service/log"
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
	PolicyChan = make(chan tp.K8sKubeArmorHostPolicyEventWithIdentity)

	go func() {
		for loop {
			select {
			case <-stream.Context().Done():
				closeEvent := tp.K8sKubeArmorHostPolicyEventWithIdentity{}
				closeEvent.Identity = GetIdentityFromContext(stream.Context())
				closeEvent.CloseConnection = true
				log.Printf("Done context received for identity %d\n", closeEvent.Identity)
				loop = false
				PolicyChan <- closeEvent
				// Client Connection interrupted
				// Remove identity from etcd
				// close(PolicyChan)
			}
		}
	}()

	for {
		select {
		case event := <-PolicyChan:
			if event.Identity == GetIdentityFromContext(stream.Context()) {
				if !event.CloseConnection {
					policyBytes, err := json.Marshal(&event.Event)
					if err != nil {
						log.Print("Failed to marshall data")
					} else {
						policy.PolicyData = policyBytes
						err := stream.Send(&policy)
						if err != nil {
							log.Print("Failed to send")
						}
						response, err := stream.Recv()
						log.Printf("Policy Enforcement status in host : %d", response.Status)
					}
				} else {
					log.Printf("Context is %d\n", GetIdentityFromContext(stream.Context()))
					log.Print("Closing the connection")
					close(PolicyChan)
					return nil
				}
				break
			}
		}
	}
}

func (s *server) RegisterAgentIdentity(ctx context.Context, in *pb.AgentIdentity) (*pb.Status, error) {
	var identity uint16
	// TODO : Which function for identity register with etcd
	value, _ := strconv.Atoi(in.Identity)
	identity = uint16(value)
	log.Printf("RegisterAgentIdentity = %v", identity)

	return &pb.Status{Status: 0}, nil
}

func InitServer(grpcPort string) error {

	// TCP connection - Listen on port specified in input
	tcpConn, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		kg.Printf("Error listening on port %s", grpcPort)
	} else {
		kg.Printf("Listening on port %s", grpcPort)
	}

	// Start gRPC server and register for protobuf
	gRPCServer := grpc.NewServer()
	pb.RegisterKVMServer(gRPCServer, &server{})

	err = gRPCServer.Serve(tcpConn)
	if err != nil {
		kg.Err("Failed to serve")
	}

	return err
}
