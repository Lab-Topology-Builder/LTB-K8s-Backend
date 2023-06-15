package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LabInstanceSpec struct {
	LabTemplateReference string `json:"labTemplateReference"`
	DNSAddress           string `json:"dnsAddress"`
}

type LabInstanceStatus struct {
	Status         string `json:"status,omitempty"`
	NumPodsRunning string `json:"numpodsrunning,omitempty"`
	NumVMsRunning  string `json:"numvmsrunning,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.status`
//+kubebuilder:printcolumn:name="PODS_RUNNING",type=string,JSONPath=`.status.numpodsrunning`
//+kubebuilder:printcolumn:name="VMS_RUNNING",type=string,JSONPath=`.status.numvmsrunning`

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
