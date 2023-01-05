package sonar

import (
	"context"
	"fmt"
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
	apiV1 "github.com/epam/edp-sonar-operator/v2/api/edp/v1"
	pMock "github.com/epam/edp-sonar-operator/v2/mocks/platform"
	sMock "github.com/epam/edp-sonar-operator/v2/mocks/service"
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

func createInstance() *apiV1.Sonar {
	return &apiV1.Sonar{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Sonar",
			APIVersion: "v1",
		},
		ObjectMeta: createObjectMeta(),
	}
}

func TestReconcileSonar_Reconcile_ShouldFailToCreateDBSecret(t *testing.T) {
	sn := apiV1.Sonar{
		ObjectMeta: metav1.ObjectMeta{Name: "sonar", Namespace: "fake"},
		TypeMeta:   metav1.TypeMeta{Kind: "Sonar", APIVersion: "v2.edp.epam.com/v1"},
		Spec:       apiV1.SonarSpec{BasePath: "path"},
	}
	scheme := runtime.NewScheme()
	if err := apiV1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	rq := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: sn.Namespace, Name: sn.Name}}
	fakeCl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&sn).Build()
	l := mock.NewLogr()
	rec, err := NewReconcileSonar(fakeCl, scheme, l, "kubernetes")
	if err != nil {
		t.Fatal(err)
	}

	rs, err := rec.Reconcile(context.Background(), rq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create secret for - sonar-db")
	assert.Equal(t, reconcile.Result{RequeueAfter: 30 * time.Second}, rs)
}

func TestReconcileSonar_Reconcile_BadClient(t *testing.T) {
	ctx := context.Background()
	client := fake.NewClientBuilder().Build()
	controller := ReconcileSonar{
		client: client,
		log:    logr.FromContextOrDiscard(ctx),
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
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &apiV1.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	controller := ReconcileSonar{
		client: client,
		log:    logr.FromContextOrDiscard(ctx),
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
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &apiV1.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()
	errTest := errors.New("test")

	secret := createSecret()
	platform := pMock.Service{}
	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(errTest)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.FromContextOrDiscard(ctx),
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
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &apiV1.Sonar{})
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
		log:      logr.FromContextOrDiscard(ctx),
		platform: &platform,
		service:  &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to checking if deployment configs is ready")
	assert.Equal(t, reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, result)
	service.AssertExpectations(t)
	platform.AssertExpectations(t)
}

func TestReconcileSonar_Reconcile_IsDeploymentReadyFalse(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &apiV1.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(false, nil)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.FromContextOrDiscard(ctx),
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
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &apiV1.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()
	errTest := errors.New("test")

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(true, nil)
	service.On("Configure", ctx,
		tMock.MatchedBy(func(sonar *apiV1.Sonar) bool {
			return sonar.Name == name && sonar.Namespace == namespace
		})).Return(errTest)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.FromContextOrDiscard(ctx),
		platform: &platform,
		service:  &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to configure")
	assert.Equal(t, reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, result)
	service.AssertExpectations(t)
	platform.AssertExpectations(t)
}

func TestReconcileSonar_Reconcile_ExposeConfigurationErr(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &apiV1.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()
	errTest := errors.New("test")

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(true, nil)

	service.On("Configure", ctx,
		tMock.MatchedBy(func(sonar *apiV1.Sonar) bool {
			return sonar.Name == name && sonar.Namespace == namespace
		})).Return(nil)

	service.On("ExposeConfiguration", ctx, tMock.MatchedBy(func(sonar *apiV1.Sonar) bool {
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(errTest)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.FromContextOrDiscard(ctx),
		platform: &platform,
		service:  &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to expose configuration")
	assert.Equal(t, reconcile.Result{RequeueAfter: 10 * time.Second}, result)
	service.AssertExpectations(t)
	platform.AssertExpectations(t)
}

func TestReconcileSonar_Reconcile_IntegrationErr(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &apiV1.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()
	errTest := errors.New("test")

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(true, nil)

	service.On("Configure", ctx,
		tMock.MatchedBy(func(sonar *apiV1.Sonar) bool {
			return sonar.Name == name && sonar.Namespace == namespace
		})).Return(nil)

	service.On("ExposeConfiguration", ctx, tMock.MatchedBy(func(sonar *apiV1.Sonar) bool {
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(nil)

	service.On("Integration", ctx, tMock.MatchedBy(func(sonar *apiV1.Sonar) bool {
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(nil, errTest)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.FromContextOrDiscard(ctx),
		platform: &platform,
		service:  &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to integrate")
	assert.Equal(t, reconcile.Result{RequeueAfter: 10 * time.Second}, result)
	service.AssertExpectations(t)
	platform.AssertExpectations(t)
}

func TestReconcileSonar_Reconcile_updateAvailableStatusErr(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &apiV1.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(true, nil)

	service.On("Configure", ctx,
		tMock.MatchedBy(func(sonar *apiV1.Sonar) bool {
			return sonar.Name == name && sonar.Namespace == namespace
		})).Return(nil)

	service.On("ExposeConfiguration", ctx, tMock.MatchedBy(func(sonar *apiV1.Sonar) bool {
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(nil)

	service.On("Integration", ctx, tMock.MatchedBy(func(sonar *apiV1.Sonar) bool {
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(instance, nil)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.FromContextOrDiscard(ctx),
		platform: &platform,
		service:  &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update availability status")
	assert.Equal(t, reconcile.Result{RequeueAfter: 30 * time.Second}, result)
	service.AssertExpectations(t)
	platform.AssertExpectations(t)
}

func TestReconcileSonar_Reconcile(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &apiV1.Sonar{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance).Build()

	service := sMock.ServiceInterface{}
	secret := createSecret()
	platform := pMock.Service{}

	platform.On("CreateSecret", name, namespace, fmt.Sprintf("%s-db", name), tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("IsDeploymentReady", instance).Return(true, nil)

	service.On("Configure", ctx,
		tMock.MatchedBy(func(sonar *apiV1.Sonar) bool {
			return sonar.Name == name && sonar.Namespace == namespace
		})).Return(nil)

	service.On("ExposeConfiguration", ctx, tMock.MatchedBy(func(sonar *apiV1.Sonar) bool {
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(nil)

	instance2 := instance.DeepCopy()
	service.On("Integration", ctx, tMock.MatchedBy(func(sonar *apiV1.Sonar) bool {
		// hack. this is done in order not to monitor the state of the instance
		instance2.ObjectMeta = sonar.ObjectMeta
		return sonar.Name == name && sonar.Namespace == namespace
	})).Return(instance2, nil)

	controller := ReconcileSonar{
		client:   client,
		log:      logr.FromContextOrDiscard(ctx),
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
