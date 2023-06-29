package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeTypeSpec defines the Kind and NodeSpec for a NodeType
type NodeTypeSpec struct {

	// Kind can be used to specify if the nodes is either a pod or a vm
	Kind string `json:"kind,omitempty"`
	// NodeSpec is the PodSpec or VirtualMachineSpec configuration for the node with the possibility to use go templating syntax to include LabTemplate variables (see [User Guide](https://lab-topology-builder.github.io/LTB-K8s-Backend/user-guide/#example-node-type))
	// See [PodSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podspec-v1-core) and [VirtualMachineSpec](https://kubevirt.io/api-reference/master/definitions.html#_v1_virtualmachinespec)
	NodeSpec string `json:"nodeSpec,omitempty"`
}

// NodeTypeStatus defines the observed state of NodeType
type NodeTypeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// NodeType defines a type of node that can be used in a lab template
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
