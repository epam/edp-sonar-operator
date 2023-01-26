package group

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	tMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/epam/edp-common/pkg/mock"

	sonarApi "github.com/epam/edp-sonar-operator/v2/api/v1"
	cMock "github.com/epam/edp-sonar-operator/v2/mocks/client"
	sMock "github.com/epam/edp-sonar-operator/v2/mocks/service"
	sonarClient "github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform"
)

const (
	name      = "name"
	namespace = "namespace"
	id        = "id"
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

func createInstance() sonarApi.SonarGroup {
	return sonarApi.SonarGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SonarGroup",
			APIVersion: "v1",
		},
		ObjectMeta: createObjectMeta(),
		Status: sonarApi.SonarGroupStatus{
			ID: id,
		},
	}
}
func TestNewReconcile(t *testing.T) {
	tm := metav1.Time{Time: time.Now()}
	ctx := context.Background()

	sg := sonarApi.SonarGroup{
		Spec: sonarApi.SonarGroupSpec{
			SonarOwner:  "sonar",
			Name:        "sc-name",
			Description: "sc-desc",
		},
		Status: sonarApi.SonarGroupStatus{ID: "id1"},
		TypeMeta: metav1.TypeMeta{
			Kind:       "SonarGroup",
			APIVersion: "v2.edp.epam.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "sg1", Namespace: "ns1",
			DeletionTimestamp: &tm},
	}

	sn := sonarApi.Sonar{
		ObjectMeta: metav1.ObjectMeta{Name: "sonar", Namespace: sg.Namespace},
		TypeMeta:   metav1.TypeMeta{Kind: "Sonar", APIVersion: "v2.edp.epam.com/v1"},
		Spec:       sonarApi.SonarSpec{BasePath: "path"},
	}

	scheme := runtime.NewScheme()
	if err := sonarApi.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	rq := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: sg.Namespace, Name: sg.Name}}
	fakeCl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&sg, &sn).Build()
	l := mock.NewLogr()

	rec, err := NewReconcile(fakeCl, scheme, l, platform.Kubernetes)
	if err != nil {
		t.Fatal(err)
	}

	serviceMock := sMock.ServiceInterface{}
	rec.service = &serviceMock
	clientMock := cMock.ClientInterface{}

	serviceMock.On("ClientForChild", ctx, tMock.AnythingOfType("*v1.SonarGroup")).Return(&clientMock, nil)
	serviceMock.
		On("DeleteResource",
			ctx,
			tMock.AnythingOfType("*v1.SonarGroup"),
			finalizer,
			tMock.AnythingOfType("func() error"),
		).
		Return(true, nil)
	clientMock.On("GetGroup", ctx, sg.Spec.Name).Return(&sonarClient.Group{}, nil)
	clientMock.
		On("UpdateGroup",
			ctx,
			sg.Spec.Name,
			&sonarClient.Group{
				Name:        sg.Spec.Name,
				Description: sg.Spec.Description,
			}).
		Return(nil)
	if _, err = rec.Reconcile(ctx, rq); err != nil {
		t.Fatal(err)
	}

	loggerSink, ok := l.GetSink().(*mock.Logger)
	require.True(t, ok, "wrong logger type")

	if err = loggerSink.LastError(); err != nil {
		t.Fatal(err)
	}

	serviceMock.AssertExpectations(t)
	clientMock.AssertExpectations(t)
}

func TestReconcile_Reconcile_BadClientErr(t *testing.T) {
	ctx := context.Background()
	client := fake.NewClientBuilder().Build()
	controller := Reconcile{
		client: client,
		log:    mock.NewLogr(),
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	result, err := controller.Reconcile(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, reconcile.Result{}, result)
}

func TestReconcile_Reconcile_NotFound(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.SonarGroup{})
	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	controller := Reconcile{
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

func TestReconcile_Reconcile_tryReconcile_ClientForChildErr(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	instance := createInstance()
	service := sMock.ServiceInterface{}
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.SonarGroup{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&instance).Build()

	errTest := errors.New("test")
	service.On("ClientForChild", ctx, &instance).Return(nil, errTest)

	controller := Reconcile{
		client:  client,
		log:     logr.FromContextOrDiscard(ctx),
		service: &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	_, err := controller.Reconcile(ctx, req)
	assert.NoError(t, err)

	err = client.Get(ctx, createNamespacedName(), &instance)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(instance.Status.Value, "unable to init sonar rest client"))
	service.AssertExpectations(t)
}

func TestReconcile_Reconcile_tryReconcile_GetGroupErr(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	instance := createInstance()
	service := sMock.ServiceInterface{}
	sonarClientMock := &cMock.ClientInterface{}
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.SonarGroup{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&instance).Build()

	errTest := errors.New("test")
	service.On("ClientForChild", ctx, &instance).Return(sonarClientMock, nil)
	sonarClientMock.On("GetGroup", ctx, instance.Spec.Name).Return(nil, errTest)

	controller := Reconcile{
		client:  client,
		log:     logr.FromContextOrDiscard(ctx),
		service: &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	_, err := controller.Reconcile(ctx, req)
	assert.NoError(t, err)

	err = client.Get(ctx, createNamespacedName(), &instance)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(instance.Status.Value, "unexpected error during get group"))
	service.AssertExpectations(t)
	sonarClientMock.AssertExpectations(t)
}

func TestReconcile_Reconcile_tryReconcile_CreateGroupErr(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	instance := createInstance()
	service := sMock.ServiceInterface{}
	sonarClientMock := &cMock.ClientInterface{}
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.SonarGroup{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&instance).Build()

	errTest := errors.New("test")
	service.On("ClientForChild", ctx, &instance).Return(sonarClientMock, nil)
	sonarClientMock.On("GetGroup", ctx, instance.Spec.Name).Return(nil, sonarClient.NotFoundError("test"))
	sonarClientMock.On("CreateGroup", ctx, &sonarClient.Group{Name: instance.Spec.Name, Description: instance.Spec.Description}).Return(errTest)

	controller := Reconcile{
		client:  client,
		log:     logr.FromContextOrDiscard(ctx),
		service: &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	_, err := controller.Reconcile(ctx, req)
	assert.NoError(t, err)

	err = client.Get(ctx, createNamespacedName(), &instance)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(instance.Status.Value, "unable to create sonar group"))
	service.AssertExpectations(t)
	sonarClientMock.AssertExpectations(t)
}

func TestReconcile_Reconcile_tryReconcile_EmptyStatusID(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	instance := sonarApi.SonarGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SonarGroup",
			APIVersion: "v1",
		},
		ObjectMeta: createObjectMeta(),
	}
	service := sMock.ServiceInterface{}
	sonarClientMock := &cMock.ClientInterface{}
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.SonarGroup{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&instance).Build()

	service.On("ClientForChild", ctx, &instance).Return(sonarClientMock, nil)
	sonarClientMock.On("GetGroup", ctx, instance.Spec.Name).Return(nil, nil)

	controller := Reconcile{
		client:  client,
		log:     logr.FromContextOrDiscard(ctx),
		service: &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	_, err := controller.Reconcile(ctx, req)
	assert.NoError(t, err)

	err = client.Get(ctx, createNamespacedName(), &instance)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(instance.Status.Value, "group already exists in sonar"))
	service.AssertExpectations(t)
	sonarClientMock.AssertExpectations(t)
}

func TestReconcile_Reconcile_tryReconcile_UpdateGroupErr(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	instance := createInstance()
	service := sMock.ServiceInterface{}
	sonarClientMock := &cMock.ClientInterface{}
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.SonarGroup{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&instance).Build()

	errTest := errors.New("test")
	service.On("ClientForChild", ctx, &instance).Return(sonarClientMock, nil)
	sonarClientMock.On("GetGroup", ctx, instance.Spec.Name).Return(nil, nil)
	sonarClientMock.On("UpdateGroup", ctx, instance.Spec.Name, &sonarClient.Group{Name: instance.Spec.Name,
		Description: instance.Spec.Description}).Return(errTest)

	controller := Reconcile{
		client:  client,
		log:     logr.FromContextOrDiscard(ctx),
		service: &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	_, err := controller.Reconcile(ctx, req)
	assert.NoError(t, err)

	err = client.Get(ctx, createNamespacedName(), &instance)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(instance.Status.Value, "unable to update group"))
	service.AssertExpectations(t)
	sonarClientMock.AssertExpectations(t)
}

func TestReconcile_Reconcile_tryReconcile_DeleteResourceErr(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	instance := createInstance()
	service := sMock.ServiceInterface{}
	sonarClientMock := &cMock.ClientInterface{}
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &sonarApi.SonarGroup{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&instance).Build()

	errTest := errors.New("test")
	service.On("ClientForChild", ctx, &instance).Return(sonarClientMock, nil)
	sonarClientMock.On("GetGroup", ctx, instance.Spec.Name).Return(nil, nil)
	sonarClientMock.On("UpdateGroup", ctx, instance.Spec.Name, &sonarClient.Group{Name: instance.Spec.Name,
		Description: instance.Spec.Description}).Return(nil)
	service.On("DeleteResource", ctx, &instance, finalizer, tMock.AnythingOfType("func() error")).Return(false, errTest)

	controller := Reconcile{
		client:  client,
		log:     logr.FromContextOrDiscard(ctx),
		service: &service,
	}
	req := reconcile.Request{
		NamespacedName: createNamespacedName(),
	}
	_, err := controller.Reconcile(ctx, req)
	assert.NoError(t, err)

	err = client.Get(ctx, createNamespacedName(), &instance)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(instance.Status.Value, "unable to delete resource"))
	service.AssertExpectations(t)
	sonarClientMock.AssertExpectations(t)
}
