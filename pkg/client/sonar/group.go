package sonar

import (
	"context"
	"fmt"
	"net/http"
)

type Group struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type groupSearchResponse struct {
	Groups []Group `json:"groups"`
}

func (sc *Client) SearchGroups(ctx context.Context, groupName string) ([]Group, error) {
	var groupResponse groupSearchResponse
	rsp, err := sc.startRequest(ctx).SetResult(&groupResponse).
		Get(fmt.Sprintf("/user_groups/search?q=%s&f=name", groupName))

	if err = sc.checkError(rsp, err); err != nil {
		return nil, fmt.Errorf("failed to search for groups: %w", err)
	}

	return groupResponse.Groups, nil
}

func (sc Client) GetGroup(ctx context.Context, groupName string) (*Group, error) {
	groups, err := sc.SearchGroups(ctx, groupName)
	if err != nil {
		return nil, fmt.Errorf("failed to search for groups: %w", err)
	}

	for _, g := range groups {
		if g.Name == groupName {
			return &g, nil
		}
	}

	return nil, NewHTTPError(http.StatusNotFound, fmt.Sprintf("group %s not found", groupName))
}

type createGroupResponse struct {
	Group Group `json:"group"`
}

func (sc *Client) CreateGroup(ctx context.Context, group *Group) error {
	var createGroupRsp createGroupResponse

	rsp, err := sc.startRequest(ctx).
		SetResult(&createGroupRsp).
		SetFormData(map[string]string{
			"name":        group.Name,
			"description": group.Description,
		}).
		Post("/user_groups/create")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to create user group: %w", err)
	}
	group.ID = createGroupRsp.Group.ID

	return nil
}

func (sc *Client) UpdateGroup(ctx context.Context, currentName string, group *Group) error {
	rqParams := map[string]string{
		"currentName": currentName,
		"description": group.Description,
	}

	if group.Name != currentName {
		rqParams["name"] = group.Name
	}

	rsp, err := sc.startRequest(ctx).SetFormData(rqParams).Post("/user_groups/update")
	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	return nil
}

func (sc *Client) DeleteGroup(ctx context.Context, groupName string) error {
	rsp, err := sc.startRequest(ctx).SetFormData(map[string]string{
		"name": groupName,
	}).Post("/user_groups/delete")
	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return nil
}

func (sc *Client) AddUserToGroup(ctx context.Context, user, groupName string) error {
	resp, err := sc.startRequest(ctx).
		SetQueryParams(map[string]string{
			nameField:  groupName,
			loginField: user,
		}).
		Post("/user_groups/add_user")
	if err != nil {
		return fmt.Errorf("failed to send requst to add user in group: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("failed to add user %s to group %s, response: %s", user, groupName, resp.String())
	}

	return nil
}

func (sc *Client) RemoveUserFromGroup(ctx context.Context, user, groupName string) error {
	resp, err := sc.startRequest(ctx).
		SetQueryParams(map[string]string{
			nameField:  groupName,
			loginField: user,
		}).
		Post("/user_groups/remove_user")
	if err != nil {
		return fmt.Errorf("failed to send requst to remove user from group: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("failed to remove user %s from group %s, response: %s", user, groupName, resp.String())
	}

	return nil
}
