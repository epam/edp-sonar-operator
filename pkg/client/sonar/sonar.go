package sonar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/epam/edp-sonar-operator/pkg/helper"
)

const (
	cantUnmarshalMsg = "failed to unmarshal %s: %w"
	nameField        = "name"
	loginField       = "login"
	jsonContentType  = "application/json"
	contentTypeField = "Content-Type"
	retryCount       = 60
	timeOut          = 10
)

// SystemHealthResponse provides status of SonarQube.
// https://next.sonarqube.com/sonarqube/web_api/api/system/health
type SystemHealthResponse struct {
	// GREEN: SonarQube is fully operational
	// YELLOW: SonarQube is usable, but it needs attention in order to be fully operational
	// RED: SonarQube is not operational
	Health string `json:"health"`
	Causes []any  `json:"causes"`
	Nodes  []any  `json:"nodes"`
}

var log = ctrl.Log.WithName("sonar_client")

type Client struct {
	resty *resty.Client
}

func NewClient(url string, user string, password string) *Client {
	u := strings.TrimSuffix(url, "/")
	if !strings.HasSuffix(url, "api") {
		u = fmt.Sprintf("%s/api", u)
	}

	return &Client{
		resty: resty.New().SetBaseURL(u).SetBasicAuth(user, password),
	}
}

func (sc *Client) jsonTypeRequest() *resty.Request {
	return sc.resty.R().SetHeader(contentTypeField, jsonContentType)
}

func (sc *Client) startRequest(ctx context.Context) *resty.Request {
	return sc.resty.R().
		SetHeaders(map[string]string{
			contentTypeField: "application/x-www-form-urlencoded",
			"Accept":         jsonContentType,
		}).
		SetContext(ctx)
}

func (sc *Client) checkError(response *resty.Response, err error) error {
	if err != nil {
		return fmt.Errorf("response error: %w", err)
	}

	if response == nil {
		return errors.New("empty response")
	}

	if response.IsError() {
		return HTTPError{message: response.String(), code: response.StatusCode()}
	}

	return nil
}

func (sc *Client) ChangePassword(ctx context.Context, user string, oldPassword string, newPassword string) error {
	resp, err := sc.startRequest(ctx).Get("/system/health")
	if err != nil {
		return fmt.Errorf("failed check sonar health: %w", err)
	}

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to check sonar health: %w", err)
	}

	var systemHealthResponse SystemHealthResponse

	if err = json.Unmarshal(resp.Body(), &systemHealthResponse); err != nil {
		return fmt.Errorf(cantUnmarshalMsg, resp.Body(), err)
	}

	// make sure that Sonar is up and running
	if systemHealthResponse.Health != "GREEN" {
		return fmt.Errorf("sonar is not in green state, current state - %s; %v", systemHealthResponse.Health, systemHealthResponse)
	}

	resp, err = sc.startRequest(ctx).
		SetFormData(map[string]string{
			loginField:         user,
			"password":         newPassword,
			"previousPassword": oldPassword,
		}).
		Post("/users/change_password")

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	// so starting from SonarQube 8.9.9 they changed flow of "/api/users/change_password" endpoint
	// after successful change of password, Sonar refresh JWT token in cookie,
	// so we need to update cookie to get new token
	// https://github.com/SonarSource/sonarqube/commit/eb6741754b2b35172012bc5b30f5b0d53a61f7be#diff-be83bcff4cfc3fb4d04542ca6eea91cfe7738b7bd754d86eb366ce3e18b0aa34
	sc.resty.SetCookies(resp.Cookies())

	return nil
}

func (sc *Client) Reboot() error {
	resp, err := sc.resty.R().
		Post("/system/restart")
	if err != nil {
		return fmt.Errorf("failed to send reboot request to Sonar: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("failed to reboot sonar with response %s", resp.Status())
	}

	return nil
}

type SystemStatusResponse struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

// WaitForStatusIsUp waits for Sonar to be up
// It retries the request for the specified number of times with the specified timeout
func (sc *Client) WaitForStatusIsUp(retryCount int, timeout time.Duration) error {
	var systemStatusResponse SystemStatusResponse

	sc.resty.SetRetryCount(retryCount).
		SetRetryWaitTime(timeout).
		AddRetryCondition(
			func(response *resty.Response, err error) bool {
				if response.IsError() || !response.IsSuccess() {
					return response.IsError()
				}

				if err := json.Unmarshal([]byte(response.String()), &systemStatusResponse); err != nil {
					return true
				}

				log.Info(fmt.Sprintf("Current Sonar status - %s", systemStatusResponse.Status))

				return systemStatusResponse.Status != "UP"
			},
		)
	defer sc.resty.SetRetryCount(0)

	resp, err := sc.resty.R().
		Get("/system/status")
	if err != nil {
		return fmt.Errorf("failed to send request for current Sonar status!: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("checking Sonar status failed. Response - %s", resp.Status())
	}

	return nil
}

func (sc Client) InstallPlugins(plugins []string) error {
	installedPlugins, err := sc.GetInstalledPlugins()
	if err != nil {
		return fmt.Errorf("failed to get list of installed plugins: %w", err)
	}

	log.Info("List of installed plugins", "plugins", installedPlugins)

	needReboot := false

	for _, plugin := range plugins {
		if helper.CheckPluginInstalled(installedPlugins, plugin) {
			continue
		}

		needReboot = true

		resp, errPost := sc.resty.R().
			SetBody(fmt.Sprintf("key=%s", plugin)).
			SetHeader(contentTypeField, "application/x-www-form-urlencoded").
			Post("/plugins/install")
		if errPost != nil {
			return fmt.Errorf("failed to send plugin installation request for %s: %w", plugin, errPost)
		}

		if resp.IsError() {
			return fmt.Errorf("failed to install plugin %s, response: %s", plugin, resp.Status())
		}

		log.Info(fmt.Sprintf("Plugin %s has been installed", plugin))
	}

	if needReboot {
		if err = sc.Reboot(); err != nil {
			return err
		}

		if err = sc.WaitForStatusIsUp(retryCount, timeOut*time.Second); err != nil {
			return err
		}
	}

	log.Info("Plugins have been installed")

	return nil
}

type InstalledPluginsResponse struct {
	Plugins []Plugin `json:"plugins"`
}

type Plugin struct {
	Key                 string `json:"key"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	Version             string `json:"version"`
	License             string `json:"license"`
	OrganizationName    string `json:"organizationName"`
	OrganizationURL     string `json:"organizationUrl"`
	EditionBundled      bool   `json:"editionBundled"`
	HomepageURL         string `json:"homepageUrl"`
	IssueTrackerURL     string `json:"issueTrackerUrl"`
	ImplementationBuild string `json:"implementationBuild"`
	Filename            string `json:"filename"`
	Hash                string `json:"hash"`
	SonarLintSupported  bool   `json:"sonarLintSupported"`
	DocumentationPath   string `json:"documentationPath,omitempty"`
	UpdatedAt           int    `json:"updatedAt"`
}

func (sc Client) GetInstalledPlugins() ([]string, error) {
	resp, err := sc.resty.R().Get("/plugins/installed")
	if err = sc.checkError(resp, err); err != nil {
		return nil, err
	}

	var installedPluginsResponse InstalledPluginsResponse

	if err = json.Unmarshal(resp.Body(), &installedPluginsResponse); err != nil {
		return nil, fmt.Errorf(cantUnmarshalMsg, resp.Body(), err)
	}

	var installedPlugins []string

	for index := range installedPluginsResponse.Plugins {
		installedPlugins = append(installedPlugins, installedPluginsResponse.Plugins[index].Key)
	}

	return installedPlugins, nil
}

type QualityGatesCreateResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type QualityGatesListResponse struct {
	QualityGates []QualityGate  `json:"qualitygates"`
	Default      interface{}    `json:"default"`
	Actions      QualityActions `json:"actions"`
}

type QualityActions struct {
	Create bool `json:"create"`
}

func (sc Client) UploadProfile(profileName string, profilePath string) (string, error) {
	log.Info("attempting to upload quality profile...", "profileName", profileName, "profilePath", profilePath)

	profileExist, profileId, isDefault, err := sc.checkProfileExist(profileName)
	if err != nil {
		return "", fmt.Errorf("failed to get quality profile: %w", err)
	}

	if profileExist && isDefault {
		return profileId, nil
	}

	if profileExist && !isDefault {
		if err = sc.setDefaultProfile("java", profileName); err != nil {
			return "", err
		}

		return profileId, nil
	}

	if !helper.FileExists(profilePath) {
		return "", fmt.Errorf("file %s does not exist in path provided: %s", profileName, profilePath)
	}

	log.Info(fmt.Sprintf("Uploading profile %s from path %s", profileName, profilePath))

	resp, err := sc.resty.R().
		SetHeader(contentTypeField, "multipart/form-data").
		SetFile("backup", profilePath).
		Post("/qualityprofiles/restore")
	if err != nil {
		return "", fmt.Errorf("failed to send upload profile request!: %w", err)
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Uploading profile %s failed. Response - %s", profileName, resp.Status())
		return "", errors.New(errMsg)
	}

	_, profileId, _, err = sc.checkProfileExist(profileName)
	if err != nil {
		return "", err
	}

	err = sc.setDefaultProfile("java", profileName)
	if err != nil {
		return "", err
	}

	log.Info(fmt.Sprintf("Profile %s in Sonar from path %v has been uploaded and is set as default", profileName, profilePath))
	return profileId, nil
}

type QualityProfilesSearchResponse struct {
	Profiles []Profiles `json:"profiles"`
	Actions  Actions    `json:"actions,omitempty"`
}

type Profiles struct {
	Key                       string         `json:"key"`
	Name                      string         `json:"name"`
	Language                  string         `json:"language"`
	LanguageName              string         `json:"languageName,omitempty"`
	IsInherited               bool           `json:"isInherited,omitempty"`
	IsBuiltIn                 bool           `json:"isBuiltIn,omitempty"`
	ActiveRuleCount           int            `json:"activeRuleCount,omitempty"`
	ActiveDeprecatedRuleCount int            `json:"activeDeprecatedRuleCount,omitempty"`
	IsDefault                 bool           `json:"isDefault"`
	RuleUpdatedAt             string         `json:"ruleUpdatedAt,omitempty"`
	LastUsed                  string         `json:"lastUsed,omitempty"`
	Actions                   ProfileActions `json:"actions,omitempty"`
	ParentKey                 string         `json:"parentKey,omitempty"`
	ParentName                string         `json:"parentName,omitempty"`
	ProjectCount              int            `json:"projectCount,omitempty"`
	UserUpdatedAt             string         `json:"userUpdatedAt,omitempty"`
}

type ProfileActions struct {
	Edit              bool `json:"edit"`
	SetAsDefault      bool `json:"setAsDefault"`
	Copy              bool `json:"copy"`
	Delete            bool `json:"delete"`
	AssociateProjects bool `json:"associateProjects"`
}

func (sc Client) checkProfileExist(requiredProfileName string) (exits bool, profileId string, isDefault bool, error error) {
	resp, err := sc.resty.R().
		Get(fmt.Sprintf("/qualityprofiles/search?qualityProfile=%v", strings.ReplaceAll(requiredProfileName, " ", "+")))
	if err != nil {
		return false, "", false, fmt.Errorf("failed to get default quality profile!: %w", err)
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Request for quality profile failed! Response - %v", resp.StatusCode())
		return false, "", false, errors.New(errMsg)
	}

	var qualityProfilesSearchResponse QualityProfilesSearchResponse

	err = json.Unmarshal(resp.Body(), &qualityProfilesSearchResponse)
	if err != nil {
		return false, "", false, fmt.Errorf("%s: %w", resp.Body(), err)
	}
	if qualityProfilesSearchResponse.Profiles == nil {
		return false, "", false, nil
	}
	profiles := qualityProfilesSearchResponse.Profiles
	for index := range profiles {
		if profiles[index].Name == requiredProfileName {
			return true, profiles[index].Key, profiles[index].IsDefault, nil
		}
	}

	return false, "", false, nil
}

func (sc Client) setDefaultProfile(language string, profileName string) error {
	resp, err := sc.jsonTypeRequest().
		SetQueryParams(map[string]string{
			"qualityProfile": profileName,
			"language":       language,
		}).
		Post("/qualityprofiles/set_default")
	if err != nil {
		return errors.New("failed to send request to set default quality profile!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Setting profile %s as default failed. Response - %s", profileName, resp.Status())
		return errors.New(errMsg)
	}
	return nil
}

func (sc Client) AddPermissionsToGroup(groupName string, permissions string) error {
	log.Info(fmt.Sprintf("Start adding permissions %v to group %v", permissions, groupName))
	resp, err := sc.jsonTypeRequest().
		SetQueryParams(map[string]string{
			"groupName":  groupName,
			"permission": permissions,
		}).
		Post("/permissions/add_group")
	if err != nil {
		return fmt.Errorf("failed to send request to add permissions to group!: %w", err)
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Adding permission %s to group %s failed. Response - %s", permissions, groupName, resp.Status())
		return errors.New(errMsg)
	}

	log.Info(fmt.Sprintf("Permissions %v to group %v has been added", permissions, groupName))
	return nil
}

type UserTokensGenerateResponse struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	Token     string `json:"token"`
}

func (sc Client) GenerateUserToken(userName string) (*string, error) {
	emptyString := ""

	log.Info(fmt.Sprintf("Start generating token for user %v in Sonar", userName))

	resp, err := sc.jsonTypeRequest().
		SetQueryParams(map[string]string{
			loginField: userName,
			nameField:  cases.Title(language.English).String(userName),
		}).
		Post("/user_tokens/generate")
	if err != nil {
		return &emptyString, fmt.Errorf("failed to send request for user token generation: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to generate token for user %s with response %s", userName, resp.Status())
	}

	log.Info(fmt.Sprintf("Token for user %v has been generated", userName))

	var userTokensGenerateResponse UserTokensGenerateResponse

	if err = json.Unmarshal(resp.Body(), &userTokensGenerateResponse); err != nil {
		return nil, fmt.Errorf(cantUnmarshalMsg, resp.Body(), err)
	}

	token := userTokensGenerateResponse.Token

	return &token, nil
}

func (sc Client) AddWebhook(webhookName string, webhookUrl string) error {
	webHookExist, err := sc.checkWebhookExist(webhookName)
	if err != nil {
		return err
	}

	if webHookExist {
		return nil
	}

	log.Info(fmt.Sprintf("Start creating webhook %v in Sonar", webhookName))

	resp, err := sc.jsonTypeRequest().
		SetQueryParams(map[string]string{
			nameField: webhookName,
			"url":     webhookUrl,
		}).
		Post("/webhooks/create")
	if err != nil {
		return fmt.Errorf("failed to send request to add webhook: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("failed to add webhook %s with response %s", webhookName, resp.Status())
	}

	log.Info(fmt.Sprintf("Webhook %v has been created", webhookName))

	return nil
}

type WebhooksListResponse struct {
	Webhooks []Webhook `json:"webhooks"`
}

type Webhook struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Secret string `json:"secret,omitempty"`
}

func (sc Client) checkWebhookExist(webhookName string) (bool, error) {
	resp, err := sc.resty.R().
		Get("/webhooks/list")
	if err != nil {
		return false, fmt.Errorf("failed to send request to list all webhooks!: %w", err)
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("failed to list webhooks on server! Response code - %v", resp.StatusCode())
		return false, errors.New(errMsg)
	}

	var raw WebhooksListResponse
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return false, fmt.Errorf(cantUnmarshalMsg, resp.Body(), err)
	}

	for _, v := range raw.Webhooks {
		if v.Name == webhookName {
			return true, nil
		}
	}

	return false, nil
}

func (sc Client) ConfigureGeneralSettings(settings ...SettingRequest) error {
	for _, setting := range settings {
		if err := sc.configureGeneralSetting(setting); err != nil {
			return fmt.Errorf("failed to configure sonar sonar setting: %w", err)
		}
	}

	return nil
}

func (sc Client) configureGeneralSetting(setting SettingRequest) error {
	generalSettingsExist, err := sc.checkGeneralSetting(setting.Key, setting.Value)
	if err != nil {
		return err
	}

	if generalSettingsExist {
		return nil
	}

	resp, err := sc.jsonTypeRequest().
		SetQueryParams(
			map[string]string{
				"key":             setting.Key,
				setting.ValueType: setting.Value,
			}).
		Post("/settings/set")
	if err != nil {
		return fmt.Errorf("failed to send request to configure general settings: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("failed to configure %s: response code - %v", setting.Key, resp.StatusCode())
	}

	log.Info(fmt.Sprintf("Setting %v has been set to %v", setting.Key, setting.Value))

	return nil
}

type SettingsValuesResponse struct {
	Settings []Setting `json:"settings"`
}

type SettingRequest struct {
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"`
	ValueType string `json:"type"`
}

type Setting struct {
	Key         string              `json:"key"`
	Value       string              `json:"value,omitempty"`
	Inherited   bool                `json:"inherited"`
	Values      []string            `json:"values,omitempty"`
	FieldValues []SettingFieldValue `json:"fieldValues,omitempty"`
}

type SettingFieldValue struct {
	Boolean string `json:"boolean"`
	Text    string `json:"text"`
}

func (sc Client) checkGeneralSetting(key string, valueToCheck string) (bool, error) {
	resp, err := sc.resty.R().
		Get("/settings/values")
	if err != nil || resp.IsError() {
		return false, err
	}

	var settingsValuesResponse SettingsValuesResponse
	err = json.Unmarshal(resp.Body(), &settingsValuesResponse)
	if err != nil {
		return false, fmt.Errorf("%s: %w", resp.Body(), err)
	}

	for _, v := range settingsValuesResponse.Settings {
		if v.Key == key {
			if v.Values != nil {
				if checkValue(v.Values, valueToCheck) {
					return true, nil
				}
			} else if v.Value != "" {
				if checkValue(v.Value, valueToCheck) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func checkValue(value interface{}, valueToCheck string) bool {
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(value)
		for i := 0; i < s.Len(); i++ {
			if valueToCheck == s.Index(i).Interface() {
				return true
			}
		}
	case reflect.String:
		if valueToCheck == value {
			return true
		}
	default:
		return false
	}
	return false
}

func (sc Client) SetProjectsDefaultVisibility(visibility string) error {
	resp, err := sc.resty.R().
		SetBody(fmt.Sprintf("organization=default-organization&projectVisibility=%v", visibility)).
		SetHeader(contentTypeField, "application/x-www-form-urlencoded").
		Post("/projects/update_default_visibility")
	if err != nil {
		return err
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("setting project visibility failed. Response - %s", resp.Status())
		return errors.New(errMsg)
	}
	return nil
}

func (sc *Client) SetSetting(ctx context.Context, setting url.Values) error {
	rsp, err := sc.startRequest(ctx).
		SetFormDataFromValues(setting).
		Post("/settings/set")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to set setting: %w", err)
	}

	return nil
}
func (sc *Client) ResetSettings(ctx context.Context, settingsKeys []string) error {
	keys := strings.Join(settingsKeys, ",")
	rsp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"keys": keys,
		}).
		Post("/settings/reset")

	if err = sc.checkError(rsp, err); err != nil {
		return fmt.Errorf("failed to reset settings %s: %w", keys, err)
	}

	return nil
}
