package v1alpha1

import (
	"github.com/epam/edp-sonar-operator/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SonarUserSpec defines the desired state of SonarUser
type SonarUserSpec struct {
	// Email is a user email.
	// +optional
	// +kubebuilder:validation:MaxLength=100
	// +kubebuilder:example="myname@email.com"
	Email string `json:"email,omitempty"`

	// Login is a user login.
	// Do not edit this field after creation. Otherwise, the user will be recreated.
	// +required
	// +kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:example="myuser"
	Login string `json:"login"`

	// Name is a username.
	// +required
	// +kubebuilder:validation:MaxLength=200
	// +kubebuilder:example="My Name"
	Name string `json:"name"`

	// Groups is a list of groups assigned to user.
	// +nullable
	// +optional
	// +kubebuilder:example={sonar-administrators, developers}
	Groups []string `json:"groups,omitempty"`

	// Permissions is a list of permissions assigned to user.
	// +nullable
	// +optional
	// +kubebuilder:example={admin, provisioning}
	Permissions []string `json:"permissions,omitempty"`

	// Secret is the name of the secret with the user password.
	// It should contain a password field with a user password.
	// User password can't be updated.
	// +required
	// +kubebuilder:example="sonar-user-password"
	Secret string `json:"secret"`

	// SonarRef is a reference to Sonar custom resource.
	// +required
	SonarRef common.SonarRef `json:"sonarRef"`
}

// SonarUserStatus defines the observed state of SonarUser
type SonarUserStatus struct {
	// Value is a status of the user.
	// +optional
	Value string `json:"value,omitempty"`

	// Error is an error message if something went wrong.
	// +optional
	Error string `json:"error,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// SonarUser is the Schema for the sonarusers API.
type SonarUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonarUserSpec   `json:"spec,omitempty"`
	Status SonarUserStatus `json:"status,omitempty"`
}

func (in *SonarUser) GetSonarRef() common.SonarRef {
	return in.Spec.SonarRef
}

// +kubebuilder:object:root=true

// SonarUserList contains a list of SonarUser
type SonarUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SonarUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SonarUser{}, &SonarUserList{})
}
