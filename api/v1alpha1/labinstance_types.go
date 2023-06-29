package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LabInstanceSpec define which LabTemplate should be used for the lab instance and the DNS address.
type LabInstanceSpec struct {
	// Reference to the name of a LabTemplate to use for the lab instance.
	LabTemplateReference string `json:"labTemplateReference"`
	// The DNS address, which will be used to expose the lab instance.
	// It should point to the Kubernetes node where the lab instance is running.
	DNSAddress string `json:"dnsAddress"`
}

type LabInstanceStatus struct {
	Status         string `json:"status,omitempty"`
	NumPodsRunning string `json:"numPodsRunning,omitempty"`
	NumVMsRunning  string `json:"numVMsRunning,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.status`
//+kubebuilder:printcolumn:name="PODS_RUNNING",type=string,JSONPath=`.status.numPodsRunning`
//+kubebuilder:printcolumn:name="VMS_RUNNING",type=string,JSONPath=`.status.numVMsRunning`

// A lab instance is created as a specific instance of a deployed lab, using the configuration from the corresponding lab template.
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
