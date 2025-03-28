package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/helper"
)

const (
	logKeyGroup = "group"
	logKeyPerm  = "permission"
)

// SyncPermissionTemplateGroups is a chain element that syncs group permission templates in sonar.
type SyncPermissionTemplateGroups struct {
	sonarApiClient sonar.PermissionTemplateInterface
}

// NewSyncPermissionTemplateGroups returns a new instance of SyncPermissionTemplateGroups.
func NewSyncPermissionTemplateGroups(sonarApiClient sonar.PermissionTemplateInterface) *SyncPermissionTemplateGroups {
	return &SyncPermissionTemplateGroups{sonarApiClient: sonarApiClient}
}

// ServeRequest handles request to permission templates group in sonar.
func (h SyncPermissionTemplateGroups) ServeRequest(ctx context.Context, template *sonarApi.SonarPermissionTemplate) error {
	log := ctrl.LoggerFrom(ctx).WithValues("name", template.Spec.Name)
	log.Info("Syncing permission template groups in sonar")

	sonarTemplate, err := h.sonarApiClient.GetPermissionTemplate(ctx, template.Spec.Name)
	if err != nil {
		return fmt.Errorf("failed to get permission template: %w", err)
	}

	existingTemplates, err := h.sonarApiClient.GetPermissionTemplateGroups(ctx, sonarTemplate.ID)
	if err != nil {
		return fmt.Errorf("failed to get permission template groups: %w", err)
	}

	for groupName, existingPermissions := range existingTemplates {
		if currentPermissions, ok := template.Spec.GroupsPermissions[groupName]; ok {
			existingPermissionsMap := helper.SliceToMap(existingPermissions)

			for _, p := range currentPermissions {
				if _, ok = existingPermissionsMap[p]; ok {
					delete(existingPermissionsMap, p)
					continue
				}

				log.Info("Adding permission template group", logKeyGroup, groupName, logKeyPerm, p)

				if err = h.sonarApiClient.AddGroupToPermissionTemplate(ctx, sonarTemplate.ID, groupName, p); err != nil {
					return fmt.Errorf("failed to add permission template group: %w", err)
				}
			}

			for p := range existingPermissionsMap {
				log.Info("Removing permission template group", logKeyGroup, groupName, logKeyPerm, p)

				if err = h.sonarApiClient.RemoveGroupFromPermissionTemplate(ctx, sonarTemplate.ID, groupName, p); err != nil {
					return fmt.Errorf("failed to remove permission template group: %w", err)
				}
			}

			continue
		}

		for _, p := range existingPermissions {
			log.Info("Removing permission template group", logKeyGroup, groupName, logKeyPerm, p)

			if err = h.sonarApiClient.RemoveGroupFromPermissionTemplate(ctx, sonarTemplate.ID, groupName, p); err != nil {
				return fmt.Errorf("failed to add permission template group: %w", err)
			}
		}
	}

	for groupName, permissions := range template.Spec.GroupsPermissions {
		if _, ok := existingTemplates[groupName]; ok {
			continue
		}

		for _, p := range permissions {
			log.Info("Adding permission template group", logKeyGroup, groupName, logKeyPerm, p)

			if err = h.sonarApiClient.AddGroupToPermissionTemplate(ctx, sonarTemplate.ID, groupName, p); err != nil {
				return fmt.Errorf("failed to add permission template group: %w", err)
			}
		}
	}

	log.Info("Permission template groups have been synced successfully")

	return nil
}
