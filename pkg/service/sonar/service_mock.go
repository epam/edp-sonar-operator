package sonar

import (
	"context"

	"github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/stretchr/testify/mock"
)

type ServiceMock struct {
	mock.Mock
}

func (s *ServiceMock) Configure(ctx context.Context, instance v1alpha1.Sonar) (*v1alpha1.Sonar, error, bool) {
	panic("not implemented")
}

func (s *ServiceMock) ExposeConfiguration(instance v1alpha1.Sonar) (*v1alpha1.Sonar, error) {
	panic("not implemented")
}

func (s *ServiceMock) Integration(instance v1alpha1.Sonar) (*v1alpha1.Sonar, error) {
	panic("not implemented")
}

func (s *ServiceMock) IsDeploymentReady(instance v1alpha1.Sonar) (bool, error) {
	panic("not implemented")
}

func (s *ServiceMock) InitSonarClient(instance *v1alpha1.Sonar, defaultPassword bool) (ClientInterface, error) {
	called := s.Called(instance, defaultPassword)
	if err := called.Error(1); err != nil {
		return nil, err
	}

	return called.Get(0).(ClientInterface), nil
}

func (s *ServiceMock) ClientForChild(ctx context.Context, instance ChildInstance) (ClientInterface, error) {
	called := s.Called()
	if err := called.Error(1); err != nil {
		return nil, err
	}

	return called.Get(0).(ClientInterface), nil
}

func (s *ServiceMock) DeleteResource(ctx context.Context, instance Deletable, finalizer string,
	deleteFunc func() error) (bool, error) {
	called := s.Called()

	if err := called.Error(1); err != nil {
		return false, err
	}

	if err := deleteFunc(); err != nil {
		return false, err
	}

	return called.Bool(0), nil
}
