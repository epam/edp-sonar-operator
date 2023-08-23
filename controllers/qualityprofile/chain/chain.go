package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
)

type SonarQualityProfileHandler interface {
	ServeRequest(context.Context, *sonarApi.SonarQualityProfile) error
}

type chain struct {
	handlers []SonarQualityProfileHandler
}

func (ch *chain) Use(handlers ...SonarQualityProfileHandler) {
	ch.handlers = append(ch.handlers, handlers...)
}

func (ch *chain) ServeRequest(ctx context.Context, s *sonarApi.SonarQualityProfile) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Starting SonarQualityProfile chain")

	for i := 0; i < len(ch.handlers); i++ {
		h := ch.handlers[i]

		err := h.ServeRequest(ctx, s)
		if err != nil {
			return fmt.Errorf("failed to serve handler: %w", err)
		}
	}

	log.Info("Handling of SonarQualityProfile has been finished")

	return nil
}
