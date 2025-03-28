package chain

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

func MakeChain(sonarApiClient sonar.ClientInterface, cl client.Client) SonarUserHandler {
	ch := &chain{}
	ch.Use(NewCreateUser(sonarApiClient, cl))
	ch.Use(NewSyncUserGroups(sonarApiClient))
	ch.Use(NewSyncUserPermissions(sonarApiClient))

	return ch
}
