package sonar

import (
	"context"
	"time"

	"github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
)

type ClientInterface interface {
	ConfigureGeneralSettings(valueType string, key string, value string) error
	AddUserToGroup(groupName string, user string) error
	GenerateUserToken(userName string) (*string, error)
	CreateUser(login string, name string, password string) error
	CreateQualityGate(qgName string, conditions []map[string]string) (*string, error)
	InstallPlugins(plugins []string) error
	GetGroup(ctx context.Context, groupName string) (*sonar.Group, error)
	AddWebhook(webhookName string, webhookUrl string) error
	AddPermissionsToGroup(groupName string, permissions string) error
	CreateGroup(ctx context.Context, gr *sonar.Group) error
	SetProjectsDefaultVisibility(visibility string) error
	AddPermissionsToUser(user string, permissions string) error
	WaitForStatusIsUp(retryCount int, timeout time.Duration) error
	ChangePassword(user string, oldPassword string, newPassword string) error
	UploadProfile(profileName string, profilePath string) (*string, error)
	UpdateGroup(ctx context.Context, currentName string, group *sonar.Group) error
	DeleteGroup(ctx context.Context, groupName string) error
}
