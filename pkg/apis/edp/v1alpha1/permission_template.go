package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SonarPermissionTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SonarPermissionTemplateSpec   `json:"spec"`
	Status            SonarPermissionTemplateStatus `json:"status"`
}

type SonarPermissionTemplateSpec struct {
	SonarOwner        string            `json:"sonarOwner"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	ProjectKeyPattern string            `json:"projectKeyPattern"`
	GroupPermissions  []GroupPermission `json:"groupPermissions"`
}

type GroupPermission struct {
	GroupName   string   `json:"groupName"`
	Permissions []string `json:"permissions"`
}

type SonarPermissionTemplateStatus struct {
	Value        string `json:"value"`
	FailureCount int64  `json:"failureCount"`
	ID           string `json:"id"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SonarPermissionTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SonarPermissionTemplate `json:"items"`
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
