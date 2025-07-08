package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

type CreateProject struct {
	sonarApiClient sonar.ClientInterface
}

func NewCreateProject(sonarApiClient sonar.ClientInterface) SonarProjectHandler {
	return &CreateProject{sonarApiClient: sonarApiClient}
}

func (h *CreateProject) ServeRequest(ctx context.Context, sonarProject *sonarApi.SonarProject) error {
	log := ctrl.LoggerFrom(ctx).WithValues("key", sonarProject.Spec.Key)

	sonarProj := &sonar.Project{
		Key:        sonarProject.Spec.Key,
		Name:       sonarProject.Spec.Name,
		Visibility: sonarProject.Spec.Visibility,
		MainBranch: sonarProject.Spec.MainBranch,
	}

	// Check if project already exists
	existingProject, err := h.sonarApiClient.GetProject(ctx, sonarProject.Spec.Key)
	if err != nil {
		// If error is "not found", project doesn't exist, so create it
		if !sonar.IsErrNotFound(err) {
			return fmt.Errorf("failed to check if project exists: %w", err)
		}

		// Project doesn't exist, create it
		log.Info("Project does not exist, creating")

		if err = h.sonarApiClient.CreateProject(ctx, sonarProj); err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}

		log.Info("Project created successfully")

		return nil
	}

	// Project exists, check if update is needed
	log.Info("Project already exists, checking for updates")

	if existingProject.Visibility != sonarProject.Spec.Visibility {
		log.Info("Updating project")

		if err = h.sonarApiClient.UpdateProject(ctx, sonarProj); err != nil {
			return fmt.Errorf("failed to update project: %w", err)
		}

		log.Info("Project updated successfully")
	}

	return nil
}
