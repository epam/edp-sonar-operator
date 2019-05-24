package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/resty.v1"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func logErrorAndReturn(err error) error {
	log.Printf("[ERROR] %v", err)
	return err
}

func checkPluginInstalled(pluginsList []string, plugin string) bool {
	for _, value := range pluginsList {
		if value == plugin {
			return true
		}
	}
	return false
}

type SonarClient struct {
	resty resty.Client
}

func (sc *SonarClient) InitNewRestClient(url string, user string, password string) error {
	sc.resty = *resty.SetHostURL(url).SetBasicAuth(user, password)
	return nil
}

func (sc *SonarClient) ChangePassword(user string, oldPassword string, newPassword string) error {
	resp, err := sc.resty.R().Get("/system/status")
	if err == nil && resp.IsError() {
		return nil
	}

	resp, err = sc.resty.R().
		SetBody(fmt.Sprintf("login=%v&password=%v&previousPassword=%v", user, newPassword, oldPassword)).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post("/users/change_password")

	if err != nil || resp.IsError() {
		return logErrorAndReturn(errors.New(fmt.Sprintf("[ERROR] Password changing for user %s unsuccessful. Err - %s. Response - %s", user, err, resp.Status())))
	}

	log.Printf("Password for user %v changed successfully", user)

	return nil
}

func (sc SonarClient) Reboot() error {
	resp, err := sc.resty.R().
		Post("/system/restart")

	if err != nil {
		return logErrorAndReturn(errors.New(fmt.Sprintf("[ERROR] Sonar rebooting failed. Err - %s. Response - %s", err, resp.Status())))
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
				json.Unmarshal([]byte(response.String()), &raw)
				log.Printf("Current Sonar status - %s", raw["status"].(string))
				if raw["status"].(string) == "UP" {
					return false, nil
				}
				return true, nil
			},
		).
		R().
		Get("/system/status")
	if err != nil || resp.IsError() {
		return logErrorAndReturn(errors.New(fmt.Sprintf("[ERROR] Checking Sonar status failed. Err - %s. Response - %s", err, resp.Status())))
	}
	return nil
}

func (sc SonarClient) InstallPlugins(plugins []string) error {
	installedPlugins, err := sc.GetInstalledPlugins()
	if err != nil {
		return logErrorAndReturn(err)
	}

	needReboot := false
	for _, plugin := range plugins {
		if !checkPluginInstalled(installedPlugins, plugin) {
			needReboot = true
			resp, err := sc.resty.R().
				SetBody(fmt.Sprintf("key=%s", plugin)).
				SetHeader("Content-Type", "application/x-www-form-urlencoded").
				Post("/plugins/install")

			if err != nil || resp.IsError() {
				return logErrorAndReturn(errors.New(fmt.Sprintf("[ERROR] Installation of plugin %s failed - %s. Err - %s. Response - %s", plugin, err, resp.Status())))
			}
			log.Printf("Plugin %s has been installed", plugin)
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
	qgExist, qgId, isDefault, err := sc.checkQualityGateExist(qgName)
	if err != nil {
		return nil, logErrorAndReturn(err)
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

	log.Printf("Creating quality gate %s", qgName)
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{"name": qgName}).
		Post("/qualitygates/create")
	if err != nil || resp.IsError() {
		return nil, logErrorAndReturn(errors.New(fmt.Sprintf("Creating quality gate %s failed. Err - %s. Response - %s", qgName, err, resp.Status())))
	}

	var raw map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)
	qgId = fmt.Sprintf("%v", raw["id"])

	for _, item := range conditions {
		item["gateId"] = qgId
		sc.createCondition(item)
		if err != nil {
			return nil, logErrorAndReturn(err)
		}
	}

	err = sc.setDefaultQualityGate(qgId)
	if err != nil {
		return nil, err
	}

	log.Printf("Quality gate %s has been created and is set as default", qgName)
	return &qgId, nil
}

func (sc SonarClient) createCondition(conditionMap map[string]string) error {
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(conditionMap).
		Post("/qualitygates/create_condition")
	if err != nil || resp.IsError() {
		return logErrorAndReturn(errors.New(fmt.Sprintf("[ERROR] Creating condition %s failed. Err - %s. Response - %s", conditionMap["metric"], err, resp.Status())))
	}
	return nil
}

func (sc SonarClient) checkQualityGateExist(qgName string) (exist bool, qgId string, isDefault bool, error error) {
	resp, err := sc.resty.R().
		Get("/qualitygates/list")
	if err != nil || resp.IsError() {
		return false, "", false, err
	}

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)

	for _, v := range raw["qualitygates"] {
		if v["name"] == qgName {
			isDefault, _ = strconv.ParseBool(fmt.Sprintf("%v", v["isDefault"]))
			return true, fmt.Sprintf("%v", v["id"]), isDefault, nil
		}
	}
	return false, "", false, nil
}

func (sc SonarClient) setDefaultQualityGate(qgId string) error {
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{"id": qgId}).
		Post("/qualitygates/set_as_default")
	if err != nil || resp.IsError() {
		return logErrorAndReturn(errors.New(fmt.Sprintf("Setting default quality gate %s failed. Err - %s. Response - %s", qgId, err, resp.Status())))
	}
	return nil
}

func (sc SonarClient) UploadProfile(profileName string, profilePath string) (*string, error) {
	profileExist, profileId, isDefault, err := sc.checkProfileExist(profileName)
	if err != nil {
		return nil, logErrorAndReturn(err)
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

	log.Printf("Uploading profile %s from path %s", profileName, profilePath)
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "multipart/form-data").
		SetFile("backup", profilePath).
		Post("/qualityprofiles/restore")
	if err != nil || resp.IsError() {
		return nil, logErrorAndReturn(errors.New(fmt.Sprintf("Uploading profile %s failed. Err - %s. Response - %s", profileName, err, resp.Status())))
	}

	_, profileId, isDefault, err = sc.checkProfileExist(profileName)
	if err != nil {
		return nil, err
	}

	err = sc.setDefaultProfile("java", profileName)
	if err != nil {
		return nil, err
	}

	log.Printf("Profile %s in Sonar from path %v has been uploaded and is set as default", profileName, profilePath)
	return &profileId, nil
}

func (sc SonarClient) checkProfileExist(profileName string) (exits bool, profileId string, isDefault bool, error error) {
	resp, err := sc.resty.R().
		Get(fmt.Sprintf("/qualityprofiles/search?qualityProfile=%v", strings.Replace(profileName, " ", "+", -1)))
	if err != nil || resp.IsError() {
		return false, "", false, err
	}

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)

	if len(raw["profiles"]) > 0 {
		isDefault, _ = strconv.ParseBool(fmt.Sprintf("%v", raw["profiles"][0]["isDefault"]))
		return true, fmt.Sprintf("%v", raw["profiles"][0]["key"]), isDefault, nil
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
	if err != nil || resp.IsError() {
		return logErrorAndReturn(errors.New(fmt.Sprintf("Setting profile %s as default failed. Err - %s. Response - %s", profileName, err, resp.Status())))
	}
	return nil
}

func (sc *SonarClient) CreateUser(login string, name string, password string) error {
	resp, err := sc.resty.R().
		Get("/users/search?q=" + login)

	if err != nil {
		logErrorAndReturn(err)
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

	if err != nil || resp.IsError() {
		logErrorAndReturn(errors.New(fmt.Sprintf("Create user %s unsuccessful\nError: %v\nResponse code: %v",
			login, err, resp.StatusCode())))
	}

	log.Printf("User %s in Sonar has been created", login)
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

	log.Printf("Start creating group %v in Sonar", groupName)
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"name": groupName}).
		Post("/user_groups/create")
	if err != nil || resp.IsError() {
		return logErrorAndReturn(errors.New(fmt.Sprintf("Creating group %s failed. Err - %s. Response - %s", groupName, err, resp.Status())))
	}
	log.Printf("Group %v in Sonar has been created", groupName)

	return nil
}

func (sc SonarClient) checkGroupExist(groupName string) (bool, error) {
	resp, err := sc.resty.R().
		Get(fmt.Sprintf("/user_groups/search?q=%v&f=name", groupName))
	if err != nil || resp.IsError() {
		return false, err
	}

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)

	for _, v := range raw["groups"] {
		if v["name"] == groupName {
			return true, nil
		}
	}

	return false, nil
}

func (sc SonarClient) AddUserToGroup(groupName string, user string) error {
	log.Printf("Start adding user %v to group %v in Sonar", user, groupName)
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"name":  groupName,
			"login": user}).
		Post("/user_groups/add_user")

	if err != nil || resp.IsError() {
		return logErrorAndReturn(errors.New(fmt.Sprintf("Adding user %s to group %s failed. Err - %s. Response - %s", user, groupName, err, resp.Status())))
	}

	log.Printf("User %v has been added to group %v in Sonar", user, groupName)

	return nil
}

func (sc SonarClient) AddPermissionsToUser(user string, permissions string) error {
	log.Printf("Start adding permissions %v to user %v", permissions, user)
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"login":      user,
			"permission": permissions}).
		Post("/permissions/add_user")
	if err != nil || resp.IsError() {
		return logErrorAndReturn(errors.New(fmt.Sprintf("Adding permission %s to user %s failed. Err - %s. Response - %s", permissions, user, err, resp.Status())))
	}

	log.Printf("Permissions %v to user %v has been added", permissions, user)

	return nil
}

func (sc SonarClient) AddPermissionsToGroup(groupName string, permissions string) error {
	log.Printf("Start adding permissions %v to group %v", permissions, groupName)
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"groupName":  groupName,
			"permission": permissions}).
		Post("/permissions/add_group")
	if err != nil || resp.IsError() {
		return logErrorAndReturn(errors.New(fmt.Sprintf("Adding permission %s to group %s failed. Err - %s. Response - %s", permissions, groupName, err, resp.Status())))
	}

	log.Printf("Permissions %v to group %v has been added", permissions, groupName)
	return nil
}

func (sc SonarClient) GenerateUserToken(userName string) (*string, error) {
	tokenExist, err := sc.checkUserTokenExist(userName)
	if err != nil {
		return nil, err
	}

	if tokenExist {
		return nil, nil
	}

	log.Printf("Start generating token for user %v in Sonar", userName)
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"login": userName,
			"name":  strings.Title(userName)}).
		Post("/user_tokens/generate")

	if err != nil || resp.IsError() {
		return nil, logErrorAndReturn(errors.New(fmt.Sprintf("Generation token for user %s failed. Err - %s. Response - %s", userName, err, resp.Status())))
	}
	log.Printf("Token for user %v has been generated", userName)

	var rawResponse map[string]string
	err = json.Unmarshal(resp.Body(), &rawResponse)
	token := rawResponse["token"]

	return &token, nil
}

func (sc SonarClient) checkUserTokenExist(userName string) (bool, error) {
	resp, err := sc.resty.R().
		Get(fmt.Sprintf("/user_tokens/search?login=%v", userName))
	if err != nil || resp.IsError() {
		return false, err
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
		return logErrorAndReturn(err)
	}

	if webHookExist {
		return nil
	}

	log.Printf("Start creating webhook %v in Sonar", webhookName)
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"name": webhookName,
			"url":  webhookUrl}).
		Post("/webhooks/create")
	if err != nil || resp.IsError() {
		return logErrorAndReturn(errors.New(fmt.Sprintf("Adding webhook %s failed. Err - %s. Response - %s", webhookName, err, resp.Status())))
	}
	log.Printf("Webhook %v has been created", webhookName)

	return nil
}

func (sc SonarClient) checkWebhookExist(webhookName string) (bool, error) {
	resp, err := sc.resty.R().
		Get("/webhooks/list")
	if err != nil || resp.IsError() {
		return false, err
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

//TODO(Serhii Shydlovskyi): Current implementation works ONLY for sonar.typescript.lcov.reportPaths. Requires effort to generalize it.
func (sc SonarClient) ConfigureGeneralSettings() error {
	key := "sonar.typescript.lcov.reportPaths"
	reportPath := "coverage/lcov.info"

	generalSettingsExist, err := sc.checkGeneralSetting(key, reportPath)
	if err != nil {
		return logErrorAndReturn(err)
	}

	if generalSettingsExist {
		return nil
	}

	log.Printf("Configuring general settings.")
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(
			map[string]string{
				"key":    key,
				"values": reportPath}).
		Post("/settings/set")
	if err != nil || resp.IsError() {
		log.Printf("%v", resp)
		return err
	}

	return nil
}

func (sc SonarClient) checkGeneralSetting(key string, value string) (bool, error) {
	resp, err := sc.resty.R().
		Get("/settings/values")
	if err != nil || resp.IsError() {
		return false, err
	}

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)

	for _, v := range raw["settings"] {
		if v["key"] == key {
			s := reflect.ValueOf(v["values"])
			for i := 0; i < s.Len(); i++ {
				if value == s.Index(i).Interface() {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
