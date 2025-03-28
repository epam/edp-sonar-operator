package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

type CheckConnection struct {
	sonarApiClient sonar.System
}

func NewCheckConnection(sonarApiClient sonar.System) *CheckConnection {
	return &CheckConnection{sonarApiClient: sonarApiClient}
}

func (h *CheckConnection) ServeRequest(ctx context.Context, sonar *sonarApi.Sonar) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Start checking connection to sonar")

	systemHealth, err := h.sonarApiClient.Health(ctx)
	if err != nil {
		return fmt.Errorf("failed to get health: %w", err)
	}

	sonar.Status.Connected = true
	sonar.Status.Value = systemHealth.Health

	log.Info("Connection to sonar is established")

	return nil
}
