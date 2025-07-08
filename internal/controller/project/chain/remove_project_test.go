package chain

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar/mocks"
)

func TestRemoveProject_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		sonarProject *sonarApi.SonarProject
		setupMocks   func(m *mocks.MockClientInterface)
		wantErr      bool
		errContains  string
	}{
		{
			name: "successful project deletion",
			sonarProject: &sonarApi.SonarProject{
				Spec: sonarApi.SonarProjectSpec{
					Key:        "test-project",
					Name:       "Test Project",
					Visibility: "private",
				},
			},
			setupMocks: func(m *mocks.MockClientInterface) {
				// Project exists
				m.On("GetProject", mock.Anything, "test-project").Return(&sonar.Project{
					Key:        "test-project",
					Name:       "Test Project",
					Visibility: "private",
				}, nil)

				// DeleteProject succeeds
				m.On("DeleteProject", mock.Anything, "test-project").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "project does not exist - no error",
			sonarProject: &sonarApi.SonarProject{
				Spec: sonarApi.SonarProjectSpec{
					Key:        "non-existent-project",
					Name:       "Non-existent Project",
					Visibility: "private",
				},
			},
			setupMocks: func(m *mocks.MockClientInterface) {
				// Project doesn't exist
				m.On("GetProject", mock.Anything, "non-existent-project").Return(nil, sonar.NewHTTPError(404, "project not found"))
			},
			wantErr: false,
		},
		{
			name: "error checking if project exists",
			sonarProject: &sonarApi.SonarProject{
				Spec: sonarApi.SonarProjectSpec{
					Key:        "error-project",
					Name:       "Error Project",
					Visibility: "private",
				},
			},
			setupMocks: func(m *mocks.MockClientInterface) {
				// GetProject returns a non-404 error
				m.On("GetProject", mock.Anything, "error-project").Return(nil, errors.New("internal server error"))
			},
			wantErr:     true,
			errContains: "failed to check if project exists",
		},
		{
			name: "error deleting project",
			sonarProject: &sonarApi.SonarProject{
				Spec: sonarApi.SonarProjectSpec{
					Key:        "delete-error-project",
					Name:       "Delete Error Project",
					Visibility: "private",
				},
			},
			setupMocks: func(m *mocks.MockClientInterface) {
				// Project exists
				m.On("GetProject", mock.Anything, "delete-error-project").Return(&sonar.Project{
					Key:        "delete-error-project",
					Name:       "Delete Error Project",
					Visibility: "private",
				}, nil)

				// DeleteProject fails
				m.On("DeleteProject", mock.Anything, "delete-error-project").Return(errors.New("failed to delete project"))
			},
			wantErr:     true,
			errContains: "failed to delete project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockClient := mocks.NewMockClientInterface(t)
			tt.setupMocks(mockClient)

			handler := &RemoveProject{sonarApiClient: mockClient}

			err := handler.ServeRequest(context.Background(), tt.sonarProject)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
