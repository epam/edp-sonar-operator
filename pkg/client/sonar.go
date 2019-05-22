package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/resty.v1"
	"log"
	"reflect"
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
	resp, err := sc.resty.R().
		SetBody(fmt.Sprintf("login=%v&password=%v&previousPassword=%v", user, newPassword, oldPassword)).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post("/users/change_password")

	if err != nil {
		return logErrorAndReturn(err)
	}

	if resp.IsError() {
		return logErrorAndReturn(errors.New(fmt.Sprintf("Password change unsuccessful - %v", resp.Status())))
	}

	log.Printf("Password for user %v changed successfully", user)

	return nil
}

func (sc SonarClient) Reboot() error {
	resp, err := sc.resty.R().
		Post("/system/restart")

	if err != nil {
		log.Printf("[ERROR] Sonar rebooting failed - %s", resp.String())
		return logErrorAndReturn(err)
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
		log.Printf("Checking Sonar status failed - %s", resp.String())
		return logErrorAndReturn(err)
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
		if !checkPluginInstalled(installedPlugins, plugin) {
			needReboot = true
			resp, err := sc.resty.R().
				SetBody(fmt.Sprintf("key=%s", plugin)).
				SetHeader("Content-Type", "application/x-www-form-urlencoded").
				Post("/plugins/install")

			if err != nil || resp.IsError() {
				log.Printf("Plugin %s installation failed - %s", plugin, resp.String())
				return logErrorAndReturn(err)
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
	if err != nil || resp.IsError() {
		return nil, logErrorAndReturn(err)
	}

	var installedPlugins []string
	for _, v := range raw["plugins"] {
		installedPlugins = append(installedPlugins, fmt.Sprintf("%v", v["key"]))
	}

	return installedPlugins, nil
}

func (sc SonarClient) UploadProfile() (*string, error) {
	profileExist, profileId, err := sc.checkProfileExist()
	if err != nil {
		return nil, nil
	}
	if profileExist {
		err = sc.setDefaultProfile()
		if err != nil {
			return nil, err
		}
		return &profileId, nil
	}

	resp, err := sc.resty.R().
		SetHeader("Content-Type", "multipart/form-data").
		SetFile("backup", "/usr/local/bin/configs/quality-profile.xml").
		Post("/qualityprofiles/restore")
	if err != nil || resp.IsError() {
		return nil, err
	}
	_, profileId, err = sc.checkProfileExist()
	if err != nil {
		return nil, nil
	}

	err = sc.setDefaultProfile()
	if err != nil {
		return nil, err
	}

	return &profileId, nil
}

func (sc SonarClient) checkProfileExist() (bool, string, error) {
	resp, err := sc.resty.R().
		Get("/qualityprofiles/search?qualityProfile=EDP+way")
	if err != nil || resp.IsError() {
		return false, "", err
	}

	var raw map[string][]map[string]interface{}
	err = json.Unmarshal(resp.Body(), &raw)

	for _, v := range raw["profiles"] {
		if v["name"] == "EDP way" {
			return true, fmt.Sprintf("%v", v["key"]), nil
		}
	}
	return false, "", nil
}

func (sc SonarClient) setDefaultProfile() error {
	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"qualityProfile": "EDP way",
			"language":       "java"}).
		Post("/qualityprofiles/set_default")
	if err != nil || resp.IsError() {
		return err
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
	log.Printf("Start creating group %v in Sonar", groupName)
	groupExist, err := sc.checkGroupExist(groupName)
	if err != nil {
		return nil
	}

	if groupExist {
		log.Printf("Group %v already exist in Sonar", groupName)
		return nil
	}

	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"name": groupName}).
		Post("/user_groups/create")
	if err != nil || resp.IsError() {
		return err
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
		return err
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
		return err
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
		return err
	}

	log.Printf("Permissions %v to group %v has been added", permissions, groupName)

	return nil
}

func (sc SonarClient) GenerateUserToken(userName string) (*string, error) {
	log.Printf("Start generating token for user %v in Sonar", userName)
	tokenExist, err := sc.checkUserTokenExist(userName)
	if err != nil {
		return nil, err
	}

	if tokenExist {
		log.Printf("Token for user %v already exist in Sonar", userName)
		return nil, nil
	}

	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"login": userName,
			"name":  strings.Title(userName)}).
		Post("/user_tokens/generate")
	if err != nil || resp.IsError() {
		return nil, err
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
	log.Printf("Start creating webhook %v in Sonar", webhookName)
	webHookExist, err := sc.checkWebhookExist(webhookName)
	if err != nil {
		return err
	}

	if webHookExist {
		log.Printf("Webhook %v already exist in Sonar", webhookName)
		return nil
	}

	resp, err := sc.resty.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"name": webhookName,
			"url":  webhookUrl}).
		Post("/webhooks/create")
	if err != nil || resp.IsError() {
		return err
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
