package platform

import (
	"context"

	"github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/stretchr/testify/mock"
	coreV1Api "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) CreateSecret(sonarName, namespace, secretName string, data map[string][]byte) (*coreV1Api.Secret, error) {
	called := m.Called(sonarName, namespace, secretName)
	if err := called.Error(1); err != nil {
		return nil, err
	}

	return called.Get(0).(*coreV1Api.Secret), nil
}

func (m *Mock) GetExternalEndpoint(ctx context.Context, namespace string, name string) (string, error) {
	called := m.Called(namespace, name)
	return called.String(0), called.Error(1)
}

func (m *Mock) CreateConfigMap(instance *v1alpha1.Sonar, configMapName string, configMapData map[string]string) error {
	panic("not implemented")
}

func (m *Mock) GetAvailableDeploymentReplicas(instance *v1alpha1.Sonar) (*int, error) {
	panic("not implemented")
}

func (m *Mock) GetSecretData(namespace string, name string) (map[string][]byte, error) {
	panic("not implemented")
}

func (m *Mock) CreateJenkinsServiceAccount(namespace string, secretName string, serviceAccountType string) error {
	panic("not implemented")
}

func (m *Mock) CreateJenkinsScript(namespace string, configMap string) error { panic("not implemented") }

func (m *Mock) CreateEDPComponentIfNotExist(sonar *v1alpha1.Sonar, url string, icon string) error {
	panic("not implemented")
}

func (m *Mock) SetOwnerReference(sonar *v1alpha1.Sonar, object client.Object) error {
	return m.Called(sonar, object).Error(0)
}
