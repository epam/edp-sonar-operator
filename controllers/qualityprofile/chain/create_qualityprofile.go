package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

// CreateQualityProfile is a handler for creating quality profile.
type CreateQualityProfile struct {
	sonarApiClient sonar.QualityProfileClient
}

// NewCreateQualityProfile creates an instance of CreateQualityProfile handler.
func NewCreateQualityProfile(sonarApiClient sonar.QualityProfileClient) *CreateQualityProfile {
	return &CreateQualityProfile{sonarApiClient: sonarApiClient}
}

// ServeRequest implements the logic of creating quality profile.
func (h CreateQualityProfile) ServeRequest(ctx context.Context, profile *sonarApi.SonarQualityProfile) error {
	log := ctrl.LoggerFrom(ctx).WithValues("name", profile.Spec.Name)
	log.Info("Start creating quality profile")

	sonarProfile, err := h.sonarApiClient.GetQualityProfile(ctx, profile.Spec.Name)
	if err != nil {
		if !sonar.IsErrNotFound(err) {
			return fmt.Errorf("failed to get quality profile: %w", err)
		}

		log.Info("Quality profile doesn't exist, creating new one")
		if sonarProfile, err = h.sonarApiClient.CreateQualityProfile(ctx, profile.Spec.Name, profile.Spec.Language); err != nil {
			return fmt.Errorf("failed to create quality profile: %w", err)
		}

		log.Info("Quality profile has been created")
	}

	if profile.Spec.Default && profile.Spec.Default != sonarProfile.IsDefault {
		log.Info("Updating default quality profile")
		if err = h.sonarApiClient.SetAsDefaultQualityProfile(ctx, profile.Spec.Name, profile.Spec.Language); err != nil {
			return fmt.Errorf("failed to set default quality profile: %w", err)
		}

		log.Info("Default quality profile has been updated")
	}

	return nil
}
