package sonar

import (
	"context"
	"fmt"
)

type User struct {
	Login    string `json:"login"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

type UserToken struct {
	Login string `json:"login"`
	Name  string `json:"name,omitempty"`
	Token string `json:"token,omitempty"`
}

type userSearchResponse struct {
	Users []User `json:"users"`
}

type createUserResponse struct {
	User User `json:"user"`
}

type userTokenSearchResponse struct {
	UserTokens []UserToken `json:"usertokens"`
}

// SearchUsers searches for users by login, name, and email.
// The parameter userQuery can either be case-sensitive and perform an exact match or case-insensitive and perform a partial match (contains), depending on the scenario:
// - If the search query is less or equal to 15 characters, then the query is case-insensitive and will match any login, name, or email that contains the search query.
// - If the search query is greater than 15 characters, then the query becomes case-sensitive and will match any login, name, or email that exactly matches the search query.
func (sc *Client) SearchUsers(ctx context.Context, userQuery string) ([]User, error) {
	var userResponse userSearchResponse
	rsp, err := sc.startRequest(ctx).SetResult(&userResponse).
		SetQueryParams(map[string]string{
			"q":  userQuery,
			"ps": "500",
		}).
		Get("/users/search")

	if err = sc.checkError(rsp, err); err != nil {
		return nil, fmt.Errorf("failed to search for users: %w", err)
	}

	return userResponse.Users, nil
}

func (sc Client) GetUserByLogin(ctx context.Context, userLogin string) (*User, error) {
	users, err := sc.SearchUsers(ctx, userLogin)
	if err != nil {
		return nil, fmt.Errorf("failed to search for users: %w", err)
	}

	for _, u := range users {
		if u.Login == userLogin {
			return &u, nil
		}
	}

	return nil, ErrNotFound
}

func (sc *Client) CreateUser(ctx context.Context, user *User) error {
	var createUserRsp createUserResponse

	rsp, err := sc.startRequest(ctx).
		SetResult(&createUserRsp).
		SetFormData(map[string]string{
			"login":    user.Login,
			"name":     user.Name,
			"password": user.Password,
		}).
		Post("/users/create")
	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to create user user: %w", err)
	}

	return nil
}

// UpdateUser updates the user with the given login.
func (sc *Client) UpdateUser(ctx context.Context, user *User) error {
	rsp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"login": user.Login,
			"name":  user.Name,
			"email": user.Email,
		}).
		Post("/users/update")
	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (sc *Client) SearchUserTokens(ctx context.Context, userLogin string) ([]UserToken, error) {
	var userTokenResponse userTokenSearchResponse
	rsp, err := sc.startRequest(ctx).SetResult(&userTokenResponse).
		Get(fmt.Sprintf("/user_tokens/search?login=%s", userLogin))

	if err = sc.checkError(rsp, err); err != nil {
		return nil, fmt.Errorf("failed to search for user tokens: %w", err)
	}

	return userTokenResponse.UserTokens, nil
}

func (sc Client) GetUserToken(ctx context.Context, userLogin, tokenName string) (*UserToken, error) {
	userTokens, err := sc.SearchUserTokens(ctx, userLogin)
	if err != nil {
		return nil, fmt.Errorf("failed to search for user tokens: %w", err)
	}

	for _, ut := range userTokens {
		if ut.Name == tokenName {
			return &ut, nil
		}
	}

	return nil, NotFoundError("Token not found")
}

// GetUserGroups returns all groups that the user is a member of.
func (sc *Client) GetUserGroups(ctx context.Context, userLogin string) ([]Group, error) {
	groups := &groupSearchResponse{}
	rsp, err := sc.startRequest(ctx).
		SetResult(groups).
		SetQueryParams(map[string]string{
			"login": userLogin,
			"ps":    "500",
		}).
		Get("/users/groups")

	if err = sc.checkError(rsp, err); err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	return groups.Groups, nil
}

func (sc *Client) DeactivateUser(ctx context.Context, userLogin string) error {
	rsp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"login": userLogin,
		}).
		Post("/users/deactivate")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}
