package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

type SetDefaultPermissionTemplate struct {
	sonarApiClient sonar.PermissionTemplateInterface
}

func NewSetDefaultPermissionTemplate(sonarApiClient sonar.PermissionTemplateInterface) *SetDefaultPermissionTemplate {
	return &SetDefaultPermissionTemplate{sonarApiClient: sonarApiClient}
}

func (h *SetDefaultPermissionTemplate) ServeRequest(ctx context.Context, sonar *sonarApi.Sonar) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Start updating default permission template to sonar")

	if sonar.Spec.DefaultPermissionTemplate == "" {
		log.Info("Default permission template is not set")
		return nil
	}

	if err := h.sonarApiClient.SetDefaultPermissionTemplate(ctx, sonar.Spec.DefaultPermissionTemplate); err != nil {
		return fmt.Errorf("failed to set default permission template %s: %w", sonar.Spec.DefaultPermissionTemplate, err)
	}

	log.Info("Sonar default permission template have been updated")

	return nil
}
