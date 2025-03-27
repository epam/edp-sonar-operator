package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

// RemoveUser is handler for removing sonar user.
type RemoveUser struct {
	sonarApiClient sonar.UserInterface
}

// NewRemoveUser returns RemoveUser handler.
func NewRemoveUser(sonarApiClient sonar.UserInterface) *RemoveUser {
	return &RemoveUser{sonarApiClient: sonarApiClient}
}

// ServeRequest handles sonar user removal.
func (r RemoveUser) ServeRequest(ctx context.Context, user *sonarApi.SonarUser) error {
	log := ctrl.LoggerFrom(ctx).WithValues("userlogin", user.Spec.Login)
	log.Info("Removing user from sonar")

	if err := r.sonarApiClient.DeactivateUser(ctx, user.Spec.Login); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	log.Info("User has been deleted")

	return nil
}
