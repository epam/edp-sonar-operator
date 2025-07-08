package sonar

import (
	"context"
	"fmt"
	"net/http"
)

// Project represents a SonarQube project.
type Project struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
	MainBranch string `json:"mainBranch,omitempty"`
}

type projectSearchResponse struct {
	Projects []Project `json:"components"`
	Paging   struct {
		PageIndex int `json:"pageIndex"`
		PageSize  int `json:"pageSize"`
		Total     int `json:"total"`
	} `json:"paging"`
}

// CreateProject creates a new project in SonarQube.
func (sc *Client) CreateProject(ctx context.Context, project *Project) error {
	formData := map[string]string{
		"project":    project.Key,
		"name":       project.Name,
		"visibility": project.Visibility,
	}

	// Add mainBranch parameter if provided
	if project.MainBranch != "" {
		formData["mainBranch"] = project.MainBranch
	}

	resp, err := sc.startRequest(ctx).
		SetFormData(formData).
		Post("/projects/create")

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// GetProject returns the project with the given key.
func (sc *Client) GetProject(ctx context.Context, projectKey string) (*Project, error) {
	var projectResponse projectSearchResponse
	resp, err := sc.startRequest(ctx).
		SetResult(&projectResponse).
		SetQueryParams(map[string]string{
			"projects": projectKey,
			"ps":       "1",
		}).
		Get("/projects/search")

	if err = sc.checkError(resp, err); err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	for _, project := range projectResponse.Projects {
		if project.Key == projectKey {
			return &project, nil
		}
	}

	return nil, NewHTTPError(http.StatusNotFound, fmt.Sprintf("project %s not found", projectKey))
}

// UpdateProject updates the project with the given key.
func (sc *Client) UpdateProject(ctx context.Context, project *Project) error {
	// Update visibility if needed
	if project.Visibility != "" {
		resp, err := sc.startRequest(ctx).
			SetFormData(map[string]string{
				"project":    project.Key,
				"visibility": project.Visibility,
			}).
			Post("/projects/update_visibility")

		if err = sc.checkError(resp, err); err != nil {
			return fmt.Errorf("failed to update project visibility: %w", err)
		}
	}

	return nil
}

// DeleteProject deletes the project with the given key.
func (sc *Client) DeleteProject(ctx context.Context, projectKey string) error {
	resp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"project": projectKey,
		}).
		Post("/projects/delete")

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}
