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

func TestCreateQualityProfile_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		profile        *sonarApi.SonarQualityProfile
		sonarApiClient func(t *testing.T) sonar.QualityProfileClient
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "quality profile doesn't exist, creating new one",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name:     "test-profile",
					Language: "go",
					Default:  true,
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityProfileClient {
				m := mocks.NewMockClientInterface(t)

				m.On("GetQualityProfile", mock.Anything, "test-profile").
					Return(nil, sonar.NewHTTPError(http.StatusNotFound, "profile doesn't exist"))
				m.On("CreateQualityProfile", mock.Anything, "test-profile", "go").
					Return(&sonar.QualityProfile{
						Name:      "test-profile",
						Language:  "go",
						IsDefault: false,
					}, nil)
				m.On("SetAsDefaultQualityProfile", mock.Anything, "test-profile", "go").
					Return(nil)

				return m
			},

			wantErr: require.NoError,
		},
		{
			name: "quality profile exists, updating default quality profile",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name:     "test-profile",
					Language: "go",
					Default:  true,
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityProfileClient {
				m := mocks.NewMockClientInterface(t)

				m.On("GetQualityProfile", mock.Anything, "test-profile").
					Return(&sonar.QualityProfile{
						Name:      "test-profile",
						Language:  "go",
						IsDefault: false,
					}, nil)
				m.On("SetAsDefaultQualityProfile", mock.Anything, "test-profile", "go").
					Return(nil)

				return m
			},

			wantErr: require.NoError,
		},
		{
			name: "failed to set default quality profile",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name:     "test-profile",
					Language: "go",
					Default:  true,
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityProfileClient {
				m := mocks.NewMockClientInterface(t)

				m.On("GetQualityProfile", mock.Anything, "test-profile").
					Return(&sonar.QualityProfile{
						Name:      "test-profile",
						Language:  "go",
						IsDefault: false,
					}, nil)
				m.On("SetAsDefaultQualityProfile", mock.Anything, "test-profile", "go").
					Return(errors.New("failed to set default quality profile"))

				return m
			},

			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to set default quality profile")
			},
		},
		{
			name: "failed to create quality profile",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name:     "test-profile",
					Language: "go",
					Default:  true,
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityProfileClient {
				m := mocks.NewMockClientInterface(t)

				m.On("GetQualityProfile", mock.Anything, "test-profile").
					Return(nil, sonar.NewHTTPError(http.StatusNotFound, "profile doesn't exist"))
				m.On("CreateQualityProfile", mock.Anything, "test-profile", "go").
					Return(nil, errors.New("failed to create quality profile"))

				return m
			},

			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to create quality profile")
			},
		},
		{
			name: "failed to get quality profile",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name:     "test-profile",
					Language: "go",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityProfileClient {
				m := mocks.NewMockClientInterface(t)

				m.On("GetQualityProfile", mock.Anything, "test-profile").
					Return(nil, sonar.NewHTTPError(http.StatusBadRequest, "failed to get quality profile"))

				return m
			},

			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get quality profile")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewCreateQualityProfile(tt.sonarApiClient(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.profile)

			tt.wantErr(t, err)
		})
	}
}
