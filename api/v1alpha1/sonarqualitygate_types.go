package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/epam/edp-sonar-operator/api/common"
)

// SonarQualityGateSpec defines the desired state of SonarQualityGate
type SonarQualityGateSpec struct {
	// Name is a name of quality gate.
	// Name should be unique across all quality gates.
	// Don't change this field after creation otherwise quality gate will be recreated.
	// +required
	// +kubebuilder:validation:MaxLength=100
	// +kubebuilder:example="My Quality Gate"
	Name string `json:"name"`

	// Default is a flag to set quality gate as default.
	// Only one quality gate can be default.
	// If several quality gates have default flag, the random one will be chosen.
	// Default quality gate can't be deleted. You need to set another quality gate as default before.
	// +optional
	// +kubebuilder:example="true"
	Default bool `json:"default"`

	// Conditions is a list of conditions for quality gate.
	// Key is a metric name, value is a condition.
	// +optional
	// +nullable
	// +kubebuilder:example={new_code_smells: {error: "10", op: "LT"}}
	Conditions map[string]Condition `json:"conditions"`

	// SonarRef is a reference to Sonar custom resource.
	// +required
	SonarRef common.SonarRef `json:"sonarRef"`
}

// Condition defines the condition for quality gate.
type Condition struct {
	// Error is condition error threshold.
	// +required
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:example="10"
	Error string `json:"error"`

	// Op is condition operator.
	// LT = is lower than
	// GT = is greater than
	// +optional
	// +kubebuilder:validation:Enum=LT;GT
	Op string `json:"op,omitempty"`
}

// SonarQualityGateStatus defines the observed state of SonarQualityGate
type SonarQualityGateStatus struct {
	// Value is a status of the user.
	// +optional
	Value string `json:"value,omitempty"`

	// Error is an error message if something went wrong.
	// +optional
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SonarQualityGate is the Schema for the sonarqualitygates API
type SonarQualityGate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonarQualityGateSpec   `json:"spec,omitempty"`
	Status SonarQualityGateStatus `json:"status,omitempty"`
}

func (in *SonarQualityGate) GetSonarRef() common.SonarRef {
	return in.Spec.SonarRef
}

//+kubebuilder:object:root=true

// SonarQualityGateList contains a list of SonarQualityGate
type SonarQualityGateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SonarQualityGate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SonarQualityGate{}, &SonarQualityGateList{})
}
