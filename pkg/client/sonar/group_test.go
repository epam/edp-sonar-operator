package sonar

import (
	"context"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestSonarClient_CreateGroup(t *testing.T) {
	cs := InitNewRestClient("", "", "")
	httpmock.ActivateNonDefault(cs.resty.GetClient())
	cs.resty.SetDisableWarn(true)

	httpmock.RegisterResponder("POST", "/user_groups/create",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, createGroupResponse{Group: Group{ID: "id1"}}))

	gr := Group{Name: "foo", Description: "bar"}
	if err := cs.CreateGroup(context.Background(), &gr); err != nil {
		t.Fatal(err)
	}

	if gr.ID != "id1" {
		t.Fatal("group id is not set")
	}

	httpmock.RegisterResponder("POST", "/user_groups/create",
		httpmock.NewStringResponder(http.StatusInternalServerError, "create fatal"))
	err := cs.CreateGroup(context.Background(), &gr)
	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to create user group: status: 500, body: create fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSonarClient_UpdateGroup(t *testing.T) {
	cs := InitNewRestClient("", "", "")
	httpmock.ActivateNonDefault(cs.resty.GetClient())
	cs.resty.SetDisableWarn(true)
	httpmock.RegisterResponder("POST", "/user_groups/update",
		httpmock.NewStringResponder(http.StatusOK, ""))

	if err := cs.UpdateGroup(context.Background(), "currentName",
		&Group{Name: "name", Description: "desc"}); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/user_groups/update",
		httpmock.NewStringResponder(http.StatusInternalServerError, "update fatal"))

	err := cs.UpdateGroup(context.Background(), "currentName",
		&Group{Name: "name", Description: "desc"})

	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to update group: status: 500, body: update fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSonarClient_DeleteGroup(t *testing.T) {
	cs := InitNewRestClient("", "", "")
	httpmock.ActivateNonDefault(cs.resty.GetClient())
	cs.resty.SetDisableWarn(true)

	httpmock.RegisterResponder("POST", "/user_groups/delete",
		httpmock.NewStringResponder(http.StatusOK, ""))
	if err := cs.DeleteGroup(context.Background(), "groupName"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/user_groups/delete",
		httpmock.NewStringResponder(http.StatusInternalServerError, "delete fatal"))

	err := cs.DeleteGroup(context.Background(), "groupName")

	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to delete group: status: 500, body: delete fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSonarClient_SearchGroups(t *testing.T) {
	cs := InitNewRestClient("", "", "")
	httpmock.ActivateNonDefault(cs.resty.GetClient())
	cs.resty.SetDisableWarn(true)
	httpmock.RegisterResponder("GET", "/user_groups/search?q=name&f=name",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, groupSearchResponse{}))

	if _, err := cs.SearchGroups(context.Background(), "name"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/user_groups/search?q=name&f=name",
		httpmock.NewStringResponder(http.StatusInternalServerError, "search fatal"))

	_, err := cs.SearchGroups(context.Background(), "name")
	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to search for groups: status: 500, body: search fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSonarClient_GetGroup(t *testing.T) {
	cs := InitNewRestClient("", "", "")

	httpmock.ActivateNonDefault(cs.resty.GetClient())
	cs.resty.SetDisableWarn(true)
	httpmock.RegisterResponder("GET", "/user_groups/search?q=groupName&f=name",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, groupSearchResponse{Groups: []Group{
			{Name: "groupName"},
		}}))

	if _, err := cs.GetGroup(context.Background(), "groupName"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/user_groups/search?q=groupNameNotFound&f=name",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, groupSearchResponse{Groups: []Group{
			{Name: "groupName"},
		}}))
	_, err := cs.GetGroup(context.Background(), "groupNameNotFound")
	if err == nil {
		t.Fatal("no error returned")
	}

	if !IsErrNotFound(err) {
		t.Fatalf("wrong error returned: %s", err.Error())
	}

	httpmock.RegisterResponder("GET", "/user_groups/search?q=groupNameNotFound&f=name",
		httpmock.NewStringResponder(http.StatusInternalServerError, "search fatal"))
	_, err = cs.GetGroup(context.Background(), "groupNameNotFound")
	if err == nil {
		t.Fatal("no error returned")
	}
	if err.Error() != "unable to search for groups: unable to search for groups: status: 500, body: search fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}
