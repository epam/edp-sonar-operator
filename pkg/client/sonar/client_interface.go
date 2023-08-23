package sonar

import (
	"context"
	"net/url"
)

// ClientInterface is an interface for Sonar client.
//
//go:generate mockery --name ClientInterface --filename client_mock.go
type ClientInterface interface {
	ConfigureGeneralSettings(settings ...SettingRequest) error
	InstallPlugins(plugins []string) error
	SetProjectsDefaultVisibility(visibility string) error

	UserInterface
	GroupInterface
	PermissionTemplateInterface
	Settings
	System
	QualityGateClient
	QualityProfileClient
	RuleClient
}

type UserInterface interface {
	CreateUser(ctx context.Context, u *User) error
	UpdateUser(ctx context.Context, u *User) error
	GenerateUserToken(userName string) (*string, error)
	GetUserByLogin(ctx context.Context, userLogin string) (*User, error)
	GetUserToken(ctx context.Context, userLogin, tokenName string) (*UserToken, error)
	GetUserGroups(ctx context.Context, userLogin string) ([]Group, error)
	DeactivateUser(ctx context.Context, userLogin string) error
}

type GroupInterface interface {
	AddPermissionsToGroup(groupName string, permissions string) error
	GetGroup(ctx context.Context, groupName string) (*Group, error)
	CreateGroup(ctx context.Context, gr *Group) error
	UpdateGroup(ctx context.Context, currentName string, group *Group) error
	DeleteGroup(ctx context.Context, groupName string) error
	AddUserToGroup(ctx context.Context, userLogin, groupName string) error
	RemoveUserFromGroup(ctx context.Context, userLogin, groupName string) error
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
	GetUserPermissions(ctx context.Context, userLogin string) ([]string, error)
	AddPermissionToUser(ctx context.Context, userLogin, permission string) error
	RemovePermissionFromUser(ctx context.Context, userLogin, permission string) error
}

type Settings interface {
	SetSetting(ctx context.Context, setting url.Values) error
	ResetSettings(ctx context.Context, settingsKeys []string) error
}

type System interface {
	Health(ctx context.Context) (*SystemHealth, error)
}

type QualityGateClient interface {
	CreateQualityGate(ctx context.Context, name string) (*QualityGate, error)
	GetQualityGate(ctx context.Context, name string) (*QualityGate, error)
	DeleteQualityGate(ctx context.Context, name string) error
	SetAsDefaultQualityGate(ctx context.Context, name string) error
	CreateQualityGateCondition(ctx context.Context, gate string, condition QualityGateCondition) error
	UpdateQualityGateCondition(ctx context.Context, condition QualityGateCondition) error
	DeleteQualityGateCondition(ctx context.Context, conditionId string) error
}

type QualityProfileClient interface {
	CreateQualityProfile(ctx context.Context, name, language string) (*QualityProfile, error)
	GetQualityProfile(ctx context.Context, name string) (*QualityProfile, error)
	DeleteQualityProfile(ctx context.Context, name, language string) error
	SetAsDefaultQualityProfile(ctx context.Context, name, language string) error
	ActivateQualityProfileRule(ctx context.Context, profileKey string, rule Rule) error
	DeactivateQualityProfileRule(ctx context.Context, profileKey, ruleKey string) error
}

type RuleClient interface {
	GetQualityProfileActiveRules(ctx context.Context, profileKey string) ([]Rule, error)
}
