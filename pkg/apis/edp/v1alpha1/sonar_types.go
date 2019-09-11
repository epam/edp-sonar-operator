package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SonarSpec defines the desired state of Sonar
// +k8s:openapi-gen=true

type SonarVolumes struct {
	Name         string `json:"name"`
	StorageClass string `json:"storage_class"`
	Capacity     string `json:"capacity"`
}

type SonarSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Version                    string                     `json:"version"`
	Image                      string                     `json:"image, omitempty"`
	Volumes                    []SonarVolumes             `json:"volumes,omitempty"`
	SonarExternalConfiguration SonarExternalConfiguration `json:"externalConfiguration,omitempty"`
}

type SonarExternalConfigurationItem struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
}

type SonarExternalConfiguration struct {
	ReadUser      *SonarExternalConfigurationItem `json:"readUser,omitempty"`
	AdminUser     *SonarExternalConfigurationItem `json:"adminUser,omitempty"`
	IsCredentials *SonarExternalConfigurationItem `json:"isCredentials,omitempty"`
}

// SonarStatus defines the observed state of Sonar
// +k8s:openapi-gen=true
type SonarStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Available       bool      `json:"available, omitempty"`
	LastTimeUpdated time.Time `json:"lastTimeUpdated, omitempty"`
	Status          string    `json:"status, omitempty"`
	ExternalUrl     string    `json:"externalUrl, omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Sonar is the Schema for the sonars API
// +k8s:openapi-gen=true
type Sonar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonarSpec   `json:"spec,omitempty"`
	Status SonarStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SonarList contains a list of Sonar
type SonarList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sonar `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sonar{}, &SonarList{})
}
