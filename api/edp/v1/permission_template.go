package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// SonarPermissionTemplateSpec defines the desired state of SonarPermissionTemplate.
type SonarPermissionTemplateSpec struct {
	// SonarOwner is a name of root sonar custom resource.
	SonarOwner string `json:"sonarOwner"`

	// Name is a group name.
	Name string `json:"name"`

	// ProjectKeyPattern is key pattern. Must be a valid Java regular expression.
	ProjectKeyPattern string `json:"projectKeyPattern"`

	// Description of sonar permission template.
	// +optional
	Description string `json:"description,omitempty"`

	// GroupPermissions adds a group to a permission template.
	// +nullable
	// +optional
	GroupPermissions []GroupPermission `json:"groupPermissions,omitempty"`
}

// GroupPermission represents the group and its permissions.
type GroupPermission struct {
	// Group name or 'anyone' (case insensitive). Example value sonar-administrators.
	GroupName string `json:"groupName"`

	// Permissions is a list of permissions.
	// Possible values: admin, codeviewer, issueadmin, securityhotspotadmin, scan, user.
	Permissions []string `json:"permissions"`
}

// SonarPermissionTemplateStatus defines the observed state of SonarPermissionTemplate.
type SonarPermissionTemplateStatus struct {
	// +optional
	Value string `json:"value,omitempty"`

	// +optional
	FailureCount int64 `json:"failureCount,omitempty"`

	// +optional
	ID string `json:"id,omitempty"`
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

func (in *SonarPermissionTemplate) GetFailureCount() int64 {
	return in.Status.FailureCount
}

func (in *SonarPermissionTemplate) SetFailureCount(count int64) {
	in.Status.FailureCount = count
}

func (in *SonarPermissionTemplate) GetStatus() string {
	return in.Status.Value
}

func (in *SonarPermissionTemplate) SetStatus(value string) {
	in.Status.Value = value
}

func (in *SonarPermissionTemplate) SonarOwner() string {
	return in.Spec.SonarOwner
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
