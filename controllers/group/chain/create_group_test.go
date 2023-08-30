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

func TestCreateGroup_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		group          *sonarApi.SonarGroup
		sonarApiClient func(t *testing.T) sonar.GroupInterface
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "group doesn't exist, creating new one",
			group: &sonarApi.SonarGroup{
				Spec: sonarApi.SonarGroupSpec{
					Name:        "test-group",
					Description: "test-description",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.GroupInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetGroup", mock.Anything, "test-group").
					Return(nil, sonar.NewHTTPError(http.StatusNotFound, "group doesn't exist"))
				m.On("CreateGroup", mock.Anything, &sonar.Group{
					Name:        "test-group",
					Description: "test-description",
				}).
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "group exists, updating it",
			group: &sonarApi.SonarGroup{
				Spec: sonarApi.SonarGroupSpec{
					Name:        "test-group",
					Description: "test-description-new",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.GroupInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetGroup", mock.Anything, "test-group").
					Return(&sonar.Group{
						Name:        "test-group",
						Description: "test-description",
					}, nil)
				m.On("UpdateGroup", mock.Anything, "test-group", &sonar.Group{
					Name:        "test-group",
					Description: "test-description-new",
				}).
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to  update group",
			group: &sonarApi.SonarGroup{
				Spec: sonarApi.SonarGroupSpec{
					Name:        "test-group",
					Description: "test-description-new",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.GroupInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetGroup", mock.Anything, "test-group").
					Return(&sonar.Group{
						Name:        "test-group",
						Description: "test-description",
					}, nil)
				m.On("UpdateGroup", mock.Anything, "test-group", &sonar.Group{
					Name:        "test-group",
					Description: "test-description-new",
				}).
					Return(errors.New("failed to update group"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to update group")
			},
		},
		{
			name: "failed to create group",
			group: &sonarApi.SonarGroup{
				Spec: sonarApi.SonarGroupSpec{
					Name:        "test-group",
					Description: "test-description",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.GroupInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetGroup", mock.Anything, "test-group").
					Return(nil, sonar.NewHTTPError(http.StatusNotFound, "group doesn't exist"))
				m.On("CreateGroup", mock.Anything, &sonar.Group{
					Name:        "test-group",
					Description: "test-description",
				}).
					Return(errors.New("failed to create group"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to create group")
			},
		},
		{
			name: "failed to get group",
			group: &sonarApi.SonarGroup{
				Spec: sonarApi.SonarGroupSpec{
					Name:        "test-group",
					Description: "test-description",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.GroupInterface {
				m := mocks.NewClientInterface(t)

				m.On("GetGroup", mock.Anything, "test-group").
					Return(nil, sonar.NewHTTPError(http.StatusBadRequest, "failed to get group"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get group")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewCreateGroup(tt.sonarApiClient(t))
			err := c.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.group)

			tt.wantErr(t, err)
		})
	}
}
