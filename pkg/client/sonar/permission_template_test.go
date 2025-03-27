package sonar

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func initClient() *Client {
	cs := NewClient("", "", "")
	httpmock.ActivateNonDefault(cs.resty.GetClient())
	cs.resty.SetDisableWarn(true)

	return cs
}

func TestSonarClient_CreatePermissionTemplate(t *testing.T) {
	cs := initClient()

	httpmock.RegisterResponder("POST", "/api/permissions/create_template",
		httpmock.NewStringResponder(http.StatusOK, ""))

	if _, err := cs.CreatePermissionTemplate(context.Background(), &PermissionTemplateData{
		Name: "foo",
	}); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/api/permissions/create_template",
		httpmock.NewStringResponder(http.StatusInternalServerError, "create fatal"))

	_, err := cs.CreatePermissionTemplate(context.Background(), &PermissionTemplateData{
		Name: "foo",
	})

	require.Error(t, err)

	if err.Error() != "failed to create permission template: status: 500, body: create fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_UpdatePermissionTemplate(t *testing.T) {
	cs := initClient()

	httpmock.RegisterResponder("POST", "/api/permissions/update_template",
		httpmock.NewStringResponder(http.StatusOK, ""))

	if err := cs.UpdatePermissionTemplate(context.Background(), &PermissionTemplate{PermissionTemplateData: PermissionTemplateData{Name: "foo"}}); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/api/permissions/update_template",
		httpmock.NewStringResponder(http.StatusInternalServerError, "update fatal"))

	err := cs.UpdatePermissionTemplate(context.Background(), &PermissionTemplate{
		PermissionTemplateData: PermissionTemplateData{
			Name: "foo",
		},
	})

	require.Error(t, err)

	if err.Error() != "failed to update permission template: status: 500, body: update fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_DeletePermissionTemplate(t *testing.T) {
	cs := initClient()

	httpmock.RegisterResponder("POST", "/api/permissions/delete_template",
		httpmock.NewStringResponder(http.StatusOK, ""))

	if err := cs.DeletePermissionTemplate(context.Background(), "id1"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/api/permissions/delete_template",
		httpmock.NewStringResponder(http.StatusInternalServerError, "delete fatal"))

	err := cs.DeletePermissionTemplate(context.Background(), "id1")

	require.Error(t, err)

	if err.Error() != "failed to delete permission template: status: 500, body: delete fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_SearchPermissionTemplates(t *testing.T) {
	cs := initClient()

	httpmock.RegisterResponder("GET", "/api/permissions/search_templates?q=test",
		httpmock.NewStringResponder(http.StatusOK, ""))

	if _, err := cs.searchPermissionTemplates(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/api/permissions/search_templates?q=test",
		httpmock.NewStringResponder(http.StatusInternalServerError, "search fatal"))

	_, err := cs.searchPermissionTemplates(context.Background(), "test")

	require.Error(t, err)

	if err.Error() != "failed to search for permission templates: status: 500, body: search fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_GetPermissionTemplate(t *testing.T) {
	cs := initClient()

	httpmock.RegisterResponder("GET", "/api/permissions/search_templates?q=test",
		httpmock.NewJsonResponderOrPanic(http.StatusOK,
			searchPermissionTemplatesResponse{PermissionTemplates: []PermissionTemplate{
				{
					PermissionTemplateData: PermissionTemplateData{
						Name: "test",
					},
				},
			}}))

	if _, err := cs.GetPermissionTemplate(context.Background(), "test"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/api/permissions/search_templates?q=test",
		httpmock.NewJsonResponderOrPanic(http.StatusOK,
			searchPermissionTemplatesResponse{PermissionTemplates: []PermissionTemplate{
				{
					PermissionTemplateData: PermissionTemplateData{
						Name: "mest",
					},
				},
			}}))

	_, err := cs.GetPermissionTemplate(context.Background(), "test")
	require.Error(t, err)

	if !IsErrNotFound(err) {
		t.Fatalf("wrong error returned: %s", err.Error())
	}

	httpmock.RegisterResponder("GET", "/api/permissions/search_templates?q=test",
		httpmock.NewStringResponder(http.StatusInternalServerError, "search fatal"))

	_, err = cs.GetPermissionTemplate(context.Background(), "test")
	require.Error(t, err, "no error returned")

	if err.Error() != "failed to search for permission templates: failed to search for permission templates: status: 500, body: search fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_AddGroupToPermissionTemplate(t *testing.T) {
	sc := initClient()

	httpmock.RegisterResponder("POST", "/api/permissions/add_group_to_template",
		httpmock.NewStringResponder(http.StatusOK, ""))

	if err := sc.AddGroupToPermissionTemplate(context.Background(), "tpl1", "test", "admin"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/api/permissions/add_group_to_template",
		httpmock.NewStringResponder(http.StatusInternalServerError, "add fatal"))

	err := sc.AddGroupToPermissionTemplate(context.Background(), "tpl1", "test", "admin")

	require.Error(t, err)
	require.Contains(t, err.Error(), "add fatal")
}

func TestClient_GetPermissionTemplateGroups(t *testing.T) {
	sc := initClient()

	httpmock.RegisterRegexpResponder("GET", regexp.MustCompile("/api/permissions/template_groups.*"),
		httpmock.NewStringResponder(http.StatusOK, ""))

	if _, err := sc.GetPermissionTemplateGroups(context.Background(), "tplid1"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterRegexpResponder("GET", regexp.MustCompile("/api/permissions/template_groups.*"),
		httpmock.NewStringResponder(http.StatusInternalServerError, "get template groups fatal"))

	_, err := sc.GetPermissionTemplateGroups(context.Background(), "tplid1")

	require.Error(t, err)
	require.Contains(t, err.Error(), "get template groups fatal")
}

func TestClient_RemoveGroupFromPermissionTemplate(t *testing.T) {
	sc := initClient()

	httpmock.RegisterResponder("POST", "/api/permissions/remove_group_from_template",
		httpmock.NewStringResponder(http.StatusOK, ""))

	if err := sc.RemoveGroupFromPermissionTemplate(context.Background(), "tpl1", "test1", "foo"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/api/permissions/remove_group_from_template",
		httpmock.NewStringResponder(http.StatusInternalServerError, "remove fatal"))

	err := sc.RemoveGroupFromPermissionTemplate(context.Background(), "tpl1", "test1", "foo")

	require.Error(t, err)

	if err.Error() != "failed to remove group from permission template: status: 500, body: remove fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestClient_SetDefaultPermissionTemplate(t *testing.T) {
	sc := initClient()

	httpmock.RegisterResponder("POST", "/api/permissions/set_default_template",
		httpmock.NewStringResponder(http.StatusOK, ""))

	if err := sc.SetDefaultPermissionTemplate(context.Background(), "test1"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("POST", "/api/permissions/set_default_template",
		httpmock.NewStringResponder(http.StatusInternalServerError, "set default fatal"))

	err := sc.SetDefaultPermissionTemplate(context.Background(), "test1")
	require.Error(t, err)

	if err.Error() != "failed to set default permission template: status: 500, body: set default fatal" {
		t.Fatalf("wrong err returned: %s", err.Error())
	}
}
