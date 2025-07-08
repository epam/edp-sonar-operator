package chain

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

func MakeChain(sonarApiClient sonar.ClientInterface, cl client.Client) SonarProjectHandler {
	ch := &chain{}
	ch.Use(NewCreateProject(sonarApiClient))

	return ch
}

func NewRemoveProject(sonarApiClient sonar.ClientInterface) SonarProjectHandler {
	return &RemoveProject{sonarApiClient: sonarApiClient}
}
