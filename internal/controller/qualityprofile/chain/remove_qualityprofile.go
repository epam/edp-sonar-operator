package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

// RemoveQualityProfile is a handler for removing quality profile.
type RemoveQualityProfile struct {
	sonarApiClient sonar.QualityProfileClient
}

// NewRemoveQualityProfile creates an instance of RemoveQualityProfile handler.
func NewRemoveQualityProfile(sonarApiClient sonar.QualityProfileClient) *RemoveQualityProfile {
	return &RemoveQualityProfile{sonarApiClient: sonarApiClient}
}

// ServeRequest implements the logic of removing quality profile.
func (r RemoveQualityProfile) ServeRequest(ctx context.Context, profile *sonarApi.SonarQualityProfile) error {
	log := ctrl.LoggerFrom(ctx).WithValues("name", profile.Spec.Name)
	log.Info("Start removing quality profile")

	if err := r.sonarApiClient.DeleteQualityProfile(ctx, profile.Spec.Name, profile.Spec.Language); err != nil {
		if !sonar.IsErrNotFound(err) {
			return fmt.Errorf("failed to delete quality profile: %w", err)
		}
	}

	log.Info("Quality profile has been removed")

	return nil
}
