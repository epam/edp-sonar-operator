package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SonarSpec defines the desired state of Sonar.
type SonarSpec struct {
	EdpSpec EdpSpec `json:"edpSpec"`

	// +optional
	BasePath string `json:"basePath,omitempty"`

	// +optional
	DefaultPermissionTemplate string `json:"defaultPermissionTemplate,omitempty"`
}

type EdpSpec struct {
	DnsWildcard string `json:"dnsWildcard"`
}

// SonarStatus defines the observed state of Sonar.
type SonarStatus struct {
	// +optional
	Available bool `json:"available,omitempty"`

	// +optional
	LastTimeUpdated metav1.Time `json:"lastTimeUpdated,omitempty"`

	// +optional
	Status string `json:"status,omitempty"`

	// +optional
	ExternalUrl string `json:"externalUrl,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// Sonar is the Schema for the sonars API.
type Sonar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonarSpec   `json:"spec,omitempty"`
	Status SonarStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SonarList contains a list of Sonar.
type SonarList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Sonar `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sonar{}, &SonarList{})
}
