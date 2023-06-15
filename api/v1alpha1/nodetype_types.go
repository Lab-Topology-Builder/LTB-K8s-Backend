package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeTypeSpec struct {
	Kind     string `json:"kind,omitempty"`
	NodeSpec string `json:"nodeSpec,omitempty"`
}

type NodeTypeStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

type NodeType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeTypeSpec   `json:"spec,omitempty"`
	Status NodeTypeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type NodeTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeType `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeType{}, &NodeTypeList{})
}
