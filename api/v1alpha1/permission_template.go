package v1alpha1

import (
	"github.com/epam/edp-sonar-operator/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SonarPermissionTemplateSpec defines the desired state of SonarPermissionTemplate.
type SonarPermissionTemplateSpec struct {
	// Name is a name of permission template.
	// Name should be unique across all permission templates.
	// Do not edit this field after creation. Otherwise, the permission template will be recreated.
	// +required
	// +kubebuilder:validation:MaxLength=100
	// +kubebuilder:example="sonar-users-tmpl"
	Name string `json:"name"`

	// Description of sonar permission template.
	// +optional
	// +kubebuilder:example="Default permission template for new users"
	Description string `json:"description,omitempty"`

	// ProjectKeyPattern is key pattern. Must be a valid Java regular expression.
	// +optional
	// +kubebuilder:example="finance.*"
	ProjectKeyPattern string `json:"projectKeyPattern"`

	// Default is a flag to set permission template as default.
	// Only one permission template can be default.
	// If several permission templates have default flag, the random one will be chosen.
	// Default permission template can't be deleted. You need to set another permission template as default before.
	// +optional
	// +kubebuilder:example="true"
	Default bool `json:"default"`

	// SonarRef is a reference to Sonar custom resource.
	// +required
	SonarRef common.SonarRef `json:"sonarRef"`
}

// SonarPermissionTemplateStatus defines the observed state of SonarPermissionTemplate.
type SonarPermissionTemplateStatus struct {
	// Value is a status of the permission template.
	// +optional
	Value string `json:"value,omitempty"`

	// Error is an error message if something went wrong.
	// +optional
	Error string `json:"error,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// SonarPermissionTemplate is the Schema for the sonar permission template API.
type SonarPermissionTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonarPermissionTemplateSpec   `json:"spec,omitempty"`
	Status SonarPermissionTemplateStatus `json:"status,omitempty"`
}

func (in *SonarPermissionTemplate) GetSonarRef() common.SonarRef {
	return in.Spec.SonarRef
}

// +kubebuilder:object:root=true

// SonarPermissionTemplateList contains a list of SonarPermissionTemplate.
type SonarPermissionTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SonarPermissionTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SonarPermissionTemplate{}, &SonarPermissionTemplateList{})
}
