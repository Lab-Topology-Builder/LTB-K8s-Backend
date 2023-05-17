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

type LabTemplateSpec struct {
	Nodes       []LabInstanceNodes `json:"nodes"`
	Connections []Connection       `json:"connections"`
}

type LabInstanceNodes struct {
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:MinLength=1
	Name        string          `json:"name"`
	NodeTypeRef NodeTypeRef     `json:"nodetyperef"`
	Interfaces  []NodeInterface `json:"interfaces,omitempty"`
	Config      string          `json:"config,omitempty"`
	Ports       []Port          `json:"ports,omitempty"`
}

type Port struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol,omitempty"`
	Port     int32  `json:"port"`
}

type NodeInterface struct {
	IPv4 string `json:"ipv4,omitempty"`
	IPv6 string `json:"ipv6,omitempty"`
}

type NodeTypeRef struct {
	Type    string `json:"type"`
	Image   string `json:"image,omitempty"`
	Version string `json:"version,omitempty"`
}

type Connection struct {
	Neighbors []string `json:"neighbors"` // comma separated list of neighbors, maybe call it endpoints?
}

type LabTemplateStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

type LabTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LabTemplateSpec   `json:"spec,omitempty"`
	Status LabTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type LabTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LabTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LabTemplate{}, &LabTemplateList{})
}
