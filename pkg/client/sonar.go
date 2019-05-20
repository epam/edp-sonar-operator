package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/resty.v1"
	"log"
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
		SetBody("login="+user+"&password="+newPassword+"&previousPassword="+oldPassword).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post("/users/change_password")

	if err != nil {
		logErrorAndReturn(err)
	}

	if resp.IsError() {
		logErrorAndReturn(errors.New(fmt.Sprintf("Password change unsuccessful - %v", resp.Status())))
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
