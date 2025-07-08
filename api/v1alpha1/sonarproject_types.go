package v1alpha1

import (
	"github.com/epam/edp-sonar-operator/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SonarProjectSpec defines the desired state of SonarProject.
type SonarProjectSpec struct {
	// Key is the SonarQube project key.
	// This is a unique identifier for the project in SonarQube.
	// Allowed characters are alphanumeric, '-' (dash), '_' (underscore), '.' (period) and ':' (colon), with at least one non-digit.
	// +required
	// +kubebuilder:validation:MaxLength=400
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +kubebuilder:example="my-project"
	Key string `json:"key"`

	// Name is the display name of the project.
	// +required
	// +kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:example="My Project"
	Name string `json:"name"`

	// Visibility defines the visibility of the project.
	// +optional
	// +kubebuilder:default=public
	// +kubebuilder:validation:Enum=private;public
	// +kubebuilder:example="private"
	Visibility string `json:"visibility,omitempty"`

	// SonarRef is a reference to Sonar custom resource.
	// +required
	SonarRef common.SonarRef `json:"sonarRef"`

	// MainBranch is the key of the main branch of the project.
	// If not provided, the default main branch key will be used.
	// +optional
	// +kubebuilder:example="develop"
	MainBranch string `json:"mainBranch,omitempty"`
}

// SonarProjectStatus defines the observed state of SonarProject.
type SonarProjectStatus struct {
	// Value is a status of the project.
	// +optional
	Value string `json:"value,omitempty"`

	// Error is an error message if something went wrong.
	// +optional
	Error string `json:"error,omitempty"`

	// ProjectKey is the actual project key in SonarQube.
	// +optional
	ProjectKey string `json:"projectKey,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Key",type="string",JSONPath=".spec.key",description="Project key"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.value",description="Project status"
// +kubebuilder:printcolumn:name="Error",type="string",JSONPath=".status.error",description="Error message"

// SonarProject is the Schema for the sonarprojects API.
type SonarProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonarProjectSpec   `json:"spec,omitempty"`
	Status SonarProjectStatus `json:"status,omitempty"`
}

func (in *SonarProject) GetSonarRef() common.SonarRef {
	return in.Spec.SonarRef
}

// +kubebuilder:object:root=true

// SonarProjectList contains a list of SonarProject.
type SonarProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SonarProject `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SonarProject{}, &SonarProjectList{})
}
