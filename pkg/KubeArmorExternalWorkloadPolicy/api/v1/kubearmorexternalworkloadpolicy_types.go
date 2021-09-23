// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:singular="kubearmorexternalworkload",path="kubearmorexternalworkloads",scope="Cluster",shortName={kew}
// +kubebuilder:subresource:status
type KubeArmorExternalWorkloadPolicy struct {
	// +k8s:openapi-gen=false
	// +deepequal-gen=false
	metav1.TypeMeta `json:",inline"`
	// +k8s:openapi-gen=false
	// +deepequal-gen=false
	metav1.ObjectMeta `json:"metadata"`

	// Status is the most recent status of the external KubeArmor workload.
	// It is a read-only field.
	//
	// +deepequal-gen=false
	// +kubebuilder:validation:Optional
	Status KubeArmorExternalWorkloadPolicyStatus `json:"status"`
}

// KubeArmorExternalWorkloadPolicyStatus is the status of a the external KubeArmor workload.
type KubeArmorExternalWorkloadPolicyStatus struct {
	// ID is the numeric identity allocated for the external workload.
	ID uint64 `json:"id,omitempty"`

	// IP is the IP address of the workload. Empty if the workload has not registered.
	IP string `json:"ip,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=false
// +deepequal-gen=false

// KubeArmorExternalWorkloadPolicyList is a list of KubeArmorExternalWorkloadPolicy objects.
type KubeArmorExternalWorkloadPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items is a list of KubeArmorExternalWorkloadPolicy
	Items []KubeArmorExternalWorkloadPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeArmorExternalWorkloadPolicy{}, &KubeArmorExternalWorkloadPolicyList{})
}
