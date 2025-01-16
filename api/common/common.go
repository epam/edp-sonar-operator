package common

import (
	corev1 "k8s.io/api/core/v1"
)

// StatusCreated is success status for Sonar resources.
const StatusCreated = "created"

// SonarRef is a reference to a Sonar instance.
type SonarRef struct {
	// Kind specifies the kind of the Sonar resource.
	// +optional
	// +kubebuilder:default=Sonar
	Kind string `json:"kind"`

	// Name specifies the name of the Sonar resource.
	// +required
	Name string `json:"name"`
}

type HasSonarRef interface {
	GetSonarRef() SonarRef
}

// SourceRef is a reference to a key in a ConfigMap or a Secret.
// +kubebuilder:object:generate=true
type SourceRef struct {
	// Selects a key of a ConfigMap.
	// +optional
	ConfigMapKeyRef *ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`

	// Selects a key of a secret.
	// +optional
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty"`
}

type ConfigMapKeySelector struct {
	// The ConfigMap to select from.
	corev1.LocalObjectReference `json:",inline"`
	// The key to select.
	Key string `json:"key"`
}

type SecretKeySelector struct {
	// The name of the secret.
	corev1.LocalObjectReference `json:",inline"`
	// The key of the secret to select from.
	Key string `json:"key"`
}
