package chain

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar/mocks"
)

func TestCreateUser_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		user           *sonarApi.SonarUser
		client         func(t *testing.T) client.Client
		sonarApiClient func(t *testing.T) sonar.UserInterface
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "user not found, create new one",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:  "test-user",
					Secret: "test-secret",
				},
			},
			client: func(t *testing.T) client.Client {
				s := runtime.NewScheme()

				require.NoError(t, sonarApi.AddToScheme(s))
				require.NoError(t, corev1.AddToScheme(s))

				secret := corev1.Secret{
					ObjectMeta: ctrl.ObjectMeta{
						Name: "test-secret",
					},
					Data: map[string][]byte{
						"password": []byte("test-password"),
					},
				}

				return fake.NewClientBuilder().WithScheme(s).WithObjects(&secret).Build()
			},
			sonarApiClient: func(t *testing.T) sonar.UserInterface {
				m := mocks.NewMockClientInterface(t)
				m.On("GetUserByLogin", mock.Anything, "test-user").
					Return(nil, sonar.NewHTTPError(http.StatusNotFound, ""))
				m.On("CreateUser", mock.Anything, mock.Anything).
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "user already exists with the same data",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:  "test-user",
					Secret: "test-secret",
				},
			},
			client: func(t *testing.T) client.Client {
				s := runtime.NewScheme()

				require.NoError(t, sonarApi.AddToScheme(s))
				require.NoError(t, corev1.AddToScheme(s))

				secret := corev1.Secret{
					ObjectMeta: ctrl.ObjectMeta{
						Name: "test-secret",
					},
					Data: map[string][]byte{
						"password": []byte("test-password"),
					},
				}

				return fake.NewClientBuilder().WithScheme(s).WithObjects(&secret).Build()
			},
			sonarApiClient: func(t *testing.T) sonar.UserInterface {
				m := mocks.NewMockClientInterface(t)
				m.On("GetUserByLogin", mock.Anything, "test-user").
					Return(&sonar.User{
						Login: "test-user",
					}, nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "user already exists, update it",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:  "test-user",
					Secret: "test-secret",
					Name:   "test-name",
				},
			},
			client: func(t *testing.T) client.Client {
				s := runtime.NewScheme()

				require.NoError(t, sonarApi.AddToScheme(s))
				require.NoError(t, corev1.AddToScheme(s))

				secret := corev1.Secret{
					ObjectMeta: ctrl.ObjectMeta{
						Name: "test-secret",
					},
					Data: map[string][]byte{
						"password": []byte("test-password"),
					},
				}

				return fake.NewClientBuilder().WithScheme(s).WithObjects(&secret).Build()
			},
			sonarApiClient: func(t *testing.T) sonar.UserInterface {
				m := mocks.NewMockClientInterface(t)
				m.On("GetUserByLogin", mock.Anything, "test-user").
					Return(&sonar.User{
						Login: "test-user",
					}, nil)
				m.On("UpdateUser", mock.Anything, mock.Anything).
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to update user",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:  "test-user",
					Secret: "test-secret",
					Name:   "test-name",
				},
			},
			client: func(t *testing.T) client.Client {
				s := runtime.NewScheme()

				require.NoError(t, sonarApi.AddToScheme(s))
				require.NoError(t, corev1.AddToScheme(s))

				secret := corev1.Secret{
					ObjectMeta: ctrl.ObjectMeta{
						Name: "test-secret",
					},
					Data: map[string][]byte{
						"password": []byte("test-password"),
					},
				}

				return fake.NewClientBuilder().WithScheme(s).WithObjects(&secret).Build()
			},
			sonarApiClient: func(t *testing.T) sonar.UserInterface {
				m := mocks.NewMockClientInterface(t)
				m.On("GetUserByLogin", mock.Anything, "test-user").
					Return(&sonar.User{
						Login: "test-user",
					}, nil)
				m.On("UpdateUser", mock.Anything, mock.Anything).
					Return(errors.New("failed to update user"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to update user")
			},
		},
		{
			name: "failed to create user",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:  "test-user",
					Secret: "test-secret",
				},
			},
			client: func(t *testing.T) client.Client {
				s := runtime.NewScheme()

				require.NoError(t, sonarApi.AddToScheme(s))
				require.NoError(t, corev1.AddToScheme(s))

				secret := corev1.Secret{
					ObjectMeta: ctrl.ObjectMeta{
						Name: "test-secret",
					},
					Data: map[string][]byte{
						"password": []byte("test-password"),
					},
				}

				return fake.NewClientBuilder().WithScheme(s).WithObjects(&secret).Build()
			},
			sonarApiClient: func(t *testing.T) sonar.UserInterface {
				m := mocks.NewMockClientInterface(t)
				m.On("GetUserByLogin", mock.Anything, "test-user").
					Return(nil, sonar.NewHTTPError(http.StatusNotFound, ""))
				m.On("CreateUser", mock.Anything, mock.Anything).
					Return(errors.New("failed to create user"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to create user")
			},
		},
		{
			name: "failed to get user by login",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:  "test-user",
					Secret: "test-secret",
				},
			},
			client: func(t *testing.T) client.Client {
				s := runtime.NewScheme()

				require.NoError(t, sonarApi.AddToScheme(s))
				require.NoError(t, corev1.AddToScheme(s))

				secret := corev1.Secret{
					ObjectMeta: ctrl.ObjectMeta{
						Name: "test-secret",
					},
					Data: map[string][]byte{
						"password": []byte("test-password"),
					},
				}

				return fake.NewClientBuilder().WithScheme(s).WithObjects(&secret).Build()
			},
			sonarApiClient: func(t *testing.T) sonar.UserInterface {
				m := mocks.NewMockClientInterface(t)
				m.On("GetUserByLogin", mock.Anything, "test-user").
					Return(nil, errors.New("failed to get user by login"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get user by login")
			},
		},
		{
			name: "user secret not found",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login:  "test-user",
					Secret: "test-secret",
				},
			},
			client: func(t *testing.T) client.Client {
				s := runtime.NewScheme()

				require.NoError(t, sonarApi.AddToScheme(s))
				require.NoError(t, corev1.AddToScheme(s))

				return fake.NewClientBuilder().WithScheme(s).Build()
			},
			sonarApiClient: func(t *testing.T) sonar.UserInterface {
				return mocks.NewMockClientInterface(t)
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to get user secret")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewCreateUser(tt.sonarApiClient(t), tt.client(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.user)

			tt.wantErr(t, err)
		})
	}
}
