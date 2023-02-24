package sonar

import (
	"context"
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

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	pMock "github.com/epam/edp-sonar-operator/mocks/platform"
	sMock "github.com/epam/edp-sonar-operator/mocks/service"
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
		Name:      name,
	}
}

func createDBSecret() v1.Secret {
	return v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: createObjectMeta(),
		Data: map[string][]byte{
			"token": []byte("token"),
		},
	}
}

func createInstance() *sonarApi.Sonar {
	return &sonarApi.Sonar{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Sonar",
			APIVersion: "v1",
		},
		ObjectMeta: createObjectMeta(),
		Spec: sonarApi.SonarSpec{
			Secret: name,
		},
	}
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
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{})
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

func TestReconcileSonar_Reconcile_ConfigureErr(t *testing.T) {
	ctx := context.Background()
	instance := createInstance()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{}, &v1.Secret{})
	secret := createDBSecret()
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance, &secret).Build()
	errTest := errors.New("test")

	service := sMock.ServiceInterface{}
	platform := pMock.Service{}

	platform.On("SetOwnerReference", instance, &secret).Return(nil)
	service.On("Configure", ctx,
		tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
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
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{}, &v1.Secret{})
	secret := createDBSecret()
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance, &secret).Build()
	errTest := errors.New("test")

	service := sMock.ServiceInterface{}
	platform := pMock.Service{}

	platform.On("SetOwnerReference", instance, &secret).Return(nil)

	service.On("Configure", ctx,
		tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
			return sonar.Name == name && sonar.Namespace == namespace
		})).Return(nil)

	service.On("ExposeConfiguration", ctx, tMock.MatchedBy(func(sonar *sonarApi.Sonar) bool {
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
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{}, &v1.Secret{})
	secret := createDBSecret()
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance, &secret).Build()
	errTest := errors.New("test")

	service := sMock.ServiceInterface{}
	platform := pMock.Service{}

	platform.On("SetOwnerReference", instance, &secret).Return(nil)

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
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{}, &v1.Secret{})
	secret := createDBSecret()
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance, &secret).Build()

	service := sMock.ServiceInterface{}
	platform := pMock.Service{}

	platform.On("SetOwnerReference", instance, &secret).Return(nil)

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
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.Sonar{}, &v1.Secret{})
	secret := createDBSecret()
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(instance, &secret).Build()

	service := sMock.ServiceInterface{}
	platform := pMock.Service{}

	platform.On("SetOwnerReference", instance, &secret).Return(nil)

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
