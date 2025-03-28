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

func TestSyncQualityGateConditions_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		gate           *sonarApi.SonarQualityGate
		sonarApiClient func(t *testing.T) sonar.QualityGateClient
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "quality gate conditions synced successfully",
			gate: &sonarApi.SonarQualityGate{
				Spec: sonarApi.SonarQualityGateSpec{
					Name: "test-gate",
					Conditions: map[string]sonarApi.Condition{
						"test-metric": {
							Error: "2",
							Op:    "GT",
						},
						"test-metric2": {
							Error: "90",
							Op:    "GT",
						},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityGateClient {
				m := mocks.NewMockClientInterface(t)

				m.On("GetQualityGate", mock.Anything, "test-gate").
					Return(&sonar.QualityGate{
						Conditions: []sonar.QualityGateCondition{
							{
								Error:  "1",
								Metric: "test-metric",
								OP:     "GT",
								ID:     "111",
							},
							{
								Error:  "80",
								Metric: "test-metric3",
								OP:     "GT",
								ID:     "113",
							},
						},
					}, nil)

				m.On("CreateQualityGateCondition", mock.Anything, "test-gate", sonar.QualityGateCondition{
					Error:  "90",
					Metric: "test-metric2",
					OP:     "GT",
				}).
					Return(nil)

				m.On("UpdateQualityGateCondition", mock.Anything, sonar.QualityGateCondition{
					ID:     "111",
					Error:  "2",
					Metric: "test-metric",
					OP:     "GT",
				}).
					Return(nil)

				m.On("DeleteQualityGateCondition", mock.Anything, "113").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to delete quality gate condition",
			gate: &sonarApi.SonarQualityGate{
				Spec: sonarApi.SonarQualityGateSpec{
					Name: "test-gate",
					Conditions: map[string]sonarApi.Condition{
						"test-metric": {
							Error: "2",
							Op:    "GT",
						},
						"test-metric2": {
							Error: "90",
							Op:    "GT",
						},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityGateClient {
				m := mocks.NewMockClientInterface(t)

				m.On("GetQualityGate", mock.Anything, "test-gate").
					Return(&sonar.QualityGate{
						Conditions: []sonar.QualityGateCondition{
							{
								Error:  "1",
								Metric: "test-metric",
								OP:     "GT",
								ID:     "111",
							},
							{
								Error:  "80",
								Metric: "test-metric3",
								OP:     "GT",
								ID:     "113",
							},
						},
					}, nil)

				m.On("CreateQualityGateCondition", mock.Anything, "test-gate", sonar.QualityGateCondition{
					Error:  "90",
					Metric: "test-metric2",
					OP:     "GT",
				}).
					Return(nil)

				m.On("UpdateQualityGateCondition", mock.Anything, sonar.QualityGateCondition{
					ID:     "111",
					Error:  "2",
					Metric: "test-metric",
					OP:     "GT",
				}).
					Return(nil)

				m.On("DeleteQualityGateCondition", mock.Anything, "113").
					Return(errors.New("failed to delete quality gate condition"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to delete quality gate condition")
			},
		},
		{
			name: "failed to update quality gate condition",
			gate: &sonarApi.SonarQualityGate{
				Spec: sonarApi.SonarQualityGateSpec{
					Name: "test-gate",
					Conditions: map[string]sonarApi.Condition{
						"test-metric": {
							Error: "2",
							Op:    "GT",
						},
						"test-metric2": {
							Error: "90",
							Op:    "GT",
						},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityGateClient {
				m := mocks.NewMockClientInterface(t)

				m.On("GetQualityGate", mock.Anything, "test-gate").
					Return(&sonar.QualityGate{
						Conditions: []sonar.QualityGateCondition{
							{
								Error:  "1",
								Metric: "test-metric",
								OP:     "GT",
								ID:     "111",
							},
						},
					}, nil)

				m.On("CreateQualityGateCondition", mock.Anything, "test-gate", sonar.QualityGateCondition{
					Error:  "90",
					Metric: "test-metric2",
					OP:     "GT",
				}).
					Return(nil).
					Maybe()

				m.On("UpdateQualityGateCondition", mock.Anything, sonar.QualityGateCondition{
					ID:     "111",
					Error:  "2",
					Metric: "test-metric",
					OP:     "GT",
				}).
					Return(errors.New("failed to update quality gate condition"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to update quality gate condition")
			},
		},
		{
			name: "failed to create quality gate condition",
			gate: &sonarApi.SonarQualityGate{
				Spec: sonarApi.SonarQualityGateSpec{
					Name: "test-gate",
					Conditions: map[string]sonarApi.Condition{
						"test-metric": {
							Error: "2",
							Op:    "GT",
						},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.QualityGateClient {
				m := mocks.NewMockClientInterface(t)

				m.On("GetQualityGate", mock.Anything, "test-gate").
					Return(&sonar.QualityGate{}, nil)

				m.On("CreateQualityGateCondition", mock.Anything, "test-gate", sonar.QualityGateCondition{
					Error:  "2",
					Metric: "test-metric",
					OP:     "GT",
				}).
					Return(errors.New("failed to create quality gate condition"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to create quality gate condition")
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
				m := mocks.NewMockClientInterface(t)

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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewSyncQualityGateConditions(tt.sonarApiClient(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.gate)

			tt.wantErr(t, err)
		})
	}
}
