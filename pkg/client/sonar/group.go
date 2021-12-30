package sonar

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
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

	if err := sc.checkError(rsp, err); err != nil {
		return nil, errors.Wrap(err, "unable to search for groups")
	}

	return groupResponse.Groups, nil
}

func (sc Client) GetGroup(ctx context.Context, groupName string) (*Group, error) {
	groups, err := sc.SearchGroups(ctx, groupName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to search for groups")
	}

	for _, g := range groups {
		if g.Name == groupName {
			return &g, nil
		}
	}

	return nil, ErrNotFound("group not found")
}

type createGroupResponse struct {
	Group Group `json:"group"`
}

func (sc *Client) CreateGroup(ctx context.Context, gr *Group) error {
	var createGroupRsp createGroupResponse

	rsp, err := sc.startRequest(ctx).SetResult(&createGroupRsp).SetFormData(map[string]string{
		"name":        gr.Name,
		"description": gr.Description,
	}).Post("/user_groups/create")
	if err := sc.checkError(rsp, err); err != nil {
		return errors.Wrap(err, "unable to create user group")
	}
	gr.ID = createGroupRsp.Group.ID

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
		return errors.Wrap(err, "unable to update group")
	}

	return nil
}

func (sc *Client) DeleteGroup(ctx context.Context, groupName string) error {
	rsp, err := sc.startRequest(ctx).SetFormData(map[string]string{
		"name": groupName,
	}).Post("/user_groups/delete")
	if err := sc.checkError(rsp, err); err != nil {
		return errors.Wrap(err, "unable to delete group")
	}

	return nil
}
