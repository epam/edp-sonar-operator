package sonar

import (
	"context"
	"time"

	"github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	"github.com/stretchr/testify/mock"
)

type ClientMock struct {
	mock.Mock
}

func (c *ClientMock) ConfigureGeneralSettings(valueType string, key string, value string) error {
	panic("not implemented")
}

func (c *ClientMock) AddUserToGroup(groupName string, user string) error {
	panic("not implemented")
}

func (c *ClientMock) GenerateUserToken(userName string) (*string, error) {
	panic("not implemented")
}

func (c *ClientMock) CreateUser(login string, name string, password string) error {
	panic("not implemented")
}

func (c *ClientMock) CreateQualityGate(qgName string, conditions []map[string]string) (*string, error) {
	panic("not implemented")
}

func (c *ClientMock) InstallPlugins(plugins []string) error {
	panic("not implemented")
}

func (c *ClientMock) GetGroup(ctx context.Context, groupName string) (*sonar.Group, error) {
	called := c.Called(groupName)
	if err := called.Error(1); err != nil {
		return nil, err
	}

	return called.Get(0).(*sonar.Group), nil
}

func (c *ClientMock) AddWebhook(webhookName string, webhookUrl string) error {
	panic("not implemented")
}

func (c *ClientMock) AddPermissionsToGroup(groupName string, permissions string) error {
	panic("not implemented")
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

func (c *ClientMock) ChangePassword(user string, oldPassword string, newPassword string) error {
	panic("not implemented")
}

func (c *ClientMock) UploadProfile(profileName string, profilePath string) (*string, error) {
	panic("not implemented")
}

func (c *ClientMock) UpdateGroup(ctx context.Context, currentName string, group *sonar.Group) error {
	return c.Called(currentName, group).Error(0)
}

func (c *ClientMock) DeleteGroup(ctx context.Context, groupName string) error {
	return c.Called(groupName).Error(0)
}

func (c *ClientMock) AddGroupToPermissionTemplate(ctx context.Context, permGroup *sonar.PermissionTemplateGroup) error {
	return c.Called(permGroup).Error(0)
}

func (c *ClientMock) CreatePermissionTemplate(ctx context.Context, tpl *sonar.PermissionTemplate) error {
	return c.Called(tpl).Error(0)
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

func (c *ClientMock) RemoveGroupFromPermissionTemplate(ctx context.Context, permGroup *sonar.PermissionTemplateGroup) error {
	return c.Called(permGroup).Error(0)
}
