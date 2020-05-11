package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CascSpec defines the desired state of Casc
type CascSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Secret         SecretRef      `json:"secret"`
	Configurations []ConfigMapRef `json:"configurations"`

	JenkinsRef JenkinsRef `json:"jenkinsRef"`
}

// SecretRef is reference to Kubernetes secret.
type SecretRef struct {
	Name string `json:"name"`
}

// ConfigMapRef is reference to Kubernetes ConfigMap.
type ConfigMapRef struct {
	Name string `json:"name"`
}

// JenkinsRef is reference to Jenkins CR.
type JenkinsRef struct {
	Name string `json:"name"`
}

// CascStatus defines the observed state of Casc
type CascStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Casc is the Schema for the cascs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=cascs,scope=Namespaced
type Casc struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CascSpec   `json:"spec,omitempty"`
	Status CascStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CascList contains a list of Casc
type CascList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Casc `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Casc{}, &CascList{})
}
