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

func TestRemoveUser_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		user           *sonarApi.SonarUser
		sonarApiClient func(t *testing.T) sonar.UserInterface
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "user removed successfully",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login: "test-user",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.UserInterface {
				m := mocks.NewClientInterface(t)

				m.On("DeactivateUser", mock.Anything, "test-user").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to remove user",
			user: &sonarApi.SonarUser{
				Spec: sonarApi.SonarUserSpec{
					Login: "test-user",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.UserInterface {
				m := mocks.NewClientInterface(t)

				m.On("DeactivateUser", mock.Anything, "test-user").
					Return(errors.New("failed to remove user"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to remove user")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := NewRemoveUser(tt.sonarApiClient(t))
			err := r.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.user)
			tt.wantErr(t, err)
		})
	}
}
