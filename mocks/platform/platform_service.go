// Code generated by mockery v2.9.4. DO NOT EDIT.

package mock

import (
	context "context"

	client "sigs.k8s.io/controller-runtime/pkg/client"

	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/api/core/v1"

	sonarApi "github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// CreateConfigMap provides a mock function with given fields: instance, configMapName, configMapData
func (_m *Service) CreateConfigMap(instance *sonarApi.Sonar, configMapName string, configMapData map[string]string) error {
	ret := _m.Called(instance, configMapName, configMapData)

	var r0 error
	if rf, ok := ret.Get(0).(func(*sonarApi.Sonar, string, map[string]string) error); ok {
		r0 = rf(instance, configMapName, configMapData)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateEDPComponentIfNotExist provides a mock function with given fields: sonar, url, icon
func (_m *Service) CreateEDPComponentIfNotExist(sonar *sonarApi.Sonar, url string, icon string) error {
	ret := _m.Called(sonar, url, icon)

	var r0 error
	if rf, ok := ret.Get(0).(func(*sonarApi.Sonar, string, string) error); ok {
		r0 = rf(sonar, url, icon)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateJenkinsScript provides a mock function with given fields: namespace, configMap
func (_m *Service) CreateJenkinsScript(namespace string, configMap string) error {
	ret := _m.Called(namespace, configMap)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(namespace, configMap)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateJenkinsServiceAccount provides a mock function with given fields: namespace, secretName, serviceAccountType
func (_m *Service) CreateJenkinsServiceAccount(namespace string, secretName string, serviceAccountType string) error {
	ret := _m.Called(namespace, secretName, serviceAccountType)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(namespace, secretName, serviceAccountType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateSecret provides a mock function with given fields: sonarName, namespace, secretName, data
func (_m *Service) CreateSecret(sonarName string, namespace string, secretName string, data map[string][]byte) (*v1.Secret, error) {
	ret := _m.Called(sonarName, namespace, secretName, data)

	var r0 *v1.Secret
	if rf, ok := ret.Get(0).(func(string, string, string, map[string][]byte) *v1.Secret); ok {
		r0 = rf(sonarName, namespace, secretName, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.Secret)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, map[string][]byte) error); ok {
		r1 = rf(sonarName, namespace, secretName, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAvailableDeploymentReplicas provides a mock function with given fields: instance
func (_m *Service) GetAvailableDeploymentReplicas(instance *sonarApi.Sonar) (*int, error) {
	ret := _m.Called(instance)

	var r0 *int
	if rf, ok := ret.Get(0).(func(*sonarApi.Sonar) *int); ok {
		r0 = rf(instance)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*int)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*sonarApi.Sonar) error); ok {
		r1 = rf(instance)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetExternalEndpoint provides a mock function with given fields: ctx, namespace, name
func (_m *Service) GetExternalEndpoint(ctx context.Context, namespace string, name string) (string, error) {
	ret := _m.Called(ctx, namespace, name)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = rf(ctx, namespace, name)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, namespace, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSecretData provides a mock function with given fields: namespace, name
func (_m *Service) GetSecretData(namespace string, name string) (map[string][]byte, error) {
	ret := _m.Called(namespace, name)

	var r0 map[string][]byte
	if rf, ok := ret.Get(0).(func(string, string) map[string][]byte); ok {
		r0 = rf(namespace, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(namespace, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetOwnerReference provides a mock function with given fields: sonar, object
func (_m *Service) SetOwnerReference(sonar *sonarApi.Sonar, object client.Object) error {
	ret := _m.Called(sonar, object)

	var r0 error
	if rf, ok := ret.Get(0).(func(*sonarApi.Sonar, client.Object) error); ok {
		r0 = rf(sonar, object)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}