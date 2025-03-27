package sonar

import (
	"context"
	"fmt"
	"net/http"
)

const templateIdName = "templateId"

type PermissionTemplateData struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	ProjectKeyPattern string `json:"projectKeyPattern"`
}

type PermissionTemplate struct {
	ID        string `json:"id,omitempty"`
	IsDefault bool   `json:"isDefault,omitempty"`
	PermissionTemplateData
}

type PermissionTemplateGroup struct {
	GroupID     string   `json:"id"`
	GroupName   string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type getPermissionGroupsResponse struct {
	Groups []PermissionTemplateGroup `json:"groups"`
}

type createPermissionTemplateResponse struct {
	PermissionTemplate PermissionTemplate `json:"permissionTemplate"`
}

type searchPermissionTemplatesResponse struct {
	PermissionTemplates []PermissionTemplate `json:"permissionTemplates"`
	DefaultTemplates    []struct {
		TemplateId string `json:"templateId"`
	} `json:"defaultTemplates"`
}

type getUserPermissionResponse struct {
	Users []struct {
		Login       string   `json:"login"`
		Permissions []string `json:"permissions"`
	} `json:"users"`
}

type getGroupPermissionResponse struct {
	Groups []struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	} `json:"groups"`
}

func (sc *Client) CreatePermissionTemplate(ctx context.Context, tpl *PermissionTemplateData) (*PermissionTemplate, error) {
	var result createPermissionTemplateResponse
	rsp, err := sc.startRequest(ctx).SetResult(&result).SetFormData(map[string]string{
		"name":              tpl.Name,
		"description":       tpl.Description,
		"projectKeyPattern": tpl.ProjectKeyPattern,
	}).Post("/permissions/create_template")

	if err = sc.checkError(rsp, err); err != nil {
		return nil, fmt.Errorf("failed to create permission template: %w", err)
	}

	return &result.PermissionTemplate, nil
}

func (sc *Client) UpdatePermissionTemplate(ctx context.Context, tpl *PermissionTemplate) error {
	rsp, err := sc.startRequest(ctx).SetFormData(map[string]string{
		"id":                tpl.ID,
		"name":              tpl.Name,
		"description":       tpl.Description,
		"projectKeyPattern": tpl.ProjectKeyPattern,
	}).Post("/permissions/update_template")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to update permission template: %w", err)
	}

	return nil
}

func (sc *Client) DeletePermissionTemplate(ctx context.Context, id string) error {
	rsp, err := sc.startRequest(ctx).SetFormData(map[string]string{
		templateIdName: id,
	}).Post("/permissions/delete_template")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to delete permission template: %w", err)
	}

	return nil
}

func (sc *Client) searchPermissionTemplates(ctx context.Context, name string) (*searchPermissionTemplatesResponse, error) {
	var result searchPermissionTemplatesResponse

	rsp, err := sc.startRequest(ctx).SetQueryParam("q", name).SetResult(&result).
		Get("/permissions/search_templates")
	if err = sc.checkError(rsp, err); err != nil {
		return nil, fmt.Errorf("failed to search for permission templates: %w", err)
	}

	return &result, nil
}

func (sc *Client) GetPermissionTemplate(ctx context.Context, name string) (*PermissionTemplate, error) {
	tpls, err := sc.searchPermissionTemplates(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to search for permission templates: %w", err)
	}

	for _, t := range tpls.PermissionTemplates {
		if t.Name == name {
			for _, dt := range tpls.DefaultTemplates {
				if dt.TemplateId == t.ID {
					t.IsDefault = true
					break
				}
			}

			return &t, nil
		}
	}

	return nil, NewHTTPError(http.StatusNotFound, fmt.Sprintf("permission template %s not found", name))
}

func (sc *Client) AddGroupToPermissionTemplate(ctx context.Context, templateID, groupName, permission string) error {
	rsp, err := sc.startRequest(ctx).SetFormData(map[string]string{
		templateIdName: templateID,
		"groupName":    groupName,
		"permission":   permission,
	}).Post("/permissions/add_group_to_template")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to add group %s to permission template: %w", groupName, err)
	}

	return nil
}

// GetPermissionTemplateGroups returns map where key is group name and value is list of permissions.
// Warning: this is a sonar internal endpoint, which may be changed in future versions.
func (sc *Client) GetPermissionTemplateGroups(ctx context.Context, templateID string) (map[string][]string, error) {
	var response getPermissionGroupsResponse
	rsp, err := sc.startRequest(ctx).
		SetResult(&response).
		SetQueryParams(map[string]string{
			"templateId": templateID,
			"ps":         "100",
		}).
		Get("/permissions/template_groups")

	if err = sc.checkError(rsp, err); err != nil {
		return nil, fmt.Errorf("failed to get permission template groups: %w", err)
	}

	result := make(map[string][]string, len(response.Groups))
	for _, g := range response.Groups {
		result[g.GroupName] = g.Permissions
	}

	return result, nil
}

func (sc *Client) RemoveGroupFromPermissionTemplate(ctx context.Context, templateID, groupName, permission string) error {
	rsp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			templateIdName: templateID,
			"groupName":    groupName,
			"permission":   permission,
		}).
		Post("/permissions/remove_group_from_template")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to remove group from permission template: %w", err)
	}

	return nil
}

func (sc *Client) SetDefaultPermissionTemplate(ctx context.Context, name string) error {
	rsp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"templateName": name,
		}).
		Post("/permissions/set_default_template")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to set default permission template: %w", err)
	}

	return nil
}

// GetUserPermissions returns user permissions.
// Warning: this is a sonar internal endpoint, which may be changed in future versions.
// nolint:dupl // this has a lot of common code with GetGroupPermissions, but it's not worth to extract it
func (sc *Client) GetUserPermissions(ctx context.Context, userLogin string) ([]string, error) {
	response := getUserPermissionResponse{}
	rsp, err := sc.startRequest(ctx).
		SetResult(&response).
		SetQueryParams(map[string]string{
			"q":  userLogin,
			"ps": "100",
		}).
		Get("/permissions/users")

	if err = sc.checkError(rsp, err); err != nil {
		return nil, fmt.Errorf("failed to get user %s permission: %w", userLogin, err)
	}

	for _, u := range response.Users {
		if u.Login == userLogin {
			return u.Permissions, nil
		}
	}

	return nil, NewHTTPError(http.StatusNotFound, fmt.Sprintf("user %s not found", userLogin))
}

// AddPermissionToUser adds permission to user.
func (sc *Client) AddPermissionToUser(ctx context.Context, userLogin, permission string) error {
	rsp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"login":      userLogin,
			"permission": permission,
		}).
		Post("/permissions/add_user")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to add permission %s to user %s: %w", permission, userLogin, err)
	}

	return nil
}

// RemovePermissionFromUser removes permission from user.
func (sc *Client) RemovePermissionFromUser(ctx context.Context, userLogin, permission string) error {
	rsp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"login":      userLogin,
			"permission": permission,
		}).
		Post("/permissions/remove_user")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to remove permission %s from user %s: %w", permission, userLogin, err)
	}

	return nil
}

// GetGroupPermissions returns group permissions.
// Warning: this is a sonar internal endpoint, which may be changed in future versions.
// nolint:dupl // this has a lot of common code with GetUserPermissions, but it's not worth to extract it
func (sc *Client) GetGroupPermissions(ctx context.Context, groupName string) ([]string, error) {
	response := getGroupPermissionResponse{}
	rsp, err := sc.startRequest(ctx).
		SetResult(&response).
		SetQueryParams(map[string]string{
			"q":  groupName,
			"ps": "100",
		}).
		Get("/permissions/groups")

	if err = sc.checkError(rsp, err); err != nil {
		return nil, fmt.Errorf("failed to get group %s permission: %w", groupName, err)
	}

	for _, g := range response.Groups {
		if g.Name == groupName {
			return g.Permissions, nil
		}
	}

	return nil, NewHTTPError(http.StatusNotFound, fmt.Sprintf("group %s not found", groupName))
}

// AddPermissionToGroup adds permission to group.
func (sc *Client) AddPermissionToGroup(ctx context.Context, groupName, permission string) error {
	rsp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"groupName":  groupName,
			"permission": permission,
		}).
		Post("/permissions/add_group")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to add permission %s to group %s: %w", permission, groupName, err)
	}

	return nil
}

// RemovePermissionFromGroup removes permission from group.
func (sc *Client) RemovePermissionFromGroup(ctx context.Context, groupName, permission string) error {
	rsp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"groupName":  groupName,
			"permission": permission,
		}).
		Post("/permissions/remove_group")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to remove permission %s from group %s: %w", permission, groupName, err)
	}

	return nil
}
