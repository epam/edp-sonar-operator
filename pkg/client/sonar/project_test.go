package sonar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		project        *Project
		serverResponse int
		serverBody     string
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful project creation",
			project: &Project{
				Key:        "test-project",
				Name:       "Test Project",
				Visibility: "private",
			},
			serverResponse: http.StatusOK,
			serverBody:     `{"project":{"key":"test-project","name":"Test Project","visibility":"private"}}`,
			wantErr:        false,
		},
		{
			name: "project already exists",
			project: &Project{
				Key:        "existing-project",
				Name:       "Existing Project",
				Visibility: "private",
			},
			serverResponse: http.StatusBadRequest,
			serverBody:     `{"errors":[{"msg":"Could not create project. A project with key 'existing-project' already exists."}]}`,
			wantErr:        true,
			errContains:    "failed to create project",
		},
		{
			name: "invalid project key",
			project: &Project{
				Key:        "invalid key",
				Name:       "Invalid Project",
				Visibility: "private",
			},
			serverResponse: http.StatusBadRequest,
			serverBody:     `{"errors":[{"msg":"Invalid project key"}]}`,
			wantErr:        true,
			errContains:    "failed to create project",
		},
		{
			name: "server error",
			project: &Project{
				Key:        "test-project",
				Name:       "Test Project",
				Visibility: "private",
			},
			serverResponse: http.StatusInternalServerError,
			serverBody:     `{"errors":[{"msg":"Internal server error"}]}`,
			wantErr:        true,
			errContains:    "failed to create project",
		},
		{
			name: "successful project creation with main branch",
			project: &Project{
				Key:        "test-project-with-branch",
				Name:       "Test Project With Branch",
				Visibility: "private",
				MainBranch: "develop",
			},
			serverResponse: http.StatusOK,
			serverBody:     `{"project":{"key":"test-project-with-branch","name":"Test Project With Branch","visibility":"private"}}`,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/projects/create", r.URL.Path)

				err := r.ParseForm()
				require.NoError(t, err)

				assert.Equal(t, tt.project.Key, r.FormValue("project"))
				assert.Equal(t, tt.project.Name, r.FormValue("name"))
				assert.Equal(t, tt.project.Visibility, r.FormValue("visibility"))

				// Check mainBranch parameter if provided
				if tt.project.MainBranch != "" {
					assert.Equal(t, tt.project.MainBranch, r.FormValue("mainBranch"))
				}

				w.WriteHeader(tt.serverResponse)
				_, err = w.Write([]byte(tt.serverBody))
				require.NoError(t, err)
			}))
			defer server.Close()

			client := NewClient(server.URL, "user", "password")

			err := client.CreateProject(context.Background(), tt.project)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_GetProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		projectKey     string
		serverResponse int
		serverBody     string
		expectedResult *Project
		wantErr        bool
		errContains    string
	}{
		{
			name:           "successful project retrieval",
			projectKey:     "test-project",
			serverResponse: http.StatusOK,
			serverBody: `{
				"components": [
					{
						"key": "test-project",
						"name": "Test Project",
						"visibility": "private"
					}
				],
				"paging": {
					"pageIndex": 1,
					"pageSize": 1,
					"total": 1
				}
			}`,
			expectedResult: &Project{
				Key:        "test-project",
				Name:       "Test Project",
				Visibility: "private",
			},
			wantErr: false,
		},
		{
			name:           "project not found",
			projectKey:     "non-existent-project",
			serverResponse: http.StatusOK,
			serverBody: `{
				"components": [],
				"paging": {
					"pageIndex": 1,
					"pageSize": 1,
					"total": 0
				}
			}`,
			expectedResult: nil,
			wantErr:        true,
			errContains:    "project non-existent-project not found",
		},
		{
			name:           "server error",
			projectKey:     "test-project",
			serverResponse: http.StatusInternalServerError,
			serverBody:     `{"errors":[{"msg":"Internal server error"}]}`,
			expectedResult: nil,
			wantErr:        true,
			errContains:    "failed to get project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/projects/search", r.URL.Path)
				assert.Equal(t, tt.projectKey, r.URL.Query().Get("projects"))
				assert.Equal(t, "1", r.URL.Query().Get("ps"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverResponse)
				_, err := w.Write([]byte(tt.serverBody))
				require.NoError(t, err)
			}))
			defer server.Close()

			client := NewClient(server.URL, "user", "password")

			result, err := client.GetProject(context.Background(), tt.projectKey)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestClient_UpdateProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		project        *Project
		serverResponse int
		serverBody     string
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful project update",
			project: &Project{
				Key:        "test-project",
				Name:       "Updated Test Project",
				Visibility: "public",
			},
			serverResponse: http.StatusOK,
			serverBody:     `{}`,
			wantErr:        false,
		},
		{
			name: "project not found",
			project: &Project{
				Key:        "non-existent-project",
				Name:       "Updated Project",
				Visibility: "public",
			},
			serverResponse: http.StatusNotFound,
			serverBody:     `{"errors":[{"msg":"Project not found"}]}`,
			wantErr:        true,
			errContains:    "failed to update project visibility",
		},
		{
			name: "server error",
			project: &Project{
				Key:        "test-project",
				Name:       "Updated Test Project",
				Visibility: "public",
			},
			serverResponse: http.StatusInternalServerError,
			serverBody:     `{"errors":[{"msg":"Internal server error"}]}`,
			wantErr:        true,
			errContains:    "failed to update project visibility",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/projects/update_visibility", r.URL.Path)

				err := r.ParseForm()
				require.NoError(t, err)

				assert.Equal(t, tt.project.Key, r.FormValue("project"))
				assert.Equal(t, tt.project.Visibility, r.FormValue("visibility"))

				w.WriteHeader(tt.serverResponse)
				_, err = w.Write([]byte(tt.serverBody))
				require.NoError(t, err)
			}))

			defer server.Close()

			client := NewClient(server.URL, "user", "password")

			err := client.UpdateProject(context.Background(), tt.project)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_DeleteProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		projectKey     string
		serverResponse int
		serverBody     string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "successful project deletion",
			projectKey:     "test-project",
			serverResponse: http.StatusOK,
			serverBody:     `{}`,
			wantErr:        false,
		},
		{
			name:           "project not found",
			projectKey:     "non-existent-project",
			serverResponse: http.StatusNotFound,
			serverBody:     `{"errors":[{"msg":"Project not found"}]}`,
			wantErr:        true,
			errContains:    "failed to delete project",
		},
		{
			name:           "server error",
			projectKey:     "test-project",
			serverResponse: http.StatusInternalServerError,
			serverBody:     `{"errors":[{"msg":"Internal server error"}]}`,
			wantErr:        true,
			errContains:    "failed to delete project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/projects/delete", r.URL.Path)

				err := r.ParseForm()
				require.NoError(t, err)

				assert.Equal(t, tt.projectKey, r.FormValue("project"))

				w.WriteHeader(tt.serverResponse)
				_, err = w.Write([]byte(tt.serverBody))
				require.NoError(t, err)
			}))
			defer server.Close()

			client := NewClient(server.URL, "user", "password")

			err := client.DeleteProject(context.Background(), tt.projectKey)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
