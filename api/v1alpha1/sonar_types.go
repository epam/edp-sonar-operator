package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SonarSpec defines the desired state of Sonar.
type SonarSpec struct {
	// Secret is the name of the k8s object Secret related to sonar.
	// Secret should contain a user field with a sonar username and a password field with a sonar password.
	// Pass the token in the user field and leave the password field empty for token authentication.
	Secret string `json:"secret"`

	// Url is the url of sonar instance.
	Url string `json:"url"`

	// Settings specify which settings should be configured.
	// +optional
	Settings []SonarSetting `json:"settings,omitempty"`

	// DefaultPermissionTemplate is the name of the default permission template.
	// +optional
	// +kubebuilder:example="Default template for projects"
	DefaultPermissionTemplate string `json:"defaultPermissionTemplate,omitempty"`
}

// SonarSetting defines the setting of sonar.
type SonarSetting struct {
	// Key is the key of the setting.
	// +kubebuilder:example=sonar.core.serverBaseURL
	Key string `json:"key"`

	// Value is the value of the setting.
	// +optional
	// +kubebuilder:validation:MaxLength=4000
	// +kubebuilder:example="https://my-sonarqube-instance.com"
	Value string `json:"value,omitempty"`

	// Setting multi value. To set several values, the parameter must be called once for each value.
	// +optional
	// +kubebuilder:example={**/vendor/**,**/tests/**}
	Values []string `json:"values,omitempty"`

	// Setting field values. To set several values, the parameter must be called once for each value.
	// +optional
	// +kubebuilder:example={beginBlockRegexp: ".*", endBlockRegexp: ".*"}
	FieldValues map[string]string `json:"fieldValues,omitempty"`
}

// SonarStatus defines the observed state of Sonar.
type SonarStatus struct {
	// Value is status of sonar instance.
	// Possible values:
	// GREEN: SonarQube is fully operational
	// YELLOW: SonarQube is usable, but it needs attention in order to be fully operational
	// RED: SonarQube is not operational
	// +optional
	Value string `json:"value,omitempty"`

	// Error represents error message if something went wrong.
	// +optional
	Error string `json:"error,omitempty"`

	// Connected shows if operator is connected to sonar.
	// +optional
	Connected bool `json:"connected"`

	// ProcessedSettings shows which settings were processed.
	// It is used to compare the current settings with the settings that were processed
	// to unset the settings that are not in the current settings.
	// +optional
	ProcessedSettings string `json:"processedSettings,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Connected",type="boolean",JSONPath=".status.connected",description="Is connected to sonar"

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
