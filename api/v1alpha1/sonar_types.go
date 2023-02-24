package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SonarSpec defines the desired state of Sonar.
type SonarSpec struct {
	// Secret is the name of the k8s object Secret related to sonar.
	Secret string `json:"secret"`

	// Url is used to explicitly specify the url of sonar. It may not be needed if the sonar is deployed in the same cluster.
	Url string `json:"url,omitempty"`

	// Users specify which users should be created.
	// +optional
	Users []User `json:"users,omitempty"`

	// Groups specify which groups should be created.
	// +optional
	Groups []Group `json:"groups,omitempty"`

	// Plugins specify which plugins should be installed to sonar.
	// +optional
	Plugins []string `json:"plugins,omitempty"`

	// QualityGates specify which quality gates should be created.
	// +optional
	QualityGates []QualityGate `json:"qualityGates,omitempty"`

	// Settings specify which settings should be configured.
	// +optional
	Settings []SonarSetting `json:"settings,omitempty"`

	// +optional
	BasePath string `json:"basePath,omitempty"`

	// +optional
	DefaultPermissionTemplate string `json:"defaultPermissionTemplate,omitempty"`
}

type QualityGate struct {
	Name string `json:"name"`

	Conditions []QualityGateCondition `json:"conditions"`

	// +optional
	SetAsDefault bool `json:"setAsDefault,omitempty"`
}

type QualityGateCondition struct {
	Error string `json:"error"`

	Metric string `json:"metric"`

	OP string `json:"op"`

	// +optional
	Period string `json:"period,omitempty"`
}

type SonarSetting struct {
	Key string `json:"key"`

	Value string `json:"value"`

	ValueType string `json:"valueType"`
}

type Group struct {
	Name string `json:"name"`

	// +optional
	Permissions []string `json:"permissions,omitempty"`
}

type User struct {
	Username string `json:"username"`

	Login string `json:"login"`

	// +optional
	Group string `json:"group,omitempty"`

	// +optional
	Permissions []string `json:"permissions,omitempty"`
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
