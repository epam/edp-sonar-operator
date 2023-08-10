package chain

import (
	"context"
	"errors"
	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar/mocks"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

func TestUpdateSettings_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		sonarApiClient func(t *testing.T) sonar.Settings
		sonar          *sonarApi.Sonar
		wantErr        require.ErrorAssertionFunc
		wantStatus     sonarApi.SonarStatus
	}{
		{
			name: "settings is set",
			sonar: &sonarApi.Sonar{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sonar",
				},
				Spec: sonarApi.SonarSpec{
					Settings: []sonarApi.SonarSetting{
						{
							Key:   "sonar.core.a",
							Value: "sonar",
						},
						{
							Key:    "sonar.core.b",
							Values: []string{"value1", "value2"},
						},
						{
							Key: "sonar.core.c",
							FieldValues: map[string]string{
								"field1": "value1",
								"field2": "value2",
							},
						},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.Settings {
				m := mocks.NewClientInterface(t)
				m.On("SetSetting", mock.Anything, mock.Anything).
					Return(nil).Times(3)

				return m
			},
			wantErr: require.NoError,
			wantStatus: sonarApi.SonarStatus{
				ProcessedSettings: "sonar.core.a,sonar.core.b,sonar.core.c",
			},
		},
		{
			name: "unset previous settings",
			sonar: &sonarApi.Sonar{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sonar",
				},
				Spec: sonarApi.SonarSpec{
					Settings: []sonarApi.SonarSetting{
						{
							Key:   "sonar.core.a",
							Value: "sonar",
						},
					},
				},
				Status: sonarApi.SonarStatus{
					ProcessedSettings: "sonar.core.a,sonar.core.b",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.Settings {
				m := mocks.NewClientInterface(t)
				m.On("SetSetting", mock.Anything, mock.Anything).
					Return(nil)
				m.On("ResetSettings", mock.Anything, []string{"sonar.core.b"}).
					Return(nil)

				return m
			},
			wantErr: require.NoError,
			wantStatus: sonarApi.SonarStatus{
				ProcessedSettings: "sonar.core.a",
			},
		},
		{
			name: "failed to unset previous settings",
			sonar: &sonarApi.Sonar{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sonar",
				},
				Spec: sonarApi.SonarSpec{
					Settings: []sonarApi.SonarSetting{
						{
							Key:   "sonar.core.a",
							Value: "sonar",
						},
					},
				},
				Status: sonarApi.SonarStatus{
					ProcessedSettings: "sonar.core.a,sonar.core.b",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.Settings {
				m := mocks.NewClientInterface(t)
				m.On("SetSetting", mock.Anything, mock.Anything).
					Return(nil)
				m.On("ResetSettings", mock.Anything, []string{"sonar.core.b"}).
					Return(errors.New("failed to unset settings"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to unset settings")
			},
			wantStatus: sonarApi.SonarStatus{
				ProcessedSettings: "sonar.core.a,sonar.core.b",
			},
		},
		{
			name: "failed to set setting",
			sonar: &sonarApi.Sonar{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sonar",
				},
				Spec: sonarApi.SonarSpec{
					Settings: []sonarApi.SonarSetting{
						{
							Key:   "sonar.core.a",
							Value: "sonar",
						},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.Settings {
				m := mocks.NewClientInterface(t)
				m.On("SetSetting", mock.Anything, mock.Anything).
					Return(errors.New("failed to set setting"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to set setting")
			},
			wantStatus: sonarApi.SonarStatus{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewUpdateSettings(tt.sonarApiClient(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.sonar)
			tt.wantErr(t, err)
			assert.Equal(t, tt.wantStatus, tt.sonar.Status)
		})
	}
}
