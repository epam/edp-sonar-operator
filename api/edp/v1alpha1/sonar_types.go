package v1alpha1

import (
	coreV1Api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SonarSpec defines the desired state of Sonar
type SonarVolumes struct {
	Name     string `json:"name"`
	Capacity string `json:"capacity"`

	// +optional
	StorageClass string `json:"storage_class,omitempty"`
}

type SonarSpec struct {
	Version   string  `json:"version"`
	Image     string  `json:"image"`
	InitImage string  `json:"initImage"`
	DBImage   string  `json:"dbImage"`
	EdpSpec   EdpSpec `json:"edpSpec"`

	// +optional
	BasePath string `json:"basePath,omitempty"`

	// +optional
	// +nullable
	Volumes []SonarVolumes `json:"volumes,omitempty"`

	// +optional
	// +nullable
	ImagePullSecrets []coreV1Api.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// +optional
	DefaultPermissionTemplate string `json:"defaultPermissionTemplate,omitempty"`
}

type EdpSpec struct {
	DnsWildcard string `json:"dnsWildcard"`
}

// SonarStatus defines the observed state of Sonar
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
// +kubebuilder:deprecatedversion

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
	SchemeBuilder.Register(&Sonar{}, &SonarList{}, &SonarGroup{}, &SonarGroupList{}, &SonarPermissionTemplate{},
		&SonarPermissionTemplateList{})
}
