package sonar

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestSonarClient_CreateUser(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("POST", "/users/create",
		httpmock.NewJsonResponderOrPanic(200, createUserResponse{
			User: User{Login: "userlogin", Name: "username"}}))

	u := User{Name: "foo", Login: "bar"}
	if err := cs.CreateUser(context.Background(), &u); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/users/create",
		httpmock.NewStringResponder(500, "create fatal"))
	err := cs.CreateUser(context.Background(), &u)
	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to create user user: status: 500, body: create fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSonarClient_SearchUsers(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("GET", "/users/search?q=name",
		httpmock.NewJsonResponderOrPanic(200, userSearchResponse{}))

	if _, err := cs.SearchUsers(context.Background(), "name"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/users/search?q=name",
		httpmock.NewStringResponder(500, "search fatal"))

	_, err := cs.SearchUsers(context.Background(), "name")
	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to search for users: status: 500, body: search fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSonarClient_GetUser(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("GET", "/users/search?q=loginName",
		httpmock.NewJsonResponderOrPanic(200, userSearchResponse{Users: []User{
			{Name: "userName", Login: "loginName"},
		}}))

	if _, err := cs.GetUser(context.Background(), "loginName"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/users/search?q=userNameNotFound",
		httpmock.NewJsonResponderOrPanic(200, userSearchResponse{Users: []User{
			{Name: "userName", Login: "loginName"},
		}}))
	_, err := cs.GetUser(context.Background(), "userNameNotFound")
	if err == nil {
		t.Fatal("no error returned")
	}

	if !IsErrNotFound(err) {
		t.Fatalf("wrong error returned: %s", err.Error())
	}

	httpmock.RegisterResponder("GET", "/users/search?q=userNameNotFound",
		httpmock.NewStringResponder(500, "search fatal"))
	_, err = cs.GetUser(context.Background(), "userNameNotFound")
	if err == nil {
		t.Fatal("no error returned")
	}
	if err.Error() != "unable to search for users: unable to search for users: status: 500, body: search fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSonarClient_SearchUserTokens(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("GET", "/user_tokens/search?login=name",
		httpmock.NewJsonResponderOrPanic(200, userTokenSearchResponse{}))

	if _, err := cs.SearchUserTokens(context.Background(), "name"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/user_tokens/search?login=name",
		httpmock.NewStringResponder(500, "search fatal"))

	_, err := cs.SearchUserTokens(context.Background(), "name")
	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to search for user tokens: status: 500, body: search fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSonarClient_GetUserToken(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("GET", "/user_tokens/search?login=loginName",
		httpmock.NewJsonResponderOrPanic(200, userTokenSearchResponse{UserTokens: []UserToken{
			{Name: "tokenName", Login: "loginName"},
		}}))

	if _, err := cs.GetUserToken(context.Background(), "loginName", "tokenName"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/user_tokens/search?login=userNameNotFound",
		httpmock.NewJsonResponderOrPanic(200, userTokenSearchResponse{UserTokens: []UserToken{
			{Name: "userName", Login: "loginName"},
		}}))
	_, err := cs.GetUserToken(context.Background(), "userNameNotFound", "someToken")
	if err == nil {
		t.Fatal("no error returned")
	}

	if !IsErrNotFound(err) {
		t.Fatalf("wrong error returned: %s", err.Error())
	}

	httpmock.RegisterResponder("GET", "/user_tokens/search?login=userNameNotFound",
		httpmock.NewStringResponder(500, "search fatal"))
	_, err = cs.GetUserToken(context.Background(), "userNameNotFound", "someToken")
	if err == nil {
		t.Fatal("no error returned")
	}
	if err.Error() != "unable to search for user tokens: unable to search for user tokens: status: 500, body: search fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}
