package chain

import (
	"context"
	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestCheckConnection_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		sonarApiClient func(t *testing.T) sonar.System
		sonar          *sonarApi.Sonar
		wantErr        require.ErrorAssertionFunc
		wantStatus     sonarApi.SonarStatus
	}{
		{
			name: "connection is established",
			sonarApiClient: func(t *testing.T) sonar.System {
				m := mocks.NewClientInterface(t)
				m.On("Health", mock.Anything).
					Return(&sonar.SystemHealth{Health: "GREEN"}, nil)

				return m
			},
			sonar: &sonarApi.Sonar{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sonar",
				},
			},
			wantErr: require.NoError,
			wantStatus: sonarApi.SonarStatus{
				Value:     "GREEN",
				Connected: true,
			},
		},
		{
			name: "failed to connect",
			sonarApiClient: func(t *testing.T) sonar.System {
				m := mocks.NewClientInterface(t)
				m.On("Health", mock.Anything).
					Return(nil, errors.New("failed to connect"))

				return m
			},
			sonar: &sonarApi.Sonar{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sonar",
				},
			},
			wantErr: func(t require.TestingT, err error, i ...interface{}) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "failed to connect")
			},
			wantStatus: sonarApi.SonarStatus{
				Connected: false,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewCheckConnection(tt.sonarApiClient(t))
			err := h.ServeRequest(context.Background(), tt.sonar)
			tt.wantErr(t, err)
			require.Equal(t, tt.wantStatus, tt.sonar.Status)
		})
	}
}
