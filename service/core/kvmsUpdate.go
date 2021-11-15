// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package core

import (
	"encoding/json"
	"io"
	"log"

	//"sort"
	"strings"

	//"math/rand"
	"context"
	"strconv"
	"time"

	kl "github.com/kubearmor/KVMService/service/common"
	ct "github.com/kubearmor/KVMService/service/constants"
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

func (dm *KVMS) mGetAllEtcdEWLabels() {
	kg.Print("Getting the External workload labels from ETCD")

	etcdLabels, err := dm.EtcdClient.EtcdGet(context.TODO(), ct.KvmOprLabelToIdentities)
	if err != nil {
		log.Fatal(err)
		return
	}

	for key, value := range etcdLabels {
		s := strings.Split(key, "/")
		identity := s[len(s)-1]
		dm.MapEtcdEWIdentityLabels[identity] = value
		dm.EtcdEWLabels = append(dm.EtcdEWLabels, value)

		idNum, _ := strconv.ParseUint(identity, 0, 16)
		_, found := Find(dm.MapLabelToIdentities[value], uint16(idNum))
		if !found {
			dm.MapLabelToIdentities[value] = append(dm.MapLabelToIdentities[value], uint16(idNum))
		}
	}
}

func (dm *KVMS) GetAllEtcdEWLabels() {
	kg.Print("Getting the External workload labels from ETCD")
	etcdLabels, err := dm.EtcdClient.EtcdGet(context.TODO(), ct.KvmOprLabelToIdentities)
	if err != nil {
		log.Fatal(err)
		return
	}

	for key, _ := range etcdLabels {
		s := strings.Split(key, "/")
		label := s[len(s)-1]
		dm.EtcdEWLabels = append(dm.EtcdEWLabels, label)
	}

	for _, label := range dm.EtcdEWLabels {
		data, err := dm.EtcdClient.EtcdGetRaw(context.TODO(), ct.KvmOprLabelToIdentities+label)
		if err != nil {
			log.Fatal(err)
			return
		}

		for _, ev := range data.Kvs {
			var arr []uint16
			err := json.Unmarshal(ev.Value, &arr)
			if err != nil {
				log.Fatal(err)
				return
			}
			s := strings.Split(string(ev.Key), "/")
			dm.MapLabelToIdentities[s[len(s)-1]] = arr
		}
	}
}

func (dm *KVMS) PassOverToKVMSAgent(event tp.K8sKubeArmorHostPolicyEvent, identities []uint16) {
	eventWithIdentity := tp.K8sKubeArmorHostPolicyEventWithIdentity{}

	eventWithIdentity.Event = event
	eventWithIdentity.CloseConnection = false
	for _, identity := range identities {
		eventWithIdentity.Identity = identity
		kg.Printf("Sending the event towards the KVMAgent of identity:%v\n", identity)
		ks.PolicyChan <- eventWithIdentity
	}
}

func (dm *KVMS) GetIdentityFromLabelPool(label string) []uint16 {
	kg.Printf("Getting the identity from the pool => label:%s\n", label)
	return dm.MapLabelToIdentities[label]
}

// ================================= //
// == Host Security Policy Update == //
// ================================= //
// UpdateHostSecurityPolicies Function
func (dm *KVMS) UpdateHostSecurityPolicies(event tp.K8sKubeArmorHostPolicyEvent) {
	var identities []uint16
	var labels []string
	secPolicy := tp.HostSecurityPolicy{}

	secPolicy.Metadata = map[string]string{}
	secPolicy.Metadata["policyName"] = event.Object.Metadata.Name

	if err := kl.Clone(event.Object.Spec, &secPolicy.Spec); err != nil {
		log.Fatal("Failed to clone a spec")
	}

	for k, v := range secPolicy.Spec.NodeSelector.MatchLabels {
		labels = append(labels, k+"="+v)
	}

	dm.GetAllEtcdEWLabels()
	if dm.EtcdEWLabels == nil {
		kg.Err("No etcd keys")
		return
	}

	if kl.MatchIdentities(labels, dm.EtcdEWLabels) {
		for _, label := range labels {
			identities = dm.GetIdentityFromLabelPool(label)
			kg.Printf("External workload CRD matched with policy identity:%+v label:%s\n", identities, label)
			if len(identities) > 0 {
				dm.PassOverToKVMSAgent(event, identities)
			}
		}
	}
}

// WatchHostSecurityPolicies Function
func (dm *KVMS) WatchHostSecurityPolicies() {
	for {
		if !K8s.CheckCustomResourceDefinition(ct.KhpCRDName) {
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

				kg.Print("Host policy got detected")
				dm.UpdateHostSecurityPolicies(event)
			}
		}
	}
}
