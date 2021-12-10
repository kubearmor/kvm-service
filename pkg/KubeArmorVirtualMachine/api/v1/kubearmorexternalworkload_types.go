// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:singular="kubearmorvirtualmachine",path="kubearmorvirtualmachines",scope="Cluster",shortName={kvm}
// +kubebuilder:subresource:status
type KubeArmorVirtualMachine struct {
	// +k8s:openapi-gen=false
	// +deepequal-gen=false
	metav1.TypeMeta `json:",inline"`
	// +k8s:openapi-gen=false
	// +deepequal-gen=false
	metav1.ObjectMeta `json:"metadata"`

	// Status is the most recent status of the KubeArmor Virtual Machine.
	// It is a read-only field.
	//
	// +deepequal-gen=false
	// +kubebuilder:validation:Optional
	Status KubeArmorVirtualMachineStatus `json:"status"`
}

// KubeArmorVirtualMachineStatus is the status of a the KubeArmor Virtual Machine.
type KubeArmorVirtualMachineStatus struct {
	// ID is the numeric identity allocated for the virtual machine.
	ID uint64 `json:"id,omitempty"`

	// IP is the IP address of the virtual machine. Empty if the virtual machine has not registered.
	IP string `json:"ip,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=false
// +deepequal-gen=false

// KubeArmorVirtualMachineList is a list of KubeArmorVirtualMachine objects.
type KubeArmorVirtualMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items is a list of KubeArmorVirtualMachine
	Items []KubeArmorVirtualMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeArmorVirtualMachine{}, &KubeArmorVirtualMachineList{})
}
