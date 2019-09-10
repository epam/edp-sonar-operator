package controller

import (
	"github.com/epmd-edp/sonar-operator/v2/pkg/controller/sonar"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, sonar.Add)
}
