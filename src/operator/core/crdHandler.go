// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package core

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"reflect"
	"strconv"
	"time"

	kl "github.com/kubearmor/KVMService/src/common"
	ct "github.com/kubearmor/KVMService/src/constants"
	kg "github.com/kubearmor/KVMService/src/log"
	tp "github.com/kubearmor/KVMService/src/types"
)

// UpdateVirtualMachineSecurityPolicies Function
func (dm *KVMSOperator) UpdateVirtualMachineSecurityPolicies() {

	dm.VirtualMachineSecurityPoliciesLock.Lock()
	defer dm.VirtualMachineSecurityPoliciesLock.Unlock()

	/* Unused code
	secPolicies := []tp.ExternalWorkloadSecurityPolicy{}

	for _, policy := range dm.VirtualMachineSecurityPolicies {
		// TODO:
		secPolicies = append(secPolicies, policy)
	}
	*/
}
func Find(identities []uint16, identity uint16) (int, bool) {
	for i, item := range identities {
		if item == identity {
			return i, true
		}
	}
	return -1, false
}

func (dm *KVMSOperator) convertLabelsToStr(labelStr map[string]string) string {
	var label string
	for k, v := range labelStr {
		label = k + "=" + v
	}
	return label
}

func (dm *KVMSOperator) UnMapLabelIdentity(identity uint16, ewName, label string) {
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

func (dm *KVMSOperator) UpdateIdentityLabelsMap(identity uint16, label string) {
	kg.Printf("Updating identity to label map identity:%d label:%s", identity, label)
	dm.MapIdentityToLabel[identity] = label
	err := dm.EtcdClient.EtcdPut(context.TODO(), ct.KvmOprIdentityToLabel+strconv.FormatUint(uint64(identity), 10), label)
	if err != nil {
		log.Fatal(err)
	}

	_, found := Find(dm.MapLabelToIdentities[label], identity)
	if !found {
		dm.MapLabelToIdentities[label] = append(dm.MapLabelToIdentities[label], identity)
		mapStr, _ := json.Marshal(dm.MapLabelToIdentities[label])
		err := dm.EtcdClient.EtcdPut(context.TODO(), ct.KvmOprLabelToIdentities+label, string(mapStr))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (dm *KVMSOperator) GenerateVirtualMachineIdentity(name string, labels map[string]string) uint16 {
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

func (dm *KVMSOperator) GetEWIdentityFromName(name string) uint16 {
	return dm.MapEWNameToIdentity[name]
}

func (dm *KVMSOperator) GetVirtualMachineIdentities(name string) []uint16 {
	return dm.MapLabelToIdentities[name]
}

func (dm *KVMSOperator) GetVirtualMachineLabel(identity uint16) string {
	return dm.MapIdentityToLabel[identity]
}

func (dm *KVMSOperator) GetVirtualMachineAllLabels() []string {
	var VirtualMachineLabels []string
	for label := range dm.MapLabelToIdentities {
		VirtualMachineLabels = append(VirtualMachineLabels, label)
	}
	return VirtualMachineLabels
}

// WatchVirtualMachineSecurityPolicies Function
func (dm *KVMSOperator) WatchVirtualMachineSecurityPolicies() {
	for {
		if !K8s.CheckCustomResourceDefinition(ct.KvmCRDName) {
			time.Sleep(time.Second * 1)
			continue
		}

		if resp := K8s.WatchK8sVirtualMachineSecurityPolicies(); resp != nil {
			defer resp.Body.Close()

			decoder := json.NewDecoder(resp.Body)
			for {
				event := tp.K8sKubeArmorVirtualMachinePolicyEvent{}
				if err := decoder.Decode(&event); err == io.EOF {
					break
				} else if err != nil {
					break
				}

				if event.Object.Status.Status != "" && event.Object.Status.Status != "OK" {
					continue
				}
				dm.VirtualMachineSecurityPoliciesLock.Lock()

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
							break
						}
					}
				}
				dm.VirtualMachineSecurityPoliciesLock.Unlock()

				// apply security policies to a Virtual Machine
				dm.UpdateVirtualMachineSecurityPolicies()
			}
		}
	}
}
