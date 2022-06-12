package cilium

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	etcd "github.com/kubearmor/KVMService/src/etcd"
	"github.com/kubearmor/KVMService/src/log"
	"github.com/kubearmor/KVMService/src/service/cilium/kvstore"
	"github.com/kubearmor/KVMService/src/service/cilium/labels"
	"github.com/kubearmor/KVMService/src/service/cilium/types"
	ct "github.com/kubearmor/KVMService/src/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkPolicyRequest struct {
	Type   string        `json:"type"`
	Object NetworkPolicy `json:"object"`
}

type NetworkPolicy struct {
	Metadata v1.ObjectMeta   `json:"metadata"`
	Spec     *types.Rule     `json:"spec"`
	Status   ct.PolicyStatus `json:"status,omitempty"`
}

type NetworkPolicyRequestCallback func(*NetworkPolicyRequest)

var (
	DefaultNamespace   = "default"
	DefaultClusterName = "default"

	DefaultClusterID = 1

	PodNameLabel      = "io.kubernetes.pod.name"
	PodNamespaceLabel = "io.kubernetes.pod.namespace"

	nodeRegisterWatcher *etcd.EtcdWatcher
)

func HandlePolicyUpdates(client *etcd.EtcdClient, req *NetworkPolicyRequest) {
	if (req.Type == "ADDED") || (req.Type == "MODIFIED") {
		policy := req.Object

		name := policy.Metadata.Name
		ccnp := policy.Spec
		ccnp.Labels = getPolicyLabels(name, uuid.New().String())

		key := path.Join(kvstore.PolicyPrefix, name)
		value, err := json.Marshal(ccnp)
		if err != nil {
			log.Err(fmt.Sprintf("Unable to JSON marshal entry %#v", value))
			log.Err(err.Error())
			return
		}

		err = client.EtcdPut(context.TODO(), key, string(value))
		if err != nil {
			log.Err(err.Error())
			return
		}

	} else if req.Type == "DELETED" {
		name := req.Object.Metadata.Name

		key := path.Join(kvstore.PolicyPrefix, name)
		err := client.EtcdDelete(context.TODO(), key)
		if err != nil {
			log.Err(err.Error())
			return
		}
	} else {
		log.Printf("Invalid network policy request. Request type=%s", req.Type)
	}
}

func getPolicyLabels(name, uid string) labels.LabelArray {
	return labels.LabelArray{
		labels.NewLabel(labels.PolicyLabelName, name, labels.LabelSourceKVM),
		labels.NewLabel(labels.PolicyLabelUID, uid, labels.LabelSourceKVM),
		labels.NewLabel(labels.PolicyLabelDerivedFrom, types.ResourceTypeCiliumClusterwideNetworkPolicy, labels.LabelSourceKVM),
	}
}

func UpdateLabels(client *etcd.EtcdClient, identity uint16, lbls []string) {
	key := path.Join(kvstore.IdentityPrefix, strconv.Itoa(int(identity)))

	// 1. Parse labels => LabelMap
	lblMap := labels.NewLabelMap(lbls)

	// 2. Stringify
	value := lblMap.SortedList()

	// 3. Push it to etcd
	err := client.EtcdPut(context.TODO(), key, value)
	if err != nil {
		log.Err(err.Error())
	}
}

func NodeRegisterWatcherInit(client *etcd.EtcdClient, idMap map[string]uint16) {
	handlerFunc := func(key string, obj interface{}) {
		if node, ok := obj.(*types.Node); !ok {
			log.Err(fmt.Sprintf("Invalid Node object type %s received: %+v", reflect.TypeOf(obj), obj))
		} else {
			handleNodeRegister(client, idMap, node)
		}
	}

	nodeRegisterWatcher = etcd.NewWatcher(
		client,
		kvstore.NodeRegisterPrefix,
		(time.Second * 3),
		nodeRegisterUnmarshal,
		handlerFunc,
		handlerFunc,
		nil,
	)

	nodeRegisterWatcher.Observe(context.TODO())
}

func handleNodeRegister(client *etcd.EtcdClient, idMap map[string]uint16, node *types.Node) {
	if node.NodeIdentity == 0 {
		// Node registration Phase 1

		// 1. Based on node.Name get the identity and labels
		id := idMap[node.Name]
		if id == 0 {
			log.Err(fmt.Sprintf("Cannot find identity of Node: %s", node.Name))
			return
		}

		labels := getNodeLabels(client, id)
		if len(labels) == 0 {
			log.Err(fmt.Sprintf("Cannot find labels of Node: %s", node.Name))
		}

		// 2. Update the node
		node.NodeIdentity = uint32(id)

		setDefaultNodeLabels(node)
		for k, v := range labels {
			node.Labels[k] = v
		}

		node.Cluster = DefaultClusterName
		node.ClusterID = DefaultClusterID
		node.IPAddresses = nil
		node.IPv4AllocCIDR = nil
		node.IPv6AllocCIDR = nil

		// 3. Push it to etcd
		key := path.Join(kvstore.NodeRegisterPrefix, node.Name)
		value, err := json.Marshal(node)
		if err != nil {
			log.Err(fmt.Sprintf("Unable to JSON marshal entry %#v", value))
			log.Err(err.Error())
			return
		}

		err = client.EtcdPut(context.TODO(), key, string(value))
		if err != nil {
			log.Err(err.Error())
			return
		}

	} else if len(node.IPAddresses) > 0 {
		// Node registration Phase 2
		updateNodeIPs(client, node)
	}
}

func setDefaultNodeLabels(node *types.Node) {
	if node != nil {
		if node.Labels == nil {
			node.Labels = map[string]string{}
		}
		node.Labels[PodNameLabel] = node.Name
		node.Labels[PodNamespaceLabel] = DefaultNamespace
	}
}

func updateNodeIPs(client *etcd.EtcdClient, node *types.Node) {
	for _, addr := range node.IPAddresses {
		key := path.Join(kvstore.IPCachePrefix, DefaultNamespace, addr.IP.String())

		// 1. Prepare IPIdentityPair
		entry := types.IPIdentityPair{
			IP:           addr.IP,
			Metadata:     "",
			HostIP:       node.GetHostIP(),
			ID:           types.NumericIdentity(node.NodeIdentity),
			Key:          node.EncryptionKey,
			K8sNamespace: DefaultNamespace,
			K8sPodName:   node.Name,
		}

		// 2. Stringify
		marshaledEntry, err := json.Marshal(entry)
		if err != nil {
			log.Err(fmt.Sprintf("Unable to JSON marshal entry %#v", entry))
			log.Err(err.Error())
			continue
		}

		// 3. Push it to etcd
		err = client.EtcdPut(context.TODO(), key, string(marshaledEntry))
		if err != nil {
			log.Err(err.Error())
			continue
		}
	}
}

func getNodeLabels(client *etcd.EtcdClient, identity uint16) map[string]string {
	labelMap := make(map[string]string)
	id := strconv.Itoa(int(identity))
	key := path.Join(kvstore.IdentityPrefix, id)
	resp, err := client.EtcdGet(context.TODO(), key)
	if err != nil {
		log.Err(err.Error())
		return nil
	}

	lbls := strings.Split(resp[key], ";")
	for _, lbl := range lbls {
		if lbl == "" {
			continue
		}
		label := labels.ParseLabel(lbl)
		labelMap[label.Key] = label.Value
	}

	return labelMap
}

func nodeRegisterUnmarshal(bytes []byte) (interface{}, error) {
	var node types.Node

	err := json.Unmarshal(bytes, &node)
	if err != nil {
		log.Err(err.Error())
		return nil, err
	}

	return &node, nil
}
