package chain

import (
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

func MakeChain(sonarApiClient sonar.ClientInterface) SonarHandler {
	ch := &chain{}
	ch.Use(NewCheckConnection(sonarApiClient))
	ch.Use(NewUpdateSettings(sonarApiClient))
	ch.Use(NewSetDefaultPermissionTemplate(sonarApiClient))

	return ch
}
