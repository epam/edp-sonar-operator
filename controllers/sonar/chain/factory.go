package chain

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

func MakeChain(sonarApiClient sonar.ClientInterface, k8sClient client.Client) SonarHandler {
	ch := &chain{}
	ch.Use(NewCheckConnection(sonarApiClient))
	ch.Use(NewUpdateSettings(sonarApiClient, k8sClient))
	ch.Use(NewSetDefaultPermissionTemplate(sonarApiClient))

	return ch
}
