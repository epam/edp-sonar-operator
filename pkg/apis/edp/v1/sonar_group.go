package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// SonarGroupSpec defines the desired state of SonarGroup.
type SonarGroupSpec struct {
	// SonarOwner is a name of root sonar custom resource.
	SonarOwner string `json:"sonarOwner"`

	// Name is a group name.
	Name string `json:"name"`

	// Description of sonar group.
	// +optional
	Description string `json:"description,omitempty"`
}

// SonarGroupStatus defines the observed state of SonarGroup.
type SonarGroupStatus struct {
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

// SonarGroup is the Schema for the sonar group API.
type SonarGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonarGroupSpec   `json:"spec,omitempty"`
	Status SonarGroupStatus `json:"status,omitempty"`
}

func (in *SonarGroup) GetFailureCount() int64 {
	return in.Status.FailureCount
}

func (in *SonarGroup) SetFailureCount(count int64) {
	in.Status.FailureCount = count
}

func (in *SonarGroup) GetStatus() string {
	return in.Status.Value
}

func (in *SonarGroup) SetStatus(value string) {
	in.Status.Value = value
}

func (in *SonarGroup) SonarOwner() string {
	return in.Spec.SonarOwner
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
