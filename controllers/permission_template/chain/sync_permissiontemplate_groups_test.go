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

func TestSyncPermissionTemplateGroups_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		template       *sonarApi.SonarPermissionTemplate
		sonarApiClient func(t *testing.T) sonar.PermissionTemplateInterface
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "syncing permission template groups in sonar successfully",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name: "test-template",
					GroupsPermissions: map[string][]string{
						"test-group":   {"test-permission", "test-permission-2"},
						"test-group-2": {"test-permission-3"},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(&sonar.PermissionTemplate{
						ID: "id1",
					}, nil)
				m.On("GetPermissionTemplateGroups", mock.Anything, "id1").
					Return(map[string][]string{
						"test-group":   {"test-permission", "test-permission-3"},
						"test-group-3": {"test-permission-4"},
					}, nil)
				m.On("AddGroupToPermissionTemplate", mock.Anything, "id1", "test-group", "test-permission-2").
					Return(nil)
				m.On("RemoveGroupFromPermissionTemplate", mock.Anything, "id1", "test-group", "test-permission-3").
					Return(nil)
				m.On("AddGroupToPermissionTemplate", mock.Anything, "id1", "test-group-2", "test-permission-3").
					Return(nil)
				m.On("RemoveGroupFromPermissionTemplate", mock.Anything, "id1", "test-group-3", "test-permission-4").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to remove permission template group - existing group",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name: "test-template",
					GroupsPermissions: map[string][]string{
						"test-group": {"test-permission"},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(&sonar.PermissionTemplate{
						ID: "id1",
					}, nil)
				m.On("GetPermissionTemplateGroups", mock.Anything, "id1").
					Return(map[string][]string{
						"test-group":   {"test-permission"},
						"test-group-3": {"test-permission-4"},
					}, nil)
				m.On("RemoveGroupFromPermissionTemplate", mock.Anything, "id1", "test-group-3", "test-permission-4").
					Return(errors.New("failed to remove permission template group"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to remove permission template group")
			},
		},
		{
			name: "failed to remove permission template group - existing permission in group",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name: "test-template",
					GroupsPermissions: map[string][]string{
						"test-group": {"test-permission"},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(&sonar.PermissionTemplate{
						ID: "id1",
					}, nil)
				m.On("GetPermissionTemplateGroups", mock.Anything, "id1").
					Return(map[string][]string{
						"test-group": {"test-permission", "test-permission-2"},
					}, nil)
				m.On("RemoveGroupFromPermissionTemplate", mock.Anything, "id1", "test-group", "test-permission-2").
					Return(errors.New("failed to remove permission template group"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to remove permission template group")
			},
		},
		{
			name: "failed to add permission template group - existing group",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name: "test-template",
					GroupsPermissions: map[string][]string{
						"test-group": {"test-permission", "test-permission-2"},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(&sonar.PermissionTemplate{
						ID: "id1",
					}, nil)
				m.On("GetPermissionTemplateGroups", mock.Anything, "id1").
					Return(map[string][]string{"test-group": {"test-permission"}}, nil)
				m.On("AddGroupToPermissionTemplate", mock.Anything, "id1", "test-group", "test-permission-2").
					Return(errors.New("failed to add permission template group"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to add permission template group")
			},
		},
		{
			name: "failed to add permission template group - new group",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name: "test-template",
					GroupsPermissions: map[string][]string{
						"test-group": {"test-permission"},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(&sonar.PermissionTemplate{
						ID: "id1",
					}, nil)
				m.On("GetPermissionTemplateGroups", mock.Anything, "id1").
					Return(map[string][]string{}, nil)
				m.On("AddGroupToPermissionTemplate", mock.Anything, "id1", "test-group", "test-permission").
					Return(errors.New("failed to add permission template group"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to add permission template group")
			},
		},
		{
			name: "failed to get permission template groups",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name: "test-template",
					GroupsPermissions: map[string][]string{
						"test-group": {"test-permission"},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(&sonar.PermissionTemplate{
						ID: "id1",
					}, nil)
				m.On("GetPermissionTemplateGroups", mock.Anything, "id1").
					Return(nil, errors.New("failed to get permission template groups"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get permission template groups")
			},
		},
		{
			name: "failed to get permission template",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name: "test-template",
					GroupsPermissions: map[string][]string{
						"test-group": {"test-permission"},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(nil, errors.New("failed to get permission template"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get permission template")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewSyncPermissionTemplateGroups(tt.sonarApiClient(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.template)

			tt.wantErr(t, err)
		})
	}
}
