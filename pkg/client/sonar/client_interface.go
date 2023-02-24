package sonar

import (
	"context"
)

type ClientInterface interface {
	ConfigureGeneralSettings(settings ...SettingRequest) error
	CreateQualityGates(qualityGates QualityGates) error
	InstallPlugins(plugins []string) error
	SetProjectsDefaultVisibility(visibility string) error

	UserInterface
	GroupInterface
	PermissionTemplateInterface
}

type UserInterface interface {
	AddPermissionToUser(user string, permissions string) error
	AddUserToGroup(groupName string, user string) error
	CreateUser(ctx context.Context, u *User) error
	GenerateUserToken(userName string) (*string, error)
	GetUser(ctx context.Context, userName string) (*User, error)
	GetUserToken(ctx context.Context, userLogin, tokenName string) (*UserToken, error)
}

type GroupInterface interface {
	AddPermissionsToGroup(groupName string, permissions string) error
	GetGroup(ctx context.Context, groupName string) (*Group, error)
	CreateGroup(ctx context.Context, gr *Group) error
	UpdateGroup(ctx context.Context, currentName string, group *Group) error
	DeleteGroup(ctx context.Context, groupName string) error
}

type PermissionTemplateInterface interface {
	CreatePermissionTemplate(ctx context.Context, tpl *PermissionTemplateData) (string, error)
	UpdatePermissionTemplate(ctx context.Context, tpl *PermissionTemplate) error
	DeletePermissionTemplate(ctx context.Context, id string) error
	GetPermissionTemplate(ctx context.Context, name string) (*PermissionTemplate, error)
	AddGroupToPermissionTemplate(ctx context.Context, templateID string, permGroup *PermissionTemplateGroup) error
	GetPermissionTemplateGroups(ctx context.Context, templateID string) ([]PermissionTemplateGroup, error)
	RemoveGroupFromPermissionTemplate(ctx context.Context, templateID string, permGroup *PermissionTemplateGroup) error
	SetDefaultPermissionTemplate(ctx context.Context, name string) error
}
