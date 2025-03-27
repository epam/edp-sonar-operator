package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

// RemoveGroup is a handler for removing group.
type RemoveGroup struct {
	sonarApiClient sonar.GroupInterface
}

// NewRemoveGroup creates an instance of RemoveGroup handler.
func NewRemoveGroup(sonarApiClient sonar.GroupInterface) *RemoveGroup {
	return &RemoveGroup{sonarApiClient: sonarApiClient}
}

// ServeRequest implements the logic of removing group.
func (c RemoveGroup) ServeRequest(ctx context.Context, group *sonarApi.SonarGroup) error {
	log := ctrl.LoggerFrom(ctx).WithValues("name", group.Spec.Name)
	log.Info("Start removing group")

	if err := c.sonarApiClient.DeleteGroup(ctx, group.Spec.Name); err != nil {
		if !sonar.IsErrNotFound(err) {
			return fmt.Errorf("failed to delete group: %w", err)
		}
	}

	return nil
}
