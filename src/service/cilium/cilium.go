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

	ci "github.com/cilium/cilium/pkg/identity"
	k8sConst "github.com/cilium/cilium/pkg/k8s/apis/cilium.io"
	cu "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/utils"
	cv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	ck8sv1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/apis/meta/v1"
	cl "github.com/cilium/cilium/pkg/labels"
	cnt "github.com/cilium/cilium/pkg/node/types"
	ca "github.com/cilium/cilium/pkg/policy/api"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/uuid"
	etcd "github.com/kubearmor/KVMService/src/etcd"
	"github.com/kubearmor/KVMService/src/log"
	"github.com/kubearmor/KVMService/src/service/cilium/kvstore"
)

type NodeInfo struct {
	Labels            cl.LabelArray
	Initialized       bool
	EgressPolicyCount uint32
}

var NodeCache map[string]*NodeInfo

func init() {
	NodeCache = make(map[string]*NodeInfo)
}

type NetworkPolicyRequest struct {
	Type   string                  `json:"type"`
	Object cv2.CiliumNetworkPolicy `json:"object"`
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
		handleDefaultCtrlPlanePolicy(client, req)
		err := updatePolicy(client, &req.Object)
		if err != nil {
			log.Err(err.Error())
			return
		}

	} else if req.Type == "DELETED" {
		err := deletePolicy(client, &req.Object)
		if err != nil {
			log.Err(err.Error())
			return
		}
		handleDefaultCtrlPlanePolicy(client, req)
	} else {
		log.Printf("Invalid network policy request. Request type=%s", req.Type)
	}
}

func handleDefaultCtrlPlanePolicy(client *etcd.EtcdClient, req *NetworkPolicyRequest) {
	ccnp := req.Object
	selectorLabels := ccnp.Spec.NodeSelector

	// 1. Get the list of nodes that matches the NodeSelector
	matchedNodes := []string{}
	if len(ccnp.Spec.Egress) > 0 {
		for nodeName, nodeInfo := range NodeCache {
			nodeLabel := nodeInfo.Labels
			if selectorLabels.Matches(nodeLabel) {
				matchedNodes = append(matchedNodes, nodeName)
			}
		}
	}

	if (req.Type == "ADDED") || (req.Type == "MODIFIED") {
		if oldSpec := getPolicy(client, ccnp.Name); oldSpec != nil {
			// user is updating an existing policy
			// Old policy and new policy might have different node selectors
			// So handle deletion of the old policy
			oldCCnp := newCCNP(ccnp.Name)
			oldCCnp.Spec = oldSpec
			handleDefaultCtrlPlanePolicy(client, &NetworkPolicyRequest{"DELETED", oldCCnp})
		}

		// 2. Increment the policy counter
		for _, nodeName := range matchedNodes {
			nodeInfo := NodeCache[nodeName]
			nodeInfo.EgressPolicyCount++

			// 3. This is when the first egress policy is applied to a node.
			//    Apply the default control plane policy before applying anything else.
			if nodeInfo.EgressPolicyCount == 1 {
				defPolicy := buildDefaultCtrlPlanePolicy(nodeName, nodeInfo.Labels)
				err := updatePolicy(client, defPolicy)
				if err != nil {
					log.Err(err.Error())
					return
				}
			}
		}
	} else if req.Type == "DELETED" {
		// 2. Decrement the policy counter
		for _, nodeName := range matchedNodes {
			nodeInfo := NodeCache[nodeName]
			if nodeInfo.EgressPolicyCount > 0 {
				nodeInfo.EgressPolicyCount--
			}

			// 2. This is when all the egress policy are deleted.
			//    Remove the default control plane policy.
			if nodeInfo.EgressPolicyCount == 0 {
				defPolicy := buildDefaultCtrlPlanePolicy(nodeName, nodeInfo.Labels)
				err := deletePolicy(client, defPolicy)
				if err != nil {
					log.Err(err.Error())
					return
				}
			}
		}
	}
}

func newCCNP(name string) cv2.CiliumNetworkPolicy {
	return cv2.CiliumNetworkPolicy{
		TypeMeta:   v1.TypeMeta{Kind: cu.ResourceTypeCiliumClusterwideNetworkPolicy},
		ObjectMeta: v1.ObjectMeta{Name: name},
	}
}

func buildDefaultCtrlPlanePolicy(node string, selector cl.LabelArray) *cv2.CiliumNetworkPolicy {
	policyName := fmt.Sprintf("00-allow-%s-egress-control-plane", node)
	description := fmt.Sprintf("Egress policy of %s to allow control plane access", node)

	ccnp := newCCNP(policyName)
	ccnp.Spec = &ca.Rule{
		Description: description,
		NodeSelector: ca.EndpointSelector{
			LabelSelector: &ck8sv1.LabelSelector{MatchLabels: make(map[string]string)},
		},
		Egress: []ca.EgressRule{{
			EgressCommonRule: ca.EgressCommonRule{ToEntities: ca.EntitySlice{"world"}},
			ToPorts:          ca.PortRules{{Ports: []ca.PortProtocol{{Port: "2379", Protocol: ca.ProtoTCP}}}},
		}},
	}

	for _, lbl := range selector {
		ccnp.Spec.NodeSelector.MatchLabels[lbl.Key] = lbl.Value
	}

	return &ccnp
}

func getPolicy(client *etcd.EtcdClient, name string) *ca.Rule {
	var spec ca.Rule

	key := path.Join(kvstore.PolicyPrefix, name)
	resp, err := client.EtcdGet(context.TODO(), key)
	if err != nil {
		log.Err(err.Error())
		return nil
	}

	if len(resp) == 0 {
		return nil
	}

	err = json.Unmarshal([]byte(resp[key]), &spec)
	if err != nil {
		log.Err(err.Error())
		return nil
	}

	return &spec
}

func updatePolicy(client *etcd.EtcdClient, ccnp *cv2.CiliumNetworkPolicy) error {
	if ccnp == nil {
		return nil
	}

	ccnp.Spec.Labels = getPolicyLabels(ccnp.Name, uuid.New().String())

	key := path.Join(kvstore.PolicyPrefix, ccnp.Name)
	value, err := json.Marshal(ccnp.Spec)
	if err != nil {
		return err
	}

	err = client.EtcdPut(context.TODO(), key, string(value))
	if err != nil {
		return err
	}

	return nil
}

func deletePolicy(client *etcd.EtcdClient, ccnp *cv2.CiliumNetworkPolicy) error {
	if ccnp == nil {
		return nil
	}

	key := path.Join(kvstore.PolicyPrefix, ccnp.Name)
	err := client.EtcdDelete(context.TODO(), key)
	if err != nil {
		return err
	}
	return nil
}

func getPolicyLabels(name, uid string) cl.LabelArray {
	return cl.LabelArray{
		cl.NewLabel(k8sConst.PolicyLabelName, name, cl.LabelSourceK8s),
		cl.NewLabel(k8sConst.PolicyLabelUID, uid, cl.LabelSourceK8s),
		cl.NewLabel(k8sConst.PolicyLabelDerivedFrom, cu.ResourceTypeCiliumClusterwideNetworkPolicy, cl.LabelSourceK8s),
	}
}

func UpdateLabels(client *etcd.EtcdClient, identity uint16, lbls []string) {
	key := path.Join(kvstore.IdentityPrefix, strconv.Itoa(int(identity)))

	// 1. Parse labels => LabelMap
	lblMap := cl.NewLabelsFromModel(lbls)

	// 2. Stringify
	value := string(lblMap.SortedList())

	// 3. Push it to etcd
	err := client.EtcdPut(context.TODO(), key, value)
	if err != nil {
		log.Err(err.Error())
	}
}

func UpdateAnnotations(client *etcd.EtcdClient, nodeName string, annotations map[string]string) {
	node := getNodeEntity(client, nodeName)
	if node == nil {
		log.Err(fmt.Sprintf("Cannot fetch information about VM %s from ETCD", nodeName))
		return
	}

	node.Annotations = annotations
	updateNodeEntity(client, node)
}

func NodeRegisterWatcherInit(client *etcd.EtcdClient, idMap map[string]uint16) {
	handlerFunc := func(key string, obj interface{}) {
		if node, ok := obj.(*cnt.Node); !ok {
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

func handleNodeRegister(client *etcd.EtcdClient, idMap map[string]uint16, node *cnt.Node) {
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
		for _, lbl := range labels {
			node.Labels[lbl.Key] = lbl.Value
		}

		node.Cluster = DefaultClusterName
		node.ClusterID = DefaultClusterID
		node.IPAddresses = nil
		node.IPv4AllocCIDR = nil
		node.IPv6AllocCIDR = nil

		// 3. Update Cache
		nodeInfo, ok := NodeCache[node.Name]
		if ok {
			nodeInfo.Initialized = false
		} else {
			nodeInfo := NodeInfo{}
			nodeInfo.Labels = labels
			NodeCache[node.Name] = &nodeInfo
		}

		// 4. Push it to etcd
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
		if nodeInfo, ok := NodeCache[node.Name]; ok && !nodeInfo.Initialized {
			updateNodeEntity(client, node.ToCiliumNode())
			updateNodeIPs(client, node)
			nodeInfo.Initialized = true
		}
	}
}

func setDefaultNodeLabels(node *cnt.Node) {
	if node != nil {
		if node.Labels == nil {
			node.Labels = map[string]string{}
		}
		node.Labels[PodNameLabel] = node.Name
		node.Labels[PodNamespaceLabel] = DefaultNamespace
	}
}

func getNodeEntity(client *etcd.EtcdClient, node string) *cv2.CiliumNode {
	var ciliumNode cv2.CiliumNode

	key := path.Join(kvstore.NodePrefix, node)

	resp, err := client.EtcdGet(context.TODO(), key)
	if err != nil {
		log.Err(err.Error())
		return nil
	}

	err = json.Unmarshal([]byte(resp[key]), &ciliumNode)
	if err != nil {
		log.Err(err.Error())
		return nil
	}

	return &ciliumNode
}

func updateNodeEntity(client *etcd.EtcdClient, node *cv2.CiliumNode) {
	key := path.Join(kvstore.NodePrefix, node.Name)

	marshaledEntry, err := json.Marshal(node)
	if err != nil {
		log.Err(fmt.Sprintf("Unable to JSON marshal entry %#v", node))
		log.Err(err.Error())
	}

	err = client.EtcdPut(context.TODO(), key, string(marshaledEntry))
	if err != nil {
		log.Err(err.Error())
	}
}

func updateNodeIPs(client *etcd.EtcdClient, node *cnt.Node) {
	for _, addr := range node.IPAddresses {
		key := path.Join(kvstore.IPCachePrefix, DefaultNamespace, addr.IP.String())

		// 1. Prepare IPIdentityPair
		entry := ci.IPIdentityPair{
			IP:           addr.IP,
			Metadata:     "",
			HostIP:       node.GetK8sNodeIP(),
			ID:           ci.NumericIdentity(node.NodeIdentity),
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

func getNodeLabels(client *etcd.EtcdClient, identity uint16) cl.LabelArray {
	labelArr := cl.LabelArray{}
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
		label := cl.ParseLabel(lbl)
		labelArr = append(labelArr, label)
	}

	return labelArr
}

func nodeRegisterUnmarshal(bytes []byte) (interface{}, error) {
	var node cnt.Node

	err := json.Unmarshal(bytes, &node)
	if err != nil {
		log.Err(err.Error())
		return nil, err
	}

	return &node, nil
}
