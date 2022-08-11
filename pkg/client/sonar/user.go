package sonar

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
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

func (sc *Client) SearchUsers(ctx context.Context, userName string) ([]User, error) {
	var userResponse userSearchResponse
	rsp, err := sc.startRequest(ctx).SetResult(&userResponse).
		Get(fmt.Sprintf("/users/search?q=%s", userName))

	if err = sc.checkError(rsp, err); err != nil {
		return nil, errors.Wrap(err, "unable to search for users")
	}

	return userResponse.Users, nil
}

func (sc Client) GetUser(ctx context.Context, userName string) (*User, error) {
	users, err := sc.SearchUsers(ctx, userName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to search for users")
	}

	for _, u := range users {
		if u.Login == userName {
			return &u, nil
		}
	}

	return nil, NotFoundError("user not found")
}

func (sc *Client) CreateUser(ctx context.Context, user *User) error {
	var createUserRsp createUserResponse

	rsp, err := sc.startRequest(ctx).SetResult(&createUserRsp).SetFormData(map[string]string{
		"login":    user.Login,
		"name":     user.Name,
		"password": user.Password,
	}).Post("/users/create")
	if err = sc.checkError(rsp, err); err != nil {
		return errors.Wrap(err, "unable to create user user")
	}

	return nil
}

func (sc *Client) SearchUserTokens(ctx context.Context, userLogin string) ([]UserToken, error) {
	var userTokenResponse userTokenSearchResponse
	rsp, err := sc.startRequest(ctx).SetResult(&userTokenResponse).
		Get(fmt.Sprintf("/user_tokens/search?login=%s", userLogin))

	if err = sc.checkError(rsp, err); err != nil {
		return nil, errors.Wrap(err, "unable to search for user tokens")
	}

	return userTokenResponse.UserTokens, nil
}

func (sc Client) GetUserToken(ctx context.Context, userLogin, tokenName string) (*UserToken, error) {
	userTokens, err := sc.SearchUserTokens(ctx, userLogin)
	if err != nil {
		return nil, errors.Wrap(err, "unable to search for user tokens")
	}

	for _, ut := range userTokens {
		if ut.Name == tokenName {
			return &ut, nil
		}
	}

	return nil, NotFoundError("Token not found")
}
