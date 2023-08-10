package sonar

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	key            = "key"
	valueType      = "string"
	gID            = "1"
	name           = "name"
	testUrl        = "https://domain"
	user           = "user"
	password       = "pwd"
	visibility     = "all"
	tokenValue     = "token"
	nonDefaultName = "nonDefault"
	newName        = "new"
	path           = "/tmp/temp.txt"
	id             = "AU-Tpxb--iU5OvuD2FLy"
)

func createListResponse() QualityGatesListResponse {
	defaultGate := QualityGate{
		IsDefault: true,
		Name:      name,
		ID:        id,
	}
	nonDefaultGate := QualityGate{
		IsDefault: false,
		Name:      nonDefaultName,
		ID:        id,
	}

	return QualityGatesListResponse{
		QualityGates: []QualityGate{defaultGate, nonDefaultGate},
	}
}

func createFileWithData(path string) error {
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err = fp.WriteString(name); err != nil {
		return err
	}
	if err = fp.Close(); err != nil {
		return err
	}
	return nil
}

func createProfileResp(profileName string, isDefault bool) QualityProfilesSearchResponse {
	respProfile := Profiles{
		Key:       key,
		IsDefault: isDefault,
		Name:      profileName,
	}
	return QualityProfilesSearchResponse{
		Profiles: []Profiles{respProfile},
	}
}

func CreateMockResty() *resty.Client {
	restyClient := resty.New().SetBaseURL(testUrl).SetBasicAuth(user, password).SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	})
	httpmock.DeactivateAndReset()
	httpmock.ActivateNonDefault(restyClient.GetClient())
	return restyClient
}

func TestClient_Reboot_PostErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	err := client.Reboot()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
}

func TestClient_Reboot_NotFoundStatus(t *testing.T) {
	restClient := CreateMockResty()
	httpmock.RegisterResponder(http.MethodPost, "https://domain/system/restart", httpmock.NewStringResponder(http.StatusNotFound, ""))

	client := Client{resty: restClient}
	err := client.Reboot()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to reboot sonar with response")
}

func TestClient_Reboot(t *testing.T) {
	restClient := CreateMockResty()
	httpmock.RegisterResponder(http.MethodPost, "https://domain/system/restart", httpmock.NewStringResponder(http.StatusOK, ""))

	client := Client{resty: restClient}
	err := client.Reboot()
	assert.NoError(t, err)
}

func TestClient_GetInstalledPlugins_GetErr(t *testing.T) {
	restClient := CreateMockResty()

	client := Client{resty: restClient}
	plugins, err := client.GetInstalledPlugins()
	assert.Error(t, err)
	assert.Nil(t, plugins)
	assert.Contains(t, err.Error(), "no responder found")
}

func TestClient_GetInstalledPlugins_UnmarshalErr(t *testing.T) {
	restClient := CreateMockResty()
	httpmock.RegisterResponder(http.MethodGet, "https://domain/plugins/installed", httpmock.NewStringResponder(http.StatusOK, ""))

	client := Client{resty: restClient}
	plugins, err := client.GetInstalledPlugins()
	errJSON := &json.SyntaxError{}
	assert.ErrorAs(t, err, &errJSON)
	assert.Nil(t, plugins)
}

func TestClient_GetInstalledPlugins(t *testing.T) {
	restClient := CreateMockResty()
	expectedPlugin := []string{name}
	plugin := Plugin{Key: name}
	body := InstalledPluginsResponse{Plugins: []Plugin{plugin}}
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	httpmock.RegisterResponder(http.MethodGet, "https://domain/plugins/installed", httpmock.NewBytesResponder(http.StatusOK, raw))

	client := Client{resty: restClient}
	plugins, err := client.GetInstalledPlugins()
	assert.NoError(t, err)
	assert.Equal(t, expectedPlugin, plugins)
}

func TestClient_InstallPlugins_GetInstalledPluginsErr(t *testing.T) {
	restClient := CreateMockResty()
	plugins := []string{name}

	client := Client{resty: restClient}
	err := client.InstallPlugins(plugins)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
}

func TestClient_InstallPlugins_PostErr(t *testing.T) {
	restClient := CreateMockResty()
	plugins := []string{"test"}
	plugin := Plugin{Key: name}
	body := InstalledPluginsResponse{Plugins: []Plugin{plugin}}
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	httpmock.RegisterResponder(http.MethodGet, "https://domain/plugins/installed", httpmock.NewBytesResponder(http.StatusOK, raw))

	client := Client{resty: restClient}
	err = client.InstallPlugins(plugins)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Post")
}

func TestClient_InstallPlugins_RebootErr(t *testing.T) {
	restClient := CreateMockResty()
	plugins := []string{"test"}
	plugin := Plugin{Key: name}
	body := InstalledPluginsResponse{Plugins: []Plugin{plugin}}
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	httpmock.RegisterResponder(http.MethodGet, "https://domain/plugins/installed", httpmock.NewBytesResponder(http.StatusOK, raw))
	httpmock.RegisterResponder(http.MethodPost, "https://domain/plugins/install", httpmock.NewStringResponder(http.StatusOK, ""))

	client := Client{resty: restClient}
	err = client.InstallPlugins(plugins)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Post")
}

func TestClient_InstallPlugins(t *testing.T) {
	status := SystemStatusResponse{Status: "UP"}
	rawStatus, err := json.Marshal(status)
	require.NoError(t, err)

	restClient := CreateMockResty()
	plugins := []string{"test"}
	plugin := Plugin{Key: name}
	body := InstalledPluginsResponse{Plugins: []Plugin{plugin}}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/plugins/installed", httpmock.NewBytesResponder(http.StatusOK, raw))
	httpmock.RegisterResponder(http.MethodPost, "https://domain/plugins/install", httpmock.NewStringResponder(http.StatusOK, ""))
	httpmock.RegisterResponder(http.MethodPost, "https://domain/system/restart", httpmock.NewStringResponder(http.StatusOK, ""))
	httpmock.RegisterResponder(http.MethodGet, "https://domain/system/status", httpmock.NewBytesResponder(http.StatusOK, rawStatus))

	client := Client{resty: restClient}
	err = client.InstallPlugins(plugins)
	assert.NoError(t, err)
}

func TestClient_AddUserToGroup_PostErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	err := client.AddUserToGroup(name, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
}

func TestClient_AddUserToGroup_BadStatus(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	httpmock.RegisterResponder(http.MethodPost, "https://domain/user_groups/add_user?login=user&name=name", httpmock.NewStringResponder(http.StatusNotFound, ""))

	err := client.AddUserToGroup(name, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add user user")
}

func TestClient_AddUserToGroup(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	httpmock.RegisterResponder(http.MethodPost, "https://domain/user_groups/add_user?login=user&name=name", httpmock.NewStringResponder(http.StatusOK, ""))

	err := client.AddUserToGroup(name, user)
	assert.NoError(t, err)
}

func TestClient_AddPermissionsToUser_PostErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	err := client.AddPermissionToUser(name, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
}

func TestClient_AddPermissionsToUser_BadStatus(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	httpmock.RegisterResponder(http.MethodPost, "https://domain/permissions/add_user?login=name&permission=user", httpmock.NewStringResponder(http.StatusNotFound, ""))

	err := client.AddPermissionToUser(name, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Adding permission")
}

func TestClient_AddPermissionsToUser(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	httpmock.RegisterResponder(http.MethodPost, "https://domain/permissions/add_user?login=name&permission=user", httpmock.NewStringResponder(http.StatusOK, ""))

	err := client.AddPermissionToUser(name, user)
	assert.NoError(t, err)
}

func TestClient_AddPermissionsToGroup_PostErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	err := client.AddPermissionsToGroup(name, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
}

func TestClient_AddPermissionsToGroup_BadStatus(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	httpmock.RegisterResponder(http.MethodPost, "https://domain/permissions/add_group?groupName=name&permission=user", httpmock.NewStringResponder(http.StatusNotFound, ""))

	err := client.AddPermissionsToGroup(name, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Adding permission")
}

func TestClient_AddPermissionsToGroup(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	httpmock.RegisterResponder(http.MethodPost, "https://domain/permissions/add_group?groupName=name&permission=user", httpmock.NewStringResponder(http.StatusOK, ""))

	err := client.AddPermissionsToGroup(name, user)
	assert.NoError(t, err)
}

func TestClient_SetProjectsDefaultVisibility_PostErr(t *testing.T) {
	restClient := CreateMockResty()

	client := Client{resty: restClient}
	err := client.SetProjectsDefaultVisibility(visibility)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
}

func TestClient_SetProjectsDefaultVisibility_BadStatus(t *testing.T) {
	restClient := CreateMockResty()
	httpmock.RegisterResponder(http.MethodPost, "https://domain/projects/update_default_visibility", httpmock.NewStringResponder(http.StatusNotFound, ""))

	client := Client{resty: restClient}
	err := client.SetProjectsDefaultVisibility(visibility)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting project visibility failed")
}

func TestClient_SetProjectsDefaultVisibility(t *testing.T) {
	restClient := CreateMockResty()
	httpmock.RegisterResponder(http.MethodPost, "https://domain/projects/update_default_visibility", httpmock.NewStringResponder(http.StatusOK, ""))

	client := Client{resty: restClient}
	err := client.SetProjectsDefaultVisibility(visibility)
	assert.NoError(t, err)
}

func TestClient_AddWebhook_checkWebhookExistErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	err := client.AddWebhook(name, testUrl)
	assert.Error(t, err)
}

func TestClient_AddWebhook_ExistWebHook(t *testing.T) {
	restClient := CreateMockResty()
	webhook := Webhook{Name: name}
	body := WebhooksListResponse{Webhooks: []Webhook{webhook}}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/webhooks/list", httpmock.NewBytesResponder(http.StatusOK, raw))

	client := Client{resty: restClient}
	err = client.AddWebhook(name, testUrl)
	assert.NoError(t, err)
}

func TestClient_AddWebhook_PostErr(t *testing.T) {
	restClient := CreateMockResty()
	body := WebhooksListResponse{}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/webhooks/list", httpmock.NewBytesResponder(http.StatusOK, raw))

	client := Client{resty: restClient}
	err = client.AddWebhook(name, testUrl)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
}

func TestClient_AddWebhook_BadStatus(t *testing.T) {
	restClient := CreateMockResty()
	body := WebhooksListResponse{}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/webhooks/list", httpmock.NewBytesResponder(http.StatusOK, raw))
	httpmock.RegisterResponder(http.MethodPost, "https://domain/webhooks/create?name=name&url=https%3A%2F%2Fdomain", httpmock.NewStringResponder(http.StatusNotFound, ""))

	client := Client{resty: restClient}
	err = client.AddWebhook(name, testUrl)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add webhook")
}

func TestClient_AddWebhook(t *testing.T) {
	restClient := CreateMockResty()
	body := WebhooksListResponse{}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/webhooks/list", httpmock.NewBytesResponder(http.StatusOK, raw))
	httpmock.RegisterResponder(http.MethodPost, "https://domain/webhooks/create?name=name&url=https%3A%2F%2Fdomain", httpmock.NewStringResponder(http.StatusOK, ""))

	client := Client{resty: restClient}
	err = client.AddWebhook(name, testUrl)
	assert.NoError(t, err)
}

func TestClient_GenerateUserToken_PostErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	token, err := client.GenerateUserToken(user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
	assert.Empty(t, token)
}

func TestClient_GenerateUserToken_BadStatus(t *testing.T) {
	restClient := CreateMockResty()
	httpmock.RegisterResponder(http.MethodPost, "https://domain/user_tokens/generate?login=user&name=User", httpmock.NewStringResponder(http.StatusNotFound, ""))

	client := Client{resty: restClient}
	token, err := client.GenerateUserToken(user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate token")
	assert.Empty(t, token)
}

func TestClient_GenerateUserToken_UnmarshallErr(t *testing.T) {
	restClient := CreateMockResty()
	httpmock.RegisterResponder(http.MethodPost, "https://domain/user_tokens/generate?login=user&name=User", httpmock.NewStringResponder(http.StatusOK, ""))

	client := Client{resty: restClient}
	token, err := client.GenerateUserToken(user)
	errJSON := &json.SyntaxError{}
	assert.ErrorAs(t, err, &errJSON)
	assert.Empty(t, token)
}

func TestClient_GenerateUserToken(t *testing.T) {
	restClient := CreateMockResty()
	body := UserTokensGenerateResponse{Token: tokenValue}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodPost, "https://domain/user_tokens/generate?login=user&name=User", httpmock.NewBytesResponder(http.StatusOK, raw))

	client := Client{resty: restClient}
	token, err := client.GenerateUserToken(user)
	assert.NoError(t, err)
	assert.Equal(t, tokenValue, *token)
}

func TestClient_CreateQualityGate_checkQualityGateExistErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	gate, err := client.CreateQualityGate(name, true, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
	assert.Empty(t, gate)
}

func TestClient_CreateQualityGate_AlreadyDefault(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	responder, err := httpmock.NewJsonResponder(http.StatusOK, createListResponse())
	require.NoError(t, err)
	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualitygates/list", responder)
	gate, err := client.CreateQualityGate(name, true, nil)
	assert.NoError(t, err)
	assert.Equal(t, "AU-Tpxb--iU5OvuD2FLy", gate)
}

func TestClient_CreateQualityGate_NotDefault_setDefaultQualityGateErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	responder, err := httpmock.NewJsonResponder(http.StatusOK, createListResponse())
	require.NoError(t, err)
	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualitygates/list", responder)
	gate, err := client.CreateQualityGate(nonDefaultName, true, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
	assert.Equal(t, "", gate)
}

func TestClient_CreateQualityGate_NotDefault(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	responder, err := httpmock.NewJsonResponder(http.StatusOK, createListResponse())
	require.NoError(t, err)
	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualitygates/list", responder)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualitygates/set_as_default?id="+id, httpmock.NewStringResponder(http.StatusOK, ""))

	gate, err := client.CreateQualityGate(nonDefaultName, true, nil)
	assert.NoError(t, err)
	assert.Equal(t, "AU-Tpxb--iU5OvuD2FLy", gate)
}

func TestClient_CreateQualityGate_PostErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	responder, err := httpmock.NewJsonResponder(http.StatusOK, createListResponse())
	require.NoError(t, err)
	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualitygates/list", responder)

	gate, err := client.CreateQualityGate(newName, true, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
	assert.Empty(t, gate)
}

func TestClient_CreateQualityGate_BadStatus(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	responder, err := httpmock.NewJsonResponder(http.StatusOK, createListResponse())
	require.NoError(t, err)
	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualitygates/list", responder)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualitygates/create?name=new", httpmock.NewStringResponder(http.StatusNotFound, ""))

	gate, err := client.CreateQualityGate(newName, true, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create quality gate")
	assert.Empty(t, gate)
}

func TestClient_CreateQualityGate_UnmarshalErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	responder, err := httpmock.NewJsonResponder(http.StatusOK, createListResponse())
	require.NoError(t, err)
	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualitygates/list", responder)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualitygates/create?name=new", httpmock.NewStringResponder(http.StatusOK, ""))

	gate, err := client.CreateQualityGate(newName, true, nil)
	errJSON := &json.SyntaxError{}
	assert.ErrorAs(t, err, &errJSON)
	assert.Empty(t, gate)
}

func TestClient_CreateQualityGate_createConditionErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	condition := []QualityGateCondition{
		{
			ID: "",
		},
	}

	resp := QualityGatesCreateResponse{ID: gID}
	raw, err := json.Marshal(resp)
	require.NoError(t, err)

	responder, err := httpmock.NewJsonResponder(http.StatusOK, createListResponse())
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualitygates/list", responder)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualitygates/create?name=new", httpmock.NewBytesResponder(http.StatusOK, raw))

	gate, err := client.CreateQualityGate(newName, true, condition)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
	assert.Empty(t, gate)
}

func TestClient_CreateQualityGate_setDefaultQualityGateErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	resp := QualityGatesCreateResponse{ID: gID}
	raw, err := json.Marshal(resp)
	require.NoError(t, err)

	responder, err := httpmock.NewJsonResponder(http.StatusOK, createListResponse())
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualitygates/list", responder)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualitygates/create?name=new", httpmock.NewBytesResponder(http.StatusOK, raw))
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualitygates/set_as_default?id=1", httpmock.NewStringResponder(http.StatusNotFound, ""))

	gate, err := client.CreateQualityGate(newName, true, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set quality gate")
	assert.Empty(t, gate)
}

func TestClient_CreateQualityGate(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	resp := QualityGatesCreateResponse{ID: gID}
	raw, err := json.Marshal(resp)
	require.NoError(t, err)
	responder, err := httpmock.NewJsonResponder(http.StatusOK, createListResponse())
	require.NoError(t, err)
	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualitygates/list", responder)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualitygates/create?name=new", httpmock.NewBytesResponder(http.StatusOK, raw))
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualitygates/set_as_default?id=1", httpmock.NewStringResponder(http.StatusOK, ""))

	gate, err := client.CreateQualityGate(newName, true, nil)
	assert.NoError(t, err)
	assert.Equal(t, gID, gate)
}

func TestClient_UploadProfile_checkProfileExistErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	profile, err := client.UploadProfile(name, path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
	assert.Empty(t, profile)
}

func TestClient_UploadProfile_AlreadyDefault(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	respBody := createProfileResp(name, true)
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualityprofiles/search?qualityProfile=name", response)
	profile, err := client.UploadProfile(name, path)
	assert.NoError(t, err)
	assert.Equal(t, key, profile)
}

func TestClient_UploadProfile_NotDefault_setDefaultProfileErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	respBody := createProfileResp(name, false)
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualityprofiles/search?qualityProfile=name", response)
	profile, err := client.UploadProfile(name, path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send request to set default quality profile")
	assert.Empty(t, profile)
}

func TestClient_UploadProfile_NotDefault(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	respBody := createProfileResp(name, false)
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualityprofiles/search?qualityProfile=name", response)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualityprofiles/set_default?language=java&qualityProfile=name", httpmock.NewStringResponder(http.StatusOK, ""))

	profile, err := client.UploadProfile(name, path)
	assert.NoError(t, err)
	assert.Equal(t, key, profile)
}

func TestClient_UploadProfile_FileNotExists(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	respBody := createProfileResp(name, false)
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualityprofiles/search?qualityProfile=new", response)

	profile, err := client.UploadProfile(newName, path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist in path provided")
	assert.Empty(t, profile)
}

func TestClient_UploadProfile_PostErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	respBody := createProfileResp(name, false)
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	err = createFileWithData(path)
	require.NoError(t, err)
	defer func() {
		if err = os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}()

	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualityprofiles/search?qualityProfile=new", response)

	profile, err := client.UploadProfile(newName, path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
	assert.Empty(t, profile)
}

func TestClient_UploadProfile_PostBadStatus(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	respBody := createProfileResp(name, false)
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	err = createFileWithData(path)
	require.NoError(t, err)
	defer func() {
		if err = os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}()

	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualityprofiles/search?qualityProfile=new", response)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualityprofiles/restore", httpmock.NewStringResponder(http.StatusNotFound, ""))

	profile, err := client.UploadProfile(newName, path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Uploading profile")
	assert.Empty(t, profile)
}

func TestClient_UploadProfile_checkProfileExistErr2(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	respBody := createProfileResp(name, false)
	resp, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	response := resp.Once()
	require.NoError(t, err)

	err = createFileWithData(path)
	require.NoError(t, err)
	defer func() {
		if err = os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}()

	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualityprofiles/search?qualityProfile=new", response)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualityprofiles/restore", httpmock.NewStringResponder(http.StatusOK, ""))

	profile, err := client.UploadProfile(newName, path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get default quality profile")
	assert.Empty(t, profile)
}

func TestClient_UploadProfile_setDefaultProfileErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	respBody := createProfileResp(name, false)
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	err = createFileWithData(path)
	require.NoError(t, err)
	defer func() {
		if err = os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}()

	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualityprofiles/search?qualityProfile=new", response)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualityprofiles/restore", httpmock.NewStringResponder(http.StatusOK, ""))
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualityprofiles/set_default?language=java&qualityProfile=new", httpmock.NewStringResponder(http.StatusNotFound, ""))

	profile, err := client.UploadProfile(newName, path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Setting profile")
	assert.Empty(t, profile)
}

func TestClient_UploadProfile(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}

	respBody := createProfileResp(name, false)
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	err = createFileWithData(path)
	require.NoError(t, err)
	defer func() {
		if err = os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}()

	httpmock.RegisterResponder(http.MethodGet, "https://domain/qualityprofiles/search?qualityProfile=new", response)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualityprofiles/restore", httpmock.NewStringResponder(http.StatusOK, ""))
	httpmock.RegisterResponder(http.MethodPost, "https://domain/qualityprofiles/set_default?language=java&qualityProfile=new", httpmock.NewStringResponder(http.StatusOK, ""))

	profile, err := client.UploadProfile(newName, path)
	assert.NoError(t, err)
	assert.Empty(t, profile)
}

func TestClient_ConfigureGeneralSettings_checkGeneralSettingErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	err := client.ConfigureGeneralSettings([]SettingRequest{{Key: key, Value: name, ValueType: valueType}}...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no responder found")
}

func TestClient_ConfigureGeneralSettings_generalSettingsExist2(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	setting := Setting{
		Key:    key,
		Values: []string{name},
	}
	respBody := SettingsValuesResponse{Settings: []Setting{setting}}
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/settings/values", response)

	err = client.ConfigureGeneralSettings([]SettingRequest{{Key: key, Value: name, ValueType: valueType}}...)
	assert.NoError(t, err)
}

func TestClient_ConfigureGeneralSettings_generalSettingsExist(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	setting := Setting{
		Key:   key,
		Value: name,
	}
	respBody := SettingsValuesResponse{Settings: []Setting{setting}}
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/settings/values", response)

	err = client.ConfigureGeneralSettings([]SettingRequest{{Key: key, Value: name, ValueType: valueType}}...)
	assert.NoError(t, err)
}

func TestClient_ConfigureGeneralSettings_PostErr(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	setting := Setting{
		Key:   key,
		Value: newName,
	}
	respBody := SettingsValuesResponse{Settings: []Setting{setting}}
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/settings/values", response)

	err = client.ConfigureGeneralSettings([]SettingRequest{{Key: key, Value: name, ValueType: valueType}}...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send request to configure general settings")
}

func TestClient_ConfigureGeneralSettings_PostBadStatus(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	setting := Setting{
		Key:   key,
		Value: newName,
	}
	respBody := SettingsValuesResponse{Settings: []Setting{setting}}
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/settings/values", response)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/settings/set?key=key&string=name", httpmock.NewStringResponder(http.StatusNotFound, ""))

	err = client.ConfigureGeneralSettings([]SettingRequest{{Key: key, Value: name, ValueType: valueType}}...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to configure")
}

func TestClient_ConfigureGeneralSettings(t *testing.T) {
	restClient := CreateMockResty()
	client := Client{resty: restClient}
	setting := Setting{
		Key:   key,
		Value: newName,
	}
	respBody := SettingsValuesResponse{Settings: []Setting{setting}}
	response, err := httpmock.NewJsonResponder(http.StatusOK, respBody)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "https://domain/settings/values", response)
	httpmock.RegisterResponder(http.MethodPost, "https://domain/settings/set?key=key&string=name", httpmock.NewStringResponder(http.StatusOK, ""))

	err = client.ConfigureGeneralSettings([]SettingRequest{{Key: key, Value: name, ValueType: valueType}}...)
	assert.NoError(t, err)
}

func TestClient_WaitForStatusIsUp(t *testing.T) {
	sc := NewClient("", "", "")

	err := sc.WaitForStatusIsUp(1, time.Nanosecond)
	require.Error(t, err)

	if sc.resty.RetryCount > 0 {
		t.Fatal("retry count is changed")
	}
}

func TestClient_ChangePassword(t *testing.T) {
	t.Parallel()

	sc := initClient()

	systemHealthResponse := SystemHealthResponse{Health: "GREEN", Causes: []any{}, Nodes: []any{}}
	httpmock.RegisterResponder("GET", "/api/system/health", httpmock.NewJsonResponderOrPanic(http.StatusOK, systemHealthResponse))
	httpmock.RegisterResponder("POST", "/api/users/change_password", httpmock.NewStringResponder(http.StatusOK, ""))

	if err := sc.ChangePassword(context.Background(), "foo", "bar", "baz"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/api/system/health",
		httpmock.NewStringResponder(http.StatusUnauthorized, ""))

	if err := sc.ChangePassword(context.Background(),
		"foo", "bar", "baz"); !IsHTTPErrorCode(err, http.StatusUnauthorized) {
		t.Fatal("no error or wrong type")
	}

	httpmock.RegisterResponder("GET", "/api/system/health", httpmock.NewJsonResponderOrPanic(http.StatusOK, systemHealthResponse))
	httpmock.RegisterResponder("POST", "/api/users/change_password", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

	if err := sc.ChangePassword(context.Background(),
		"foo", "bar", "baz"); !IsHTTPErrorCode(err, http.StatusInternalServerError) {
		t.Fatal("no error or wrong type")
	}
}
