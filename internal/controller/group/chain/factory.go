package chain

import (
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

type sonarApiClient interface {
	sonar.GroupInterface
	sonar.PermissionTemplateInterface
}

func MakeChain(sonarApiClient sonarApiClient) SonarGroupHandler {
	ch := &chain{}

	ch.Use(NewCreateGroup(sonarApiClient))
	ch.Use(NewSyncGroupPermissions(sonarApiClient))

	return ch
}
