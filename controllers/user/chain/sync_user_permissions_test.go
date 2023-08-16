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

func TestSyncUserPermissions_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		user           *sonarApi.SonarUser
		sonarApiClient func(t *testing.T) sonar.PermissionTemplateInterface
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "user permissions synced successfully",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:       "test-user",
					Permissions: []string{"test-permission", "test-permission-2"},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetUserPermissions", mock.Anything, "test-user").
					Return([]string{"test-permission", "test-permission-3"}, nil)
				m.On("RemovePermissionFromUser", mock.Anything, "test-user", "test-permission-3").
					Return(nil)
				m.On("AddPermissionToUser", mock.Anything, "test-user", "test-permission-2").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to add user permission",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:       "test-user",
					Permissions: []string{"test-permission", "test-permission-2"},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetUserPermissions", mock.Anything, "test-user").
					Return([]string{"test-permission", "test-permission-3"}, nil)
				m.On("RemovePermissionFromUser", mock.Anything, "test-user", "test-permission-3").
					Return(nil)
				m.On("AddPermissionToUser", mock.Anything, "test-user", "test-permission-2").
					Return(errors.New("failed to add user permission"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to add user permission")
			},
		},
		{
			name: "failed to remove user permission",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:       "test-user",
					Permissions: []string{"test-permission", "test-permission-2"},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetUserPermissions", mock.Anything, "test-user").
					Return([]string{"test-permission", "test-permission-3"}, nil)
				m.On("RemovePermissionFromUser", mock.Anything, "test-user", "test-permission-3").
					Return(errors.New("failed to remove user permission"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to remove user permission")
			},
		},
		{
			name: "failed to get user permissions",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:       "test-user",
					Permissions: []string{"test-permission", "test-permission-2"},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetUserPermissions", mock.Anything, "test-user").
					Return(nil, errors.New("failed to get user permissions"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get user permissions")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewSyncUserPermissions(tt.sonarApiClient(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.user)
			tt.wantErr(t, err)
		})
	}
}
