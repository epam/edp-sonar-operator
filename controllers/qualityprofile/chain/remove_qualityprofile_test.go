package chain

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar/mocks"
)

func TestRemoveQualityProfile_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		profile        *sonarApi.SonarQualityProfile
		sonarApiClient func(t *testing.T) sonar.QualityProfileClient
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "remove quality profile successfully",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name:     "test-profile",
					Language: "go",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityProfileClient {
				m := mocks.NewClientInterface(t)

				m.On("DeleteQualityProfile", mock.Anything, "test-profile", "go").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "quality profile doesn't exist, ignoring error",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name:     "test-profile",
					Language: "go",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityProfileClient {
				m := mocks.NewClientInterface(t)

				m.On("DeleteQualityProfile", mock.Anything, "test-profile", "go").
					Return(sonar.NewHTTPError(http.StatusNotFound, "profile doesn't exist"))

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to remove quality profile",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name:     "test-profile",
					Language: "go",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityProfileClient {
				m := mocks.NewClientInterface(t)

				m.On("DeleteQualityProfile", mock.Anything, "test-profile", "go").
					Return(errors.New("failed to remove quality profile"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to remove quality profile")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := NewRemoveQualityProfile(tt.sonarApiClient(t))
			err := r.ServeRequest(context.Background(), tt.profile)

			tt.wantErr(t, err)
		})
	}
}
