package kubernetes

import (
	"context"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	tMock "github.com/stretchr/testify/mock"
	v1 "k8s.io/api/admission/v1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1Api "k8s.io/api/core/v1"
	networkingV1 "k8s.io/api/networking/v1"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	edpCompApi "github.com/epam/edp-component-operator/api/v1"
	jenkinsV1Api "github.com/epam/edp-jenkins-operator/v2/pkg/apis/v2/v1"

	sonarApi "github.com/epam/edp-sonar-operator/v2/api/v1"
	kMock "github.com/epam/edp-sonar-operator/v2/mocks/k8s"
	"github.com/epam/edp-sonar-operator/v2/pkg/helper"
)

const (
	name        = "name"
	namespace   = "ns"
	secretName  = "secret"
	host        = "domain"
	path        = "/path"
	serviceType = "ssh"
	icon        = "test.png"
)

func createObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}
}

func createSecret(name, namespace string, data map[string][]byte) *coreV1Api.Secret {
	labels := helper.GenerateLabels(name)

	return &coreV1Api.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: data,
		Type: "Opaque",
	}
}

func TestK8SService_Init(t *testing.T) {
	config := rest.Config{}
	scheme := runtime.NewScheme()
	client := fake.NewClientBuilder().Build()
	service := K8SService{}
	err := service.Init(&config, scheme, client)
	assert.NoError(t, err)
}

func TestK8SService_GetSecretData_GetErr(t *testing.T) {
	errTest := errors.New("Test")
	coreClient := kMock.K8SClusterClient{}
	secrets := &kMock.SecretInterface{}
	coreClient.On("Secrets", namespace).Return(secrets)
	secrets.On("Get", context.Background(), name, metav1.GetOptions{}).Return(nil, errTest)
	service := K8SService{
		k8sClusterClient: &coreClient,
	}
	data, err := service.GetSecretData(namespace, name)
	assert.Equal(t, errTest, err)
	assert.Nil(t, data)
	coreClient.AssertExpectations(t)
	secrets.AssertExpectations(t)
}

func TestK8SService_GetSecretData_NotFound(t *testing.T) {
	coreClient := kMock.K8SClusterClient{}
	secrets := &kMock.SecretInterface{}
	coreClient.On("Secrets", namespace).Return(secrets)
	secrets.On("Get", context.Background(), name, metav1.GetOptions{}).Return(nil, k8errors.NewNotFound(schema.GroupResource{}, name))
	service := K8SService{
		k8sClusterClient: &coreClient,
	}
	data, err := service.GetSecretData(namespace, name)
	assert.NoError(t, err)
	assert.Nil(t, data)
	coreClient.AssertExpectations(t)
	secrets.AssertExpectations(t)
}

func TestK8SService_GetSecretData(t *testing.T) {
	data := map[string][]byte{"data": []byte("data")}
	secret := coreV1Api.Secret{
		Data: data,
	}
	coreClient := kMock.K8SClusterClient{}
	secrets := &kMock.SecretInterface{}
	coreClient.On("Secrets", namespace).Return(secrets)
	secrets.On("Get", context.Background(), name, metav1.GetOptions{}).Return(&secret, nil)
	service := K8SService{
		k8sClusterClient: &coreClient,
	}
	secretData, err := service.GetSecretData(namespace, name)
	assert.NoError(t, err)
	assert.Equal(t, data, secretData)
	coreClient.AssertExpectations(t)
	secrets.AssertExpectations(t)
}

func TestK8SService_CreateSecret_GetErr(t *testing.T) {
	data := map[string][]byte{"data": []byte("crazy-random-data for secret")}
	errTest := errors.New("Test")
	coreClient := kMock.K8SClusterClient{}
	secrets := &kMock.SecretInterface{}
	coreClient.On("Secrets", namespace).Return(secrets)
	secrets.On("Get", context.Background(), secretName, metav1.GetOptions{}).Return(nil, errTest)
	service := K8SService{
		k8sClusterClient: &coreClient,
	}
	s, err := service.CreateSecret(name, namespace, secretName, data)
	assert.Equal(t, errTest, err)
	assert.Nil(t, s)
	coreClient.AssertExpectations(t)
	secrets.AssertExpectations(t)
}

func TestK8SService_CreateSecret_CreateErr(t *testing.T) {
	data := map[string][]byte{"data": []byte("data")}
	secret := createSecret(name, namespace, data)
	errTest := errors.New("Test")
	coreClient := kMock.K8SClusterClient{}
	secrets := &kMock.SecretInterface{}
	coreClient.On("Secrets", namespace).Return(secrets)
	secrets.On("Get", context.Background(), secretName, metav1.GetOptions{}).Return(nil, k8errors.NewNotFound(schema.GroupResource{}, name))
	secrets.On("Create", context.Background(), secret, metav1.CreateOptions{}).Return(nil, errTest)
	service := K8SService{
		k8sClusterClient: &coreClient,
	}
	createdSecret, err := service.CreateSecret(name, namespace, secretName, data)
	assert.Equal(t, errTest, err)
	assert.Nil(t, createdSecret)
	coreClient.AssertExpectations(t)
	secrets.AssertExpectations(t)
}

func TestK8SService_CreateSecret_AlreadyExist(t *testing.T) {
	data := map[string][]byte{"data": []byte("data")}
	existedSecret := coreV1Api.Secret{
		Data: data,
	}
	coreClient := kMock.K8SClusterClient{}
	secrets := &kMock.SecretInterface{}
	coreClient.On("Secrets", namespace).Return(secrets)
	secrets.On("Get", context.Background(), secretName, metav1.GetOptions{}).Return(&existedSecret, nil)
	service := K8SService{
		k8sClusterClient: &coreClient,
	}
	createdSecret, err := service.CreateSecret(name, namespace, secretName, data)
	assert.NoError(t, err)
	assert.Equal(t, &existedSecret, createdSecret)
	coreClient.AssertExpectations(t)
	secrets.AssertExpectations(t)
}

func TestK8SService_CreateSecret(t *testing.T) {
	data := map[string][]byte{"data": []byte("data")}
	secret := createSecret(name, namespace, data)
	coreClient := kMock.K8SClusterClient{}
	secrets := &kMock.SecretInterface{}
	coreClient.On("Secrets", namespace).Return(secrets)
	secrets.On("Get", context.Background(), secretName, metav1.GetOptions{}).Return(nil, k8errors.NewNotFound(schema.GroupResource{}, name))
	secrets.On("Create", context.Background(), secret, metav1.CreateOptions{}).Return(secret, nil)
	service := K8SService{
		k8sClusterClient: &coreClient,
	}
	createdSecret, err := service.CreateSecret(name, namespace, secretName, data)
	assert.NoError(t, err)
	assert.Equal(t, secret, createdSecret)
	coreClient.AssertExpectations(t)
	secrets.AssertExpectations(t)
}

func TestK8SService_GetExternalEndpoint_Err(t *testing.T) {
	errTest := errors.New("test")
	ctx := context.Background()
	networkingClient := kMock.NetworkingClient{}
	ingresses := &kMock.IngressInterface{}
	networkingClient.On("Ingresses", namespace).Return(ingresses)
	ingresses.On("Get", ctx, name, metav1.GetOptions{}).Return(nil, errTest)

	service := K8SService{
		NetworkingV1Client: &networkingClient,
	}
	endpoint, err := service.GetExternalEndpoint(ctx, namespace, name)
	assert.Equal(t, errTest, err)
	assert.Empty(t, endpoint)

	networkingClient.AssertExpectations(t)
	ingresses.AssertExpectations(t)
}

func TestK8SService_GetExternalEndpoint(t *testing.T) {
	ingressCR := networkingV1.Ingress{
		Spec: networkingV1.IngressSpec{
			Rules: []networkingV1.IngressRule{
				{
					Host: host,
					IngressRuleValue: networkingV1.IngressRuleValue{
						HTTP: &networkingV1.HTTPIngressRuleValue{
							Paths: []networkingV1.HTTPIngressPath{
								{Path: path},
							},
						},
					},
				},
			},
		},
	}
	ctx := context.Background()

	networkingClient := kMock.NetworkingClient{}
	ingresses := &kMock.IngressInterface{}

	networkingClient.On("Ingresses", namespace).Return(ingresses)
	ingresses.On("Get", ctx, name, metav1.GetOptions{}).Return(&ingressCR, nil)

	service := K8SService{
		NetworkingV1Client: &networkingClient,
	}
	endpoint, err := service.GetExternalEndpoint(ctx, namespace, name)
	assert.NoError(t, err)
	assert.Equal(t, "https://domain/path", endpoint)

	networkingClient.AssertExpectations(t)
	ingresses.AssertExpectations(t)
}

func TestK8SService_CreateConfigMap_BadScheme(t *testing.T) {
	scheme := runtime.NewScheme()
	configMapData := map[string]string{}
	sonarCR := sonarApi.Sonar{ObjectMeta: metav1.ObjectMeta{Name: name}}
	coreClient := kMock.K8SClusterClient{}
	configMapClient := &kMock.ConfigMapInterface{}

	service := K8SService{
		k8sClusterClient: &coreClient,
		Scheme:           scheme,
	}
	err := service.CreateConfigMap(&sonarCR, name, configMapData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set reference for config map")
	configMapClient.AssertExpectations(t)
	coreClient.AssertExpectations(t)
}

func TestK8SService_CreateConfigMap_GetErr(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
	errTest := errors.New("test")
	configMapData := map[string]string{}
	sonarCR := sonarApi.Sonar{ObjectMeta: metav1.ObjectMeta{Name: name}}
	coreClient := kMock.K8SClusterClient{}
	configMapClient := &kMock.ConfigMapInterface{}
	coreClient.On("ConfigMaps", "").Return(configMapClient)
	configMapClient.On("Get", context.Background(), name, metav1.GetOptions{}).Return(nil, errTest)

	service := K8SService{
		k8sClusterClient: &coreClient,
		Scheme:           scheme,
	}
	err := service.CreateConfigMap(&sonarCR, name, configMapData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get config map")
	configMapClient.AssertExpectations(t)
	coreClient.AssertExpectations(t)
}

func TestK8SService_CreateConfigMap_AlreadyExist(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
	configMapData := map[string]string{}
	sonarCR := sonarApi.Sonar{ObjectMeta: metav1.ObjectMeta{Name: name}}
	coreClient := kMock.K8SClusterClient{}
	configMapClient := &kMock.ConfigMapInterface{}
	coreClient.On("ConfigMaps", "").Return(configMapClient)
	configMapClient.On("Get", context.Background(), name, metav1.GetOptions{}).Return(nil, nil)

	service := K8SService{
		k8sClusterClient: &coreClient,
		Scheme:           scheme,
	}
	err := service.CreateConfigMap(&sonarCR, name, configMapData)
	assert.NoError(t, err)
	configMapClient.AssertExpectations(t)
	coreClient.AssertExpectations(t)
}

func TestK8SService_CreateConfigMap_CreateErr(t *testing.T) {
	errTest := errors.New("test")
	configMapData := map[string]string{}
	configMapObject := &coreV1Api.ConfigMap{}
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})

	sonarCR := sonarApi.Sonar{ObjectMeta: metav1.ObjectMeta{Name: name}}
	coreClient := kMock.K8SClusterClient{}
	configMapClient := &kMock.ConfigMapInterface{}
	coreClient.On("ConfigMaps", "").Return(configMapClient)
	configMapClient.On("Get", context.Background(), name, metav1.GetOptions{}).Return(nil, k8errors.NewNotFound(schema.GroupResource{}, name))
	configMapClient.On("Create", context.Background(), tMock.MatchedBy(func(configMap *coreV1Api.ConfigMap) bool {
		return configMap.Name == name && configMap.Namespace == ""
	}), metav1.CreateOptions{}).Return(configMapObject, errTest)

	service := K8SService{
		k8sClusterClient: &coreClient,
		Scheme:           scheme,
	}
	err := service.CreateConfigMap(&sonarCR, name, configMapData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create config map")
	configMapClient.AssertExpectations(t)
	coreClient.AssertExpectations(t)
}

func TestK8SService_CreateConfigMap(t *testing.T) {
	configMapData := map[string]string{}
	configMapObject := &coreV1Api.ConfigMap{ObjectMeta: createObjectMeta()}
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})

	sonarCR := sonarApi.Sonar{ObjectMeta: metav1.ObjectMeta{Name: name}}
	coreClient := kMock.K8SClusterClient{}
	configMapClient := &kMock.ConfigMapInterface{}
	coreClient.On("ConfigMaps", "").Return(configMapClient)
	configMapClient.On("Get", context.Background(), name, metav1.GetOptions{}).Return(nil, k8errors.NewNotFound(schema.GroupResource{}, name))
	configMapClient.On("Create", context.Background(), tMock.MatchedBy(func(configMap *coreV1Api.ConfigMap) bool {
		return configMap.Name == name && configMap.Namespace == ""
	}), metav1.CreateOptions{}).Return(configMapObject, nil)

	service := K8SService{
		k8sClusterClient: &coreClient,
		Scheme:           scheme,
	}
	err := service.CreateConfigMap(&sonarCR, name, configMapData)
	assert.NoError(t, err)
	configMapClient.AssertExpectations(t)
	coreClient.AssertExpectations(t)
}

func TestK8SService_CreateJenkinsScript_GetErr(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion)
	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects().Build()

	service := K8SService{
		Scheme: scheme,
		client: client,
	}
	err := service.CreateJenkinsScript(namespace, name)
	assert.Error(t, err)
	assert.True(t, runtime.IsNotRegisteredError(err))
}

func TestK8SService_CreateJenkinsScript_AlreadyExist(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &jenkinsV1Api.JenkinsScript{})
	jenkinsScriptCR := jenkinsV1Api.JenkinsScript{ObjectMeta: createObjectMeta()}
	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&jenkinsScriptCR).Build()

	service := K8SService{
		Scheme: scheme,
		client: client,
	}
	err := service.CreateJenkinsScript(namespace, name)
	assert.NoError(t, err)
}

func TestK8SService_CreateJenkinsScript(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &jenkinsV1Api.JenkinsScript{})
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	service := K8SService{
		Scheme: scheme,
		client: client,
	}
	err := service.CreateJenkinsScript(namespace, name)
	assert.NoError(t, err)
}

func TestK8SService_CreateJenkinsServiceAccount_GetErr(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion)
	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects().Build()

	service := K8SService{
		Scheme: scheme,
		client: client,
	}
	err := service.CreateJenkinsServiceAccount(namespace, name, serviceType)
	assert.Error(t, err)
	assert.True(t, runtime.IsNotRegisteredError(err))
}

func TestK8SService_CreateJenkinsServiceAccount_AlreadyExist(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &jenkinsV1Api.JenkinsServiceAccount{})
	jenkinsServiceAccountCR := jenkinsV1Api.JenkinsServiceAccount{ObjectMeta: createObjectMeta()}
	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&jenkinsServiceAccountCR).Build()

	service := K8SService{
		Scheme: scheme,
		client: client,
	}
	err := service.CreateJenkinsServiceAccount(namespace, name, serviceType)
	assert.NoError(t, err)
}

func TestK8SService_CreateJenkinsServiceAccount(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &jenkinsV1Api.JenkinsServiceAccount{})
	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	service := K8SService{
		Scheme: scheme,
		client: client,
	}
	err := service.CreateJenkinsServiceAccount(namespace, name, serviceType)
	assert.NoError(t, err)
}

func TestK8SService_CreateEDPComponentIfNotExist_GetErr(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects().Build()
	sonarCR := sonarApi.Sonar{ObjectMeta: createObjectMeta()}

	service := K8SService{
		Scheme: scheme,
		client: client,
	}
	err := service.CreateEDPComponentIfNotExist(&sonarCR, host, icon)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to get edp component"))
}

func TestK8SService_CreateEDPComponentIfNotExist_AlreadyExist(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &edpCompApi.EDPComponent{}, &sonarApi.Sonar{})
	edpComponentCR := edpCompApi.EDPComponent{ObjectMeta: createObjectMeta()}
	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&edpComponentCR).Build()
	sonarCR := sonarApi.Sonar{ObjectMeta: createObjectMeta()}

	service := K8SService{
		Scheme: scheme,
		client: client,
	}
	err := service.CreateEDPComponentIfNotExist(&sonarCR, host, icon)
	assert.NoError(t, err)
}

func TestK8SService_CreateEDPComponentIfNotExist(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &edpCompApi.EDPComponent{}, &sonarApi.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	sonarCR := sonarApi.Sonar{ObjectMeta: createObjectMeta()}

	service := K8SService{
		Scheme: scheme,
		client: client,
	}
	err := service.CreateEDPComponentIfNotExist(&sonarCR, host, icon)
	assert.NoError(t, err)
}

func TestK8SService_GetAvailableDeploymentReplicas_GetErr(t *testing.T) {
	sonarCR := sonarApi.Sonar{ObjectMeta: createObjectMeta()}
	appClient := kMock.PodsStateClient{}
	deploymentClient := &kMock.DeploymentInterface{}
	errTest := errors.New("test")

	appClient.On("Deployments", namespace).Return(deploymentClient)
	deploymentClient.On("Get", context.Background(), name, metav1.GetOptions{}).Return(nil, errTest)

	service := K8SService{
		AppsClient: &appClient,
	}
	replicas, err := service.GetAvailableDeploymentReplicas(&sonarCR)
	assert.Error(t, err)
	assert.Nil(t, replicas)
}

func TestK8SService_GetAvailableDeploymentReplicas(t *testing.T) {
	sonarCR := sonarApi.Sonar{ObjectMeta: createObjectMeta()}
	appClient := kMock.PodsStateClient{}
	deploymentClient := &kMock.DeploymentInterface{}
	deploymentCR := appsV1.Deployment{Status: appsV1.DeploymentStatus{
		AvailableReplicas: 1,
	}}
	appClient.On("Deployments", namespace).Return(deploymentClient)
	deploymentClient.On("Get", context.Background(), name, metav1.GetOptions{}).Return(&deploymentCR, nil)

	service := K8SService{
		AppsClient: &appClient,
	}
	replicas, err := service.GetAvailableDeploymentReplicas(&sonarCR)
	assert.NoError(t, err)
	assert.Equal(t, 1, *replicas)
}

func TestK8SService_SetOwnerReference(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &edpCompApi.EDPComponent{}, &sonarApi.Sonar{})
	edpComponentCR := edpCompApi.EDPComponent{ObjectMeta: createObjectMeta()}
	sonarCR := sonarApi.Sonar{ObjectMeta: createObjectMeta()}

	service := K8SService{
		Scheme: scheme,
	}
	err := service.SetOwnerReference(&sonarCR, &edpComponentCR)
	assert.NoError(t, err)
	assert.Equal(t, edpComponentCR.OwnerReferences[0].Kind, "Sonar")
	assert.Equal(t, edpComponentCR.OwnerReferences[0].Name, name)
}
