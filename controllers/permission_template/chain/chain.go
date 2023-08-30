package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
)

type SonarPermissionTemplateHandler interface {
	ServeRequest(context.Context, *sonarApi.SonarPermissionTemplate) error
}

type chain struct {
	handlers []SonarPermissionTemplateHandler
}

func (ch *chain) Use(handlers ...SonarPermissionTemplateHandler) {
	ch.handlers = append(ch.handlers, handlers...)
}

func (ch *chain) ServeRequest(ctx context.Context, s *sonarApi.SonarPermissionTemplate) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Starting PermissionTemplate chain")

	for i := 0; i < len(ch.handlers); i++ {
		h := ch.handlers[i]

		err := h.ServeRequest(ctx, s)
		if err != nil {
			return fmt.Errorf("failed to serve handler: %w", err)
		}
	}

	log.Info("Handling of SonarPermissionTemplate has been finished")

	return nil
}
