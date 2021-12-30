package sonar

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
)

type ClientMock struct {
	mock.Mock
}

func (c *ClientMock) ConfigureGeneralSettings(valueType string, key string, value string) error {
	return c.Called(valueType, key, value).Error(0)
}

func (c *ClientMock) AddUserToGroup(groupName string, user string) error {
	panic("not implemented")
}

func (c *ClientMock) GenerateUserToken(userName string) (*string, error) {
	panic("not implemented")
}

func (c *ClientMock) CreateUser(ctx context.Context, u *sonar.User) error {
	return c.Called(u).Error(0)
}

func (c *ClientMock) GetUser(ctx context.Context, userName string) (*sonar.User, error) {
	called := c.Called(userName)
	if err := called.Error(1); err != nil {
		return nil, err
	}

	return called.Get(0).(*sonar.User), nil
}

func (c *ClientMock) GetUserToken(ctx context.Context, userLogin, tokenName string) (*sonar.UserToken, error) {
	called := c.Called(userLogin, tokenName)
	if err := called.Error(1); err != nil {
		return nil, err
	}

	return called.Get(0).(*sonar.UserToken), nil
}

func (c *ClientMock) CreateQualityGate(qgName string, conditions []map[string]string) (string, error) {
	called := c.Called(qgName)
	return called.String(0), called.Error(1)
}

func (c *ClientMock) InstallPlugins(plugins []string) error {
	return c.Called(plugins).Error(0)
}

func (c *ClientMock) GetGroup(ctx context.Context, groupName string) (*sonar.Group, error) {
	called := c.Called(groupName)
	if err := called.Error(1); err != nil {
		return nil, err
	}

	return called.Get(0).(*sonar.Group), nil
}

func (c *ClientMock) AddWebhook(webhookName string, webhookUrl string) error {
	return c.Called(webhookName, webhookUrl).Error(0)
}

func (c *ClientMock) AddPermissionsToGroup(groupName string, permissions string) error {
	return c.Called(groupName, permissions).Error(0)
}

func (c *ClientMock) CreateGroup(ctx context.Context, gr *sonar.Group) error {
	return c.Called(gr).Error(0)
}

func (c *ClientMock) SetProjectsDefaultVisibility(visibility string) error {
	panic("not implemented")
}

func (c *ClientMock) AddPermissionsToUser(user string, permissions string) error {
	panic("not implemented")
}

func (c *ClientMock) WaitForStatusIsUp(retryCount int, timeout time.Duration) error {
	panic("not implemented")
}

func (c *ClientMock) ChangePassword(ctx context.Context, user string, oldPassword string, newPassword string) error {
	return c.Called(user, oldPassword, newPassword).Error(0)
}

func (c *ClientMock) UploadProfile(profileName string, profilePath string) (string, error) {
	called := c.Called(profileName)
	return called.String(0), called.Error(1)
}

func (c *ClientMock) UpdateGroup(ctx context.Context, currentName string, group *sonar.Group) error {
	return c.Called(currentName, group).Error(0)
}

func (c *ClientMock) DeleteGroup(ctx context.Context, groupName string) error {
	return c.Called(groupName).Error(0)
}

func (c *ClientMock) AddGroupToPermissionTemplate(ctx context.Context, templateID string,
	permGroup *sonar.PermissionTemplateGroup) error {
	return c.Called(templateID, permGroup).Error(0)
}

func (c *ClientMock) CreatePermissionTemplate(ctx context.Context, tpl *sonar.PermissionTemplateData) (string, error) {
	called := c.Called(tpl)
	return called.String(0), called.Error(1)
}

func (c *ClientMock) UpdatePermissionTemplate(ctx context.Context, tpl *sonar.PermissionTemplate) error {
	return c.Called(tpl).Error(0)
}

func (c *ClientMock) DeletePermissionTemplate(ctx context.Context, id string) error {
	return c.Called(id).Error(0)
}

func (c *ClientMock) SearchPermissionTemplates(ctx context.Context, name string) ([]sonar.PermissionTemplate, error) {
	panic("not implemented")
}

func (c *ClientMock) GetPermissionTemplate(ctx context.Context, name string) (*sonar.PermissionTemplate, error) {
	called := c.Called(name)
	if err := called.Error(1); err != nil {
		return nil, err
	}

	return called.Get(0).(*sonar.PermissionTemplate), nil
}

func (c *ClientMock) GetPermissionTemplateGroups(ctx context.Context, templateID string) ([]sonar.PermissionTemplateGroup, error) {
	called := c.Called(templateID)
	if err := called.Error(1); err != nil {
		return nil, err
	}

	return called.Get(0).([]sonar.PermissionTemplateGroup), nil
}

func (c *ClientMock) RemoveGroupFromPermissionTemplate(ctx context.Context, templateID string,
	permGroup *sonar.PermissionTemplateGroup) error {
	return c.Called(templateID, permGroup).Error(0)
}

func (c *ClientMock) SetDefaultPermissionTemplate(ctx context.Context, name string) error {
	return c.Called(name).Error(0)
}
