package openshift

import (
	"context"
	"os"
	"testing"

	appsV1 "github.com/openshift/api/apps/v1"
	routeV1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	oMock "github.com/epam/edp-sonar-operator/mocks/openshift"
)

const (
	name      = "name"
	namespace = "ns"
	host      = "domain"
)

func TestK8SService_Init(t *testing.T) {
	config := rest.Config{}
	scheme := runtime.NewScheme()
	client := fake.NewClientBuilder().Build()
	service := OpenshiftService{}
	err := service.Init(&config, scheme, client)
	assert.NoError(t, err)
}

func TestOpenshiftService_GetExternalEndpoint_GetErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	routeClient := oMock.RouteClient{}
	routes := &oMock.Route{}
	routeClient.On("Routes", namespace).Return(routes)
	routes.On("Get", ctx, name, metav1.GetOptions{}).Return(nil, errTest)
	service := OpenshiftService{
		routeClient: &routeClient,
	}
	endpoint, err := service.GetExternalEndpoint(ctx, namespace, name)
	assert.Error(t, err)
	assert.Empty(t, endpoint)

	routes.AssertExpectations(t)
	routeClient.AssertExpectations(t)
}

func TestOpenshiftService_GetExternalEndpoint_NotFound(t *testing.T) {
	ctx := context.Background()
	routeClient := oMock.RouteClient{}
	routes := &oMock.Route{}
	routeClient.On("Routes", namespace).Return(routes)
	routes.On("Get", ctx, name, metav1.GetOptions{}).Return(nil, k8errors.NewNotFound(schema.GroupResource{}, name))
	service := OpenshiftService{
		routeClient: &routeClient,
	}
	endpoint, err := service.GetExternalEndpoint(ctx, namespace, name)
	assert.Error(t, err)
	assert.Empty(t, endpoint)

	routes.AssertExpectations(t)
	routeClient.AssertExpectations(t)
}

func TestOpenshiftService_GetExternalEndpoint(t *testing.T) {
	route := routeV1.Route{
		Spec: routeV1.RouteSpec{
			TLS:  &routeV1.TLSConfig{Termination: "yes"},
			Host: host,
		},
	}
	ctx := context.Background()
	routeClient := oMock.RouteClient{}
	routes := &oMock.Route{}
	routeClient.On("Routes", namespace).Return(routes)
	routes.On("Get", ctx, name, metav1.GetOptions{}).Return(&route, nil)
	service := OpenshiftService{
		routeClient: &routeClient,
	}
	endpoint, err := service.GetExternalEndpoint(ctx, namespace, name)
	assert.NoError(t, err)
	assert.Equal(t, "https://domain", endpoint)

	routes.AssertExpectations(t)
	routeClient.AssertExpectations(t)
}

func TestOpenshiftService_GetAvailableDeploymentReplicas_GetErr(t *testing.T) {
	err := os.Setenv(deploymentTypeEnvName, deploymentConfigsDeploymentType)
	if err != nil {
		t.Fatal(t)
	}
	appClient := oMock.OpenshiftPodsStateClient{}
	deploymentClient := &oMock.DeploymentConfig{}
	errTest := errors.New("test")
	sonarCR := sonarApi.Sonar{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: namespace,
		},
	}
	appClient.On("DeploymentConfigs", namespace).Return(deploymentClient)
	deploymentClient.On("Get", context.Background(), name, metav1.GetOptions{}).Return(nil, errTest)

	service := OpenshiftService{
		appClient: &appClient,
	}
	replicas, err := service.GetAvailableDeploymentReplicas(&sonarCR)
	assert.Error(t, err)
	assert.Nil(t, replicas)
	appClient.AssertExpectations(t)
	deploymentClient.AssertExpectations(t)
	err = os.Unsetenv(deploymentTypeEnvName)
	if err != nil {
		t.Fatal(t)
	}
}

func TestOpenshiftService_GetAvailableDeploymentReplicas(t *testing.T) {
	expectedReplicasCount := 1
	err := os.Setenv(deploymentTypeEnvName, deploymentConfigsDeploymentType)
	if err != nil {
		t.Fatal(t)
	}
	configCR := appsV1.DeploymentConfig{Status: appsV1.DeploymentConfigStatus{AvailableReplicas: int32(expectedReplicasCount)}}
	appClient := oMock.OpenshiftPodsStateClient{}
	deploymentClient := &oMock.DeploymentConfig{}
	sonarCR := sonarApi.Sonar{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: namespace,
		},
	}
	appClient.On("DeploymentConfigs", namespace).Return(deploymentClient)
	deploymentClient.On("Get", context.Background(), name, metav1.GetOptions{}).Return(&configCR, nil)

	service := OpenshiftService{
		appClient: &appClient,
	}
	replicas, err := service.GetAvailableDeploymentReplicas(&sonarCR)
	assert.NoError(t, err)
	assert.Equal(t, expectedReplicasCount, *replicas)
	appClient.AssertExpectations(t)
	deploymentClient.AssertExpectations(t)
	err = os.Unsetenv(deploymentTypeEnvName)
	if err != nil {
		t.Fatal(t)
	}
}
