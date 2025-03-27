package chain

import (
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

type sonarApiClient interface {
	sonar.QualityProfileClient
	sonar.RuleClient
}

func MakeChain(sonarApiClient sonarApiClient) SonarQualityProfileHandler {
	ch := &chain{}

	ch.Use(NewCreateQualityProfile(sonarApiClient))
	ch.Use(NewSyncQualityProfileRules(sonarApiClient))

	return ch
}
