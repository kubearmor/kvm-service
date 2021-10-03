// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package core

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

	//"math/rand"
	"context"
	"strconv"
	"time"

	kl "github.com/kubearmor/KVMService/service/common"
	kg "github.com/kubearmor/KVMService/service/log"
	ks "github.com/kubearmor/KVMService/service/server"
	tp "github.com/kubearmor/KVMService/service/types"
)

func Find(slice []uint16, val uint16) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func (dm *KVMS) GetAllEtcdEWLabels() {
	fmt.Println("Getting the External workload labels from ETCD")

	etcdLabels, err := dm.EtcdClient.EtcdGet(context.TODO(), "/externalworkloads")
	if err != nil {
		log.Fatal(err)
		return
	}

	for key, value := range etcdLabels {
		s := strings.Split(key, "/")
		identity := s[len(s)-1]
		dm.MapEtcdEWIdentityLabels[identity] = value
		dm.EtcdEWLabels = append(dm.EtcdEWLabels, value)

		idNum, _ := strconv.ParseUint(identity, 0, 10)
		_, found := Find(dm.MapLabelToIdentity[value], uint16(idNum))
		if !found {
			dm.MapLabelToIdentity[value] = append(dm.MapLabelToIdentity[value], uint16(idNum))
		}
	}

	fmt.Println("MDEBUG:", dm.EtcdEWLabels)
	fmt.Println("MDEBUG:", dm.MapEtcdEWIdentityLabels)
	fmt.Println("MDEBUG:", dm.MapLabelToIdentity)
}

// ================================= //
// == Host Security Policy Update == //
// ================================= //
func (dm *KVMS) PassOverToKVMSAgent(event tp.K8sKubeArmorHostPolicyEvent, identities []uint16) {
	eventWithIdentity := tp.K8sKubeArmorHostPolicyEventWithIdentity{}
	fmt.Println(event, identities)

	eventWithIdentity.Event = event
	eventWithIdentity.CloseConnection = false
	for _, identity := range identities {
		eventWithIdentity.Identity = identity
		ks.PolicyChan <- eventWithIdentity
	}
}

func (dm *KVMS) GetIdentityFromLabelPool(label string) []uint16 {
	fmt.Println(label)
	return dm.MapLabelToIdentity[label]
}

// UpdateHostSecurityPolicies Function
func (dm *KVMS) UpdateHostSecurityPolicies(event tp.K8sKubeArmorHostPolicyEvent, labels []string) {
	var identities []uint16
	dm.GetAllEtcdEWLabels()

	if dm.EtcdEWLabels == nil {
		fmt.Println("No etcd keys")
		return
	}

	if kl.MatchIdentities(labels, dm.EtcdEWLabels) {
		for _, label := range labels {
			identities = dm.GetIdentityFromLabelPool(label)
			fmt.Println("External workload CRD matched with policy")
			if len(identities) > 0 {
				dm.PassOverToKVMSAgent(event, identities)
			}
		}
	}
}

// WatchHostSecurityPolicies Function
func (dm *KVMS) WatchHostSecurityPolicies() {
	for {
		if !K8s.CheckCustomResourceDefinition("kubearmorhostpolicies") {
			time.Sleep(time.Second * 1)
			continue
		}

		if resp := K8s.WatchK8sHostSecurityPolicies(); resp != nil {
			defer resp.Body.Close()

			decoder := json.NewDecoder(resp.Body)
			for {
				event := tp.K8sKubeArmorHostPolicyEvent{}
				if err := decoder.Decode(&event); err == io.EOF {
					break
				} else if err != nil {
					break
				}

				if event.Object.Status.Status != "" && event.Object.Status.Status != "OK" {
					continue
				}

				dm.HostSecurityPoliciesLock.Lock()

				// create a host security policy
				kg.Print("Host policy got detected")

				secPolicy := tp.HostSecurityPolicy{}

				secPolicy.Metadata = map[string]string{}
				secPolicy.Metadata["policyName"] = event.Object.Metadata.Name

				if err := kl.Clone(event.Object.Spec, &secPolicy.Spec); err != nil {
					log.Fatal("Failed to clone a spec")
				}

				// add identities

				secPolicy.Spec.NodeSelector.Identities = []string{}

				for k, v := range secPolicy.Spec.NodeSelector.MatchLabels {
					secPolicy.Spec.NodeSelector.Identities = append(secPolicy.Spec.NodeSelector.Identities, k+"="+v)
				}

				sort.Slice(secPolicy.Spec.NodeSelector.Identities, func(i, j int) bool {
					return secPolicy.Spec.NodeSelector.Identities[i] < secPolicy.Spec.NodeSelector.Identities[j]
				})

				dm.HostSecurityPoliciesLock.Unlock()

				//dm.Logger.Printf("Detected a Host Security Policy (%s/%s)", strings.ToLower(event.Type), secPolicy.Metadata["policyName"])

				// apply security policies to a host
				log.Print("Host Policy Labels:", secPolicy.Spec.NodeSelector.Identities)
				dm.UpdateHostSecurityPolicies(event, secPolicy.Spec.NodeSelector.Identities)
			}
		}
	}
}
