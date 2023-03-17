//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2023 Jan Untersander, Tsigereda Nebai Kidane.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BasicLabTemplate) DeepCopyInto(out *BasicLabTemplate) {
	*out = *in
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BasicLabTemplate.
func (in *BasicLabTemplate) DeepCopy() *BasicLabTemplate {
	if in == nil {
		return nil
	}
	out := new(BasicLabTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BasicLabTemplateSpec) DeepCopyInto(out *BasicLabTemplateSpec) {
	*out = *in
	if in.Hosts != nil {
		in, out := &in.Hosts, &out.Hosts
		*out = make([]LabInstanceHost, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BasicLabTemplateSpec.
func (in *BasicLabTemplateSpec) DeepCopy() *BasicLabTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(BasicLabTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HostImage) DeepCopyInto(out *HostImage) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HostImage.
func (in *HostImage) DeepCopy() *HostImage {
	if in == nil {
		return nil
	}
	out := new(HostImage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HostInterface) DeepCopyInto(out *HostInterface) {
	*out = *in
	out.Connects = in.Connects
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HostInterface.
func (in *HostInterface) DeepCopy() *HostInterface {
	if in == nil {
		return nil
	}
	out := new(HostInterface)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabInstance) DeepCopyInto(out *LabInstance) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabInstance.
func (in *LabInstance) DeepCopy() *LabInstance {
	if in == nil {
		return nil
	}
	out := new(LabInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LabInstance) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabInstanceElement) DeepCopyInto(out *LabInstanceElement) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabInstanceElement.
func (in *LabInstanceElement) DeepCopy() *LabInstanceElement {
	if in == nil {
		return nil
	}
	out := new(LabInstanceElement)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabInstanceGenerator) DeepCopyInto(out *LabInstanceGenerator) {
	*out = *in
	if in.LabInstances != nil {
		in, out := &in.LabInstances, &out.LabInstances
		*out = make([]LabInstanceElement, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabInstanceGenerator.
func (in *LabInstanceGenerator) DeepCopy() *LabInstanceGenerator {
	if in == nil {
		return nil
	}
	out := new(LabInstanceGenerator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabInstanceHost) DeepCopyInto(out *LabInstanceHost) {
	*out = *in
	out.Image = in.Image
	if in.Interfaces != nil {
		in, out := &in.Interfaces, &out.Interfaces
		*out = make([]HostInterface, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabInstanceHost.
func (in *LabInstanceHost) DeepCopy() *LabInstanceHost {
	if in == nil {
		return nil
	}
	out := new(LabInstanceHost)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabInstanceList) DeepCopyInto(out *LabInstanceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]LabInstance, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabInstanceList.
func (in *LabInstanceList) DeepCopy() *LabInstanceList {
	if in == nil {
		return nil
	}
	out := new(LabInstanceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LabInstanceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabInstanceSet) DeepCopyInto(out *LabInstanceSet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabInstanceSet.
func (in *LabInstanceSet) DeepCopy() *LabInstanceSet {
	if in == nil {
		return nil
	}
	out := new(LabInstanceSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LabInstanceSet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabInstanceSetList) DeepCopyInto(out *LabInstanceSetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]LabInstanceSet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabInstanceSetList.
func (in *LabInstanceSetList) DeepCopy() *LabInstanceSetList {
	if in == nil {
		return nil
	}
	out := new(LabInstanceSetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LabInstanceSetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabInstanceSetSpec) DeepCopyInto(out *LabInstanceSetSpec) {
	*out = *in
	in.Generator.DeepCopyInto(&out.Generator)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabInstanceSetSpec.
func (in *LabInstanceSetSpec) DeepCopy() *LabInstanceSetSpec {
	if in == nil {
		return nil
	}
	out := new(LabInstanceSetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabInstanceSetStatus) DeepCopyInto(out *LabInstanceSetStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabInstanceSetStatus.
func (in *LabInstanceSetStatus) DeepCopy() *LabInstanceSetStatus {
	if in == nil {
		return nil
	}
	out := new(LabInstanceSetStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabInstanceSpec) DeepCopyInto(out *LabInstanceSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabInstanceSpec.
func (in *LabInstanceSpec) DeepCopy() *LabInstanceSpec {
	if in == nil {
		return nil
	}
	out := new(LabInstanceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabInstanceStatus) DeepCopyInto(out *LabInstanceStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabInstanceStatus.
func (in *LabInstanceStatus) DeepCopy() *LabInstanceStatus {
	if in == nil {
		return nil
	}
	out := new(LabInstanceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabTemplate) DeepCopyInto(out *LabTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabTemplate.
func (in *LabTemplate) DeepCopy() *LabTemplate {
	if in == nil {
		return nil
	}
	out := new(LabTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LabTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabTemplateList) DeepCopyInto(out *LabTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]LabTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabTemplateList.
func (in *LabTemplateList) DeepCopy() *LabTemplateList {
	if in == nil {
		return nil
	}
	out := new(LabTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LabTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabTemplateSpec) DeepCopyInto(out *LabTemplateSpec) {
	*out = *in
	in.BasicTemplate.DeepCopyInto(&out.BasicTemplate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabTemplateSpec.
func (in *LabTemplateSpec) DeepCopy() *LabTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(LabTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabTemplateStatus) DeepCopyInto(out *LabTemplateStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabTemplateStatus.
func (in *LabTemplateStatus) DeepCopy() *LabTemplateStatus {
	if in == nil {
		return nil
	}
	out := new(LabTemplateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NeighborInterface) DeepCopyInto(out *NeighborInterface) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NeighborInterface.
func (in *NeighborInterface) DeepCopy() *NeighborInterface {
	if in == nil {
		return nil
	}
	out := new(NeighborInterface)
	in.DeepCopyInto(out)
	return out
}
