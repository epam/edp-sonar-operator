package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/helper"
)

// SyncGroupPermissions is a chain element that syncs group permissions in sonar.
type SyncGroupPermissions struct {
	sonarApiClient sonar.PermissionTemplateInterface
}

// NewSyncGroupPermissions returns a new instance of SyncGroupPermissions.
func NewSyncGroupPermissions(sonarApiClient sonar.PermissionTemplateInterface) *SyncGroupPermissions {
	return &SyncGroupPermissions{sonarApiClient: sonarApiClient}
}

// ServeRequest handles request to sync group permissions in sonar.
func (h SyncGroupPermissions) ServeRequest(ctx context.Context, group *sonarApi.SonarGroup) error {
	log := ctrl.LoggerFrom(ctx).WithValues("name", group.Spec.Name)
	log.Info("Syncing group permissions in sonar")

	existingPermissions, err := h.getExistingGroupPermissions(ctx, group.Spec.Name)
	if err != nil {
		return err
	}

	currentPermissions := helper.SliceToMap(group.Spec.Permissions)

	for p := range existingPermissions {
		if _, ok := currentPermissions[p]; ok {
			delete(currentPermissions, p)
			continue
		}

		if err = h.sonarApiClient.RemovePermissionFromGroup(ctx, group.Spec.Name, p); err != nil {
			return fmt.Errorf("failed to remove group permission: %w", err)
		}

		log.Info("Group permission has been removed", "permission", p)
	}

	for g := range currentPermissions {
		if err = h.sonarApiClient.AddPermissionToGroup(ctx, group.Spec.Name, g); err != nil {
			return fmt.Errorf("failed to add group permission: %w", err)
		}

		log.Info("Group permission has been added", "permission", g)
	}

	log.Info("Group permissions have been synced")

	return nil
}

func (h SyncGroupPermissions) getExistingGroupPermissions(ctx context.Context, groupName string) (map[string]struct{}, error) {
	existingPermissions, err := h.sonarApiClient.GetGroupPermissions(ctx, groupName)
	if err != nil {
		return nil, fmt.Errorf("failed to get group permissions: %w", err)
	}

	return helper.SliceToMap(existingPermissions), nil
}
