package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LabTemplateSpec struct {
	Nodes     []LabInstanceNodes `json:"nodes"`
	Neighbors []string           `json:"neighbors"`
}

type LabInstanceNodes struct {
	Name             string          `json:"name"`
	NodeTypeRef      NodeTypeRef     `json:"nodeTypeRef"`
	Interfaces       []NodeInterface `json:"interfaces,omitempty"`
	Config           string          `json:"config,omitempty"`
	Ports            []Port          `json:"ports,omitempty"`
	RenderedNodeSpec string          `json:"renderedNodeSpec,omitempty"`
}

type Port struct {
	Name     string          `json:"name"`
	Protocol corev1.Protocol `json:"protocol"`
	Port     int32           `json:"port"`
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
