// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package core

import (
	"context"
	"encoding/json"
	"fmt"
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
)

func (dm *KVMSOperator) PassOverToKVMS(secPolicies []tp.HostSecurityPolicy) {
	fmt.Println(secPolicies)
}

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

func (dm *KVMSOperator) convertLabelsToStr(labels map[string]string) string {
	var label string
	for k, v := range labels {
		label = k + "=" + v
	}
	return label
}

func (dm *KVMSOperator) updateEtcdIdentityLabelsMap(identity uint16, labels map[string]string) {
	dm.EtcdClient.EtcdPut(context.TODO(), "/externalworkloads/" + strconv.FormatUint(uint64(identity), 10),
		dm.convertLabelsToStr(labels))
}

func (dm *KVMSOperator) UpdateIdentityLabelsMap(identity uint16, labels string) {
	fmt.Println("Updating identity to labels map")
	dm.MapIdentityToLabel[identity] = labels
	dm.MapLabelToIdentity[labels] = append(dm.MapLabelToIdentity[labels], identity)

}

func (dm *KVMSOperator) GenerateExternalWorkloadIdentity(name string, labels map[string]string) uint16 {
	label := dm.convertLabelsToStr(labels)
	for {
		identity := uint16(rand.Uint32())
		if dm.MapIdentityToEWName[identity] == "" {
			dm.MapIdentityToEWName[identity] = name
			dm.MapEWNameToIdentity[name] = identity
			dm.UpdateIdentityLabelsMap(identity, label)
			return identity
		}
	}
}

func (dm *KVMSOperator) GetEWIdentityFromName(name string) uint16 {
	return dm.MapEWNameToIdentity[name]
}

func (dm *KVMSOperator) GetExternalWorkloadIdentities(name string) []uint16 {
	return dm.MapLabelToIdentity[name]
}

func (dm *KVMSOperator) GetExternalWorkloadLabel(identity uint16) string {
	return dm.MapIdentityToLabel[identity]
}

func (dm *KVMSOperator) GetExternalWorkLoadAllLabels() []string {
	var externalWorkloadLabels []string
	for label, _ := range dm.MapLabelToIdentity {
		externalWorkloadLabels = append(externalWorkloadLabels, label)
	}
	return externalWorkloadLabels
}

// WatchExternalWorkloadSecurityPolicies Function
func (dm *KVMSOperator) WatchExternalWorkloadSecurityPolicies() {
	for {
		if !K8s.CheckCustomResourceDefinition("kubearmorexternalworkloads") {
			time.Sleep(time.Second * 1)
			continue
		}

		if resp := K8s.WatchK8sExternalWorkloadSecurityPolicies(); resp != nil {
			defer resp.Body.Close()

			//kg.Print("Watching ExternalWorkload policies")
			decoder := json.NewDecoder(resp.Body)
			for {
				event := tp.K8sKubeArmorExternalWorkloadPolicyEvent{}
				if err := decoder.Decode(&event); err == io.EOF {
					kg.Print("No CRDS")
					break
				} else if err != nil {
					kg.Print("Err: no CRDs")
					break
				}

				if event.Object.Status.Status != "" && event.Object.Status.Status != "OK" {
					continue
				}
				kg.Print("Recieved the CRD")

				dm.ExternalWorkloadSecurityPoliciesLock.Lock()

				// create a security policy

				secPolicy := tp.ExternalWorkloadSecurityPolicy{}
				fmt.Println("External workload policy configured", secPolicy)

				secPolicy.Metadata.NodeSelector.MatchLabels = event.Object.Metadata.Labels
				secPolicy.Metadata.Name = event.Object.Metadata.Name

				// update a security policy into the policy list

				if event.Type == "ADDED" {
					if !kl.ContainsElement(dm.ExternalWorkloadSecurityPolicies, secPolicy) {
						dm.ExternalWorkloadSecurityPolicies = append(dm.ExternalWorkloadSecurityPolicies, secPolicy)
					}
					identity := dm.GenerateExternalWorkloadIdentity(secPolicy.Metadata.Name, secPolicy.Metadata.NodeSelector.MatchLabels)
					log.Print("Generated the identity for this CRD:", secPolicy.Metadata.Name, identity)
					gs.GenerateEWInstallationScript(dm.Port, dm.ClusterIp, secPolicy.Metadata.Name, identity)
					dm.updateEtcdIdentityLabelsMap(identity, secPolicy.Metadata.NodeSelector.MatchLabels)

					// TODO: Handle this map of identity to grpc connection seperate context
					//dm.MapExternalWorkloadConnIdentity[identity] = conn
				} else if event.Type == "MODIFIED" {
					for idx, policy := range dm.ExternalWorkloadSecurityPolicies {
						if policy.Metadata.Name == secPolicy.Metadata.Name {
							dm.ExternalWorkloadSecurityPolicies[idx] = secPolicy
							identity := dm.GetEWIdentityFromName(secPolicy.Metadata.Name)
							dm.updateEtcdIdentityLabelsMap(identity, secPolicy.Metadata.NodeSelector.MatchLabels)
							break
						}
					}
				} else if event.Type == "DELETED" {
					for idx, policy := range dm.ExternalWorkloadSecurityPolicies {
						if reflect.DeepEqual(secPolicy, policy) {
							dm.ExternalWorkloadSecurityPolicies = append(dm.ExternalWorkloadSecurityPolicies[:idx], dm.ExternalWorkloadSecurityPolicies[idx+1:]...)
							identity := dm.GetEWIdentityFromName(secPolicy.Metadata.Name)
							err := dm.EtcdClient.EtcdDelete(context.TODO(), strconv.FormatUint(uint64(identity), 10))
							if err != nil {
								log.Fatal(err)
							}
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
