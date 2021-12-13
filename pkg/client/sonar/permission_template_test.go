package sonar

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
)

func initClient() *Client {
	cs := InitNewRestClient("", "", "")
	httpmock.ActivateNonDefault(cs.resty.GetClient())
	cs.resty.SetDisableWarn(true)

	return cs
}

func TestSonarClient_CreatePermissionTemplate(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("POST", "/permissions/create_template",
		httpmock.NewStringResponder(200, ""))
	if err := cs.CreatePermissionTemplate(context.Background(), &PermissionTemplate{
		Name: "foo",
	}); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/permissions/create_template",
		httpmock.NewStringResponder(500, "create fatal"))

	err := cs.CreatePermissionTemplate(context.Background(), &PermissionTemplate{
		Name: "foo",
	})

	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to create permission template: status: 500, body: create fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_UpdatePermissionTemplate(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("POST", "/permissions/update_template",
		httpmock.NewStringResponder(200, ""))
	if err := cs.UpdatePermissionTemplate(context.Background(), &PermissionTemplate{Name: "foo"}); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/permissions/update_template",
		httpmock.NewStringResponder(500, "update fatal"))

	err := cs.UpdatePermissionTemplate(context.Background(), &PermissionTemplate{
		Name: "foo",
	})

	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to update permission template: status: 500, body: update fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_DeletePermissionTemplate(t *testing.T) {
	cs := initClient()

	httpmock.RegisterResponder("POST", "/permissions/delete_template",
		httpmock.NewStringResponder(200, ""))
	if err := cs.DeletePermissionTemplate(context.Background(), "id1"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/permissions/delete_template",
		httpmock.NewStringResponder(500, "delete fatal"))
	err := cs.DeletePermissionTemplate(context.Background(), "id1")

	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to delete permission template: status: 500, body: delete fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_SearchPermissionTemplates(t *testing.T) {
	cs := initClient()
	httpmock.RegisterResponder("GET", "/permissions/search_templates?q=test",
		httpmock.NewStringResponder(200, ""))
	if _, err := cs.SearchPermissionTemplates(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/permissions/search_templates?q=test",
		httpmock.NewStringResponder(500, "search fatal"))
	_, err := cs.SearchPermissionTemplates(context.Background(), "test")

	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to search for permission templates: status: 500, body: search fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_GetPermissionTemplate(t *testing.T) {
	cs := initClient()

	httpmock.RegisterResponder("GET", "/permissions/search_templates?q=test",
		httpmock.NewJsonResponderOrPanic(200,
			searchPermissionTemplatesResponse{PermissionTemplates: []PermissionTemplate{
				{
					Name: "test",
				},
			}}))
	if _, err := cs.GetPermissionTemplate(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/permissions/search_templates?q=test",
		httpmock.NewJsonResponderOrPanic(200,
			searchPermissionTemplatesResponse{PermissionTemplates: []PermissionTemplate{
				{
					Name: "mest",
				},
			}}))

	_, err := cs.GetPermissionTemplate(context.Background(), "test")
	if err == nil {
		t.Fatal("no error returned")
	}

	if !IsErrNotFound(err) {
		t.Fatalf("wrong error returned: %s", err.Error())
	}

	httpmock.RegisterResponder("GET", "/permissions/search_templates?q=test",
		httpmock.NewStringResponder(500, "search fatal"))

	_, err = cs.GetPermissionTemplate(context.Background(), "test")
	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to search for permission templates: unable to search for permission templates: status: 500, body: search fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_AddGroupToPermissionTemplate(t *testing.T) {
	sc := initClient()

	httpmock.RegisterResponder("POST", "/permissions/add_group_to_template",
		httpmock.NewStringResponder(200, ""))
	if err := sc.AddGroupToPermissionTemplate(context.Background(),
		&PermissionTemplateGroup{GroupName: "test", Permissions: []string{"admin"}}); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/permissions/add_group_to_template",
		httpmock.NewStringResponder(500, "add fatal"))

	err := sc.AddGroupToPermissionTemplate(context.Background(),
		&PermissionTemplateGroup{GroupName: "test", Permissions: []string{"admin"}})

	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to add group to permission template: status: 500, body: add fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_GetPermissionTemplateGroups(t *testing.T) {
	sc := initClient()
	httpmock.RegisterResponder("GET", "/permissions/template_groups?templateId=tplid1",
		httpmock.NewStringResponder(200, ""))
	if _, err := sc.GetPermissionTemplateGroups(context.Background(), "tplid1"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/permissions/template_groups?templateId=tplid1",
		httpmock.NewStringResponder(500, "get template groups fatal"))

	_, err := sc.GetPermissionTemplateGroups(context.Background(), "tplid1")
	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to get permission template groups: status: 500, body: get template groups fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_RemoveGroupFromPermissionTemplate(t *testing.T) {
	sc := initClient()

	httpmock.RegisterResponder("POST", "/permissions/remove_group_from_template",
		httpmock.NewStringResponder(200, ""))
	if err := sc.RemoveGroupFromPermissionTemplate(context.Background(),
		&PermissionTemplateGroup{GroupName: "test1", Permissions: []string{"foo"}}); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/permissions/remove_group_from_template",
		httpmock.NewStringResponder(500, "remove fatal"))

	err := sc.RemoveGroupFromPermissionTemplate(context.Background(),
		&PermissionTemplateGroup{GroupName: "test1", Permissions: []string{"foo"}})

	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to remove group from permission template: status: 500, body: remove fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_SetDefaultPermissionTemplate(t *testing.T) {
	sc := initClient()
	httpmock.RegisterResponder("POST", "/permissions/set_default_template",
		httpmock.NewStringResponder(200, ""))
	if err := sc.SetDefaultPermissionTemplate(context.Background(), "test1"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/permissions/set_default_template",
		httpmock.NewStringResponder(500, "set default fatal"))

	err := sc.SetDefaultPermissionTemplate(context.Background(), "test1")
	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to set default permission template: status: 500, body: set default fatal" {
		t.Fatalf("wrong err returned: %s", err.Error())
	}
}
