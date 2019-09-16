package sonar

import (
	"encoding/json"
	"errors"
	"fmt"
	sonarClientHelper "github.com/epmd-edp/sonar-operator/v2/pkg/client/helper"
	errorsf "github.com/pkg/errors"
	"github.com/totherme/unstructured"
	"gopkg.in/resty.v1"
	"reflect"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
	"time"
)

var log = logf.Log.WithName("sonar_client")

type SonarClient struct {
	resty  resty.Client
	ApiUrl string
}

func (sc *SonarClient) InitNewRestClient(url string, user string, password string) error {
	sc.resty = *resty.SetHostURL(url).SetBasicAuth(user, password)
	sc.ApiUrl = url
	return nil
}

func (sc *SonarClient) ChangePassword(user string, oldPassword string, newPassword string) error {
	resp, err := sc.resty.R().Get("/system/health")
	if err == nil && resp.IsError() {
		return nil
	}

	resp, err = sc.resty.R().
		SetBody(fmt.Sprintf("login=%v&password=%v&previousPassword=%v", user, newPassword, oldPassword)).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post("/users/change_password")

	if err != nil {
		return errorsf.Wrap(err, "Failed to send request for password change in Sonar!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Password changing for user %s unsuccessful. Response - %s", user, resp.Status())
		return errors.New(errMsg)
	}

	log.Info(fmt.Sprintf("Password for user %v changed successfully", user))

	return nil
}

func (sc SonarClient) Reboot() error {
	resp, err := sc.resty.R().
		Post("/system/restart")

	if err != nil {
		return errorsf.Wrap(err, "Failed to send reboot request to Sonar!")
	}
	if resp.IsError() {
		return errors.New(fmt.Sprintf("Sonar rebooting failed. Response - %s", resp.Status()))
	}

	return nil
}

func (sc SonarClient) WaitForStatusIsUp(retryCount int, timeout time.Duration) error {
	var raw map[string]interface{}
	resp, err := sc.resty.
		SetRetryCount(retryCount).
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
		).
		R().
		Get("/system/status")
	if err != nil {
		return errorsf.Wrap(err, "Failed to send request for current Sonar status!")
	}
	if resp.IsError() {
		return errors.New(fmt.Sprintf("Checking Sonar status failed. Response - %s", resp.Status()))
	}

	return nil
}

func (sc SonarClient) InstallPlugins(plugins []string) error {
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
				return errorsf.Wrapf(err, "Failed to send plugin installation request for %s", plugin)
			}
			if resp.IsError() {
				errMsg := fmt.Sprintf("Installation of plugin %s failed. Response - %s", plugin, resp.Status())
				return errors.New(errMsg)
			}
			log.Info(fmt.Sprintf("Plugin %s has been installed", plugin))
		}
	}
	if needReboot {
		sc.Reboot()
		sc.WaitForStatusIsUp(60, 10)
	}
	return nil
}

func (sc SonarClient) GetInstalledPlugins() ([]string, error) {
	resp, err := sc.resty.R().Get("/plugins/installed")
	if err != nil || resp.IsError() {
		return nil, err
	}

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)

	var installedPlugins []string
	for _, v := range raw["plugins"] {
		installedPlugins = append(installedPlugins, fmt.Sprintf("%v", v["key"]))
	}

	return installedPlugins, nil
}

func (sc SonarClient) CreateQualityGate(qgName string, conditions []map[string]string) (*string, error) {
	emptyString := ""
	qgExist, qgId, isDefault, err := sc.checkQualityGateExist(qgName)
	if err != nil {
		return nil, err
	}

	if qgExist && isDefault {
		return &qgId, nil
	}

	if qgExist && !isDefault {
		err = sc.setDefaultQualityGate(qgId)
		if err != nil {
			return nil, err
		}
		return &qgId, nil
	}

	log.Info(fmt.Sprintf("Creating quality gate %s", qgName))
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{"name": qgName}).
		Post("/qualitygates/create")
	if err != nil {
		return &emptyString, errorsf.Wrap(err, "Failed to send request to create quality gates!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Creating quality gate %s failed. Response - %s", qgName, resp.Status())
		return nil, errors.New(errMsg)
	}

	var raw map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)
	qgId = fmt.Sprintf("%v", raw["id"])

	for _, item := range conditions {
		item["gateId"] = qgId
		sc.createCondition(item)
		if err != nil {
			return nil, err
		}
	}

	err = sc.setDefaultQualityGate(qgId)
	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("Quality gate %s has been created and is set as default", qgName))
	return &qgId, nil
}

func (sc SonarClient) createCondition(conditionMap map[string]string) error {
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(conditionMap).
		Post("/qualitygates/create_condition")
	if err != nil {
		return errorsf.Wrap(err, "Failed to send request to create condition in Sonar!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Creating condition %s failed. Response - %s", conditionMap["metric"], resp.Status())
		return errors.New(errMsg)
	}
	return nil
}

func (sc SonarClient) checkQualityGateExist(qgName string) (exist bool, qgId string, isDefault bool, error error) {
	resp, err := sc.resty.R().
		Get("/qualitygates/list")
	if err != nil {
		return false, "", false, errorsf.Wrap(err, "Requesting quality gates list failed!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Listing quality gates in Sonar failed. Response code -%v", resp.StatusCode())
		return false, "", false, errorsf.Wrap(err, errMsg)
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

func (sc SonarClient) setDefaultQualityGate(qgId string) error {
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{"id": qgId}).
		Post("/qualitygates/set_as_default")
	if err != nil {
		return errorsf.Wrap(err, "Failed to send request to set default quality gates!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Setting default quality gate %s failed. Response - %s", qgId, resp.Status())
		return errors.New(errMsg)
	}
	return nil
}

func (sc SonarClient) UploadProfile(profileName string, profilePath string) (*string, error) {
	emptyString := ""
	profileExist, profileId, isDefault, err := sc.checkProfileExist(profileName)
	if err != nil {
		return nil, err
	}

	if profileExist && isDefault {
		return &profileId, nil
	}

	if profileExist && !isDefault {
		err = sc.setDefaultProfile("java", profileName)
		if err != nil {
			return nil, err
		}
		return &profileId, nil
	}

	if !sonarClientHelper.FileExists(profilePath) {
		return &emptyString, errors.New(fmt.Sprintf("File %s does not exist in path provided: %s !", profileName, profilePath))
	}

	log.Info(fmt.Sprintf("Uploading profile %s from path %s", profileName, profilePath))
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "multipart/form-data").
		SetFile("backup", profilePath).
		Post("/qualityprofiles/restore")
	if err != nil {
		return &emptyString, errorsf.Wrap(err, "Failed to send upload profile request!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Uploading profile %s failed. Response - %s", profileName, resp.Status())
		return &emptyString, errors.New(errMsg)
	}

	_, profileId, isDefault, err = sc.checkProfileExist(profileName)
	if err != nil {
		return nil, err
	}

	err = sc.setDefaultProfile("java", profileName)
	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("Profile %s in Sonar from path %v has been uploaded and is set as default", profileName, profilePath))
	return &profileId, nil
}

func (sc SonarClient) checkProfileExist(requiredProfileName string) (exits bool, profileId string, isDefault bool, error error) {
	resp, err := sc.resty.R().
		Get(fmt.Sprintf("/qualityprofiles/search?qualityProfile=%v", strings.Replace(requiredProfileName, " ", "+", -1)))
	if err != nil {
		return false, "", false, errorsf.Wrap(err, "Failed to get default quality profile!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Request for quality profile failed! Response - %v", resp.StatusCode())
		return false, "", false, errorsf.New(errMsg)
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

func (sc SonarClient) setDefaultProfile(language string, profileName string) error {
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"qualityProfile": profileName,
			"language":       language}).
		Post("/qualityprofiles/set_default")
	if err != nil {
		return errorsf.New("Failed to send request to set default quality profile!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Setting profile %s as default failed. Response - %s", profileName, resp.Status())
		return errors.New(errMsg)
	}
	return nil
}

func (sc *SonarClient) CreateUser(login string, name string, password string) error {
	resp, err := sc.resty.R().
		Get("/users/search?q=" + login)

	if err != nil {
		return err
	}

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)

	for _, v := range raw["users"] {
		if v["login"] == login {
			return nil
		}
	}

	resp, err = sc.resty.R().
		SetBody("login="+login+"&name="+name+"&password="+password).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post("/users/create")

	if err != nil {
		return errorsf.Wrap(err, "Failed to send user creation request to Sonar!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Failed to create user %s. Response code: %v", login, resp.StatusCode())
		return errors.New(errMsg)
	}

	log.Info(fmt.Sprintf("User %s has been created", login))
	return nil
}

func (sc SonarClient) CreateGroup(groupName string) error {
	groupExist, err := sc.checkGroupExist(groupName)
	if err != nil {
		return nil
	}

	if groupExist {
		return nil
	}

	log.Info(fmt.Sprintf("Start creating group %v in Sonar", groupName))
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"name": groupName}).
		Post("/user_groups/create")
	if err != nil {
		return errorsf.Wrap(err, "Failed to send group creation request!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Creating group %s failed. Err - %v. Response - %s", groupName, err, resp.Status())
		return errors.New(errMsg)
	}
	log.Info(fmt.Sprintf("Group %v in Sonar has been created", groupName))

	return nil
}

func (sc SonarClient) checkGroupExist(groupName string) (bool, error) {
	resp, err := sc.resty.R().
		Get(fmt.Sprintf("/user_groups/search?q=%v&f=name", groupName))
	if err != nil {
		return false, errorsf.Wrap(err, "Failed to send request to check group existence!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Checking group existence failed! Response - %v", resp.StatusCode())
		return false, errorsf.New(errMsg)
	}

	responseJson, err := unstructured.ParseJSON(string(resp.Body()))
	if err != nil {
		return false, err
	}

	if ok := responseJson.HasKey("groups"); ok {
		groups, _ := responseJson.GetByPointer("/groups")
		groupsList, err := groups.ListValue()
		if err != nil || len(groupsList) == 0 {
			return false, err
		}
		for _, group := range groupsList {
			currentGroupName, _ := group.F("name").StringValue()
			if currentGroupName == groupName {
				return true, nil
			}
			return false, nil
		}
	}

	return false, nil
}

func (sc SonarClient) AddUserToGroup(groupName string, user string) error {
	log.Info(fmt.Sprintf("Start adding user %v to group %v in Sonar", user, groupName))
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"name":  groupName,
			"login": user}).
		Post("/user_groups/add_user")

	if err != nil {
		return errorsf.Wrap(err, "Failed to send requst to add user in group! ")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Adding user %s to group %s failed. Response - %s", user, groupName, resp.Status())
		return errors.New(errMsg)
	}

	log.Info(fmt.Sprintf("User %v has been added to group %v in Sonar", user, groupName))

	return nil
}

func (sc SonarClient) AddPermissionsToUser(user string, permissions string) error {
	log.Info(fmt.Sprintf("Start adding permissions %v to user %v", permissions, user))
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"login":      user,
			"permission": permissions}).
		Post("/permissions/add_user")
	if err != nil {
		return errorsf.Wrap(err, "Failed to send request to add permission to user!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Adding permission %s to user %s failed. Response - %s", permissions, user, resp.Status())
		return errors.New(errMsg)
	}

	log.Info(fmt.Sprintf("Permissions %v to user %v has been added", permissions, user))

	return nil
}

func (sc SonarClient) AddPermissionsToGroup(groupName string, permissions string) error {
	log.Info(fmt.Sprintf("Start adding permissions %v to group %v", permissions, groupName))
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"groupName":  groupName,
			"permission": permissions}).
		Post("/permissions/add_group")
	if err != nil {
		return errorsf.Wrap(err, "Failed to send request to add permissions to group!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Adding permission %s to group %s failed. Response - %s", permissions, groupName, resp.Status())
		return errors.New(errMsg)
	}

	log.Info(fmt.Sprintf("Permissions %v to group %v has been added", permissions, groupName))
	return nil
}

func (sc SonarClient) GenerateUserToken(userName string) (*string, error) {
	emptyString := ""
	tokenExist, err := sc.checkUserTokenExist(userName)
	if err != nil {
		return nil, err
	}

	if tokenExist {
		return nil, nil
	}

	log.Info(fmt.Sprintf("Start generating token for user %v in Sonar", userName))
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"login": userName,
			"name":  strings.Title(userName)}).
		Post("/user_tokens/generate")

	if err != nil {
		return &emptyString, errorsf.Wrap(err, "Failed to send request for user token generation!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Generation token for user %s failed. Response - %s", userName, resp.Status())
		return nil, errors.New(errMsg)
	}
	log.Info(fmt.Sprintf("Token for user %v has been generated", userName))

	var rawResponse map[string]string
	err = json.Unmarshal(resp.Body(), &rawResponse)
	token := rawResponse["token"]

	return &token, nil
}

func (sc SonarClient) checkUserTokenExist(userName string) (bool, error) {
	resp, err := sc.resty.R().
		Get(fmt.Sprintf("/user_tokens/search?login=%v", userName))
	if err != nil {
		return false, errorsf.Wrap(err, "Failed to send request to check user token existence!")
	}

	if resp.IsError() {
		errMsg := fmt.Sprintf("Failed to check user token existence for %s! Response code - %v", userName, resp.StatusCode())
		return false, errorsf.Wrap(err, errMsg)
	}

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)

	if len(raw["userTokens"]) == 0 {
		return false, nil
	}

	return true, nil
}

func (sc SonarClient) AddWebhook(webhookName string, webhookUrl string) error {
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
		return errorsf.Wrap(err, "Failed to send request to add webhook!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Adding webhook %s failed. Response - %s", webhookName, resp.Status())
		return errors.New(errMsg)
	}
	log.Info(fmt.Sprintf("Webhook %v has been created", webhookName))

	return nil
}

func (sc SonarClient) checkWebhookExist(webhookName string) (bool, error) {
	resp, err := sc.resty.R().
		Get("/webhooks/list")
	if err != nil {
		return false, errorsf.Wrap(err, "Failed to send request to list all webhooks!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Failed to list webhooks on server! Response code - %v", resp.StatusCode())
		return false, errorsf.New(errMsg)
	}

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)

	for _, v := range raw["webhooks"] {
		if v["name"] == webhookName {
			return true, nil
		}
	}

	return false, nil
}

//TODO(Serhii Shydlovskyi): Current implementation works ONLY for single value setting. Requires effort to generalize it and use for several values simultaneously.
func (sc SonarClient) ConfigureGeneralSettings(valueType string, key string, value string) error {
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
		//errMsg := fmt.Sprintf("Failed to configure %s")
		return errorsf.Wrap(err, "Failed to send request to configure general settings!")
	}
	if resp.IsError() {
		errMsg := fmt.Sprintf("Failed to configure %s! Response code - %v", key, resp.StatusCode())
		return errorsf.New(errMsg)
	}
	log.Info(fmt.Sprintf("Setting %v has been set to %v", key, value))

	return nil
}

func (sc SonarClient) checkGeneralSetting(key string, valueToCheck string) (bool, error) {
	resp, err := sc.resty.R().
		Get("/settings/values")
	if err != nil || resp.IsError() {
		return false, err
	}

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)
	if err != nil {
		return false, errorsf.Wrap(err, "Failed to unmarshal response body!")
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
