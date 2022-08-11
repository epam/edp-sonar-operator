package sonar

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	tMock "github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/epam/edp-common/pkg/mock"

	pMock "github.com/epam/edp-sonar-operator/v2/mocks/platform"
	sMock "github.com/epam/edp-sonar-operator/v2/mocks/service"
	sonarApi "github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1"
)

const (
	name      = "name"
	namespace = "namespace"
)

func createNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
}

func createObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: namespace,
		Name:      name}
}

func createSecret() v1.Secret {
	return v1.Secret{
		ObjectMeta: createObjectMeta(),
	}
}

func createInstance() *sonarApi.Sonar {
	return &sonarApi.Sonar{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Sonar",
			APIVersion: "v1",
		},
		ObjectMeta: createObjectMeta(),
	}
}

func TestReconcileSonar_Reconcile_ShouldFailToCreateDBSecret(t *testing.T) {
	sn := sonarApi.Sonar{
		ObjectMeta: metav1.ObjectMeta{Name: "sonar", Namespace: "fake"},
		TypeMeta:   metav1.TypeMeta{Kind: "Sonar", APIVersion: "v2.edp.epam.com/v1"},
		Spec:       sonarApi.SonarSpec{BasePath: "path"},
	}
	scheme := runtime.NewScheme()
	if err := sonarApi.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	rq := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: sn.Namespace, Name: sn.Name}}
	fakeCl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&sn).Build()
	l := mock.Logger{}
	rec, err := NewReconcileSonar(fakeCl, scheme, &l, "kubernetes")
	if err != nil {
		t.Fatal(err)
	}

	rs, err := rec.Reconcile(context.Background(), rq)
	assert.Error(t, err)
	if !strings.Contains(err.Error(), "Failed to create secret for sonar-db") {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
	assert.Equal(t, reconcile.Result{RequeueAfter: 30 * time.Second}, rs)
}

func TestReconcileSonar_Reconcile_BadClient(t *testing.T) {
	ctx := context.Background()
	client := fake.NewClientBuilder().Build()
	controller := ReconcileSonar{
		client: client,
		log:    logr.DiscardLogger{},
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, reconcile.Result{}, result)
}

func TestReconcileSonar_Reconcile_NotFound(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	controller := ReconcileSonar{
		client: client,
		log:    logr.DiscardLogger{},
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, reconcile.Result{}, result)
}

func TestReconcileSonar_Reconcile_SetOwnerReferenceErr(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()
	errTest := errors.New("test")

	secret := createSecret()
	platform := pMock.Service{}
	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(errTest)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.DiscardLogger{},
		platform: &platform,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, result)
	platform.AssertExpectations(t)
}

func TestReconcileSonar_Reconcile_IsDeploymentReadyErr(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()
	errTest := errors.New("test")

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(false, errTest)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.DiscardLogger{},
		platform: &platform,
		service:  &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Checking if Deployment configs is ready has been failed"))
	assert.Equal(t, reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, result)
	service.AssertExpectations(t)
	platform.AssertExpectations(t)
}

func TestReconcileSonar_Reconcile_IsDeploymentReadyFalse(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(false, nil)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.DiscardLogger{},
		platform: &platform,
		service:  &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, result)
	service.AssertExpectations(t)
	platform.AssertExpectations(t)
}

func TestReconcileSonar_Reconcile_ConfigureErr(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()
	errTest := errors.New("test")

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(true, nil)
	service.On("Configure", ctx,
		tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
			return sonar.Name == name && sonar.Namespace == namespace
		})).Return(errTest)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.DiscardLogger{},
		platform: &platform,
		service:  &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Configuration failed"))
	assert.Equal(t, reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, result)
	service.AssertExpectations(t)
	platform.AssertExpectations(t)
}

func TestReconcileSonar_Reconcile_ExposeConfigurationErr(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()
	errTest := errors.New("test")

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(true, nil)

	service.On("Configure", ctx,
		tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
			return sonar.Name == name && sonar.Namespace == namespace
		})).Return(nil)

	service.On("ExposeConfiguration", ctx, tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(errTest)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.DiscardLogger{},
		platform: &platform,
		service:  &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Exposing configuration failed"))
	assert.Equal(t, reconcile.Result{RequeueAfter: 10 * time.Second}, result)
	service.AssertExpectations(t)
	platform.AssertExpectations(t)
}

func TestReconcileSonar_Reconcile_IntegrationErr(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()
	errTest := errors.New("test")

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(true, nil)

	service.On("Configure", ctx,
		tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
			return sonar.Name == name && sonar.Namespace == namespace
		})).Return(nil)

	service.On("ExposeConfiguration", ctx, tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(nil)

	service.On("Integration", ctx, tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(nil, errTest)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.DiscardLogger{},
		platform: &platform,
		service:  &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Integration failed"))
	assert.Equal(t, reconcile.Result{RequeueAfter: 10 * time.Second}, result)
	service.AssertExpectations(t)
	platform.AssertExpectations(t)
}

func TestReconcileSonar_Reconcile_updateAvailableStatusErr(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(true, nil)

	service.On("Configure", ctx,
		tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
			return sonar.Name == name && sonar.Namespace == namespace
		})).Return(nil)

	service.On("ExposeConfiguration", ctx, tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(nil)

	service.On("Integration", ctx, tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(instance, nil)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.DiscardLogger{},
		platform: &platform,
		service:  &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Couldn't update availability status"))
	assert.Equal(t, reconcile.Result{RequeueAfter: 30 * time.Second}, result)
	service.AssertExpectations(t)
	platform.AssertExpectations(t)
}

func TestReconcileSonar_Reconcile(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(true, nil)

	service.On("Configure", ctx,
		tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
			return sonar.Name == name && sonar.Namespace == namespace
		})).Return(nil)

	service.On("ExposeConfiguration", ctx, tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(nil)

	instance2 := instance.DeepCopy()
	service.On("Integration", ctx, tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
		// hack. this is done in order not to monitor the state of the instance
		instance2.ObjectMeta = sonar.ObjectMeta
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(instance2, nil)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.DiscardLogger{},
		platform: &platform,
		service:  &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, reconcile.Result{Requeue: false}, result)
	service.AssertExpectations(t)
	platform.AssertExpectations(t)
}
