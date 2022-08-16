package sonar

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/resty.v1"
	ctrl "sigs.k8s.io/controller-runtime"

	sonarClientHelper "github.com/epam/edp-sonar-operator/v2/pkg/client/helper"
)

const (
	cantUnmarshalMsg = "cant unmarshal %s"
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

func InitNewRestClient(url string, user string, password string) *Client {
	return &Client{
		resty: resty.SetHostURL(url).SetBasicAuth(user, password),
	}
}

func (sc *Client) jsonTypeRequest() *resty.Request {
	return sc.resty.R().SetHeader(contentTypeField, jsonContentType)
}

func (sc *Client) startRequest(ctx context.Context) *resty.Request {
	return sc.resty.R().SetHeaders(map[string]string{
		contentTypeField: "application/x-www-form-urlencoded",
		"Accept":         jsonContentType,
	}).SetContext(ctx)
}

func (sc *Client) checkError(response *resty.Response, err error) error {
	if err != nil {
		return errors.Wrap(err, "response error")
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
		return errors.Wrap(err, "unable check sonar health")
	}

	if err = sc.checkError(resp, err); err != nil {
		return errors.Wrap(err, "unable to check sonar health")
	}

	var systemHealthResponse SystemHealthResponse
	err = json.Unmarshal(resp.Body(), &systemHealthResponse)
	if err != nil {
		return errors.Wrapf(err, cantUnmarshalMsg, resp.Body())
	}

	// make sure that Sonar is up and running
	if systemHealthResponse.Health != "GREEN" {
		return errors.Errorf("sonar is not in green state, current state - %s; %v", systemHealthResponse.Health, systemHealthResponse)
	}

	resp, err = sc.startRequest(ctx).SetFormData(map[string]string{
		loginField:         user,
		"password":         newPassword,
		"previousPassword": oldPassword,
	}).Post("/users/change_password")

	if err = sc.checkError(resp, err); err != nil {
		return errors.Wrap(err, "unable to change password")
	}

	// so starting from SonarQube 8.9.9 they changed flow of "/api/users/change_password" endpoint
	// after successful change of password, Sonar refresh JWT token in cookie,
	// so we need to update cookie to get new token
	// https://github.com/SonarSource/sonarqube/commit/eb6741754b2b35172012bc5b30f5b0d53a61f7be#diff-be83bcff4cfc3fb4d04542ca6eea91cfe7738b7bd754d86eb366ce3e18b0aa34
	sc.resty.SetCookies(resp.Cookies())

	return nil
}

func (sc Client) Reboot() error {
	resp, err := sc.resty.R().
		Post("/system/restart")

	if err != nil {
		return errors.Wrap(err, "Failed to send reboot request to Sonar!")
	}
	if resp.IsError() {
		return errors.New(fmt.Sprintf("Sonar rebooting failed. Response - %s", resp.Status()))
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
func (sc Client) WaitForStatusIsUp(retryCount int, timeout time.Duration) error {
	var systemStatusResponse SystemStatusResponse

	sc.resty.SetRetryCount(retryCount).
		SetRetryWaitTime(timeout).
		AddRetryCondition(
			func(response *resty.Response) (bool, error) {
				if response.IsError() || !response.IsSuccess() {
					return response.IsError(), nil
				}
				err := json.Unmarshal([]byte(response.String()), &systemStatusResponse)
				if err != nil {
					return true, errors.Wrap(err, response.String())
				}
				log.Info(fmt.Sprintf("Current Sonar status - %s", systemStatusResponse.Status))
				if systemStatusResponse.Status == "UP" {
					return false, nil
				}
				return true, nil
			},
		)
	defer sc.resty.SetRetryCount(0)

	resp, err := sc.resty.R().
		Get("/system/status")
	if err != nil {
		return errors.Wrap(err, "Failed to send request for current Sonar status!")
	}
	if resp.IsError() {
		return errors.New(fmt.Sprintf("checking Sonar status failed. Response - %s", resp.Status()))
	}

	return nil
}

func (sc Client) InstallPlugins(plugins []string) error {
	installedPlugins, err := sc.GetInstalledPlugins()
	if err != nil {
		return errors.Wrap(err, "failed to get list of installed plugins")
	}

	log.Info("List of installed plugins", "plugins", installedPlugins)

	needReboot := false
	for _, plugin := range plugins {
		if sonarClientHelper.CheckPluginInstalled(installedPlugins, plugin) {
			continue
		} else {
			needReboot = true
			resp, errPost := sc.resty.R().
				SetBody(fmt.Sprintf("key=%s", plugin)).
				SetHeader(contentTypeField, "application/x-www-form-urlencoded").
				Post("/plugins/install")

			if errPost != nil {
				return errors.Wrapf(errPost, "Failed to send plugin installation request for %s", plugin)
			}
			if resp.IsError() {
				errMsg := fmt.Sprintf("Installation of plugin %s failed. Response - %s", plugin, resp.Status())
				return errors.New(errMsg)
			}
			log.Info(fmt.Sprintf("Plugin %s has been installed", plugin))
		}
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
	err = json.Unmarshal(resp.Body(), &installedPluginsResponse)
	if err != nil {
		return nil, errors.Wrapf(err, cantUnmarshalMsg, resp.Body())
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

func (sc Client) CreateQualityGate(qgName string, conditions []map[string]string) (string, error) {
	qgExist, qgId, isDefault, err := sc.checkQualityGateExist(qgName)
	if err != nil {
		return "", err
	}

	if qgExist && isDefault {
		return qgId, nil
	}

	if qgExist && !isDefault {
		err = sc.setDefaultQualityGate(qgId)
		if err != nil {
			return "", err
		}
		return qgId, nil
	}

	log.Info(fmt.Sprintf("Creating quality gate %s", qgName))
	resp, err := sc.jsonTypeRequest().
		SetQueryParams(map[string]string{nameField: qgName}).
		Post("/qualitygates/create")
	if err != nil {
		return "", errors.Wrap(err, "Failed to send request to create quality gates!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Creating quality gate %s failed. Response - %s", qgName, resp.Status())
		return "", errors.New(errMsg)
	}

	var qualityGate QualityGatesCreateResponse
	err = json.Unmarshal(resp.Body(), &qualityGate)
	if err != nil {
		return "", errors.Wrapf(err, cantUnmarshalMsg, resp.Body())
	}
	qgId = fmt.Sprintf("%v", qualityGate.ID)

	for _, item := range conditions {
		item["gateId"] = qgId
		err = sc.createCondition(item)
		if err != nil {
			return "", err
		}
	}

	err = sc.setDefaultQualityGate(qgId)
	if err != nil {
		return "", err
	}

	log.Info(fmt.Sprintf("Quality gate %s has been created and is set as default", qgName))
	return qgId, nil
}

func (sc Client) createCondition(conditionMap map[string]string) error {
	resp, err := sc.jsonTypeRequest().
		SetQueryParams(conditionMap).
		Post("/qualitygates/create_condition")
	if err != nil {
		return errors.Wrap(err, "Failed to send request to create condition in Sonar!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Creating condition %s failed. Response - %s", conditionMap["metric"], resp.Status())
		return errors.New(errMsg)
	}
	return nil
}

type QualityGatesListResponse struct {
	QualityGates []QualityGate  `json:"qualitygates"`
	Default      interface{}    `json:"default"`
	Actions      QualityActions `json:"actions"`
}

type QualityActions struct {
	Create bool `json:"create"`
}

type QualityGate struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	IsDefault bool    `json:"isDefault"`
	IsBuiltIn bool    `json:"isBuiltIn"`
	Actions   Actions `json:"actions"`
}

type Actions struct {
	Rename            bool `json:"rename"`
	SetAsDefault      bool `json:"setAsDefault"`
	Copy              bool `json:"copy"`
	AssociateProjects bool `json:"associateProjects"`
	Delete            bool `json:"delete"`
	ManageConditions  bool `json:"manageConditions"`
}

func (sc Client) checkQualityGateExist(qgName string) (exist bool, qgId string, isDefault bool, error error) {
	resp, err := sc.resty.R().
		Get("/qualitygates/list")
	if err != nil {
		return false, "", false, errors.Wrap(err, "Requesting quality gates list failed!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Listing quality gates in Sonar failed. Response code -%v", resp.StatusCode())
		return false, "", false, errors.Wrap(err, errMsg)
	}
	var qualityGatesListResponse QualityGatesListResponse
	err = json.Unmarshal(resp.Body(), &qualityGatesListResponse)
	if err != nil {
		return false, "", false, errors.Wrap(err, string(resp.Body()))
	}

	if qualityGatesListResponse.QualityGates == nil || len(qualityGatesListResponse.QualityGates) == 0 {
		return false, "", false, err
	}
	for _, qGate := range qualityGatesListResponse.QualityGates {
		if qGate.Name == qgName {
			return true, qGate.ID, qGate.IsDefault, nil
		}
	}

	return false, "", false, nil
}

func (sc Client) setDefaultQualityGate(qgId string) error {
	resp, err := sc.jsonTypeRequest().
		SetQueryParams(map[string]string{"id": qgId}).
		Post("/qualitygates/set_as_default")
	if err != nil {
		return errors.Wrap(err, "Failed to send request to set default quality gates!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Setting default quality gate %s failed. Response - %s", qgId, resp.Status())
		return errors.New(errMsg)
	}
	return nil
}

func (sc Client) UploadProfile(profileName string, profilePath string) (string, error) {
	log.Info("Attempt to uploading quality profile...", "profileName", profileName, "profilePath", profilePath)

	profileExist, profileId, isDefault, err := sc.checkProfileExist(profileName)
	if err != nil {
		return "", errors.Wrap(err, "failed to get quality profile")
	}

	if profileExist && isDefault {
		return profileId, nil
	}

	if profileExist && !isDefault {
		err = sc.setDefaultProfile("java", profileName)
		if err != nil {
			return "", err
		}
		return profileId, nil
	}

	if !sonarClientHelper.FileExists(profilePath) {
		return "", fmt.Errorf("file %s does not exist in path provided: %s", profileName, profilePath)
	}

	log.Info(fmt.Sprintf("Uploading profile %s from path %s", profileName, profilePath))

	resp, err := sc.resty.R().
		SetHeader(contentTypeField, "multipart/form-data").
		SetFile("backup", profilePath).
		Post("/qualityprofiles/restore")
	if err != nil {
		return "", errors.Wrap(err, "Failed to send upload profile request!")
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
		return false, "", false, errors.Wrap(err, "Failed to get default quality profile!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Request for quality profile failed! Response - %v", resp.StatusCode())
		return false, "", false, errors.New(errMsg)
	}

	var qualityProfilesSearchResponse QualityProfilesSearchResponse

	err = json.Unmarshal(resp.Body(), &qualityProfilesSearchResponse)
	if err != nil {
		return false, "", false, errors.Wrap(err, string(resp.Body()))
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
			"language":       language}).
		Post("/qualityprofiles/set_default")
	if err != nil {
		return errors.New("Failed to send request to set default quality profile!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Setting profile %s as default failed. Response - %s", profileName, resp.Status())
		return errors.New(errMsg)
	}
	return nil
}

func (sc Client) AddUserToGroup(groupName string, user string) error {
	log.Info(fmt.Sprintf("Start adding user %v to group %v in Sonar", user, groupName))
	resp, err := sc.jsonTypeRequest().
		SetQueryParams(map[string]string{
			nameField:  groupName,
			loginField: user}).
		Post("/user_groups/add_user")

	if err != nil {
		return errors.Wrap(err, "Failed to send requst to add user in group! ")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Adding user %s to group %s failed. Response - %s", user, groupName, resp.Status())
		return errors.New(errMsg)
	}

	log.Info(fmt.Sprintf("User %v has been added to group %v in Sonar", user, groupName))

	return nil
}

func (sc Client) AddPermissionsToUser(user string, permissions string) error {
	log.Info(fmt.Sprintf("Start adding permissions %v to user %v", permissions, user))
	resp, err := sc.jsonTypeRequest().
		SetQueryParams(map[string]string{
			loginField:   user,
			"permission": permissions}).
		Post("/permissions/add_user")
	if err != nil {
		return errors.Wrap(err, "Failed to send request to add permission to user!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Adding permission %s to user %s failed. Response - %s", permissions, user, resp.Status())
		return errors.New(errMsg)
	}

	log.Info(fmt.Sprintf("Permissions %v to user %v has been added", permissions, user))

	return nil
}

func (sc Client) AddPermissionsToGroup(groupName string, permissions string) error {
	log.Info(fmt.Sprintf("Start adding permissions %v to group %v", permissions, groupName))
	resp, err := sc.jsonTypeRequest().
		SetQueryParams(map[string]string{
			"groupName":  groupName,
			"permission": permissions}).
		Post("/permissions/add_group")
	if err != nil {
		return errors.Wrap(err, "Failed to send request to add permissions to group!")
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
			nameField:  cases.Title(language.English).String(userName)}).
		Post("/user_tokens/generate")

	if err != nil {
		return &emptyString, errors.Wrap(err, "Failed to send request for user token generation!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Generation token for user %s failed. Response - %s", userName, resp.Status())
		return nil, errors.New(errMsg)
	}
	log.Info(fmt.Sprintf("Token for user %v has been generated", userName))

	var userTokensGenerateResponse UserTokensGenerateResponse
	err = json.Unmarshal(resp.Body(), &userTokensGenerateResponse)
	if err != nil {
		return nil, errors.Wrapf(err, cantUnmarshalMsg, resp.Body())
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
			"url":     webhookUrl}).
		Post("/webhooks/create")
	if err != nil {
		return errors.Wrap(err, "Failed to send request to add webhook!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Adding webhook %s failed. Response - %s", webhookName, resp.Status())
		return errors.New(errMsg)
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
		return false, errors.Wrap(err, "Failed to send request to list all webhooks!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Failed to list webhooks on server! Response code - %v", resp.StatusCode())
		return false, errors.New(errMsg)
	}

	var raw WebhooksListResponse
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return false, errors.Wrapf(err, cantUnmarshalMsg, resp.Body())
	}

	for _, v := range raw.Webhooks {
		if v.Name == webhookName {
			return true, nil
		}
	}

	return false, nil
}

// TODO(Serhii Shydlovskyi): Current implementation works ONLY for single value setting. Requires effort to generalize it and use for several values simultaneously.
func (sc Client) ConfigureGeneralSettings(valueType string, key string, value string) error {
	generalSettingsExist, err := sc.checkGeneralSetting(key, value)
	if err != nil {
		return err
	}

	if generalSettingsExist {
		return nil
	}

	resp, err := sc.jsonTypeRequest().
		SetQueryParams(
			map[string]string{
				"key":     key,
				valueType: value}).
		Post("/settings/set")
	if err != nil {
		return errors.Wrap(err, "Failed to send request to configure general settings!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Failed to configure %s! Response code - %v", key, resp.StatusCode())
		return errors.New(errMsg)
	}
	log.Info(fmt.Sprintf("Setting %v has been set to %v", key, value))

	return nil
}

type SettingsValuesResponse struct {
	Settings []Setting `json:"settings"`
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
		return false, errors.Wrap(err, string(resp.Body()))
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
