package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/helper"
)

// SyncUserPermissions is a chain element that syncs user permissions in sonar.
type SyncUserPermissions struct {
	sonarApiClient sonar.PermissionTemplateInterface
}

// NewSyncUserPermissions returns a new instance of SyncUserPermissions.
func NewSyncUserPermissions(sonarApiClient sonar.PermissionTemplateInterface) *SyncUserPermissions {
	return &SyncUserPermissions{sonarApiClient: sonarApiClient}
}

// ServeRequest handles request to sync user permissions in sonar.
func (h SyncUserPermissions) ServeRequest(ctx context.Context, user *sonarApi.SonarUser) error {
	log := ctrl.LoggerFrom(ctx).WithValues("userlogin", user.Spec.Login)
	log.Info("Syncing user permissions in sonar")

	existingPermissions, err := h.getExistingUserPermissions(ctx, user.Spec.Login)
	if err != nil {
		return err
	}

	currentPermissions := helper.SliceToMap(user.Spec.Permissions)

	for p := range existingPermissions {
		if _, ok := currentPermissions[p]; ok {
			delete(currentPermissions, p)
			continue
		}

		if err = h.sonarApiClient.RemovePermissionFromUser(ctx, user.Spec.Login, p); err != nil {
			return fmt.Errorf("failed to remove usr permission: %w", err)
		}

		log.Info("User permission has been removed", "permission", p)
	}

	for g := range currentPermissions {
		if err = h.sonarApiClient.AddPermissionToUser(ctx, user.Spec.Login, g); err != nil {
			return fmt.Errorf("failed to add user permission: %w", err)
		}

		log.Info("User permission has been added", "permission", g)
	}

	log.Info("User permissions have been synced")

	return nil
}

func (h SyncUserPermissions) getExistingUserPermissions(ctx context.Context, userLogin string) (map[string]struct{}, error) {
	existingPermissions, err := h.sonarApiClient.GetUserPermissions(ctx, userLogin)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return helper.SliceToMap(existingPermissions), nil
}
