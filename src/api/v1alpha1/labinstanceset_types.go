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

// LabInstanceSetSpec defines the desired state of LabInstanceSet
type LabInstanceSetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Generator LabInstanceGenerator `json:"generator"`
}

type LabInstanceGenerator struct {
	LabInstances []LabInstanceElement `json:"labInstances,omitempty"`
}

type LabInstanceElement struct {
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// LabInstanceSetStatus defines the observed state of LabInstanceSet
type LabInstanceSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LabInstanceSet is the Schema for the labinstancesets API
type LabInstanceSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LabInstanceSetSpec   `json:"spec,omitempty"`
	Status LabInstanceSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LabInstanceSetList contains a list of LabInstanceSet
type LabInstanceSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LabInstanceSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LabInstanceSet{}, &LabInstanceSetList{})
}
