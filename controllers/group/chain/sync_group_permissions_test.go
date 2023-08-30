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

func TestSyncGroupPermissions_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		group          *sonarApi.SonarGroup
		sonarApiClient func(t *testing.T) sonar.PermissionTemplateInterface
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "group permissions synced successfully",
			group: &sonarApi.SonarGroup{
				Spec: sonarApi.SonarGroupSpec{
					Name:        "test-group",
					Permissions: []string{"test-permission1", "test-permission2"},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)
				m.On("GetGroupPermissions", mock.Anything, "test-group").
					Return([]string{"test-permission1", "test-permission3"}, nil)
				m.On("RemovePermissionFromGroup", mock.Anything, "test-group", "test-permission3").
					Return(nil)
				m.On("AddPermissionToGroup", mock.Anything, "test-group", "test-permission2").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to add group permission",
			group: &sonarApi.SonarGroup{
				Spec: sonarApi.SonarGroupSpec{
					Name:        "test-group",
					Permissions: []string{"test-permission1", "test-permission2"},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)
				m.On("GetGroupPermissions", mock.Anything, "test-group").
					Return([]string{"test-permission1", "test-permission3"}, nil)
				m.On("RemovePermissionFromGroup", mock.Anything, "test-group", "test-permission3").
					Return(nil)
				m.On("AddPermissionToGroup", mock.Anything, "test-group", "test-permission2").
					Return(errors.New("failed to add group permission"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to add group permission")
			},
		},
		{
			name: "failed to remove group permission",
			group: &sonarApi.SonarGroup{
				Spec: sonarApi.SonarGroupSpec{
					Name:        "test-group",
					Permissions: []string{"test-permission1", "test-permission2"},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)
				m.On("GetGroupPermissions", mock.Anything, "test-group").
					Return([]string{"test-permission1", "test-permission3"}, nil)
				m.On("RemovePermissionFromGroup", mock.Anything, "test-group", "test-permission3").
					Return(errors.New("failed to remove group permission"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to remove group permission")
			},
		},
		{
			name: "failed to get group permission",
			group: &sonarApi.SonarGroup{
				Spec: sonarApi.SonarGroupSpec{
					Name:        "test-group",
					Permissions: []string{"test-permission1", "test-permission2"},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)
				m.On("GetGroupPermissions", mock.Anything, "test-group").
					Return(nil, errors.New("failed to get group permissions"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get group permission")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewSyncGroupPermissions(tt.sonarApiClient(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.group)

			tt.wantErr(t, err)
		})
	}
}
