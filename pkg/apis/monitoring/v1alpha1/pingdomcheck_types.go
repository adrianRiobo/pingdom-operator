package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PingdomCheckSpec defines the desired state of PingdomCheck
type PingdomCheckSpec struct {
	Name string `json:"name"`
        // +kubebuilder:validation:Pattern=`^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`
        URL  string `json:"url"`
}

// PingdomCheckStatus defines the observed state of PingdomCheck
type PingdomCheckStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PingdomCheck is the Schema for the pingdomchecks API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=pingdomchecks,scope=Namespaced
type PingdomCheck struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PingdomCheckSpec   `json:"spec,omitempty"`
	Status PingdomCheckStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PingdomCheckList contains a list of PingdomCheck
type PingdomCheckList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PingdomCheck `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PingdomCheck{}, &PingdomCheckList{})
}
