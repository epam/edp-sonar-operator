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

func TestSyncQualityProfileRules_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		profile        *sonarApi.SonarQualityProfile
		sonarApiClient func(t *testing.T) sonarApiClient
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "rules are synced",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name: "test-profile",
					Rules: map[string]sonarApi.Rule{
						"rule1": {
							Severity: "MAJOR",
							Params:   "key1=v1;key2=v2",
						},
						"rule2": {
							Severity: "CRITICAL",
							Params:   "key3=v3;key4=v4",
						},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonarApiClient {
				m := mocks.NewClientInterface(t)

				m.On("GetQualityProfile", mock.Anything, "test-profile").
					Return(&sonar.QualityProfile{
						Key: "test-profile-key",
					}, nil)
				m.On("GetQualityProfileActiveRules", mock.Anything, "test-profile-key").
					Return([]sonar.Rule{
						{
							Key:      "rule1",
							Severity: "MAJOR",
							Params:   "key1=v1;key2=v2",
						},
						{
							Key:      "rule3",
							Severity: "MAJOR",
							Params:   "key5=v5;key6=v6",
						},
					}, nil)
				m.On("ActivateQualityProfileRule", mock.Anything, "test-profile-key", sonar.Rule{
					Rule:     "rule2",
					Severity: "CRITICAL",
					Params:   "key3=v3;key4=v4",
				}).
					Return(nil)
				m.On("DeactivateQualityProfileRule", mock.Anything, "test-profile-key", "rule3").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to deactivate rule",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name:  "test-profile",
					Rules: map[string]sonarApi.Rule{},
				},
			},
			sonarApiClient: func(t *testing.T) sonarApiClient {
				m := mocks.NewClientInterface(t)

				m.On("GetQualityProfile", mock.Anything, "test-profile").
					Return(&sonar.QualityProfile{
						Key: "test-profile-key",
					}, nil)
				m.On("GetQualityProfileActiveRules", mock.Anything, "test-profile-key").
					Return([]sonar.Rule{
						{
							Key:      "rule1",
							Severity: "MAJOR",
							Params:   "key1=v1;key2=v2",
						},
					}, nil)
				m.On("DeactivateQualityProfileRule", mock.Anything, "test-profile-key", "rule1").
					Return(errors.New("deactivate rule error"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "deactivate rule error")
			},
		},
		{
			name: "failed to activate rule",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name: "test-profile",
					Rules: map[string]sonarApi.Rule{
						"rule1": {
							Severity: "MAJOR",
							Params:   "key1=v1;key2=v2",
						},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonarApiClient {
				m := mocks.NewClientInterface(t)

				m.On("GetQualityProfile", mock.Anything, "test-profile").
					Return(&sonar.QualityProfile{
						Key: "test-profile-key",
					}, nil)
				m.On("GetQualityProfileActiveRules", mock.Anything, "test-profile-key").
					Return([]sonar.Rule{}, nil)
				m.On("ActivateQualityProfileRule", mock.Anything, "test-profile-key", sonar.Rule{
					Rule:     "rule1",
					Severity: "MAJOR",
					Params:   "key1=v1;key2=v2",
				}).
					Return(errors.New("activate rule error"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "activate rule error")
			},
		},
		{
			name: "failed to get active rules",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name: "test-profile",
				},
			},
			sonarApiClient: func(t *testing.T) sonarApiClient {
				m := mocks.NewClientInterface(t)

				m.On("GetQualityProfile", mock.Anything, "test-profile").
					Return(&sonar.QualityProfile{
						Key: "test-profile-key",
					}, nil)
				m.On("GetQualityProfileActiveRules", mock.Anything, "test-profile-key").
					Return(nil, errors.New("get active rules error"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "get active rules error")
			},
		},
		{
			name: "failed to get quality profile",
			profile: &sonarApi.SonarQualityProfile{
				Spec: sonarApi.SonarQualityProfileSpec{
					Name: "test-profile",
				},
			},
			sonarApiClient: func(t *testing.T) sonarApiClient {
				m := mocks.NewClientInterface(t)

				m.On("GetQualityProfile", mock.Anything, "test-profile").
					Return(nil, errors.New("get quality profile error"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "get quality profile error")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewSyncQualityProfileRules(tt.sonarApiClient(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.profile)
			tt.wantErr(t, err)
		})
	}
}
