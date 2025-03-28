package chain

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar/mocks"
)

func TestRemovePermissionTemplate_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		template       *sonarApi.SonarPermissionTemplate
		sonarApiClient func(t *testing.T) sonar.PermissionTemplateInterface
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "remove permission template successfully",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name: "test-template",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewMockClientInterface(t)

				m.On("DeletePermissionTemplate", mock.Anything, "test-template").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "permission template doesn't exist",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name: "test-template",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewMockClientInterface(t)

				m.On("DeletePermissionTemplate", mock.Anything, "test-template").
					Return(sonar.NewHTTPError(http.StatusNotFound, "permission template doesn't exist"))

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to delete permission template",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name: "test-template",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewMockClientInterface(t)

				m.On("DeletePermissionTemplate", mock.Anything, "test-template").
					Return(sonar.NewHTTPError(http.StatusInternalServerError, "failed to delete permission template"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to delete permission template")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewRemovePermissionTemplate(tt.sonarApiClient(t))
			err := c.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.template)

			tt.wantErr(t, err)
		})
	}
}
