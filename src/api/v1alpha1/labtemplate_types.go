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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LabTemplateSpec defines the desired state of LabTemplate
type LabTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	BasicTemplate BasicLabTemplate `json:"template"`
}

type BasicLabTemplate struct {
	Name      string               `json:"name"`
	Namespace string               `json:"namespace,omitempty"`
	Label     string               `json:"label,omitempty"`
	Spec      BasicLabTemplateSpec `json:"spec"`
}

type BasicLabTemplateSpec struct {
	Hosts []LabInstanceHost `json:"hosts"`
}

type LabInstanceHost struct {
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	Image      HostImage       `json:"image"`
	Interfaces []HostInterface `json:"interfaces"`
	Config     string          `json:"config,omitempty"`
}

type HostInterface struct {
	Connects NeighborInterface `json:"connects"`
	Ipv4     string            `json:"ipv4,omitempty"`
	Ipv6     string            `json:"ipv6,omitempty"`
}

type HostImage struct {
	Type    string `json:"type"`
	Version string `json:"version"`
}

type NeighborInterface struct {
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:MinLength=1
	NeighborName string `json:"name"`

	Interface int `json:"interface"`
}

// LabTemplateStatus defines the observed state of LabTemplate
type LabTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LabTemplate is the Schema for the labtemplates API
type LabTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LabTemplateSpec   `json:"spec,omitempty"`
	Status LabTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LabTemplateList contains a list of LabTemplate
type LabTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LabTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LabTemplate{}, &LabTemplateList{})
}
