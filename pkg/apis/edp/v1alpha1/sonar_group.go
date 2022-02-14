package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SonarGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SonarGroupSpec   `json:"spec"`
	Status            SonarGroupStatus `json:"status"`
}

type SonarGroupSpec struct {
	SonarOwner  string `json:"sonarOwner"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SonarGroupStatus struct {
	Value        string `json:"value"`
	FailureCount int64  `json:"failureCount"`
	ID           string `json:"id"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SonarGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SonarGroup `json:"items"`
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
