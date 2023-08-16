package common

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
