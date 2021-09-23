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
	//kg "github.com/kubearmor/KVMService/service/log"
	tp "github.com/kubearmor/KVMService/service/types"
)

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
		dm.MapLabelToIdentity[value] = append(dm.MapLabelToIdentity[value], uint16(idNum))
	}

	fmt.Println("MDEBUG:", dm.EtcdEWLabels)
	fmt.Println("MDEBUG:", dm.MapEtcdEWIdentityLabels)
	fmt.Println("MDEBUG:", dm.MapLabelToIdentity)
}

// ================================= //
// == Host Security Policy Update == //
// ================================= //
func (dm *KVMS) PassOverToKVMSAgent(event tp.K8sKubeArmorHostPolicyEvent, conn *ClientConn) {
	if conn != nil {
		fmt.Println(event, conn)
	}
}

func (dm *KVMS) GetConnFromIdentityPool(labels []string) *ClientConn {
	fmt.Println(labels)

	return &ClientConn{conn: nil}
}

// UpdateHostSecurityPolicies Function
func (dm *KVMS) UpdateHostSecurityPolicies(event tp.K8sKubeArmorHostPolicyEvent, labels []string) {
	var conn *ClientConn
	// get node identities
	dm.GetAllEtcdEWLabels()

	if dm.EtcdEWLabels == nil {
		fmt.Println("No etcd keys")
		return
	}

	if kl.MatchIdentities(labels, dm.EtcdEWLabels) {
		conn = dm.GetConnFromIdentityPool(labels)
		fmt.Println("External workload CRD matched with policy")
	}

	// Configure these policies over External workload
	dm.PassOverToKVMSAgent(event, conn)
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
				dm.UpdateHostSecurityPolicies(event, secPolicy.Spec.NodeSelector.Identities)
			}
		}
	}
}
