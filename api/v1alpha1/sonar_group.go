package v1alpha1

import (
	"github.com/epam/edp-sonar-operator/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SonarGroupSpec defines the desired state of SonarGroup.
type SonarGroupSpec struct {
	// Name is a group name.
	// Name should be unique across all groups.
	// Do not edit this field after creation. Otherwise, the group will be recreated.
	// +required
	// +kubebuilder:validation:MaxLength=255
	// +kubebuilder:example="sonar-users"
	Name string `json:"name"`

	// Description of sonar group.
	// +optional
	// +kubebuilder:validation:MaxLength=200
	// +kubebuilder:example="Default group for new users"
	Description string `json:"description,omitempty"`

	// Permissions is a list of permissions assigned to group.
	// +nullable
	// +optional
	// +kubebuilder:example={admin, provisioning}
	Permissions []string `json:"permissions,omitempty"`

	// SonarRef is a reference to Sonar custom resource.
	// +required
	SonarRef common.SonarRef `json:"sonarRef"`
}

// SonarGroupStatus defines the observed state of SonarGroup.
type SonarGroupStatus struct {
	// Value is a status of the group.
	// +optional
	Value string `json:"value,omitempty"`

	// Error is an error message if something went wrong.
	// +optional
	Error string `json:"error,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// SonarGroup is the Schema for the sonar group API.
type SonarGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonarGroupSpec   `json:"spec,omitempty"`
	Status SonarGroupStatus `json:"status,omitempty"`
}

func (in *SonarGroup) GetSonarRef() common.SonarRef {
	return in.Spec.SonarRef
}

// +kubebuilder:object:root=true

// SonarGroupList contains a list of SonarGroup.
type SonarGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SonarGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SonarGroup{}, &SonarGroupList{})
}
