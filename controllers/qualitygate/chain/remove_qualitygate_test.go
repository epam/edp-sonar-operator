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

func TestRemoveQualityGate_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		gate           *sonarApi.SonarQualityGate
		sonarApiClient func(t *testing.T) sonar.QualityGateClient
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "quality gate removed successfully",
			gate: &sonarApi.SonarQualityGate{
				Spec: sonarApi.SonarQualityGateSpec{
					Name: "test-gate",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityGateClient {
				m := mocks.NewClientInterface(t)

				m.On("DeleteQualityGate", mock.Anything, "test-gate").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "quality not found, ignoring",
			gate: &sonarApi.SonarQualityGate{
				Spec: sonarApi.SonarQualityGateSpec{
					Name: "test-gate",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityGateClient {
				m := mocks.NewClientInterface(t)

				m.On("DeleteQualityGate", mock.Anything, "test-gate").
					Return(sonar.NewHTTPError(http.StatusNotFound, "not found"))

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to remove quality gate",
			gate: &sonarApi.SonarQualityGate{
				Spec: sonarApi.SonarQualityGateSpec{
					Name: "test-gate",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityGateClient {
				m := mocks.NewClientInterface(t)

				m.On("DeleteQualityGate", mock.Anything, "test-gate").
					Return(sonar.NewHTTPError(http.StatusBadRequest, "bad request"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "bad request")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := NewRemoveQualityGate(tt.sonarApiClient(t))
			err := r.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.gate)

			tt.wantErr(t, err)
		})
	}
}
