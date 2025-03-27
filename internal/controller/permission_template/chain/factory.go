package chain

import (
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

type sonarApiClient interface {
	sonar.PermissionTemplateInterface
}

func MakeChain(sonarApiClient sonarApiClient) SonarPermissionTemplateHandler {
	ch := &chain{}

	ch.Use(NewCreatePermissionTemplate(sonarApiClient))
	ch.Use(NewSyncPermissionTemplateGroups(sonarApiClient))

	return ch
}
