package sonar

import (
	"gotest.tools/assert"
	"log"
	"testing"
)

const (
	url                 = "https://sonar-mr-1944-1-edp-cicd.delivery.aws.main.edp.projects.epam.com/api"
	username            = "admin"
	jenkinsUsername     = "kostenko"
	jenkinsUserLogin    = "jenkins"
	jenkinsUserpassword = "password"
	groupName           = "non-interactive-users"
	token               = ""
	webhookName         = "jenkins"
	webhookUrl          = "http://jenkins:8080/sonarqube-webhook/"
	defaultProfileName  = "Sonar way"
	profileName         = "EDP way"
	ProfilePath         = "../configs/quality-profile.xml"
)

func TestExampleConfiguration_checkProfileExist(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	exist, result, _, err := cs.checkProfileExist(defaultProfileName)
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

	id, err := cs.UploadProfile(profileName, ProfilePath)
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
}

func TestExampleConfiguration_checkGroupExist(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	exist, err := cs.checkGroupExist(groupName)
	if err != nil {
		log.Print(err)
	}
	log.Println(exist)

	assert.Equal(t, exist, true)
}

func TestExampleConfiguration_CreateGroup(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	err = cs.CreateGroup(groupName)
	if err != nil {
		log.Print(err)
	}
}

func TestExampleConfiguration_AddUserToGroup(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	err = cs.AddUserToGroup(groupName, "jenkins")
	if err != nil {
		log.Print(err)
	}
}

func TestExampleConfiguration_AddPermissionsToUser(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	err = cs.AddPermissionsToUser(jenkinsUsername, "admin")
	if err != nil {
		log.Print(err)
	}
}

func TestExampleConfiguration_AddPermissionsToGroup(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	err = cs.AddPermissionsToGroup(groupName, "scan")
	if err != nil {
		log.Print(err)
	}
}

func TestExampleConfiguration_CreateUser(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, "password")
	if err != nil {
		log.Print(err)
	}

	err = cs.CreateUser(jenkinsUserLogin, jenkinsUsername, jenkinsUserpassword)
	if err != nil {
		log.Print(err)
	}
}

func TestExampleConfiguration_CheckUserToken(t *testing.T) {
	sc := SonarClient{}
	err := sc.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	exist, err := sc.checkUserTokenExist(jenkinsUsername)
	if err != nil {
		log.Print(err)
	}

	log.Println(exist)
}

func TestExampleConfiguration_GenerateUserToken(t *testing.T) {
	sc := SonarClient{}
	err := sc.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	tokenPass, err := sc.GenerateUserToken(jenkinsUsername)
	if err != nil {
		log.Print(*tokenPass, err)
	}
}

func TestExampleConfiguration_checkWebhook(t *testing.T) {
	sc := SonarClient{}
	err := sc.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	exist, err := sc.checkWebhookExist(webhookName)
	if err != nil {
		log.Print(err)
	}

	log.Println(exist)
}

func TestExampleConfiguration_AddWebhook(t *testing.T) {
	sc := SonarClient{}
	err := sc.InitNewRestClient(url, username, token)
	if err != nil {
		log.Print(err)
	}

	err = sc.AddWebhook(webhookName, webhookUrl)
	if err != nil {
		log.Print(err)
	}
}
