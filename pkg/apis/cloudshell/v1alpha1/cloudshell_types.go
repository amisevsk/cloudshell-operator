package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CloudShellSpec defines the desired state of CloudShell
// +k8s:openapi-gen=true
type CloudShellSpec struct {
	Image string `json:"image"`
}

// CloudShellStatus defines the observed state of CloudShell
// +k8s:openapi-gen=true
type CloudShellStatus struct {
	Id    string `json:"id"`
	Ready bool   `json:"ready"`
	Url   string `json:"url"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CloudShell is the Schema for the cloudshells API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=cloudshells,scope=Namespaced
type CloudShell struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudShellSpec   `json:"spec,omitempty"`
	Status CloudShellStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CloudShellList contains a list of CloudShell
type CloudShellList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudShell `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudShell{}, &CloudShellList{})
}
