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

func TestCreateQualityGate_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		gate           *sonarApi.SonarQualityGate
		sonarApiClient func(t *testing.T) sonar.QualityGateClient
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "quality gate doesn't exist, creating new one",
			gate: &sonarApi.SonarQualityGate{
				Spec: sonarApi.SonarQualityGateSpec{
					Name: "test-gate",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityGateClient {
				m := mocks.NewClientInterface(t)

				m.On("GetQualityGate", mock.Anything, "test-gate").
					Return(nil, sonar.NewHTTPError(http.StatusNotFound, "not found"))

				m.On("CreateQualityGate", mock.Anything, "test-gate").
					Return(&sonar.QualityGate{
						Name: "test-gate",
						ID:   "1",
					}, nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "updating default quality gate",
			gate: &sonarApi.SonarQualityGate{
				Spec: sonarApi.SonarQualityGateSpec{
					Name:    "test-gate",
					Default: true,
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityGateClient {
				m := mocks.NewClientInterface(t)

				m.On("GetQualityGate", mock.Anything, "test-gate").
					Return(&sonar.QualityGate{
						Name:      "test-gate",
						ID:        "1",
						IsDefault: false,
					}, nil)

				m.On("SetAsDefaultQualityGate", mock.Anything, "test-gate").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to set default quality gate",
			gate: &sonarApi.SonarQualityGate{
				Spec: sonarApi.SonarQualityGateSpec{
					Name:    "test-gate",
					Default: true,
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityGateClient {
				m := mocks.NewClientInterface(t)

				m.On("GetQualityGate", mock.Anything, "test-gate").
					Return(&sonar.QualityGate{
						Name:      "test-gate",
						ID:        "1",
						IsDefault: false,
					}, nil)

				m.On("SetAsDefaultQualityGate", mock.Anything, "test-gate").
					Return(errors.New("failed to set default quality gate"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to set default quality gate")
			},
		},
		{
			name: "failed to get create quality gate",
			gate: &sonarApi.SonarQualityGate{
				Spec: sonarApi.SonarQualityGateSpec{
					Name: "test-gate",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityGateClient {
				m := mocks.NewClientInterface(t)

				m.On("GetQualityGate", mock.Anything, "test-gate").
					Return(nil, sonar.NewHTTPError(http.StatusNotFound, "not found"))

				m.On("CreateQualityGate", mock.Anything, "test-gate").
					Return(nil, errors.New("failed to create quality gate"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to create quality gate")
			},
		},
		{
			name: "failed to get quality gate",
			gate: &sonarApi.SonarQualityGate{
				Spec: sonarApi.SonarQualityGateSpec{
					Name: "test-gate",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityGateClient {
				m := mocks.NewClientInterface(t)

				m.On("GetQualityGate", mock.Anything, "test-gate").
					Return(nil, errors.New("failed to get quality gate"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get quality gate")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewCreateQualityGate(tt.sonarApiClient(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.gate)

			tt.wantErr(t, err)
		})
	}
}
