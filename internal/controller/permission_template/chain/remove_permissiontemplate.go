package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

// RemovePermissionTemplate is a handler for removing permission template.
type RemovePermissionTemplate struct {
	sonarApiClient sonar.PermissionTemplateInterface
}

// NewRemovePermissionTemplate creates an instance of RemovePermissionTemplate handler.
func NewRemovePermissionTemplate(sonarApiClient sonar.PermissionTemplateInterface) *RemovePermissionTemplate {
	return &RemovePermissionTemplate{sonarApiClient: sonarApiClient}
}

// ServeRequest implements the logic of removing permission template.
func (c RemovePermissionTemplate) ServeRequest(ctx context.Context, template *sonarApi.SonarPermissionTemplate) error {
	log := ctrl.LoggerFrom(ctx).WithValues("name", template.Spec.Name)
	log.Info("Start removing permission template")

	if err := c.sonarApiClient.DeletePermissionTemplate(ctx, template.Spec.Name); err != nil {
		if !sonar.IsErrNotFound(err) {
			return fmt.Errorf("failed to delete template: %w", err)
		}
	}

	log.Info("Permission template has been removed")

	return nil
}
