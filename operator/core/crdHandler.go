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

	kl "github.com/kubearmor/KVMService/operator/common"
	gs "github.com/kubearmor/KVMService/operator/genscript"
	kg "github.com/kubearmor/KVMService/operator/log"
	tp "github.com/kubearmor/KVMService/operator/types"
    ct "github.com/kubearmor/KVMService/operator/constants"
)

// UpdateExternalWorkloadSecurityPolicies Function
func (dm *KVMSOperator) UpdateExternalWorkloadSecurityPolicies() {

	dm.ExternalWorkloadSecurityPoliciesLock.Lock()
	defer dm.ExternalWorkloadSecurityPoliciesLock.Unlock()

	secPolicies := []tp.ExternalWorkloadSecurityPolicy{}

	for _, policy := range dm.ExternalWorkloadSecurityPolicies {
		// TODO:
		secPolicies = append(secPolicies, policy)
	}
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

func (dm *KVMSOperator) GenerateExternalWorkloadIdentity(name string, labels map[string]string) uint16 {
	for {
		identity := uint16(rand.Uint32())
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

func (dm *KVMSOperator) GetExternalWorkloadIdentities(name string) []uint16 {
	return dm.MapLabelToIdentities[name]
}

func (dm *KVMSOperator) GetExternalWorkloadLabel(identity uint16) string {
	return dm.MapIdentityToLabel[identity]
}

func (dm *KVMSOperator) GetExternalWorkLoadAllLabels() []string {
	var externalWorkloadLabels []string
	for label, _ := range dm.MapLabelToIdentities {
		externalWorkloadLabels = append(externalWorkloadLabels, label)
	}
	return externalWorkloadLabels
}

// WatchExternalWorkloadSecurityPolicies Function
func (dm *KVMSOperator) WatchExternalWorkloadSecurityPolicies() {
	for {
		if !K8s.CheckCustomResourceDefinition(ct.KewCRDName) {
			time.Sleep(time.Second * 1)
			continue
		}

		if resp := K8s.WatchK8sExternalWorkloadSecurityPolicies(); resp != nil {
			defer resp.Body.Close()

			decoder := json.NewDecoder(resp.Body)
			for {
				event := tp.K8sKubeArmorExternalWorkloadPolicyEvent{}
				if err := decoder.Decode(&event); err == io.EOF {
					break
				} else if err != nil {
					break
				}

				if event.Object.Status.Status != "" && event.Object.Status.Status != "OK" {
					continue
				}
				dm.ExternalWorkloadSecurityPoliciesLock.Lock()

				secPolicy := tp.ExternalWorkloadSecurityPolicy{}
				kg.Print("Recieved external workload policy request!!!")

				secPolicy.Metadata.NodeSelector.MatchLabels = event.Object.Metadata.Labels
				secPolicy.Metadata.Name = event.Object.Metadata.Name

				// update a security policy into the policy list

				if event.Type == "ADDED" {
					kg.Printf("New External Workload CRD is configured! => %s", secPolicy.Metadata.Name)
					if !kl.ContainsElement(dm.ExternalWorkloadSecurityPolicies, secPolicy) {
						dm.ExternalWorkloadSecurityPolicies = append(dm.ExternalWorkloadSecurityPolicies, secPolicy)
						identity := dm.GenerateExternalWorkloadIdentity(secPolicy.Metadata.Name, secPolicy.Metadata.NodeSelector.MatchLabels)
						kg.Printf("Generated the identity(%s) for this CRD:%d", secPolicy.Metadata.Name, identity)
						gs.GenerateEWInstallationScript(dm.Port, dm.ClusterIp, secPolicy.Metadata.Name, identity)
						dm.UpdateIdentityLabelsMap(identity, dm.convertLabelsToStr(secPolicy.Metadata.NodeSelector.MatchLabels))
					}
				} else if event.Type == "MODIFIED" {
					kg.Printf("External Workload CRD is Modified! => %s", secPolicy.Metadata.Name)
					for idx, policy := range dm.ExternalWorkloadSecurityPolicies {
						if policy.Metadata.Name == secPolicy.Metadata.Name {
							dm.ExternalWorkloadSecurityPolicies[idx] = secPolicy
							identity := dm.GetEWIdentityFromName(secPolicy.Metadata.Name)
							dm.UpdateIdentityLabelsMap(identity, dm.convertLabelsToStr(secPolicy.Metadata.NodeSelector.MatchLabels))
							break
						}
					}
				} else if event.Type == "DELETED" {
					kg.Printf("External Workload CRD is Deleted! => %s", secPolicy.Metadata.Name)
					for idx, policy := range dm.ExternalWorkloadSecurityPolicies {
						if reflect.DeepEqual(secPolicy, policy) {
							dm.ExternalWorkloadSecurityPolicies = append(dm.ExternalWorkloadSecurityPolicies[:idx], dm.ExternalWorkloadSecurityPolicies[idx+1:]...)
							identity := dm.GetEWIdentityFromName(secPolicy.Metadata.Name)
							kg.Printf("Before: %+v\n", dm.MapLabelToIdentities[dm.convertLabelsToStr(secPolicy.Metadata.NodeSelector.MatchLabels)])
							dm.UnMapLabelIdentity(identity, secPolicy.Metadata.Name, dm.convertLabelsToStr(secPolicy.Metadata.NodeSelector.MatchLabels))
							kg.Printf("After: %+v\n", dm.MapLabelToIdentities[dm.convertLabelsToStr(secPolicy.Metadata.NodeSelector.MatchLabels)])
							break
						}
					}
				}
				dm.ExternalWorkloadSecurityPoliciesLock.Unlock()

				// apply security policies to a external workload
				dm.UpdateExternalWorkloadSecurityPolicies()
			}
		}
	}
}
