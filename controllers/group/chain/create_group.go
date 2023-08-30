package chain

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

// CreateGroup is a handler for creating group.
type CreateGroup struct {
	sonarApiClient sonar.GroupInterface
}

// NewCreateGroup creates an instance of CreateGroup handler.
func NewCreateGroup(sonarApiClient sonar.GroupInterface) *CreateGroup {
	return &CreateGroup{sonarApiClient: sonarApiClient}
}

// ServeRequest implements the logic of creating group.
func (c CreateGroup) ServeRequest(ctx context.Context, group *sonarApi.SonarGroup) error {
	log := ctrl.LoggerFrom(ctx).WithValues("name", group.Spec.Name)
	log.Info("Start creating group")

	sonarGroup, err := c.sonarApiClient.GetGroup(ctx, group.Spec.Name)
	if err != nil {
		if !sonar.IsErrNotFound(err) {
			return err
		}

		log.Info("Group doesn't exist, creating new one")
		if err = c.sonarApiClient.CreateGroup(ctx, &sonar.Group{
			Name:        group.Spec.Name,
			Description: group.Spec.Description,
		}); err != nil {
			return err
		}

		log.Info("Group has been created")
		return nil
	}

	if group.Spec.Description != sonarGroup.Description {
		log.Info("Updating group")
		if err = c.sonarApiClient.UpdateGroup(ctx, group.Spec.Name, &sonar.Group{
			Name:        group.Spec.Name,
			Description: group.Spec.Description,
		}); err != nil {
			return err
		}

		log.Info("Group has been updated")
	}

	return nil
}
