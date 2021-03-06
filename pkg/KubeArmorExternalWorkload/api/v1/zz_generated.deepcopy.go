// +build !ignore_autogenerated

// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeArmorExternalWorkload) DeepCopyInto(out *KubeArmorExternalWorkload) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeArmorExternalWorkload.
func (in *KubeArmorExternalWorkload) DeepCopy() *KubeArmorExternalWorkload {
	if in == nil {
		return nil
	}
	out := new(KubeArmorExternalWorkload)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KubeArmorExternalWorkload) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeArmorExternalWorkloadList) DeepCopyInto(out *KubeArmorExternalWorkloadList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KubeArmorExternalWorkload, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeArmorExternalWorkloadList.
func (in *KubeArmorExternalWorkloadList) DeepCopy() *KubeArmorExternalWorkloadList {
	if in == nil {
		return nil
	}
	out := new(KubeArmorExternalWorkloadList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KubeArmorExternalWorkloadList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubeArmorExternalWorkloadStatus) DeepCopyInto(out *KubeArmorExternalWorkloadStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubeArmorExternalWorkloadStatus.
func (in *KubeArmorExternalWorkloadStatus) DeepCopy() *KubeArmorExternalWorkloadStatus {
	if in == nil {
		return nil
	}
	out := new(KubeArmorExternalWorkloadStatus)
	in.DeepCopyInto(out)
	return out
}
