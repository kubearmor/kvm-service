package server

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"strings"

	etcd "github.com/kubearmor/KVMService/service/etcd"
	kg "github.com/kubearmor/KVMService/service/log"
	pb "github.com/kubearmor/KVMService/service/protobuf"
	tp "github.com/kubearmor/KVMService/service/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	PolicyChan  chan tp.K8sKubeArmorHostPolicyEventWithIdentity
	ClusterIp   string
	Clusterport string
	podIp       string
	EtcdClient  *etcd.EtcdClient
)

// Variables / Struct
type Server struct {
	pb.UnimplementedKVMServer
	podIp      string
	port       string
	EtcdClient *etcd.EtcdClient
}

func NewServerInit(ipAddress, ClusterIpAddress, portVal string, Etcd *etcd.EtcdClient) *Server {
	kg.Printf("Initiliazing the KVMServer => podip:%v clusterIP:%v clusterPort:%v", ipAddress, ClusterIpAddress, portVal)
	podIp = ipAddress
	Clusterport = portVal
	EtcdClient = Etcd
	ClusterIp = ClusterIpAddress
	return &Server{podIp: ipAddress, port: portVal, EtcdClient: EtcdClient}
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

func (s *Server) SendPolicy(stream pb.KVM_SendPolicyServer) error {
	var policy pb.PolicyData
	var loop bool
	loop = true

	kg.Print("Started Policy Streamer")
	PolicyChan = make(chan tp.K8sKubeArmorHostPolicyEventWithIdentity)

	go func() {
		for loop {
			select {
			case <-stream.Context().Done():
				closeEvent := tp.K8sKubeArmorHostPolicyEventWithIdentity{}
				closeEvent.Identity = GetIdentityFromContext(stream.Context())
				closeEvent.CloseConnection = true
				kg.Printf("Done context received for identity %d", closeEvent.Identity)
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
						kg.Print("Failed to marshall data")
					} else {
						policy.PolicyData = policyBytes
						err := stream.Send(&policy)
						if err != nil {
							kg.Print("Failed to send")
						}
						response, err := stream.Recv()
						if err != nil {
							kg.Print("Failed to recv")
						}
						kg.Printf("Policy Enforcement status in host :%d", response.Status)
					}
				} else {
					kg.Printf("Context is %d", GetIdentityFromContext(stream.Context()))
					kg.Print("Closing the connection")
					close(PolicyChan)
					return nil
				}
				break
			}
		}
	}
}

func IsIdentityServing(identity string) int {
	kvPair, err := EtcdClient.EtcdGet(context.Background(), "/ew-identities/"+identity)
	if err != nil {
		log.Fatal(err)
		return 0
	}

	if len(kvPair) > 0 {
		kg.Printf("This Identity is already served by this podIP:%s", kvPair["/ew-identities/"+identity])
		return 0
	}

	etcdLabels, err := EtcdClient.EtcdGet(context.Background(), "/externalworkloads")
	if err != nil {
		log.Fatal(err)
	}
	for key, value := range etcdLabels {
		s := strings.Split(key, "/")
		id := s[len(s)-1]
		if id == identity {
            kg.Printf("Validated the recieved identity from the etcd DB identity:%s label:%s", identity, value)
			return 1
		}
	}
    kg.Printf("Recieved the invalid identity:%s", identity)
	return 0
}

func (s *Server) RegisterAgentIdentity(ctx context.Context, in *pb.AgentIdentity) (*pb.Status, error) {
	kg.Print("Recieved the connection from the identity")
	var identity uint16
	// TODO : Which function for identity register with etcd
	if IsIdentityServing(in.Identity) == 0 {
		kg.Print("Connection refused due to already busy identity")
		return &pb.Status{Status: -1}, nil
	}

	value, _ := strconv.Atoi(in.Identity)
	identity = uint16(value)
	kg.Printf("New connection recieved RegisterAgentIdentity: %v podIp: %v", identity, podIp)

	EtcdClient.EtcdPutWithTTL(context.Background(), "/ew-identities/"+in.Identity, podIp)

	return &pb.Status{Status: 0}, nil
}

func (s *Server) InitServer() error {
	// TCP connection - Listen on port specified in input
	tcpConn, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		kg.Printf("Error listening on port %s", s.port)
	} else {
		kg.Printf("Successfully KVMServer Listening on port %s", s.port)
	}

	// Start gRPC server and register for protobuf
	gRPCServer := grpc.NewServer()
	if gRPCServer == nil {
		kg.Err("Failed to serve gRPCServer is null")
	}
	pb.RegisterKVMServer(gRPCServer, &Server{})

	err = gRPCServer.Serve(tcpConn)
	if err != nil {
		kg.Err("Failed to serve")
	}

	return err
}
