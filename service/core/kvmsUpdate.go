// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package core

import (
	"encoding/json"
	"io"
	"sort"
	//"strings"
	"time"
    "fmt"
    "reflect"
	"math/rand"
    "log"

	kl "github.com/kubearmor/KVMService/service/common"
	tp "github.com/kubearmor/KVMService/service/types"
)

// ================================= //
// == Host Security Policy Update == //
// ================================= //

// UpdateHostSecurityPolicies Function
func (dm *KVMS) UpdateHostSecurityPolicies() {
	// get node identities
	externalWorkloadLabels := dm.GetExternalWorkLoadAllLabels()

	dm.HostSecurityPoliciesLock.Lock()
	defer dm.HostSecurityPoliciesLock.Unlock()

	secEWPolicies := []tp.HostSecurityPolicy{}

	for _, policy := range dm.HostSecurityPolicies {
		if kl.MatchIdentities(policy.Spec.NodeSelector.Identities, externalWorkloadLabels) {
			secEWPolicies = append(secEWPolicies, policy)
		}
	}

	// Configure these policies over External workload
	dm.PassOverToKVMS(secEWPolicies)
}

func (dm *KVMS) PassOverToKVMS(secPolicies []tp.HostSecurityPolicy) {
	fmt.Println(secPolicies)
}

// UpdateExternalWorkloadSecurityPolicies Function
func (dm *KVMS) UpdateExternalWorkloadSecurityPolicies() {

	dm.ExternalWorkloadSecurityPoliciesLock.Lock()
	defer dm.ExternalWorkloadSecurityPoliciesLock.Unlock()

	secPolicies := []tp.ExternalWorkloadSecurityPolicy{}

	for _, policy := range dm.ExternalWorkloadSecurityPolicies {
		// TODO:
		secPolicies = append(secPolicies, policy)
	}

	// TODO: Pass to the kubearmor agent on external workload
	//dm.RuntimeEnforcer.UpdateHostSecurityPolicies(secPolicies)
}

func (dm *KVMS) GenerateExternalWorkloadIdentity(label string) {
	for {
		identity := uint16(rand.Uint32())
		if dm.MapIdentityToLabel[identity] == "" {
			dm.MapIdentityToLabel[identity] = label
			dm.MapLabelToIdentity[label] = append(dm.MapLabelToIdentity[label], identity)
			break
		}
	}
}

func (dm *KVMS) GetExternalWorkloadIdentities(label string) []uint16 {
	return dm.MapLabelToIdentity[label]
}

func (dm *KVMS) GetExternalWorkloadLabel(identity uint16) string {
	return dm.MapIdentityToLabel[identity]
}

func (dm *KVMS) GetExternalWorkLoadAllLabels() []string {
	var externalWorkloadLabels []string
	for label, _ := range dm.MapLabelToIdentity {
		externalWorkloadLabels = append(externalWorkloadLabels, label)
	}
	return externalWorkloadLabels
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

				kl.ObjCommaExpandFirstDupOthers(&secPolicy.Spec.Network.MatchProtocols)
				kl.ObjCommaExpandFirstDupOthers(&secPolicy.Spec.Capabilities.MatchCapabilities)

				if secPolicy.Spec.Severity == 0 {
					secPolicy.Spec.Severity = 1 // the lowest severity, by default
				}

				switch secPolicy.Spec.Action {
				case "allow":
					secPolicy.Spec.Action = "Allow"
				case "audit":
					secPolicy.Spec.Action = "Audit"
				case "block":
					secPolicy.Spec.Action = "Block"
				case "":
					secPolicy.Spec.Action = "Block" // by default
				}

				// add identities

				secPolicy.Spec.NodeSelector.Identities = []string{}

				for k, v := range secPolicy.Spec.NodeSelector.MatchLabels {
					secPolicy.Spec.NodeSelector.Identities = append(secPolicy.Spec.NodeSelector.Identities, k+"="+v)
				}

				sort.Slice(secPolicy.Spec.NodeSelector.Identities, func(i, j int) bool {
					return secPolicy.Spec.NodeSelector.Identities[i] < secPolicy.Spec.NodeSelector.Identities[j]
				})

				// add severities, tags, messages, and actions

				if len(secPolicy.Spec.Process.MatchPaths) > 0 {
					for idx, path := range secPolicy.Spec.Process.MatchPaths {
						if path.Severity == 0 {
							if secPolicy.Spec.Process.Severity != 0 {
								secPolicy.Spec.Process.MatchPaths[idx].Severity = secPolicy.Spec.Process.Severity
							} else {
								secPolicy.Spec.Process.MatchPaths[idx].Severity = secPolicy.Spec.Severity
							}
						}

						if len(path.Tags) == 0 {
							if len(secPolicy.Spec.Process.Tags) > 0 {
								secPolicy.Spec.Process.MatchPaths[idx].Tags = secPolicy.Spec.Process.Tags
							} else {
								secPolicy.Spec.Process.MatchPaths[idx].Tags = secPolicy.Spec.Tags
							}
						}

						if len(path.Message) == 0 {
							if len(secPolicy.Spec.Process.Message) > 0 {
								secPolicy.Spec.Process.MatchPaths[idx].Message = secPolicy.Spec.Process.Message
							} else {
								secPolicy.Spec.Process.MatchPaths[idx].Message = secPolicy.Spec.Message
							}
						}

						if len(path.Action) == 0 {
							if len(secPolicy.Spec.Process.Action) > 0 {
								secPolicy.Spec.Process.MatchPaths[idx].Action = secPolicy.Spec.Process.Action
							} else {
								secPolicy.Spec.Process.MatchPaths[idx].Action = secPolicy.Spec.Action
							}
						}
					}
				} else if len(secPolicy.Spec.Process.MatchDirectories) > 0 {
					for idx, dir := range secPolicy.Spec.Process.MatchDirectories {
						if dir.Severity == 0 {
							if secPolicy.Spec.Process.Severity != 0 {
								secPolicy.Spec.Process.MatchDirectories[idx].Severity = secPolicy.Spec.Process.Severity
							} else {
								secPolicy.Spec.Process.MatchDirectories[idx].Severity = secPolicy.Spec.Severity
							}
						}

						if len(dir.Tags) == 0 {
							if len(secPolicy.Spec.Process.Tags) > 0 {
								secPolicy.Spec.Process.MatchDirectories[idx].Tags = secPolicy.Spec.Process.Tags
							} else {
								secPolicy.Spec.Process.MatchDirectories[idx].Tags = secPolicy.Spec.Tags
							}
						}

						if len(dir.Message) == 0 {
							if len(secPolicy.Spec.Process.Message) > 0 {
								secPolicy.Spec.Process.MatchDirectories[idx].Message = secPolicy.Spec.Process.Message
							} else {
								secPolicy.Spec.Process.MatchDirectories[idx].Message = secPolicy.Spec.Message
							}
						}

						if len(dir.Action) == 0 {
							if len(secPolicy.Spec.Process.Action) > 0 {
								secPolicy.Spec.Process.MatchDirectories[idx].Action = secPolicy.Spec.Process.Action
							} else {
								secPolicy.Spec.Process.MatchDirectories[idx].Action = secPolicy.Spec.Action
							}
						}
					}
				} else if len(secPolicy.Spec.Process.MatchPatterns) > 0 {
					for idx, pat := range secPolicy.Spec.Process.MatchPatterns {
						if pat.Severity == 0 {
							if secPolicy.Spec.Process.Severity != 0 {
								secPolicy.Spec.Process.MatchPatterns[idx].Severity = secPolicy.Spec.Process.Severity
							} else {
								secPolicy.Spec.Process.MatchPatterns[idx].Severity = secPolicy.Spec.Severity
							}
						}

						if len(pat.Tags) == 0 {
							if len(secPolicy.Spec.Process.Tags) > 0 {
								secPolicy.Spec.Process.MatchPatterns[idx].Tags = secPolicy.Spec.Process.Tags
							} else {
								secPolicy.Spec.Process.MatchPatterns[idx].Tags = secPolicy.Spec.Tags
							}
						}

						if len(pat.Message) == 0 {
							if len(secPolicy.Spec.Process.Message) > 0 {
								secPolicy.Spec.Process.MatchPatterns[idx].Message = secPolicy.Spec.Process.Message
							} else {
								secPolicy.Spec.Process.MatchPatterns[idx].Message = secPolicy.Spec.Message
							}
						}

						if len(pat.Action) == 0 {
							if len(secPolicy.Spec.Process.Action) > 0 {
								secPolicy.Spec.Process.MatchPatterns[idx].Action = secPolicy.Spec.Process.Action
							} else {
								secPolicy.Spec.Process.MatchPatterns[idx].Action = secPolicy.Spec.Action
							}
						}
					}
				}

				if len(secPolicy.Spec.File.MatchPaths) > 0 {
					for idx, path := range secPolicy.Spec.File.MatchPaths {
						if path.Severity == 0 {
							if secPolicy.Spec.File.Severity != 0 {
								secPolicy.Spec.File.MatchPaths[idx].Severity = secPolicy.Spec.File.Severity
							} else {
								secPolicy.Spec.File.MatchPaths[idx].Severity = secPolicy.Spec.Severity
							}
						}

						if len(path.Tags) == 0 {
							if len(secPolicy.Spec.File.Tags) > 0 {
								secPolicy.Spec.File.MatchPaths[idx].Tags = secPolicy.Spec.File.Tags
							} else {
								secPolicy.Spec.File.MatchPaths[idx].Tags = secPolicy.Spec.Tags
							}
						}

						if len(path.Message) == 0 {
							if len(secPolicy.Spec.File.Message) > 0 {
								secPolicy.Spec.File.MatchPaths[idx].Message = secPolicy.Spec.File.Message
							} else {
								secPolicy.Spec.File.MatchPaths[idx].Message = secPolicy.Spec.Message
							}
						}

						if len(path.Action) == 0 {
							if len(secPolicy.Spec.File.Action) > 0 {
								secPolicy.Spec.File.MatchPaths[idx].Action = secPolicy.Spec.File.Action
							} else {
								secPolicy.Spec.File.MatchPaths[idx].Action = secPolicy.Spec.Action
							}
						}
					}
				} else if len(secPolicy.Spec.File.MatchDirectories) > 0 {
					for idx, dir := range secPolicy.Spec.File.MatchDirectories {
						if dir.Severity == 0 {
							if secPolicy.Spec.File.Severity != 0 {
								secPolicy.Spec.File.MatchDirectories[idx].Severity = secPolicy.Spec.File.Severity
							} else {
								secPolicy.Spec.File.MatchDirectories[idx].Severity = secPolicy.Spec.Severity
							}
						}

						if len(dir.Tags) == 0 {
							if len(secPolicy.Spec.File.Tags) > 0 {
								secPolicy.Spec.File.MatchDirectories[idx].Tags = secPolicy.Spec.File.Tags
							} else {
								secPolicy.Spec.File.MatchDirectories[idx].Tags = secPolicy.Spec.Tags
							}
						}

						if len(dir.Message) == 0 {
							if len(secPolicy.Spec.File.Message) > 0 {
								secPolicy.Spec.File.MatchDirectories[idx].Message = secPolicy.Spec.File.Message
							} else {
								secPolicy.Spec.File.MatchDirectories[idx].Message = secPolicy.Spec.Message
							}
						}

						if len(dir.Action) == 0 {
							if len(secPolicy.Spec.File.Action) > 0 {
								secPolicy.Spec.File.MatchDirectories[idx].Action = secPolicy.Spec.File.Action
							} else {
								secPolicy.Spec.File.MatchDirectories[idx].Action = secPolicy.Spec.Action
							}
						}
					}
				} else if len(secPolicy.Spec.File.MatchPatterns) > 0 {
					for idx, pat := range secPolicy.Spec.File.MatchPatterns {
						if pat.Severity == 0 {
							if secPolicy.Spec.File.Severity != 0 {
								secPolicy.Spec.File.MatchPatterns[idx].Severity = secPolicy.Spec.File.Severity
							} else {
								secPolicy.Spec.File.MatchPatterns[idx].Severity = secPolicy.Spec.Severity
							}
						}

						if len(pat.Tags) == 0 {
							if len(secPolicy.Spec.File.Tags) > 0 {
								secPolicy.Spec.File.MatchPatterns[idx].Tags = secPolicy.Spec.File.Tags
							} else {
								secPolicy.Spec.File.MatchPatterns[idx].Tags = secPolicy.Spec.Tags
							}
						}

						if len(pat.Message) == 0 {
							if len(secPolicy.Spec.File.Message) > 0 {
								secPolicy.Spec.File.MatchPatterns[idx].Message = secPolicy.Spec.File.Message
							} else {
								secPolicy.Spec.File.MatchPatterns[idx].Message = secPolicy.Spec.Message
							}
						}

						if len(pat.Action) == 0 {
							if len(secPolicy.Spec.File.Action) > 0 {
								secPolicy.Spec.File.MatchPatterns[idx].Action = secPolicy.Spec.File.Action
							} else {
								secPolicy.Spec.File.MatchPatterns[idx].Action = secPolicy.Spec.Action
							}
						}
					}
				}

				if len(secPolicy.Spec.Network.MatchProtocols) > 0 {
					for idx, proto := range secPolicy.Spec.Network.MatchProtocols {
						if proto.Severity == 0 {
							if secPolicy.Spec.Network.Severity != 0 {
								secPolicy.Spec.Network.MatchProtocols[idx].Severity = secPolicy.Spec.Network.Severity
							} else {
								secPolicy.Spec.Network.MatchProtocols[idx].Severity = secPolicy.Spec.Severity
							}
						}

						if len(proto.Tags) == 0 {
							if len(secPolicy.Spec.Network.Tags) > 0 {
								secPolicy.Spec.Network.MatchProtocols[idx].Tags = secPolicy.Spec.Network.Tags
							} else {
								secPolicy.Spec.Network.MatchProtocols[idx].Tags = secPolicy.Spec.Tags
							}
						}

						if len(proto.Message) == 0 {
							if len(secPolicy.Spec.Network.Message) > 0 {
								secPolicy.Spec.Network.MatchProtocols[idx].Message = secPolicy.Spec.Network.Message
							} else {
								secPolicy.Spec.Network.MatchProtocols[idx].Message = secPolicy.Spec.Message
							}
						}

						if len(proto.Action) == 0 {
							if len(secPolicy.Spec.Network.Action) > 0 {
								secPolicy.Spec.Network.MatchProtocols[idx].Action = secPolicy.Spec.Network.Action
							} else {
								secPolicy.Spec.Network.MatchProtocols[idx].Action = secPolicy.Spec.Action
							}
						}
					}
				}

				if len(secPolicy.Spec.Capabilities.MatchCapabilities) > 0 {
					for idx, cap := range secPolicy.Spec.Capabilities.MatchCapabilities {
						if cap.Severity == 0 {
							if secPolicy.Spec.Capabilities.Severity != 0 {
								secPolicy.Spec.Capabilities.MatchCapabilities[idx].Severity = secPolicy.Spec.Capabilities.Severity
							} else {
								secPolicy.Spec.Capabilities.MatchCapabilities[idx].Severity = secPolicy.Spec.Severity
							}
						}

						if len(cap.Tags) == 0 {
							if len(secPolicy.Spec.Capabilities.Tags) > 0 {
								secPolicy.Spec.Capabilities.MatchCapabilities[idx].Tags = secPolicy.Spec.Capabilities.Tags
							} else {
								secPolicy.Spec.Capabilities.MatchCapabilities[idx].Tags = secPolicy.Spec.Tags
							}
						}

						if len(cap.Message) == 0 {
							if len(secPolicy.Spec.Capabilities.Message) > 0 {
								secPolicy.Spec.Capabilities.MatchCapabilities[idx].Message = secPolicy.Spec.Capabilities.Message
							} else {
								secPolicy.Spec.Capabilities.MatchCapabilities[idx].Message = secPolicy.Spec.Message
							}
						}

						if len(cap.Action) == 0 {
							if len(secPolicy.Spec.Capabilities.Action) > 0 {
								secPolicy.Spec.Capabilities.MatchCapabilities[idx].Action = secPolicy.Spec.Capabilities.Action
							} else {
								secPolicy.Spec.Capabilities.MatchCapabilities[idx].Action = secPolicy.Spec.Action
							}
						}
					}
				}

				// update a security policy into the policy list

				if event.Type == "ADDED" {
					new := true
					for _, policy := range dm.HostSecurityPolicies {
						if policy.Metadata["policyName"] == secPolicy.Metadata["policyName"] {
							new = false
							break
						}
					}
					if new {
						dm.HostSecurityPolicies = append(dm.HostSecurityPolicies, secPolicy)
					}
				} else if event.Type == "MODIFIED" {
					for idx, policy := range dm.HostSecurityPolicies {
						if policy.Metadata["policyName"] == secPolicy.Metadata["policyName"] {
							dm.HostSecurityPolicies[idx] = secPolicy
							break
						}
					}
				} else if event.Type == "DELETED" {
					for idx, policy := range dm.HostSecurityPolicies {
						if policy.Metadata["policyName"] == secPolicy.Metadata["policyName"] {
							dm.HostSecurityPolicies = append(dm.HostSecurityPolicies[:idx], dm.HostSecurityPolicies[idx+1:]...)
							break
						}
					}
				}

				dm.HostSecurityPoliciesLock.Unlock()

				//dm.Logger.Printf("Detected a Host Security Policy (%s/%s)", strings.ToLower(event.Type), secPolicy.Metadata["policyName"])

				// apply security policies to a host
				dm.UpdateHostSecurityPolicies()
			}
		}
	}
}

// WatchExternalWorkloadSecurityPolicies Function
func (dm *KVMS) WatchExternalWorkloadSecurityPolicies() {
	for {
		if !K8s.CheckCustomResourceDefinition("kubearmorexternalworkloadpolicies") {
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
				defer dm.ExternalWorkloadSecurityPoliciesLock.Unlock()

				// create a security policy

				secPolicy := tp.ExternalWorkloadSecurityPolicy{}
				fmt.Println("External workload policy configured")

				secPolicy.Metadata = map[string]string{}
				secPolicy.Metadata["namespaceName"] = event.Object.Metadata.Namespace
				secPolicy.Metadata["policyName"] = event.Object.Metadata.Name

				if err := kl.Clone(event.Object.Spec, &secPolicy.Spec); err != nil {
					fmt.Println("Failed to clone a spec")
				}
				// update a security policy into the policy list

				if event.Type == "ADDED" {
					if !kl.ContainsElement(dm.ExternalWorkloadSecurityPolicies, secPolicy) {
						dm.ExternalWorkloadSecurityPolicies = append(dm.ExternalWorkloadSecurityPolicies, secPolicy)
					}
					dm.GenerateExternalWorkloadIdentity(secPolicy.Metadata["labels"])
					// TODO: Handle this map of identity to grpc connection seperate context
					//dm.MapExternalWorkloadConnIdentity[identity] = conn
				} else if event.Type == "MODIFIED" {
					for idx, policy := range dm.ExternalWorkloadSecurityPolicies {
						if policy.Metadata["policyName"] == secPolicy.Metadata["policyName"] {
							dm.ExternalWorkloadSecurityPolicies[idx] = secPolicy
							break
						}
					}
				} else if event.Type == "DELETED" {
					for idx, policy := range dm.ExternalWorkloadSecurityPolicies {
						if reflect.DeepEqual(secPolicy, policy) {
							dm.ExternalWorkloadSecurityPolicies = append(dm.ExternalWorkloadSecurityPolicies[:idx], dm.ExternalWorkloadSecurityPolicies[idx+1:]...)
							break
						}
					}
				}

				// apply security policies to a external workload
				dm.UpdateExternalWorkloadSecurityPolicies()
			}
		}
	}
}
