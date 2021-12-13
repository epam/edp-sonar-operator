package sonar

import (
	"context"
	"time"

	"github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
)

type ClientInterface interface {
	AddWebhook(webhookName string, webhookUrl string) error
	ConfigureGeneralSettings(valueType string, key string, value string) error
	CreateQualityGate(qgName string, conditions []map[string]string) (string, error)
	InstallPlugins(plugins []string) error
	SetProjectsDefaultVisibility(visibility string) error
	UploadProfile(profileName string, profilePath string) (string, error)
	WaitForStatusIsUp(retryCount int, timeout time.Duration) error

	User
	Group
	PermissionTemplate
}

type User interface {
	AddPermissionsToUser(user string, permissions string) error
	AddUserToGroup(groupName string, user string) error
	ChangePassword(ctx context.Context, user string, oldPassword string, newPassword string) error
	CreateUser(ctx context.Context, u *sonar.User) error
	GenerateUserToken(userName string) (*string, error)
	GetUser(ctx context.Context, userName string) (*sonar.User, error)
	GetUserToken(ctx context.Context, userLogin, tokenName string) (*sonar.UserToken, error)
}

type Group interface {
	AddPermissionsToGroup(groupName string, permissions string) error
	GetGroup(ctx context.Context, groupName string) (*sonar.Group, error)
	CreateGroup(ctx context.Context, gr *sonar.Group) error
	UpdateGroup(ctx context.Context, currentName string, group *sonar.Group) error
	DeleteGroup(ctx context.Context, groupName string) error
}

type PermissionTemplate interface {
	CreatePermissionTemplate(ctx context.Context, tpl *sonar.PermissionTemplate) error
	UpdatePermissionTemplate(ctx context.Context, tpl *sonar.PermissionTemplate) error
	DeletePermissionTemplate(ctx context.Context, id string) error
	SearchPermissionTemplates(ctx context.Context, name string) ([]sonar.PermissionTemplate, error)
	GetPermissionTemplate(ctx context.Context, name string) (*sonar.PermissionTemplate, error)
	AddGroupToPermissionTemplate(ctx context.Context, permGroup *sonar.PermissionTemplateGroup) error
	GetPermissionTemplateGroups(ctx context.Context, templateID string) ([]sonar.PermissionTemplateGroup, error)
	RemoveGroupFromPermissionTemplate(ctx context.Context, permGroup *sonar.PermissionTemplateGroup) error
	SetDefaultPermissionTemplate(ctx context.Context, name string) error
}
