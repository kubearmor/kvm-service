package server

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	ct "github.com/kubearmor/KVMService/src/constants"
	kg "github.com/kubearmor/KVMService/src/log"
	pb "github.com/kubearmor/KVMService/src/service/protobuf"
	tp "github.com/kubearmor/KVMService/src/types"
	"google.golang.org/grpc/metadata"
)

type KVMServer struct {
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

func UpdateETCDLabelToIdentitiesMaps(identity uint16) {

	err := EtcdClient.EtcdDelete(context.Background(), ct.KvmSvcIdentitiToPodIps+strconv.Itoa(int(identity)))
	if err != nil {
		kg.Err(err.Error())
		return
	}

	labelKV, err := EtcdClient.EtcdGet(context.Background(), ct.KvmOprIdentityToLabel+strconv.Itoa(int(identity)))
	if err != nil {
		kg.Err(err.Error())
		return
	}
	label := labelKV[ct.KvmOprIdentityToLabel+strconv.Itoa(int(identity))]

	data, err := EtcdClient.EtcdGetRaw(context.Background(), ct.KvmOprLabelToIdentities+label)
	if err != nil {
		kg.Err(err.Error())
		return
	}

	var arr []uint16
	for _, ev := range data.Kvs {
		err := json.Unmarshal(ev.Value, &arr)
		if err != nil {
			kg.Err(err.Error())
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
			kg.Err(err.Error())
			return
		}
	}
}

func (s *KVMServer) SendPolicy(stream pb.KVM_SendPolicyServer) error {
	var policy pb.PolicyData
	var loop bool
	loop = true

	kg.Print("Started Policy Streamer\n")

	go func() {
		for loop {
			<-stream.Context().Done()
			closeEvent := tp.KubeArmorHostPolicyEventWithIdentity{}
			closeEvent.Identity = GetIdentityFromContext(stream.Context())
			closeEvent.CloseConnection = true
			closeEvent.Err = errors.New("connection")
			kg.Errf("Closing client connection for identity %d\n", closeEvent.Identity)
			UpdateETCDLabelToIdentitiesMaps(GetIdentityFromContext(stream.Context()))
			loop = false
			PolicyChan <- closeEvent
		}
	}()

	for {
		select {
		case event := <-PolicyChan:
			if event.Identity == GetIdentityFromContext(stream.Context()) {
				if !event.CloseConnection {
					policyBytes, err := json.Marshal(&event.Event)
					if err != nil {
						kg.Warn("Failed to marshall data")
					} else {
						policy.PolicyData = policyBytes
						err := stream.Send(&policy)
						if err != nil {
							kg.Warn("Failed to send")
						}
						response, err := stream.Recv()
						if err != nil {
							kg.Warn("Failed to recv")
						}
						kg.Printf("Policy Enforcement status in host :%d", response.Status)
					}
				} else {
					kg.Warnf("Closing connection for client [%d]", GetIdentityFromContext(stream.Context()))
					return event.Err
				}
			}
		default:
			continue
		}
	}
}

func IsIdentityServing(identity string) int {
	kvPair, err := EtcdClient.EtcdGet(context.Background(), ct.KvmSvcIdentitiToPodIps+identity)
	if err != nil {
		kg.Err(err.Error())
		return 0
	}

	if len(kvPair) > 0 {
		kg.Printf("This Identity is already served by this podIP:%s", kvPair[ct.KvmSvcIdentitiToPodIps+identity])
		return 0
	}

	etcdLabels, err := EtcdClient.EtcdGet(context.Background(), ct.KvmOprIdentityToLabel)
	if err != nil {
		kg.Err(err.Error())
		return 0
	}
	for key, value := range etcdLabels {
		s := strings.Split(key, "/")
		id := s[len(s)-1]
		if id == identity {
			kg.Printf("Validated the identity from the etcd DB identity:%s is unique for label:%s", identity, value)
			return 1
		}
	}
	kg.Printf("Received the invalid identity:%s", identity)
	return 0
}

func (s *KVMServer) RegisterAgentIdentity(ctx context.Context, in *pb.AgentIdentity) (*pb.Status, error) {
	kg.Print("Received the connection from the identity")
	var identity uint16

	if IsIdentityServing(in.Identity) == 0 {
		kg.Print("Connection refused due to already busy or invalid identity")
		return &pb.Status{Status: -1}, nil
	}

	value, _ := strconv.Atoi(in.Identity)
	identity = uint16(value)
	kg.Printf("New connection received RegisterAgentIdentity: %v podIp: %v", identity, podIp)

	err := EtcdClient.EtcdPutWithTTL(context.Background(), ct.KvmSvcIdentitiToPodIps+in.Identity, podIp)
	if err != nil {
		kg.Err(err.Error())
		return &pb.Status{Status: -1}, err
	}

	return &pb.Status{Status: 0}, nil
}
