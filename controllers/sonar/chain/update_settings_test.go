package chain

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/epam/edp-sonar-operator/api/common"
	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar/mocks"
)

func TestUpdateSettings_ServeRequest(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))

	tests := []struct {
		name           string
		sonarApiClient func(t *testing.T) sonar.Settings
		sonar          *sonarApi.Sonar
		k8sClient      func(t *testing.T) client.Client
		wantErr        require.ErrorAssertionFunc
		wantStatus     sonarApi.SonarStatus
	}{
		{
			name: "settings is set",
			sonar: &sonarApi.Sonar{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sonar",
					Namespace: "default",
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
						{
							Key: "sonar.secret",
							ValueRef: &common.SourceRef{
								SecretKeyRef: &common.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "sonar-secret",
									},
									Key: "secret-key",
								},
							},
						},
					},
				},
			},
			sonarApiClient: func(t *testing.T) sonar.Settings {
				m := mocks.NewClientInterface(t)
				m.On("SetSetting", mock.Anything, mock.Anything).
					Return(nil).Times(4)

				return m
			},
			k8sClient: func(t *testing.T) client.Client {
				return fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "sonar-secret",
							Namespace: "default",
						},
						Data: map[string][]byte{"secret-key": []byte("secret-value")},
					}).
					Build()
			},
			wantErr: require.NoError,
			wantStatus: sonarApi.SonarStatus{
				ProcessedSettings: "sonar.core.a,sonar.core.b,sonar.core.c,sonar.secret",
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
			k8sClient: func(t *testing.T) client.Client {
				return fake.NewClientBuilder().WithScheme(scheme).Build()
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
			k8sClient: func(t *testing.T) client.Client {
				return fake.NewClientBuilder().WithScheme(scheme).Build()
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
			k8sClient: func(t *testing.T) client.Client {
				return fake.NewClientBuilder().WithScheme(scheme).Build()
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

			h := NewUpdateSettings(tt.sonarApiClient(t), tt.k8sClient(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.sonar)

			tt.wantErr(t, err)
			assert.Equal(t, tt.wantStatus, tt.sonar.Status)
		})
	}
}
