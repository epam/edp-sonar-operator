package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/epam/edp-sonar-operator/api/common"
)

// SonarQualityProfileSpec defines the desired state of SonarQualityProfile
type SonarQualityProfileSpec struct {
	// Name is a name of quality profile.
	// Name should be unique across all quality profiles.
	// Don't change this field after creation otherwise quality profile will be recreated.
	// +required
	// +kubebuilder:validation:MaxLength=100
	// +kubebuilder:example="My Quality Profile"
	Name string `json:"name"`

	// Language is a language of quality profile.
	// +required
	// +kubebuilder:example="go"
	Language string `json:"language"`

	// Default is a flag to set quality profile as default.
	// Only one quality profile can be default.
	// If several quality profiles have default flag, the random one will be chosen.
	// Default quality profile can't be deleted. You need to set another quality profile as default before.
	// +optional
	// +kubebuilder:example="true"
	Default bool `json:"default"`

	// Rules is a list of rules for quality profile.
	// Key is a rule key, value is a rule.
	// +optional
	// +nullable
	// +kubebuilder:example={S5547: {severity: "MAJOR", params: "key1=v1;key2=v2"}}
	Rules map[string]Rule `json:"rules,omitempty"`

	// SonarRef is a reference to Sonar custom resource.
	// +required
	SonarRef common.SonarRef `json:"sonarRef"`
}

// Rule defines a rule of quality profile.
type Rule struct {
	// Severity is a severity of rule.
	// +optional
	// +kubebuilder:example="MAJOR"
	// +kubebuilder:validation:Enum=INFO;MINOR;MAJOR;CRITICAL;BLOCKER
	Severity string `json:"severity,omitempty"`

	// Params is as semicolon list of key=value.
	// +optional
	// +kubebuilder:example="key1=v1;key2=v2"
	Params string `json:"params,omitempty"`
}

// SonarQualityProfileStatus defines the observed state of SonarQualityProfile
type SonarQualityProfileStatus struct {
	// Value is a status of the quality profile.
	// +optional
	Value string `json:"value,omitempty"`

	// Error is an error message if something went wrong.
	// +optional
	Error string `json:"error,omitempty"`
}

// SonarQualityProfile is the Schema for the sonarqualityprofiles API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type SonarQualityProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonarQualityProfileSpec   `json:"spec,omitempty"`
	Status SonarQualityProfileStatus `json:"status,omitempty"`
}

func (in *SonarQualityProfile) GetSonarRef() common.SonarRef {
	return in.Spec.SonarRef
}

// SonarQualityProfileList contains a list of SonarQualityProfile
// +kubebuilder:object:root=true
type SonarQualityProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SonarQualityProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SonarQualityProfile{}, &SonarQualityProfileList{})
}
