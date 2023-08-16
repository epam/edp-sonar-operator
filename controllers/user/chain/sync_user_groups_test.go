package chain

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar/mocks"
)

func TestSyncUserGroups_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		user           *sonarApi.SonarUser
		sonarApiClient func(t *testing.T) sonarApiUserGroupClient
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "user groups synced successfully",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:  "test-user",
					Groups: []string{"test-group", "test-group-2"},
				},
			},
			sonarApiClient: func(t *testing.T) sonarApiUserGroupClient {
				m := mocks.NewClientInterface(t)

				m.On("GetUserGroups", mock.Anything, "test-user").
					Return([]sonar.Group{
						{Name: "test-group"},
						{Name: "test-group-3"},
					}, nil)

				m.On("RemoveUserFromGroup", mock.Anything, "test-user", "test-group-3").
					Return(nil)

				m.On("AddUserToGroup", mock.Anything, "test-user", "test-group-2").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to add user to group",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:  "test-user",
					Groups: []string{"test-group", "test-group-2"},
				},
			},
			sonarApiClient: func(t *testing.T) sonarApiUserGroupClient {
				m := mocks.NewClientInterface(t)

				m.On("GetUserGroups", mock.Anything, "test-user").
					Return([]sonar.Group{
						{Name: "test-group"},
						{Name: "test-group-3"},
					}, nil)

				m.On("RemoveUserFromGroup", mock.Anything, "test-user", "test-group-3").
					Return(nil)

				m.On("AddUserToGroup", mock.Anything, "test-user", "test-group-2").
					Return(errors.New("failed to add user to group"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to add user to group")
			},
		},
		{
			name: "failed to remove user from group",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:  "test-user",
					Groups: []string{"test-group", "test-group-2"},
				},
			},
			sonarApiClient: func(t *testing.T) sonarApiUserGroupClient {
				m := mocks.NewClientInterface(t)

				m.On("GetUserGroups", mock.Anything, "test-user").
					Return([]sonar.Group{
						{Name: "test-group"},
						{Name: "test-group-3"},
					}, nil)

				m.On("RemoveUserFromGroup", mock.Anything, "test-user", "test-group-3").
					Return(errors.New("failed to remove user from group"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to remove user from group")
			},
		},
		{
			name: "failed to get user groups",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:  "test-user",
					Groups: []string{"test-group", "test-group-2"},
				},
			},
			sonarApiClient: func(t *testing.T) sonarApiUserGroupClient {
				m := mocks.NewClientInterface(t)

				m.On("GetUserGroups", mock.Anything, "test-user").
					Return(nil, errors.New("failed to get user groups"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get user groups")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewSyncUserGroups(tt.sonarApiClient(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.user)
			tt.wantErr(t, err)
		})
	}
}
