package chain

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar/mocks"
)

func TestCreateProject_ServeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		sonarProject *sonarApi.SonarProject
		setupMocks   func(m *mocks.MockClientInterface)
		wantErr      bool
		errContains  string
	}{
		{
			name: "successful project creation",
			sonarProject: &sonarApi.SonarProject{
				Spec: sonarApi.SonarProjectSpec{
					Key:        "test-project",
					Name:       "Test Project",
					Visibility: "private",
				},
			},
			setupMocks: func(m *mocks.MockClientInterface) {
				// Project doesn't exist, so GetProject returns NotFound error
				m.On("GetProject", mock.Anything, "test-project").Return(nil, sonar.NewHTTPError(404, "project not found"))

				// CreateProject succeeds
				m.On("CreateProject", mock.Anything, &sonar.Project{
					Key:        "test-project",
					Name:       "Test Project",
					Visibility: "private",
					MainBranch: "",
				}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "project already exists and is up to date",
			sonarProject: &sonarApi.SonarProject{
				Spec: sonarApi.SonarProjectSpec{
					Key:        "existing-project",
					Name:       "Existing Project",
					Visibility: "private",
				},
			},
			setupMocks: func(m *mocks.MockClientInterface) {
				// Project exists and is up to date
				m.On("GetProject", mock.Anything, "existing-project").Return(&sonar.Project{
					Key:        "existing-project",
					Name:       "Existing Project",
					Visibility: "private",
					MainBranch: "",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "project exists but needs update",
			sonarProject: &sonarApi.SonarProject{
				Spec: sonarApi.SonarProjectSpec{
					Key:        "update-project",
					Name:       "Updated Project Name",
					Visibility: "public",
				},
			},
			setupMocks: func(m *mocks.MockClientInterface) {
				// Project exists but with different name and visibility
				m.On("GetProject", mock.Anything, "update-project").Return(&sonar.Project{
					Key:        "update-project",
					Name:       "Old Project Name",
					Visibility: "private",
				}, nil)

				// UpdateProject succeeds
				m.On("UpdateProject", mock.Anything, &sonar.Project{
					Key:        "update-project",
					Name:       "Updated Project Name",
					Visibility: "public",
				}).Return(nil)
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
			name: "error creating project",
			sonarProject: &sonarApi.SonarProject{
				Spec: sonarApi.SonarProjectSpec{
					Key:        "create-error-project",
					Name:       "Create Error Project",
					Visibility: "private",
				},
			},
			setupMocks: func(m *mocks.MockClientInterface) {
				// Project doesn't exist
				m.On("GetProject", mock.Anything, "create-error-project").Return(nil, sonar.NewHTTPError(404, "project not found"))

				// CreateProject fails
				m.On("CreateProject", mock.Anything, &sonar.Project{
					Key:        "create-error-project",
					Name:       "Create Error Project",
					Visibility: "private",
					MainBranch: "",
				}).Return(errors.New("failed to create project"))
			},
			wantErr:     true,
			errContains: "failed to create project",
		},
		{
			name: "error updating project",
			sonarProject: &sonarApi.SonarProject{
				Spec: sonarApi.SonarProjectSpec{
					Key:        "update-error-project",
					Name:       "Update Error Project",
					Visibility: "public",
				},
			},
			setupMocks: func(m *mocks.MockClientInterface) {
				// Project exists but with different properties
				m.On("GetProject", mock.Anything, "update-error-project").Return(&sonar.Project{
					Key:        "update-error-project",
					Name:       "Old Name",
					Visibility: "private",
					MainBranch: "",
				}, nil)

				// UpdateProject fails
				m.On("UpdateProject", mock.Anything, &sonar.Project{
					Key:        "update-error-project",
					Name:       "Update Error Project",
					Visibility: "public",
					MainBranch: "",
				}).Return(errors.New("failed to update project"))
			},
			wantErr:     true,
			errContains: "failed to update project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockClient := mocks.NewMockClientInterface(t)
			tt.setupMocks(mockClient)

			handler := NewCreateProject(mockClient)

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

func TestCreateProject_NewCreateProject(t *testing.T) {
	t.Parallel()

	mockClient := mocks.NewMockClientInterface(t)
	handler := NewCreateProject(mockClient)

	require.NotNil(t, handler)
	assert.IsType(t, &CreateProject{}, handler)
}
