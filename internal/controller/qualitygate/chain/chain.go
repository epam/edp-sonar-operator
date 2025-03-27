package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
)

type SonarQualityGateHandler interface {
	ServeRequest(context.Context, *sonarApi.SonarQualityGate) error
}

type chain struct {
	handlers []SonarQualityGateHandler
}

func (ch *chain) Use(handlers ...SonarQualityGateHandler) {
	ch.handlers = append(ch.handlers, handlers...)
}

func (ch *chain) ServeRequest(ctx context.Context, s *sonarApi.SonarQualityGate) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Starting SonarQualityGate chain")

	for i := 0; i < len(ch.handlers); i++ {
		h := ch.handlers[i]

		err := h.ServeRequest(ctx, s)
		if err != nil {
			return fmt.Errorf("failed to serve handler: %w", err)
		}
	}

	log.Info("Handling of SonarQualityGate has been finished")

	return nil
}
