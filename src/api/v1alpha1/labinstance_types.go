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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LabInstanceSpec struct {
	LabTemplateReference string `json:"labTemplateReference"`
}

type LabInstanceStatus struct {
	Status    string `json:"status,omitempty"`
	PodStatus string `json:"podstatus,omitempty"`
	VMStatus  string `json:"vmstatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.status`
//+kubebuilder:printcolumn:name="PODS_STATUS",type=string,JSONPath=`.status.podstatus`
//+kubebuilder:printcolumn:name="VMS_STATUS",type=string,JSONPath=`.status.vmstatus`

type LabInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LabInstanceSpec   `json:"spec,omitempty"`
	Status LabInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type LabInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LabInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LabInstance{}, &LabInstanceList{})
}
