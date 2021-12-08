package sonar

import (
	"context"

	"github.com/pkg/errors"
)

type PermissionTemplate struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	ProjectKeyPattern string `json:"projectKeyPattern"`
}

type PermissionTemplateGroup struct {
	TemplateID  string   `json:"id"`
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
}

func (sc *Client) CreatePermissionTemplate(ctx context.Context, tpl *PermissionTemplate) error {
	var result createPermissionTemplateResponse
	rsp, err := sc.startRequest(ctx).SetResult(&result).SetFormData(map[string]string{
		"name":              tpl.Name,
		"description":       tpl.Description,
		"projectKeyPattern": tpl.ProjectKeyPattern,
	}).Post("/permissions/create_template")

	if err := sc.checkError(rsp, err); err != nil {
		return errors.Wrap(err, "unable to create permission template")
	}

	tpl.ID = result.PermissionTemplate.ID
	return nil
}

func (sc *Client) UpdatePermissionTemplate(ctx context.Context, tpl *PermissionTemplate) error {
	rsp, err := sc.startRequest(ctx).SetFormData(map[string]string{
		"id":                tpl.ID,
		"name":              tpl.Name,
		"description":       tpl.Description,
		"projectKeyPattern": tpl.ProjectKeyPattern,
	}).Post("permissions/update_template")

	if err := sc.checkError(rsp, err); err != nil {
		return errors.Wrap(err, "unable to update permission template")
	}

	return nil
}

func (sc *Client) DeletePermissionTemplate(ctx context.Context, id string) error {
	rsp, err := sc.startRequest(ctx).SetFormData(map[string]string{
		"templateId": id,
	}).Post("/permissions/delete_template")

	if err := sc.checkError(rsp, err); err != nil {
		return errors.Wrap(err, "unable to delete permission template")
	}

	return nil
}

func (sc *Client) SearchPermissionTemplates(ctx context.Context, name string) ([]PermissionTemplate, error) {
	var result searchPermissionTemplatesResponse
	rsp, err := sc.startRequest(ctx).SetQueryParam("q", name).SetResult(&result).
		Get("/permissions/search_templates")
	if err := sc.checkError(rsp, err); err != nil {
		return nil, errors.Wrap(err, "unable to search for permission templates")
	}

	return result.PermissionTemplates, nil
}

func (sc *Client) GetPermissionTemplate(ctx context.Context, name string) (*PermissionTemplate, error) {
	tpls, err := sc.SearchPermissionTemplates(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to search for permission templates")
	}

	for _, t := range tpls {
		if t.Name == name {
			return &t, nil
		}
	}

	return nil, ErrNotFound("permission template not found")
}

func (sc *Client) AddGroupToPermissionTemplate(ctx context.Context, permGroup *PermissionTemplateGroup) error {
	for _, perm := range permGroup.Permissions {
		rsp, err := sc.startRequest(ctx).SetFormData(map[string]string{
			"templateId": permGroup.TemplateID,
			"groupName":  permGroup.GroupName,
			"permission": perm,
		}).Post("/permissions/add_group_to_template")

		if err := sc.checkError(rsp, err); err != nil {
			return errors.Wrap(err, "unable to add group to permission template")
		}
	}

	return nil
}

func (sc *Client) GetPermissionTemplateGroups(ctx context.Context, templateID string) ([]PermissionTemplateGroup, error) {
	var response getPermissionGroupsResponse
	rsp, err := sc.startRequest(ctx).SetResult(&response).
		SetQueryParam("templateId", templateID).Get("/permissions/template_groups")
	if err := sc.checkError(rsp, err); err != nil {
		return nil, errors.Wrap(err, "unable to get permission template groups")
	}

	return response.Groups, nil
}

func (sc *Client) RemoveGroupFromPermissionTemplate(ctx context.Context, permGroup *PermissionTemplateGroup) error {
	for _, perm := range permGroup.Permissions {
		rsp, err := sc.startRequest(ctx).SetFormData(map[string]string{
			"templateId": permGroup.TemplateID,
			"groupName":  permGroup.GroupName,
			"permission": perm,
		}).Post("/permissions/remove_group_from_template")

		if err := sc.checkError(rsp, err); err != nil {
			return errors.Wrap(err, "unable to remove group from permission template")
		}
	}

	return nil
}
