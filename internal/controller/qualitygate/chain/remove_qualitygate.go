package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

// RemoveQualityGate is a handler for removing quality gate.
type RemoveQualityGate struct {
	sonarApiClient sonar.QualityGateClient
}

// NewRemoveQualityGate creates an instance of RemoveQualityGate handler.
func NewRemoveQualityGate(sonarApiClient sonar.QualityGateClient) *RemoveQualityGate {
	return &RemoveQualityGate{sonarApiClient: sonarApiClient}
}

// ServeRequest implements the logic of removing quality gate.
func (r RemoveQualityGate) ServeRequest(ctx context.Context, gate *sonarApi.SonarQualityGate) error {
	log := ctrl.LoggerFrom(ctx).WithValues("name", gate.Spec.Name)
	log.Info("Start removing quality gate")

	if err := r.sonarApiClient.DeleteQualityGate(ctx, gate.Spec.Name); err != nil {
		if !sonar.IsErrNotFound(err) {
			return fmt.Errorf("failed to delete quality gate: %w", err)
		}
	}

	log.Info("Quality gate has been removed")

	return nil
}
