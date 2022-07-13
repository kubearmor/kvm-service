package core

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	cl "github.com/cilium/cilium/pkg/labels"
	ca "github.com/cilium/cilium/pkg/policy/api"
	ct "github.com/kubearmor/KVMService/src/constants"
	kg "github.com/kubearmor/KVMService/src/log"
	"github.com/kubearmor/KVMService/src/service/cilium"
	"github.com/kubearmor/KVMService/src/service/cilium/kvstore"
	ks "github.com/kubearmor/KVMService/src/service/server"
	tp "github.com/kubearmor/KVMService/src/types"
)

func find(identities []uint16, identity uint16) (int, bool) {
	for i, item := range identities {
		if item == identity {
			return i, true
		}
	}
	return -1, false
}

func (dm *KVMS) convertLabelsToStr(labelStr map[string]string) []string {
	var labels []string
	for k, v := range labelStr {
		label := k + "=" + v
		labels = append(labels, label)
	}
	return labels
}

func (dm *KVMS) terminateClientConnection(clientIdentity uint16) {
	kg.Printf("Terminating client connection for %d", clientIdentity)

	terminateConnection := tp.KubeArmorHostPolicyEventWithIdentity{}
	terminateConnection.CloseConnection = true
	terminateConnection.Identity = clientIdentity
	terminateConnection.Err = errors.New("err-identity-removed")

	ks.PolicyChan <- terminateConnection
}

func (dm *KVMS) UnMapLabelIdentity(identity uint16, ewName string, labels []string) {
	delete(dm.MapIdentityToEWName, identity)
	delete(dm.MapEWNameToIdentity, ewName)
	delete(dm.MapIdentityToLabel, identity)

	err := dm.EtcdClient.EtcdDelete(context.TODO(), ct.KvmOprIdentityToEWName+strconv.FormatUint(uint64(identity), 10))
	if err != nil {
		kg.Err(err.Error())
		return
	}

	err = dm.EtcdClient.EtcdDelete(context.TODO(), ct.KvmOprEWNameToIdentity+ewName)
	if err != nil {
		kg.Err(err.Error())
		return
	}

	err = dm.EtcdClient.EtcdDelete(context.TODO(), ct.KvmOprIdentityToLabel+strconv.FormatUint(uint64(identity), 10))
	if err != nil {
		kg.Err(err.Error())
		return
	}

	for _, label := range labels {
		deleted := false
		identities := dm.MapLabelToIdentities[label]
		if len(identities) > 1 {
			for index, value := range dm.MapLabelToIdentities[label] {
				if value == identity {
					identities[index] = identities[len(identities)-1]
					dm.MapLabelToIdentities[label] = identities[:len(identities)-1]
					break
				}
			}
		} else {
			delete(dm.MapLabelToIdentities, label)
			deleted = true
		}

		// After deleting the identity from the label map
		// update the etcd with updated label to identities map
		if deleted {
			err = dm.EtcdClient.EtcdDelete(context.TODO(), ct.KvmOprLabelToIdentities+label)
		} else {
			mapStr, _ := json.Marshal(dm.MapLabelToIdentities[label])
			err = dm.EtcdClient.EtcdPut(context.TODO(), ct.KvmOprLabelToIdentities+label, string(mapStr))
		}
		if err != nil {
			kg.Err(err.Error())
			return
		}
	}
}

func (dm *KVMS) UpdateIdentityLabelsMap(identity uint16, labels []string) {

	for _, label := range labels {
		kg.Printf("Updating identity to label map identity:%d label:%s", identity, label)
		dm.MapIdentityToLabel[identity] = append(dm.MapIdentityToLabel[identity], label)
		labelsStr, _ := json.Marshal(dm.MapIdentityToLabel[identity])
		err := dm.EtcdClient.EtcdPut(context.TODO(), ct.KvmOprIdentityToLabel+strconv.FormatUint(uint64(identity), 10), string(labelsStr))
		if err != nil {
			kg.Err(err.Error())
			return
		}

		_, found := find(dm.MapLabelToIdentities[label], identity)
		if !found {
			dm.MapLabelToIdentities[label] = append(dm.MapLabelToIdentities[label], identity)
			mapStr, _ := json.Marshal(dm.MapLabelToIdentities[label])
			err := dm.EtcdClient.EtcdPut(context.TODO(), ct.KvmOprLabelToIdentities+label, string(mapStr))
			if err != nil {
				kg.Err(err.Error())
				return
			}
		}
	}

	cilium.UpdateLabels(dm.EtcdClient, identity, labels)
}

func (dm *KVMS) UpdateAnnotations(identity uint16, annotations map[string]string) {
	// Currently VM annotations are only used for Cilium
	cilium.UpdateAnnotations(dm.EtcdClient, dm.MapIdentityToEWName[identity], annotations)
}

func (dm *KVMS) GenerateVirtualMachineIdentity(name string, labels map[string]string) uint16 {
	for {
		identity := uint16(rand.Uint32()) // #nosec
		if dm.MapIdentityToEWName[identity] == "" {
			dm.MapIdentityToEWName[identity] = name
			dm.MapEWNameToIdentity[name] = identity
			kg.Printf("Mappings identity to ewName=> %v", dm.MapIdentityToEWName)
			err := dm.EtcdClient.EtcdPut(context.TODO(), ct.KvmOprIdentityToEWName+strconv.FormatUint(uint64(identity), 10), name)
			if err != nil {
				kg.Err(err.Error())
				return 0
			}

			kg.Printf("Mappings ewName to identity => %v", dm.MapEWNameToIdentity)
			err = dm.EtcdClient.EtcdPut(context.TODO(), ct.KvmOprEWNameToIdentity+name, strconv.FormatUint(uint64(identity), 10))
			if err != nil {
				kg.Err(err.Error())
				return 0
			}

			return identity
		}
	}
}

func (dm *KVMS) GetEWIdentityFromName(name string) uint16 {
	return dm.MapEWNameToIdentity[name]
}

func (dm *KVMS) GetVirtualMachineIdentities(name string) []uint16 {
	return dm.MapLabelToIdentities[name]
}

func (dm *KVMS) GetVirtualMachineLabel(identity uint16) []string {
	return dm.MapIdentityToLabel[identity]
}

func (dm *KVMS) GetVirtualMachineAllLabels() []string {
	var VirtualMachineLabels []string
	for label := range dm.MapLabelToIdentities {
		VirtualMachineLabels = append(VirtualMachineLabels, label)
	}
	return VirtualMachineLabels
}

func (dm *KVMS) ListOnboardedVms() string {

	var vmList []tp.KVMSEndpoint

	for vm, id := range dm.MapEWNameToIdentity {
		labels := dm.MapIdentityToLabel[id]
		endpoint := tp.KVMSEndpoint{
			VMName:    vm,
			Identity:  id,
			Labels:    labels,
			Namespace: "default",
		}
		vmList = append(vmList, endpoint)
	}

	result, _ := json.Marshal(vmList)

	return string(result)
}

func (dm *KVMS) HandleVm(event tp.KubeArmorVirtualMachinePolicyEvent) {

	secPolicy := tp.VirtualMachineSecurityPolicy{}
	kg.Print("Received Virtual Machine policy request!!!")

	secPolicy.Metadata.NodeSelector.MatchLabels = event.Object.Metadata.Labels
	secPolicy.Metadata.Name = event.Object.Metadata.Name

	// update a security policy into the policy list

	if event.Type == "ADDED" {
		kg.Printf("New Virtual Machine CRD is configured! => %s", secPolicy.Metadata.Name)
		if _, ok := dm.MapEWNameToIdentity[secPolicy.Metadata.Name]; !ok {
			identity := dm.GenerateVirtualMachineIdentity(secPolicy.Metadata.Name, secPolicy.Metadata.NodeSelector.MatchLabels)
			kg.Printf("Generated the identity(%s) for this CRD:%d", secPolicy.Metadata.Name, identity)
			dm.UpdateIdentityLabelsMap(identity, dm.convertLabelsToStr(secPolicy.Metadata.NodeSelector.MatchLabels))
		}
	} else if event.Type == "MODIFIED" {
		kg.Printf("Virtual Machine CRD is Modified! => %s", secPolicy.Metadata.Name)
		if identity, ok := dm.MapEWNameToIdentity[secPolicy.Metadata.Name]; ok {
			dm.UpdateAnnotations(identity, event.Object.Metadata.Annotations)
		}
	} else if event.Type == "DELETED" {
		kg.Printf("Virtual Machine CRD is Deleted! => %s", secPolicy.Metadata.Name)
		if identity, ok := dm.MapEWNameToIdentity[secPolicy.Metadata.Name]; ok {
			if ks.IsIdentityServing(strconv.Itoa(int(identity))) == 0 {
				// If kubearmor is connected, close the connection
				dm.terminateClientConnection(identity)
			}
			dm.UnMapLabelIdentity(identity, secPolicy.Metadata.Name, dm.convertLabelsToStr(secPolicy.Metadata.NodeSelector.MatchLabels))
		}
	}
}

func (dm *KVMS) deleteVmLabels(identity uint16, labels []string) {

	kg.Print("Calling deleteVmLabels")

	for _, label := range labels {
		identities := dm.MapLabelToIdentities[label]
		for index, value := range dm.MapLabelToIdentities[label] {
			if value == identity {
				identities[index] = identities[len(identities)-1]
				identities[len(identities)-1] = 0
				identities = identities[:len(identities)-1]
			}
		}
		// After deleting the identity from the label map
		// update the etcd with updated label to identities map
		mapStr, _ := json.Marshal(dm.MapLabelToIdentities[label])
		err := dm.EtcdClient.EtcdPut(context.TODO(), ct.KvmOprLabelToIdentities+label, string(mapStr))
		if err != nil {
			kg.Err(err.Error())
			return
		}

		for index, identityLabel := range dm.MapIdentityToLabel[identity] {
			kg.Printf(identityLabel)
			if identityLabel == label {
				dm.MapIdentityToLabel[identity][index] = ""
				break
			}
		}
	}
}

func (dm *KVMS) HandleVMLabels(event tp.KubeArmorVirtualMachineLabel) string {
	var labelStr string

	identity := dm.GetEWIdentityFromName(event.Name)
	if identity == 0 {
		kg.Warnf("%s is not configured in DB", event.Name)
		return ""
	}

	if event.Type == "LIST" {
		labelStr = ""
		for _, label := range dm.MapIdentityToLabel[identity] {
			labelStr += label + " "
		}
		kg.Printf("The list of assigned labels to VM %s is %s", event.Name, labelStr)
		return labelStr
	} else {
		// To add or remove labels to the identity/vm
		for _, label := range event.Labels {
			if event.Type == "ADD" {
				dm.UpdateIdentityLabelsMap(identity, dm.convertLabelsToStr(label))
			} else {
				// Delete labels from VM
				dm.deleteVmLabels(identity, dm.convertLabelsToStr(label))
			}
		}
	}
	return ""
}

func (dm *KVMS) HandleNetworkPolicyUpdates(req *cilium.NetworkPolicyRequest) {
	cilium.HandlePolicyUpdates(dm.EtcdClient, req)
}

func (dm *KVMS) RestoreFromEtcd() error {
	// 1. Restore MapIdentityToEWName
	id2Ew, err := dm.EtcdClient.EtcdGet(context.Background(), ct.KvmOprIdentityToEWName)
	if err != nil {
		kg.Errf("Failed to retrieve information from etcd key prefix %s", ct.KvmOprIdentityToEWName)
		return err
	}
	for k, v := range id2Ew {
		idStr := strings.TrimPrefix(k, ct.KvmOprIdentityToEWName)
		id, _ := strconv.ParseUint(idStr, 10, 0)
		dm.MapIdentityToEWName[uint16(id)] = v
	}

	// 2. Restore MapEWNameToIdentity
	ew2Id, err := dm.EtcdClient.EtcdGet(context.Background(), ct.KvmOprEWNameToIdentity)
	if err != nil {
		kg.Errf("Failed to retrieve information from etcd key prefix %s", ct.KvmOprEWNameToIdentity)
		return err
	}
	for k, v := range ew2Id {
		ew := strings.TrimPrefix(k, ct.KvmOprEWNameToIdentity)
		id, _ := strconv.ParseUint(v, 10, 0)
		dm.MapEWNameToIdentity[ew] = uint16(id)
	}

	// 3. Restore MapIdentityToLabel
	id2Label, err := dm.EtcdClient.EtcdGet(context.Background(), ct.KvmOprIdentityToLabel)
	if err != nil {
		kg.Errf("Failed to retrieve information from etcd key prefix %s", ct.KvmOprIdentityToLabel)
		return err
	}
	for k, v := range id2Label {
		idStr := strings.TrimPrefix(k, ct.KvmOprIdentityToLabel)
		id, _ := strconv.ParseUint(idStr, 10, 0)
		var labels []string
		err = json.Unmarshal([]byte(v), &labels)
		if err != nil {
			kg.Errf("Failed to parse information from etcd key prefix %s", ct.KvmOprIdentityToLabel)
			return err
		}
		dm.MapIdentityToLabel[uint16(id)] = labels
	}

	// 4. Restore MapLabelToIdentities
	label2Id, err := dm.EtcdClient.EtcdGet(context.Background(), ct.KvmOprLabelToIdentities)
	if err != nil {
		kg.Errf("Failed to retrieve information from etcd key prefix %s", ct.KvmOprLabelToIdentities)
		return err
	}
	for k, v := range label2Id {
		label := strings.TrimPrefix(k, ct.KvmOprLabelToIdentities)
		var ids []uint16
		err = json.Unmarshal([]byte(v), &ids)
		if err != nil {
			kg.Errf("Failed to parse information from etcd key prefix %s", ct.KvmOprLabelToIdentities)
			return err
		}
		dm.MapLabelToIdentities[label] = ids
	}

	// 5. Restore Labels in Cilium Node Cache
	for id, name := range dm.MapIdentityToEWName {
		if labels, ok := dm.MapIdentityToLabel[id]; ok {
			ciliumLabels := cl.NewLabelsFromModel(labels).LabelArray()
			cilium.NodeCache[name] = &cilium.NodeInfo{Labels: ciliumLabels}
		}
	}

	// 6. Restore Policy Count in Cilium Node Cache
	policies, err := dm.EtcdClient.EtcdGet(context.Background(), kvstore.PolicyPrefix)
	if err != nil {
		kg.Errf("Failed to retrieve information from etcd key prefix %s", kvstore.PolicyPrefix)
		return err
	}
	for key, policy := range policies {
		policyName := strings.TrimPrefix(key, (kvstore.PolicyPrefix + "/"))
		regex, _ := regexp.Compile("^00-allow-[a-z0-9-]*-egress-control-plane$")
		if regex.MatchString(policyName) {
			// default egress control plane policy
			continue
		}
		var ccnp ca.Rule
		err = json.Unmarshal([]byte(policy), &ccnp)
		if err != nil {
			kg.Errf("Failed to parse information from etcd key prefix %s", key)
			return err
		}
		if len(ccnp.Egress) > 0 {
			nodeSelector := ccnp.NodeSelector
			for _, nodeInfo := range cilium.NodeCache {
				nodeLabel := nodeInfo.Labels
				if nodeSelector.Matches(nodeLabel) {
					nodeInfo.EgressPolicyCount++
				}
			}
		}
	}

	kg.Print("Restored information from ETCD")
	return nil
}
