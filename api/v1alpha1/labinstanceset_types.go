package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LabInstanceSetSpec struct {
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

type LabInstanceSetStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type LabInstanceSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LabInstanceSetSpec   `json:"spec,omitempty"`
	Status LabInstanceSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type LabInstanceSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LabInstanceSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LabInstanceSet{}, &LabInstanceSetList{})
}
