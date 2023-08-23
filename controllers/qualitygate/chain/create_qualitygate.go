package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

// CreateQualityGate is a handler for creating quality gate.
type CreateQualityGate struct {
	sonarApiClient sonar.QualityGateClient
}

// NewCreateQualityGate creates an instance of CreateQualityGate handler.
func NewCreateQualityGate(sonarApiClient sonar.QualityGateClient) *CreateQualityGate {
	return &CreateQualityGate{sonarApiClient: sonarApiClient}
}

// ServeRequest implements the logic of creating quality gate.
func (h CreateQualityGate) ServeRequest(ctx context.Context, gate *sonarApi.SonarQualityGate) error {
	log := ctrl.LoggerFrom(ctx).WithValues("name", gate.Spec.Name)
	log.Info("Start creating quality gate")

	sonarGate, err := h.sonarApiClient.GetQualityGate(ctx, gate.Spec.Name)
	if err != nil {
		if !sonar.IsErrNotFound(err) {
			return fmt.Errorf("failed to get quality gate: %w", err)
		}

		log.Info("Quality gate doesn't exist, creating new one")
		if sonarGate, err = h.sonarApiClient.CreateQualityGate(ctx, gate.Spec.Name); err != nil {
			return fmt.Errorf("failed to create quality gate: %w", err)
		}

		log.Info("Quality gate has been created")
	}

	if gate.Spec.Default && gate.Spec.Default != sonarGate.IsDefault {
		log.Info("Updating default quality gate")
		if err = h.sonarApiClient.SetAsDefaultQualityGate(ctx, gate.Spec.Name); err != nil {
			return fmt.Errorf("failed to set default quality gate: %w", err)
		}

		log.Info("Default quality gate has been updated")
	}

	return nil
}
