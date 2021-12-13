package sonar

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	sonarClientHelper "github.com/epam/edp-sonar-operator/v2/pkg/client/helper"
	"github.com/pkg/errors"
	"github.com/totherme/unstructured"
	"gopkg.in/resty.v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("sonar_client")

type Client struct {
	resty *resty.Client
}

func InitNewRestClient(url string, user string, password string) *Client {
	return &Client{
		resty: resty.SetHostURL(url).SetBasicAuth(user, password),
	}
}

func (sc *Client) startRequest(ctx context.Context) *resty.Request {
	return sc.resty.R().SetHeaders(map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"Accept":       "application/json",
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

	if err := sc.checkError(resp, err); err != nil {
		return errors.Wrap(err, "unable to check sonar health")
	}

	resp, err = sc.startRequest(ctx).SetFormData(map[string]string{
		"login":            user,
		"password":         newPassword,
		"previousPassword": oldPassword,
	}).Post("/users/change_password")

	if err := sc.checkError(resp, err); err != nil {
		return errors.Wrap(err, "unable to change password")
	}

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

func (sc Client) WaitForStatusIsUp(retryCount int, timeout time.Duration) error {
	var raw map[string]interface{}

	sc.resty.SetRetryCount(retryCount).
		SetRetryWaitTime(timeout * time.Second).
		AddRetryCondition(
			func(response *resty.Response) (bool, error) {
				if response.IsError() || !response.IsSuccess() {
					return response.IsError(), nil
				}
				err := json.Unmarshal([]byte(response.String()), &raw)
				if err != nil {
					return true, err
				}
				log.Info(fmt.Sprintf("Current Sonar status - %s", raw["status"].(string)))
				if raw["status"].(string) == "UP" {
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
		return err
	}

	needReboot := false
	for _, plugin := range plugins {
		if !sonarClientHelper.CheckPluginInstalled(installedPlugins, plugin) {
			needReboot = true
			resp, err := sc.resty.R().
				SetBody(fmt.Sprintf("key=%s", plugin)).
				SetHeader("Content-Type", "application/x-www-form-urlencoded").
				Post("/plugins/install")

			if err != nil {
				return errors.Wrapf(err, "Failed to send plugin installation request for %s", plugin)
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
		if err = sc.WaitForStatusIsUp(60, 10); err != nil {
			return err
		}
	}
	return nil
}

func (sc Client) GetInstalledPlugins() ([]string, error) {
	resp, err := sc.resty.R().Get("/plugins/installed")
	if err != nil || resp.IsError() {
		return nil, err
	}

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return nil, errors.Wrapf(err, "cant unmarshal %s", resp.Body())
	}

	var installedPlugins []string
	for _, v := range raw["plugins"] {
		installedPlugins = append(installedPlugins, fmt.Sprintf("%v", v["key"]))
	}

	return installedPlugins, nil
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
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{"name": qgName}).
		Post("/qualitygates/create")
	if err != nil {
		return "", errors.Wrap(err, "Failed to send request to create quality gates!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Creating quality gate %s failed. Response - %s", qgName, resp.Status())
		return "", errors.New(errMsg)
	}

	var raw map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return "", errors.Wrapf(err, "cant unmarshal %s", resp.Body())
	}
	qgId = fmt.Sprintf("%v", raw["id"])

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
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
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

	responseJson, err := unstructured.ParseJSON(string(resp.Body()))
	if err != nil {
		return false, "", false, err
	}

	if ok := responseJson.HasKey("qualitygates"); ok {
		qualityGates, _ := responseJson.GetByPointer("/qualitygates")
		qualityGatesList, err := qualityGates.ListValue()
		if err != nil || len(qualityGatesList) == 0 {
			return false, "", false, err
		}
		for _, qualityGate := range qualityGatesList {
			currentQualityGateName, _ := qualityGate.F("name").StringValue()
			if currentQualityGateName == qgName {
				qualityGateId, _ := qualityGate.F("id").NumValue()
				if ok, _ := qualityGate.F("isDefault").BoolValue(); ok {
					return true, fmt.Sprintf("%v", qualityGateId), ok, nil
				}
				return true, fmt.Sprintf("%v", qualityGateId), ok, nil
			}
		}
	}

	return false, "", false, nil
}

func (sc Client) setDefaultQualityGate(qgId string) error {
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
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
	profileExist, profileId, isDefault, err := sc.checkProfileExist(profileName)
	if err != nil {
		return "", err
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
		return "", fmt.Errorf("File %s does not exist in path provided: %s !", profileName, profilePath)
	}

	log.Info(fmt.Sprintf("Uploading profile %s from path %s", profileName, profilePath))
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "multipart/form-data").
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

func (sc Client) checkProfileExist(requiredProfileName string) (exits bool, profileId string, isDefault bool, error error) {
	resp, err := sc.resty.R().
		Get(fmt.Sprintf("/qualityprofiles/search?qualityProfile=%v", strings.Replace(requiredProfileName, " ", "+", -1)))
	if err != nil {
		return false, "", false, errors.Wrap(err, "Failed to get default quality profile!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Request for quality profile failed! Response - %v", resp.StatusCode())
		return false, "", false, errors.New(errMsg)
	}

	responseJson, err := unstructured.ParseJSON(string(resp.Body()))
	if err != nil {
		return false, "", false, err
	}

	if ok := responseJson.HasKey("profiles"); ok {
		profilesData, _ := responseJson.GetByPointer("/profiles")
		profileList, err := profilesData.ListValue()
		if err != nil || len(profileList) == 0 {
			return false, "", false, nil
		}
		for _, profile := range profileList {
			currentProfileName, _ := profile.F("name").StringValue()
			if currentProfileName == requiredProfileName {
				profileId, _ := profile.F("key").StringValue()
				if ok, _ := profile.F("isDefault").BoolValue(); ok {
					return true, profileId, ok, nil
				}
				return true, profileId, false, nil
			}
		}
	}

	return false, "", false, nil
}

func (sc Client) setDefaultProfile(language string, profileName string) error {
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
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
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"name":  groupName,
			"login": user}).
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
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"login":      user,
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
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
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

func (sc Client) GenerateUserToken(userName string) (*string, error) {
	emptyString := ""

	log.Info(fmt.Sprintf("Start generating token for user %v in Sonar", userName))
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"login": userName,
			"name":  strings.Title(userName)}).
		Post("/user_tokens/generate")

	if err != nil {
		return &emptyString, errors.Wrap(err, "Failed to send request for user token generation!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Generation token for user %s failed. Response - %s", userName, resp.Status())
		return nil, errors.New(errMsg)
	}
	log.Info(fmt.Sprintf("Token for user %v has been generated", userName))

	var rawResponse map[string]string
	err = json.Unmarshal(resp.Body(), &rawResponse)
	if err != nil {
		return nil, errors.Wrapf(err, "cant unmarshal %s", resp.Body())
	}
	token := rawResponse["token"]

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
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"name": webhookName,
			"url":  webhookUrl}).
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

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return false, errors.Wrapf(err, "cant unmarshal %s", resp.Body())
	}

	for _, v := range raw["webhooks"] {
		if v["name"] == webhookName {
			return true, nil
		}
	}

	return false, nil
}

//TODO(Serhii Shydlovskyi): Current implementation works ONLY for single value setting. Requires effort to generalize it and use for several values simultaneously.
func (sc Client) ConfigureGeneralSettings(valueType string, key string, value string) error {
	generalSettingsExist, err := sc.checkGeneralSetting(key, value)
	if err != nil {
		return err
	}

	if generalSettingsExist {
		return nil
	}

	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
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

func (sc Client) checkGeneralSetting(key string, valueToCheck string) (bool, error) {
	resp, err := sc.resty.R().
		Get("/settings/values")
	if err != nil || resp.IsError() {
		return false, err
	}

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return false, errors.Wrap(err, "Failed to unmarshal response body!")
	}

	for _, v := range raw["settings"] {
		if v["key"] == key {
			if value, exists := v["value"]; exists {
				if checkValue(value, valueToCheck) {
					return true, nil
				}
			} else if value, exists := v["values"]; exists {
				if checkValue(value, valueToCheck) {
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
	}
	return false
}

func (sc Client) SetProjectsDefaultVisibility(visibility string) error {
	resp, err := sc.resty.R().
		SetBody(fmt.Sprintf("organization=default-organization&projectVisibility=%v", visibility)).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post("/projects/update_default_visibility")
	if err != nil || resp.IsError() {
		return err
	}
	return nil
}
