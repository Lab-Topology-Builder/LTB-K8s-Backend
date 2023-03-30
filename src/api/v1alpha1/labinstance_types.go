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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LabInstanceSpec defines the desired state of LabInstance
type LabInstanceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	LabTemplateReference string `json:"labTemplateReference"`
}

// LabInstanceStatus defines the observed state of LabInstance
type LabInstanceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase      corev1.PodStatus `json:"phase,omitempty"`
	LastUpdate string           `json:"lastUpdate,omitempty"`
	AppVersion string           `json:"appVersion,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LabInstance is the Schema for the labinstances API
type LabInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LabInstanceSpec   `json:"spec,omitempty"`
	Status LabInstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LabInstanceList contains a list of LabInstance
type LabInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LabInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LabInstance{}, &LabInstanceList{})
}
