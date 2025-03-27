package chain

import (
	"context"
	"errors"
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

func TestCreatePermissionTemplate_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		template       *sonarApi.SonarPermissionTemplate
		sonarApiClient func(t *testing.T) sonar.PermissionTemplateInterface
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "template doesn't exist, creating new one",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name:              "test-template",
					Description:       "test-description",
					ProjectKeyPattern: ".*.finance",
					Default:           true,
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewMockClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(nil, sonar.NewHTTPError(http.StatusNotFound, "template doesn't exist"))
				m.On("CreatePermissionTemplate", mock.Anything, &sonar.PermissionTemplateData{
					Name:              "test-template",
					Description:       "test-description",
					ProjectKeyPattern: ".*.finance",
				}).
					Return(&sonar.PermissionTemplate{
						ID:        "id1",
						IsDefault: false,
						PermissionTemplateData: sonar.PermissionTemplateData{
							Name:              "test-template",
							Description:       "test-description",
							ProjectKeyPattern: ".*.finance",
						},
					}, nil)
				m.On("SetDefaultPermissionTemplate", mock.Anything, "test-template").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "template exists, updating it",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name:              "test-template",
					Description:       "test-description-new",
					ProjectKeyPattern: ".*.finance2",
					Default:           true,
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewMockClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(&sonar.PermissionTemplate{
						ID:        "id1",
						IsDefault: false,
						PermissionTemplateData: sonar.PermissionTemplateData{
							Name:              "test-template",
							Description:       "test-description",
							ProjectKeyPattern: ".*.finance",
						},
					}, nil)
				m.On("UpdatePermissionTemplate", mock.Anything, &sonar.PermissionTemplate{
					ID: "id1",
					PermissionTemplateData: sonar.PermissionTemplateData{
						Name:              "test-template",
						Description:       "test-description-new",
						ProjectKeyPattern: ".*.finance2",
					},
				}).
					Return(nil)
				m.On("SetDefaultPermissionTemplate", mock.Anything, "test-template").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to update template",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name:              "test-template",
					Description:       "test-description-new",
					ProjectKeyPattern: ".*.finance2",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewMockClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(&sonar.PermissionTemplate{
						ID:        "id1",
						IsDefault: false,
						PermissionTemplateData: sonar.PermissionTemplateData{
							Name:              "test-template",
							Description:       "test-description",
							ProjectKeyPattern: ".*.finance",
						},
					}, nil)
				m.On("UpdatePermissionTemplate", mock.Anything, &sonar.PermissionTemplate{
					ID: "id1",
					PermissionTemplateData: sonar.PermissionTemplateData{
						Name:              "test-template",
						Description:       "test-description-new",
						ProjectKeyPattern: ".*.finance2",
					},
				}).
					Return(errors.New("failed to update template"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to update template")
			},
		},
		{
			name: "failed to set default template",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name:              "test-template",
					Description:       "test-description",
					ProjectKeyPattern: ".*.finance",
					Default:           true,
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewMockClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(&sonar.PermissionTemplate{
						ID:        "id1",
						IsDefault: false,
						PermissionTemplateData: sonar.PermissionTemplateData{
							Name:              "test-template",
							Description:       "test-description",
							ProjectKeyPattern: ".*.finance",
						},
					}, nil)
				m.On("SetDefaultPermissionTemplate", mock.Anything, "test-template").
					Return(errors.New("failed to set default template"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to set default template")
			},
		},
		{
			name: "failed to create template",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name:              "test-template",
					Description:       "test-description",
					ProjectKeyPattern: ".*.finance",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewMockClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(nil, sonar.NewHTTPError(http.StatusNotFound, "template doesn't exist"))
				m.On("CreatePermissionTemplate", mock.Anything, &sonar.PermissionTemplateData{
					Name:              "test-template",
					Description:       "test-description",
					ProjectKeyPattern: ".*.finance",
				}).
					Return(nil, errors.New("failed to create template"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to create template")
			},
		},
		{
			name: "failed to get template",
			template: &sonarApi.SonarPermissionTemplate{
				Spec: sonarApi.SonarPermissionTemplateSpec{
					Name:              "test-template",
					Description:       "test-description",
					ProjectKeyPattern: ".*.finance",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewMockClientInterface(t)

				m.On("GetPermissionTemplate", mock.Anything, "test-template").
					Return(nil, sonar.NewHTTPError(http.StatusBadRequest, "failed to get template"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get template")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewCreatePermissionTemplate(tt.sonarApiClient(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.template)

			tt.wantErr(t, err)
		})
	}
}
