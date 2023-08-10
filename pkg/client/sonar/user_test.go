package sonar

import (
	"context"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSonarClient_CreateUser(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("POST", "/api/users/create",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, createUserResponse{
			User: User{Login: "userlogin", Name: "username"},
		}))

	u := User{Name: "foo", Login: "bar"}
	if err := cs.CreateUser(context.Background(), &u); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/api/users/create",
		httpmock.NewStringResponder(http.StatusInternalServerError, "create fatal"))
	err := cs.CreateUser(context.Background(), &u)
	require.Error(t, err)

	if err.Error() != "failed to create user user: status: 500, body: create fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSonarClient_SearchUsers(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("GET", "/api/users/search?q=name",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, userSearchResponse{}))

	if _, err := cs.SearchUsers(context.Background(), "name"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/api/users/search?q=name",
		httpmock.NewStringResponder(http.StatusInternalServerError, "search fatal"))

	_, err := cs.SearchUsers(context.Background(), "name")
	require.Error(t, err)

	if err.Error() != "failed to search for users: status: 500, body: search fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSonarClient_GetUser(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("GET", "/api/users/search?q=loginName",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, userSearchResponse{Users: []User{
			{Name: "userName", Login: "loginName"},
		}}))

	if _, err := cs.GetUser(context.Background(), "loginName"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/api/users/search?q=userNameNotFound",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, userSearchResponse{Users: []User{
			{Name: "userName", Login: "loginName"},
		}}))
	_, err := cs.GetUser(context.Background(), "userNameNotFound")
	require.Error(t, err)

	if !IsErrNotFound(err) {
		t.Fatalf("wrong error returned: %s", err.Error())
	}

	httpmock.RegisterResponder("GET", "/api/users/search?q=userNameNotFound",
		httpmock.NewStringResponder(http.StatusInternalServerError, "search fatal"))

	_, err = cs.GetUser(context.Background(), "userNameNotFound")
	require.Error(t, err)

	assert.Equal(t, "failed to search for users: failed to search for users: status: 500, body: search fatal", err.Error())
}

func TestSonarClient_SearchUserTokens(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("GET", "/api/user_tokens/search?login=name",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, userTokenSearchResponse{}))

	if _, err := cs.SearchUserTokens(context.Background(), "name"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/api/user_tokens/search?login=name",
		httpmock.NewStringResponder(http.StatusInternalServerError, "search fatal"))

	_, err := cs.SearchUserTokens(context.Background(), "name")
	require.Error(t, err)

	if err.Error() != "failed to search for user tokens: status: 500, body: search fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSonarClient_GetUserToken(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("GET", "/api/user_tokens/search?login=loginName",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, userTokenSearchResponse{UserTokens: []UserToken{
			{Name: "tokenName", Login: "loginName"},
		}}))

	if _, err := cs.GetUserToken(context.Background(), "loginName", "tokenName"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/api/user_tokens/search?login=userNameNotFound",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, userTokenSearchResponse{UserTokens: []UserToken{
			{Name: "userName", Login: "loginName"},
		}}))

	_, err := cs.GetUserToken(context.Background(), "userNameNotFound", "someToken")
	require.Error(t, err)

	if !IsErrNotFound(err) {
		t.Fatalf("wrong error returned: %s", err.Error())
	}

	httpmock.RegisterResponder("GET", "/api/user_tokens/search?login=userNameNotFound",
		httpmock.NewStringResponder(http.StatusInternalServerError, "search fatal"))

	_, err = cs.GetUserToken(context.Background(), "userNameNotFound", "someToken")
	require.Error(t, err)

	assert.Equal(t, "failed to search for user tokens: failed to search for user tokens: status: 500, body: search fatal", err.Error())
}
