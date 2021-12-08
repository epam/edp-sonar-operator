package sonar

import (
	"context"
	logger "log"
	"testing"
)

const (
	jenkinsUsername     = "kostenko"
	jenkinsUserLogin    = "jenkins"
	jenkinsUserpassword = "password"
	groupName           = "non-interactive-users"
	webhookName         = "jenkins"
	defaultProfileName  = "Sonar way"
)

//TODO: refactor all tests, replace logger with t.Fatal
func TestExampleConfiguration_checkProfileExist(t *testing.T) {
	cs := InitNewRestClient("", "", "")

	exist, result, _, err := cs.checkProfileExist(defaultProfileName)
	if err != nil {
		logger.Print(err)
	}

	logger.Println(result, exist)
}

func TestExampleConfiguration_CreateGroup(t *testing.T) {
	cs := InitNewRestClient("", "", "")

	err := cs.CreateGroup(context.Background(), &Group{Name: groupName})
	if err != nil {
		logger.Print(err)
	}
}

func TestExampleConfiguration_AddUserToGroup(t *testing.T) {
	cs := InitNewRestClient("", "", "")

	err := cs.AddUserToGroup(groupName, "jenkins")
	if err != nil {
		logger.Print(err)
	}
}

func TestExampleConfiguration_AddPermissionsToUser(t *testing.T) {
	cs := InitNewRestClient("", "", "")

	err := cs.AddPermissionsToUser(jenkinsUsername, "admin")
	if err != nil {
		logger.Print(err)
	}
}

func TestExampleConfiguration_AddPermissionsToGroup(t *testing.T) {
	cs := InitNewRestClient("", "", "")

	err := cs.AddPermissionsToGroup(groupName, "scan")
	if err != nil {
		logger.Print(err)
	}
}

func TestExampleConfiguration_CreateUser(t *testing.T) {
	cs := InitNewRestClient("", "", "")

	err := cs.CreateUser(jenkinsUserLogin, jenkinsUsername, jenkinsUserpassword)
	if err != nil {
		logger.Print(err)
	}
}

func TestExampleConfiguration_CheckUserToken(t *testing.T) {
	sc := InitNewRestClient("", "", "")

	exist, err := sc.checkUserTokenExist(jenkinsUsername)
	if err != nil {
		logger.Print(err)
	}

	logger.Println(exist)
}

func TestExampleConfiguration_checkWebhook(t *testing.T) {
	sc := InitNewRestClient("", "", "")

	exist, err := sc.checkWebhookExist(webhookName)
	if err != nil {
		logger.Print(err)
	}

	logger.Println(exist)
}
