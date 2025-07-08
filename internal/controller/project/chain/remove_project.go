package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

type RemoveProject struct {
	sonarApiClient sonar.ClientInterface
}

func (h *RemoveProject) ServeRequest(ctx context.Context, sonarProject *sonarApi.SonarProject) error {
	log := ctrl.LoggerFrom(ctx).WithValues("key", sonarProject.Spec.Key)

	// Check if project exists before attempting deletion
	_, err := h.sonarApiClient.GetProject(ctx, sonarProject.Spec.Key)
	if err != nil {
		if sonar.IsErrNotFound(err) {
			log.Info("Project does not exist, nothing to delete")

			return nil
		}

		return fmt.Errorf("failed to check if project exists: %w", err)
	}

	// Project exists, delete it
	log.Info("Deleting project from SonarQube")

	if err = h.sonarApiClient.DeleteProject(ctx, sonarProject.Spec.Key); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	log.Info("Project deleted successfully")

	return nil
}
