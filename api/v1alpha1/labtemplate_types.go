package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LabTemplateSpec defines the Lab nodes and their connections.
type LabTemplateSpec struct {
	// Array of lab nodes and their configuration.
	Nodes []LabInstanceNodes `json:"nodes"`
	// Array of connections between lab nodes. (currently not supported)
	Neighbors []string `json:"neighbors"`
}

// Configuration for a lab node.
type LabInstanceNodes struct {
	// The name of the lab node.
	Name string `json:"name"`
	// The type of the lab node.
	NodeTypeRef NodeTypeRef `json:"nodeTypeRef"`
	// Array of interface configurations for the lab node. (currently not supported)
	Interfaces []NodeInterface `json:"interfaces,omitempty"`
	// The configuration for the lab node.
	Config string `json:"config,omitempty"`
	// Array of ports which should be publicly exposed for the lab node.
	Ports            []Port `json:"ports,omitempty"`
	RenderedNodeSpec string `json:"renderedNodeSpec,omitempty"`
}

// Port of a lab node which should be publicly exposed.
type Port struct {
	// Arbitrary name for the port.
	Name string `json:"name"`
	// Choose either TCP or UDP.
	Protocol corev1.Protocol `json:"protocol"`
	// The port number to expose.
	Port int32 `json:"port"`
}

// Interface configuration for the lab node (currently not supported)
type NodeInterface struct {
	// IPv4 address of the interface.
	IPv4 string `json:"ipv4,omitempty"`
	// IPv6 address of the interface.
	IPv6 string `json:"ipv6,omitempty"`
}

// NodeTypeRef references a NodeType with the possibility to provide additional information to the NodeType.
type NodeTypeRef struct {
	// Reference to the name of a NodeType.
	Type string `json:"type"`
	// Image to use for the NodeType. Is available as variable in the NodeType and functionality depends on its usage.
	Image string `json:"image,omitempty"`
	// Version of the NodeType. Is available as variable in the NodeType and functionality depends on its usage.
	Version string `json:"version,omitempty"`
}

type LabTemplateStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// Defines the lab topology, its nodes and their configuration.
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
