package chain

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar/mocks"
)

func TestSetDefaultPermissionTemplate_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		sonar          *sonarApi.Sonar
		sonarApiClient func(t *testing.T) sonar.PermissionTemplateInterface
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "default permission template is set",
			sonar: &sonarApi.Sonar{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sonar",
				},
				Spec: sonarApi.SonarSpec{
					DefaultPermissionTemplate: "default",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewMockClientInterface(t)
				m.On("SetDefaultPermissionTemplate", mock.Anything, "default").
					Return(nil)

				return m
			},
			wantErr: require.NoError,
		},
		{
			name: "default permission template is not set",
			sonar: &sonarApi.Sonar{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sonar",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				return mocks.NewMockClientInterface(t)
			},
			wantErr: require.NoError,
		},
		{
			name: "failed to set default permission template",
			sonar: &sonarApi.Sonar{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sonar",
				},
				Spec: sonarApi.SonarSpec{
					DefaultPermissionTemplate: "default",
				},
			},
			sonarApiClient: func(t *testing.T) sonar.PermissionTemplateInterface {
				m := mocks.NewMockClientInterface(t)
				m.On("SetDefaultPermissionTemplate", mock.Anything, "default").
					Return(errors.New("failed"))

				return m
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to set default permission template default")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewSetDefaultPermissionTemplate(tt.sonarApiClient(t))
			err := h.ServeRequest(ctrl.LoggerInto(context.Background(), logr.Discard()), tt.sonar)
			tt.wantErr(t, err)
		})
	}
}
