package sonar

import (
	"context"
	logger "log"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

const (
	jenkinsUsername    = "kostenko"
	groupName          = "non-interactive-users"
	webhookName        = "jenkins"
	defaultProfileName = "Sonar way"
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

func TestExampleConfiguration_checkWebhook(t *testing.T) {
	sc := InitNewRestClient("", "", "")

	exist, err := sc.checkWebhookExist(webhookName)
	if err != nil {
		logger.Print(err)
	}

	logger.Println(exist)
}

func TestClient_WaitForStatusIsUp(t *testing.T) {
	sc := InitNewRestClient("", "", "")

	err := sc.WaitForStatusIsUp(1, time.Nanosecond)
	if err == nil {
		t.Fatal("no error returned")
	}

	if sc.resty.RetryCount > 0 {
		t.Fatal("retry count is changed")
	}
}

func TestClient_ChangePassword(t *testing.T) {
	sc := initClient()

	httpmock.RegisterResponder("GET", "/system/health", httpmock.NewStringResponder(200, ""))
	httpmock.RegisterResponder("POST", "/users/change_password", httpmock.NewStringResponder(200, ""))

	if err := sc.ChangePassword(context.Background(), "foo", "bar", "baz"); err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "/system/health",
		httpmock.NewStringResponder(http.StatusUnauthorized, ""))

	if err := sc.ChangePassword(context.Background(),
		"foo", "bar", "baz"); !IsHTTPErrorCode(err, http.StatusUnauthorized) {
		t.Fatal("no error or wrong type")
	}

	httpmock.RegisterResponder("GET", "/system/health", httpmock.NewStringResponder(200, ""))
	httpmock.RegisterResponder("POST", "/users/change_password", httpmock.NewStringResponder(500, ""))

	if err := sc.ChangePassword(context.Background(),
		"foo", "bar", "baz"); !IsHTTPErrorCode(err, http.StatusInternalServerError) {
		t.Fatal("no error or wrong type")
	}
}
