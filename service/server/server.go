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
	ct "github.com/kubearmor/KVMService/service/constants"
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

func (s *Server) UpdateETCDLabelToIdentitiesMaps(identity uint16) {
	////////
	EtcdClient.EtcdDelete(context.Background(), ct.KvmSvcIdentitiToPodIps+strconv.Itoa(int(identity)))

	labelKV, err := EtcdClient.EtcdGet(context.Background(), ct.KvmOprIdentityToLabel+strconv.Itoa(int(identity)))
	if err != nil {
		log.Fatal(err)
		return
	}
	label := labelKV[ct.KvmOprIdentityToLabel+strconv.Itoa(int(identity))]

	data, err := EtcdClient.EtcdGetRaw(context.Background(), ct.KvmOprLabelToIdentities+label)
	if err != nil {
		log.Fatal(err)
		return
	}

	var arr []uint16
	for _, ev := range data.Kvs {
		err := json.Unmarshal(ev.Value, &arr)
		if err != nil {
			log.Fatal(err)
			return
		}
		kg.Printf("Removing the identity(%d) from the labels map of ETCD arr:%+v", identity, arr)
		for index, value := range arr {
			if identity == value {
				arr[index] = arr[len(arr)-1]
				arr[len(arr)-1] = 0
				arr = arr[:len(arr)-1]
			}
		}
		kg.Printf("After removing the identity(%d) from the labels map of ETCD arr:%+v", identity, arr)
		mapStr, _ := json.Marshal(arr)
		err = EtcdClient.EtcdPut(context.Background(), ct.KvmOprLabelToIdentities+label, string(mapStr))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (s *Server) SendPolicy(stream pb.KVM_SendPolicyServer) error {
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
					return nil
				}
				break
			}
		}
	}
}

func IsIdentityServing(identity string) int {
	kvPair, err := EtcdClient.EtcdGet(context.Background(), ct.KvmSvcIdentitiToPodIps+identity)
	if err != nil {
		log.Fatal(err)
		return 0
	}

	if len(kvPair) > 0 {
		kg.Printf("This Identity is already served by this podIP:%s", kvPair[ct.KvmSvcIdentitiToPodIps+identity])
		return 0
	}

	etcdLabels, err := EtcdClient.EtcdGet(context.Background(), ct.KvmOprIdentityToLabel)
	if err != nil {
		log.Fatal(err)
	}
	for key, value := range etcdLabels {
		s := strings.Split(key, "/")
		id := s[len(s)-1]
		if id == identity {
			kg.Printf("Validated the identity from the etcd DB identity:%s is unique for label:%s", identity, value)
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
		kg.Print("Connection refused due to already busy or invalid identity")
		return &pb.Status{Status: -1}, nil
	}

	value, _ := strconv.Atoi(in.Identity)
	identity = uint16(value)
	kg.Printf("New connection recieved RegisterAgentIdentity: %v podIp: %v", identity, podIp)

	EtcdClient.EtcdPutWithTTL(context.Background(), ct.KvmSvcIdentitiToPodIps+in.Identity, podIp)

	return &pb.Status{Status: 0}, nil
}

func (s *Server) InitServer() error {
	// TCP connection - Listen on port specified in input
	PolicyChan = make(chan tp.K8sKubeArmorHostPolicyEventWithIdentity)
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
