package client

import (
	"encoding/json"
	"fmt"
	"gopkg.in/resty.v1"
)

type SonarClient struct {
	resty resty.Client
}

func (sc *SonarClient) InitNewRestClient(url string, user string, password string) error {
	sc.resty = *resty.SetHostURL(url).SetBasicAuth(user, password)
	return nil
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
