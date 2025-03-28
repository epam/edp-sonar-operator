package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/helper"
)

type sonarApiUserGroupClient interface {
	sonar.UserInterface
	sonar.GroupInterface
}

// SyncUserGroups is a chain element that syncs user groups in sonar.
type SyncUserGroups struct {
	sonarApiClient sonarApiUserGroupClient
}

// NewSyncUserGroups returns a new instance of SyncUserGroups.
func NewSyncUserGroups(sonarApiClient sonarApiUserGroupClient) *SyncUserGroups {
	return &SyncUserGroups{sonarApiClient: sonarApiClient}
}

// ServeRequest handles request to sync user groups in sonar.
func (h SyncUserGroups) ServeRequest(ctx context.Context, user *sonarApi.SonarUser) error {
	log := ctrl.LoggerFrom(ctx).WithValues("userlogin", user.Spec.Login)
	log.Info("Syncing user groups in sonar")

	existingGroups, err := h.getExistingUserGroups(ctx, user.Spec.Login)
	if err != nil {
		return err
	}

	currentGroups := helper.SliceToMap(user.Spec.Groups)

	for g := range existingGroups {
		if _, ok := currentGroups[g]; ok || g == "sonar-users" {
			delete(currentGroups, g)
			continue
		}

		if err = h.sonarApiClient.RemoveUserFromGroup(ctx, user.Spec.Login, g); err != nil {
			return fmt.Errorf("failed to remove user from group: %w", err)
		}

		log.Info("User has been removed from group", "group", g)
	}

	for g := range currentGroups {
		if err = h.sonarApiClient.AddUserToGroup(ctx, user.Spec.Login, g); err != nil {
			return fmt.Errorf("failed to add user to group: %w", err)
		}

		log.Info("User has been added to group", "group", g)
	}

	log.Info("User groups have been synced")

	return nil
}

func (h SyncUserGroups) getExistingUserGroups(ctx context.Context, userLogin string) (map[string]struct{}, error) {
	existingGroups, err := h.sonarApiClient.GetUserGroups(ctx, userLogin)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	m := make(map[string]struct{}, len(existingGroups))

	for _, g := range existingGroups {
		m[g.Name] = struct{}{}
	}

	return m, nil
}
