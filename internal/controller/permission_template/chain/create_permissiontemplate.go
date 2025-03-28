package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

// CreatePermissionTemplate is a handler for creating permission template.
type CreatePermissionTemplate struct {
	sonarApiClient sonar.PermissionTemplateInterface
}

// NewCreatePermissionTemplate creates an instance of CreatePermissionTemplate handler.
func NewCreatePermissionTemplate(sonarApiClient sonar.PermissionTemplateInterface) *CreatePermissionTemplate {
	return &CreatePermissionTemplate{sonarApiClient: sonarApiClient}
}

// ServeRequest implements the logic of creating permission template.
func (h CreatePermissionTemplate) ServeRequest(ctx context.Context, template *sonarApi.SonarPermissionTemplate) error {
	log := ctrl.LoggerFrom(ctx).WithValues("name", template.Spec.Name)
	log.Info("Start creating permission template")

	sonarTemplate, err := h.sonarApiClient.GetPermissionTemplate(ctx, template.Spec.Name)
	if err != nil {
		if !sonar.IsErrNotFound(err) {
			return fmt.Errorf("failed to get permission template: %w", err)
		}

		log.Info("Permission template doesn't exist, creating new one")

		if sonarTemplate, err = h.sonarApiClient.CreatePermissionTemplate(ctx, &sonar.PermissionTemplateData{
			Name:              template.Spec.Name,
			Description:       template.Spec.Description,
			ProjectKeyPattern: template.Spec.ProjectKeyPattern,
		}); err != nil {
			return fmt.Errorf("failed to create permission template: %w", err)
		}

		log.Info("Permission template has been created")
	}

	if template.Spec.Description != sonarTemplate.Description || template.Spec.ProjectKeyPattern != sonarTemplate.ProjectKeyPattern {
		log.Info("Updating permission template")

		sonarTemplate.Description = template.Spec.Description
		sonarTemplate.ProjectKeyPattern = template.Spec.ProjectKeyPattern

		if err = h.sonarApiClient.UpdatePermissionTemplate(ctx, sonarTemplate); err != nil {
			return fmt.Errorf("failed to update permission template: %w", err)
		}

		log.Info("Permission template has been updated")
	}

	if template.Spec.Default && template.Spec.Default != sonarTemplate.IsDefault {
		log.Info("Updating default permission template")

		if err = h.sonarApiClient.SetDefaultPermissionTemplate(ctx, template.Spec.Name); err != nil {
			return fmt.Errorf("failed to set default permission template: %w", err)
		}

		log.Info("Default permission template has been updated")
	}

	return nil
}
