package core

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"reflect"
	"strconv"

	kl "github.com/kubearmor/KVMService/src/common"
	ct "github.com/kubearmor/KVMService/src/constants"
	kg "github.com/kubearmor/KVMService/src/log"
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

func (dm *KVMS) convertLabelsToStr(labelStr map[string]string) string {
	var label string
	for k, v := range labelStr {
		label = k + "=" + v
	}
	return label
}

func (dm *KVMS) terminateClientConnection(clientIdentity uint16) {
	kg.Printf("Terminating client connection for %d", clientIdentity)

	terminateConnection := tp.KubeArmorHostPolicyEventWithIdentity{}
	terminateConnection.CloseConnection = true
	terminateConnection.Identity = clientIdentity
	terminateConnection.Err = errors.New("err-identity-removed")

	ks.PolicyChan <- terminateConnection
}

func (dm *KVMS) UnMapLabelIdentity(identity uint16, ewName, label string) {
	delete(dm.MapIdentityToEWName, identity)
	delete(dm.MapEWNameToIdentity, ewName)
	delete(dm.MapIdentityToLabel, identity)

	err := dm.EtcdClient.EtcdDelete(context.TODO(), ct.KvmOprIdentityToEWName+strconv.FormatUint(uint64(identity), 10))
	if err != nil {
		log.Fatal(err)
	}

	err = dm.EtcdClient.EtcdDelete(context.TODO(), ct.KvmOprEWNameToIdentity+ewName)
	if err != nil {
		log.Fatal(err)
	}

	err = dm.EtcdClient.EtcdDelete(context.TODO(), ct.KvmOprIdentityToLabel+strconv.FormatUint(uint64(identity), 10))
	if err != nil {
		log.Fatal(err)
	}

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
	err = dm.EtcdClient.EtcdPut(context.TODO(), ct.KvmOprLabelToIdentities+label, string(mapStr))
	if err != nil {
		log.Fatal(err)
	}
}

func (dm *KVMS) UpdateIdentityLabelsMap(identity uint16, label string) {
	kg.Printf("Updating identity to label map identity:%d label:%s", identity, label)
	dm.MapIdentityToLabel[identity] = label
	err := dm.EtcdClient.EtcdPut(context.TODO(), ct.KvmOprIdentityToLabel+strconv.FormatUint(uint64(identity), 10), label)
	if err != nil {
		log.Fatal(err)
	}

	_, found := find(dm.MapLabelToIdentities[label], identity)
	if !found {
		dm.MapLabelToIdentities[label] = append(dm.MapLabelToIdentities[label], identity)
		mapStr, _ := json.Marshal(dm.MapLabelToIdentities[label])
		err := dm.EtcdClient.EtcdPut(context.TODO(), ct.KvmOprLabelToIdentities+label, string(mapStr))
		if err != nil {
			log.Fatal(err)
		}
	}
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
				log.Fatal(err)
			}

			kg.Printf("Mappings ewName to identity => %v", dm.MapEWNameToIdentity)
			err = dm.EtcdClient.EtcdPut(context.TODO(), ct.KvmOprEWNameToIdentity+name, strconv.FormatUint(uint64(identity), 10))
			if err != nil {
				log.Fatal(err)
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

func (dm *KVMS) GetVirtualMachineLabel(identity uint16) string {
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
	var vmList string

	for _, vm := range dm.VirtualMachineSecurityPolicies {
		vmList = vmList + "\n[" + vm.Metadata.Name + "]"
	}

	return vmList
}

func (dm *KVMS) HandleVm(event tp.KubeArmorVirtualMachinePolicyEvent) {

	secPolicy := tp.VirtualMachineSecurityPolicy{}
	kg.Print("Received Virtual Machine policy request!!!")

	secPolicy.Metadata.NodeSelector.MatchLabels = event.Object.Metadata.Labels
	secPolicy.Metadata.Name = event.Object.Metadata.Name

	// update a security policy into the policy list

	if event.Type == "ADDED" {
		kg.Printf("New Virtual Machine CRD is configured! => %s", secPolicy.Metadata.Name)
		if !kl.ContainsElement(dm.VirtualMachineSecurityPolicies, secPolicy) {
			dm.VirtualMachineSecurityPolicies = append(dm.VirtualMachineSecurityPolicies, secPolicy)
			identity := dm.GenerateVirtualMachineIdentity(secPolicy.Metadata.Name, secPolicy.Metadata.NodeSelector.MatchLabels)
			kg.Printf("Generated the identity(%s) for this CRD:%d", secPolicy.Metadata.Name, identity)
			//gs.GenerateEWInstallationScript(dm.Port, dm.ClusterIp, secPolicy.Metadata.Name, identity)
			dm.UpdateIdentityLabelsMap(identity, dm.convertLabelsToStr(secPolicy.Metadata.NodeSelector.MatchLabels))
		}
	} else if event.Type == "MODIFIED" {
		kg.Printf("Virtual Machine CRD is Modified! => %s", secPolicy.Metadata.Name)
		for idx, policy := range dm.VirtualMachineSecurityPolicies {
			if policy.Metadata.Name == secPolicy.Metadata.Name {
				dm.VirtualMachineSecurityPolicies[idx] = secPolicy
				identity := dm.GetEWIdentityFromName(secPolicy.Metadata.Name)
				dm.UpdateIdentityLabelsMap(identity, dm.convertLabelsToStr(secPolicy.Metadata.NodeSelector.MatchLabels))
				break
			}
		}
	} else if event.Type == "DELETED" {
		kg.Printf("Virtual Machine CRD is Deleted! => %s", secPolicy.Metadata.Name)
		for idx, policy := range dm.VirtualMachineSecurityPolicies {
			if reflect.DeepEqual(secPolicy, policy) {
				dm.VirtualMachineSecurityPolicies = append(dm.VirtualMachineSecurityPolicies[:idx], dm.VirtualMachineSecurityPolicies[idx+1:]...)
				identity := dm.GetEWIdentityFromName(secPolicy.Metadata.Name)
				kg.Printf("Before: %+v\n", dm.MapLabelToIdentities[dm.convertLabelsToStr(secPolicy.Metadata.NodeSelector.MatchLabels)])
				dm.UnMapLabelIdentity(identity, secPolicy.Metadata.Name, dm.convertLabelsToStr(secPolicy.Metadata.NodeSelector.MatchLabels))
				kg.Printf("After: %+v\n", dm.MapLabelToIdentities[dm.convertLabelsToStr(secPolicy.Metadata.NodeSelector.MatchLabels)])
				dm.terminateClientConnection(identity)
				break
			}
		}
	}
}
