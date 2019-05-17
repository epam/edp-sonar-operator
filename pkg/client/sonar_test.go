package client

import (
	"log"
	"testing"
)

const (
	url      = "https://sonar-qa-389-edp-cicd.delivery.aws.main.edp.projects.epam.com/api"
	username = "admin"
	token    = ""
)

func TestExampleConfiguration_checkProfileExist(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	exist, result, err := cs.checkProfileExist()
	if err != nil {
		log.Print(err)
	}

	log.Println(result, exist)
}

func TestExampleConfiguration_UploadProfile(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	id, err := cs.UploadProfile()
	if err != nil {
		log.Print(err)
	}

	log.Println(*id)
}

func TestExampleConfiguration_checkInstallPlugins(t *testing.T) {
	sc := SonarClient{}
	err := sc.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	plugins := []string{"pmd"}
	err = sc.InstallPlugins(plugins)
	if err != nil {
		log.Print(err)
	}

	//assert.Assert(t, err == nil)
}
