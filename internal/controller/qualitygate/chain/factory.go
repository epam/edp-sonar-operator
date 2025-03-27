package chain

import (
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

func MakeChain(sonarApiClient sonar.QualityGateClient) SonarQualityGateHandler {
	ch := &chain{}
	ch.Use(NewCreateQualityGate(sonarApiClient))
	ch.Use(NewSyncQualityGateConditions(sonarApiClient))

	return ch
}
