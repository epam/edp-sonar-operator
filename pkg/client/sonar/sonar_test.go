package sonar

import (
	logger "log"
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
		logger.Print(err)
	}

	exist, result, _, err := cs.checkProfileExist(defaultProfileName)
	if err != nil {
		logger.Print(err)
	}

	logger.Println(result, exist)
}

func TestExampleConfiguration_CreateGroup(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		logger.Print(err)
	}

	err = cs.CreateGroup(groupName)
	if err != nil {
		logger.Print(err)
	}
}

func TestExampleConfiguration_AddUserToGroup(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		logger.Print(err)
	}

	err = cs.AddUserToGroup(groupName, "jenkins")
	if err != nil {
		logger.Print(err)
	}
}

func TestExampleConfiguration_AddPermissionsToUser(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		logger.Print(err)
	}

	err = cs.AddPermissionsToUser(jenkinsUsername, "admin")
	if err != nil {
		logger.Print(err)
	}
}

func TestExampleConfiguration_AddPermissionsToGroup(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, token)
	if err != nil {
		logger.Print(err)
	}

	err = cs.AddPermissionsToGroup(groupName, "scan")
	if err != nil {
		logger.Print(err)
	}
}

func TestExampleConfiguration_CreateUser(t *testing.T) {
	cs := SonarClient{}
	err := cs.InitNewRestClient(url, username, "password")
	if err != nil {
		logger.Print(err)
	}

	err = cs.CreateUser(jenkinsUserLogin, jenkinsUsername, jenkinsUserpassword)
	if err != nil {
		logger.Print(err)
	}
}

func TestExampleConfiguration_CheckUserToken(t *testing.T) {
	sc := SonarClient{}
	err := sc.InitNewRestClient(url, username, token)
	if err != nil {
		logger.Print(err)
	}

	exist, err := sc.checkUserTokenExist(jenkinsUsername)
	if err != nil {
		logger.Print(err)
	}

	logger.Println(exist)
}

func TestExampleConfiguration_checkWebhook(t *testing.T) {
	sc := SonarClient{}
	err := sc.InitNewRestClient(url, username, token)
	if err != nil {
		logger.Print(err)
	}

	exist, err := sc.checkWebhookExist(webhookName)
	if err != nil {
		logger.Print(err)
	}

	logger.Println(exist)
}
