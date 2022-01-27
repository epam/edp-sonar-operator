package sonar

import (
	"context"
	"encoding/json"
	"fmt"
	keycloakApi "github.com/epam/edp-keycloak-operator/pkg/apis/v1/v1alpha1"
	v1 "k8s.io/api/admission/v1"
	"strings"
	"testing"
	"time"

	jenkinsV1Api "github.com/epam/edp-jenkins-operator/v2/pkg/apis/v2/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	tMock "github.com/stretchr/testify/mock"
	coreV1Api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	cMock "github.com/epam/edp-sonar-operator/v2/mocks/client"
	pMock "github.com/epam/edp-sonar-operator/v2/mocks/platform"
	"github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1alpha1"
	sonarClient "github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
)

const (
	namespace = "ns"
	name      = "name"
	url       = "https://domain"
	edpWay    = "EDP way"
	basePath  = "path"
	template  = "EDP default"
)

func createSonarInstance() v1alpha1.Sonar {
	return v1alpha1.Sonar{
		ObjectMeta: metav1.ObjectMeta{
			Name:      main,
			Namespace: namespace,
		},
		Spec: v1alpha1.SonarSpec{
			DefaultPermissionTemplate: template,
		},
	}
}

func TestSonarServiceImpl_DeleteResource(t *testing.T) {
	secret := coreV1Api.Secret{ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "ns"}}
	s := Service{
		k8sClient: fake.NewClientBuilder().WithRuntimeObjects(&secret).Build(),
	}

	if _, err := s.DeleteResource(context.Background(), &secret, "fin", func() error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	secret.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	secret.Finalizers = []string{"fin"}
	s.k8sClient = fake.NewClientBuilder().WithRuntimeObjects(&secret).Build()

	if _, err := s.DeleteResource(context.Background(), &secret, "fin", func() error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestServiceMock_Configure(t *testing.T) {
	ctx := context.Background()
	sch := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}
	if err := coreV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}
	if err := jenkinsV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}

	snr := v1alpha1.Sonar{ObjectMeta: metav1.ObjectMeta{
		Namespace: "ns", Name: "snr1",
	}, Spec: v1alpha1.SonarSpec{DefaultPermissionTemplate: "tpl123"}}

	jns := jenkinsV1Api.Jenkins{Spec: jenkinsV1Api.JenkinsSpec{BasePath: "zabagdo"}, ObjectMeta: metav1.ObjectMeta{
		Name: "js1", Namespace: snr.Namespace,
	}}
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	s := Service{
		k8sClient:       fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(&jns).Build(),
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar,
			useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}

	adminSecret := coreV1Api.Secret{Data: map[string][]byte{
		"password": []byte("pwd123"),
	}}

	plMock.On("CreateSecret", snr.Name, snr.Namespace,
		fmt.Sprintf("%s-admin-password", snr.Name), tMock.AnythingOfType("map[string][]uint8")).Return(&adminSecret, nil)
	plMock.On("SetOwnerReference", &snr, &adminSecret).Return(nil)
	clMock.On("ChangePassword", ctx, "admin", "admin", "pwd123").Return(nil)
	clMock.On("InstallPlugins",
		[]string{"authoidc", "checkstyle", "findbugs", "pmd", "jacoco", "xml", "javascript", "go", "ansible", "yaml",
			"python", "csharp", "groovy"}).Return(nil)
	clMock.On("UploadProfile", "EDP way").
		Return("profile123", nil)
	clMock.On("CreateQualityGate", "EDP way").Return("qg1", nil)
	clMock.On("GetGroup", ctx, nonInteractiveGroupName).Return(nil, sonarClient.ErrNotFound("not found"))
	clMock.On("GetGroup", ctx, sonarDevelopersGroupName).Return(nil, sonarClient.ErrNotFound("not found"))
	clMock.On("CreateGroup", ctx, &sonarClient.Group{Name: nonInteractiveGroupName}).Return(nil)
	clMock.On("CreateGroup", ctx, &sonarClient.Group{Name: sonarDevelopersGroupName}).Return(nil)
	clMock.On("AddPermissionsToGroup", nonInteractiveGroupName, "scan").Return(nil)
	clMock.On("AddWebhook", "jenkins",
		"http://jenkins.ns:8080/zabagdo/sonarqube-webhook/").Return(nil)
	clMock.On("ConfigureGeneralSettings", "values", "sonar.typescript.lcov.reportPaths",
		"coverage/lcov.info").Return(nil)
	clMock.On("ConfigureGeneralSettings", "values", "sonar.coverage.jacoco.xmlReportPaths",
		"target/site/jacoco/jacoco.xml").Return(nil)
	clMock.On("SetDefaultPermissionTemplate", ctx, snr.Spec.DefaultPermissionTemplate).Return(nil)

	if err := s.Configure(ctx, &snr); err != nil {
		t.Fatalf("%+v", err)
	}
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestServiceMock_Configure_FailGetGroupForCreation(t *testing.T) {
	ctx := context.Background()
	sch := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}
	if err := coreV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}
	if err := jenkinsV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}

	snr := v1alpha1.Sonar{ObjectMeta: metav1.ObjectMeta{
		Namespace: "ns", Name: "snr1",
	}, Spec: v1alpha1.SonarSpec{DefaultPermissionTemplate: "tpl123"}}

	jns := jenkinsV1Api.Jenkins{Spec: jenkinsV1Api.JenkinsSpec{BasePath: "zabagdo"}, ObjectMeta: metav1.ObjectMeta{
		Name: "js1", Namespace: snr.Namespace,
	}}
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	s := Service{
		k8sClient:       fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(&jns).Build(),
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar,
			useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}

	adminSecret := coreV1Api.Secret{Data: map[string][]byte{
		"password": []byte("pwd123"),
	}}

	plMock.On("CreateSecret", snr.Name, snr.Namespace,
		fmt.Sprintf("%s-admin-password", snr.Name), tMock.AnythingOfType("map[string][]uint8")).Return(&adminSecret, nil)
	plMock.On("SetOwnerReference", &snr, &adminSecret).Return(nil)
	plMock.On("GetExternalEndpoint", snr.Namespace, snr.Name).Return("url", nil)
	clMock.On("ChangePassword", ctx, "admin", "admin", "pwd123").Return(nil)
	clMock.On("InstallPlugins",
		[]string{"authoidc", "checkstyle", "findbugs", "pmd", "jacoco", "xml", "javascript", "go", "ansible", "yaml",
			"python", "csharp", "groovy"}).Return(nil)
	clMock.On("UploadProfile", "EDP way").
		Return("profile123", nil)
	clMock.On("CreateQualityGate", "EDP way").Return("qg1", nil)
	clMock.On("GetGroup", ctx, nonInteractiveGroupName).Return(nil, sonarClient.ErrNotFound("not found"))
	clMock.On("GetGroup", ctx, sonarDevelopersGroupName).Return(nil, errors.New("FATAL:GETGROUPS"))
	clMock.On("CreateGroup", ctx, &sonarClient.Group{Name: nonInteractiveGroupName}).Return(nil)

	err := s.Configure(ctx, &snr)
	assert.Error(t, err)
	if !strings.Contains(err.Error(), "unexpected error during group check: FATAL:GETGROUPS") {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestServiceMock_Configure_FailCreateGroup(t *testing.T) {
	ctx := context.Background()
	sch := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}
	if err := coreV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}
	if err := jenkinsV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}

	snr := v1alpha1.Sonar{ObjectMeta: metav1.ObjectMeta{
		Namespace: "ns", Name: "snr1",
	}, Spec: v1alpha1.SonarSpec{DefaultPermissionTemplate: "tpl123"}}

	jns := jenkinsV1Api.Jenkins{Spec: jenkinsV1Api.JenkinsSpec{BasePath: "zabagdo"}, ObjectMeta: metav1.ObjectMeta{
		Name: "js1", Namespace: snr.Namespace,
	}}
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	s := Service{
		k8sClient:       fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(&jns).Build(),
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar,
			useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}

	adminSecret := coreV1Api.Secret{Data: map[string][]byte{
		"password": []byte("pwd123"),
	}}

	plMock.On("CreateSecret", snr.Name, snr.Namespace,
		fmt.Sprintf("%s-admin-password", snr.Name), tMock.AnythingOfType("map[string][]uint8")).Return(&adminSecret, nil)
	plMock.On("SetOwnerReference", &snr, &adminSecret).Return(nil)
	plMock.On("GetExternalEndpoint", snr.Namespace, snr.Name).Return("url", nil)
	clMock.On("ChangePassword", ctx, "admin", "admin", "pwd123").Return(nil)
	clMock.On("InstallPlugins",
		[]string{"authoidc", "checkstyle", "findbugs", "pmd", "jacoco", "xml", "javascript", "go", "ansible", "yaml",
			"python", "csharp", "groovy"}).Return(nil)
	clMock.On("UploadProfile", "EDP way").
		Return("profile123", nil)
	clMock.On("CreateQualityGate", "EDP way").Return("qg1", nil)
	clMock.On("GetGroup", ctx, nonInteractiveGroupName).Return(nil, sonarClient.ErrNotFound("not found"))
	clMock.On("GetGroup", ctx, sonarDevelopersGroupName).Return(nil, sonarClient.ErrNotFound("not found"))
	clMock.On("CreateGroup", ctx, &sonarClient.Group{Name: nonInteractiveGroupName}).Return(errors.New("FATAL:CREATE"))

	err := s.Configure(ctx, &snr)
	assert.Error(t, err)
	if !strings.Contains(err.Error(), "Failed to create non-interactive-users group!: FATAL:CREATE") {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestServiceMock_Configure_FailAddPermissions(t *testing.T) {
	ctx := context.Background()
	sch := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}
	if err := coreV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}
	if err := jenkinsV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}

	snr := v1alpha1.Sonar{ObjectMeta: metav1.ObjectMeta{
		Namespace: "ns", Name: "snr1",
	}, Spec: v1alpha1.SonarSpec{DefaultPermissionTemplate: "tpl123"}}

	jns := jenkinsV1Api.Jenkins{Spec: jenkinsV1Api.JenkinsSpec{BasePath: "zabagdo"}, ObjectMeta: metav1.ObjectMeta{
		Name: "js1", Namespace: snr.Namespace,
	}}
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	s := Service{
		k8sClient:       fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(&jns).Build(),
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar,
			useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}

	adminSecret := coreV1Api.Secret{Data: map[string][]byte{
		"password": []byte("pwd123"),
	}}

	plMock.On("CreateSecret", snr.Name, snr.Namespace,
		fmt.Sprintf("%s-admin-password", snr.Name), tMock.AnythingOfType("map[string][]uint8")).Return(&adminSecret, nil)
	plMock.On("SetOwnerReference", &snr, &adminSecret).Return(nil)
	plMock.On("GetExternalEndpoint", snr.Namespace, snr.Name).Return("url", nil)
	clMock.On("ChangePassword", ctx, "admin", "admin", "pwd123").Return(nil)
	clMock.On("InstallPlugins",
		[]string{"authoidc", "checkstyle", "findbugs", "pmd", "jacoco", "xml", "javascript", "go", "ansible", "yaml",
			"python", "csharp", "groovy"}).Return(nil)
	clMock.On("UploadProfile", "EDP way").
		Return("profile123", nil)
	clMock.On("CreateQualityGate", "EDP way").Return("qg1", nil)
	clMock.On("GetGroup", ctx, nonInteractiveGroupName).Return(nil, sonarClient.ErrNotFound("not found"))
	clMock.On("GetGroup", ctx, sonarDevelopersGroupName).Return(nil, sonarClient.ErrNotFound("not found"))
	clMock.On("CreateGroup", ctx, &sonarClient.Group{Name: nonInteractiveGroupName}).Return(nil)
	clMock.On("CreateGroup", ctx, &sonarClient.Group{Name: sonarDevelopersGroupName}).Return(nil)
	clMock.On("AddPermissionsToGroup", nonInteractiveGroupName, "scan").Return(errors.New("FATAL:ADDPERM"))

	err := s.Configure(ctx, &snr)
	assert.Error(t, err)
	if !strings.Contains(err.Error(), "Failed to add scan permission for non-interactive-users group!: FATAL:ADDPERM") {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestService_Integration_BadBuilder(t *testing.T) {
	ctx := context.Background()
	instance := v1alpha1.Sonar{}
	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return nil, errors.New("test")
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.Error(t, err)
}

func TestService_Integration_getKeycloakRealmErr(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()

	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	service := Service{
		k8sClient: client,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.Error(t, err)
}

func TestService_Integration_NilRealmAnnotation(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{})
	keycloakRealmInstance := keycloakApi.KeycloakRealm{
		ObjectMeta: metav1.ObjectMeta{
			Name:      main,
			Namespace: namespace,
		},
	}
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&keycloakRealmInstance).Build()

	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	service := Service{
		k8sClient: client,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "realm main does not have required annotations"))
}

func TestService_Integration_EmptyAnnotationErr(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{})
	keycloakRealmInstance := keycloakApi.KeycloakRealm{
		ObjectMeta: metav1.ObjectMeta{
			Name:        main,
			Namespace:   namespace,
			Annotations: map[string]string{annotation: ""},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&keycloakRealmInstance).Build()

	ctx := context.Background()
	instance := v1alpha1.Sonar{
		ObjectMeta: metav1.ObjectMeta{
			Name:      main,
			Namespace: namespace,
		},
	}
	clMock := cMock.ClientInterface{}
	service := Service{
		k8sClient: client,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to unmarshal OpenID configuration"))
}

func TestService_Integration_ConfigureGeneralSettingsErr(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{})
	data := map[string]string{"issuer": name}
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal()
	}

	keycloakRealmInstance := keycloakApi.KeycloakRealm{
		ObjectMeta: metav1.ObjectMeta{
			Name:        main,
			Namespace:   namespace,
			Annotations: map[string]string{annotation: string(raw)},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&keycloakRealmInstance).Build()

	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	errTest := errors.New("test")
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.issuerUri", name).Return(errTest)

	service := Service{
		k8sClient: client,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err = service.Integration(ctx, instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to to configure sonar.auth.oidc.issuerUri"))
	clMock.AssertExpectations(t)
}

func TestService_Integration_EmptyAnnotation(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{})
	data := map[string]string{"issuer": ""}
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal()
	}

	keycloakRealmInstance := keycloakApi.KeycloakRealm{
		ObjectMeta: metav1.ObjectMeta{
			Name:        main,
			Namespace:   namespace,
			Annotations: map[string]string{annotation: string(raw)},
		},
	}
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&keycloakRealmInstance).Build()

	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}

	service := Service{
		k8sClient: client,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err = service.Integration(ctx, instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "issuer field in oidc configuration is empty or configuration is invalid"))
	clMock.AssertExpectations(t)
}

func TestService_Integration_GetExternalEndpointErr(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()
	errTest := errors.New("test")

	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}
	plMock.On("GetExternalEndpoint", ctx, namespace, main).Return("", errTest)
	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.Equal(t, errTest, err)
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Integration_ConfigureGeneralSettingsErr2(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()
	errTest := errors.New("test")

	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}
	plMock.On("GetExternalEndpoint", ctx, namespace, main).Return(url, nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.core.serverBaseURL", url).Return(errTest)
	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to configure sonar.core.serverBaseURL!"))
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Integration_getKeycloakClientErr(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()

	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}
	plMock.On("GetExternalEndpoint", ctx, namespace, main).Return(url, nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.core.serverBaseURL", url).Return(nil)
	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.Error(t, err)
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Integration_ConfigureGeneralSettingsErr3(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{}, &keycloakApi.KeycloakClient{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()

	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}
	plMock.On("GetExternalEndpoint", ctx, namespace, main).Return(url, nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.core.serverBaseURL", url).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.clientId.secured", instance.Name).Return(errTest)
	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to configure sonar.auth.oidc.clientId.secured!"))
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Integration_ConfigureGeneralSettingsErr4(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{}, &keycloakApi.KeycloakClient{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()

	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}
	plMock.On("GetExternalEndpoint", ctx, namespace, main).Return(url, nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.core.serverBaseURL", url).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.clientId.secured", instance.Name).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.groupsSync.claimName", claimName).Return(errTest)
	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to configure sonar.auth.oidc.groupsSync.claimName!"))
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Integration_ConfigureGeneralSettingsErr5(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{}, &keycloakApi.KeycloakClient{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()

	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}
	plMock.On("GetExternalEndpoint", ctx, namespace, main).Return(url, nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.core.serverBaseURL", url).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.clientId.secured", instance.Name).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.groupsSync.claimName", claimName).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.groupsSync", "true").Return(errTest)
	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to configure sonar.auth.oidc.groupsSync!"))
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Integration_ConfigureGeneralSettingsErr6(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{}, &keycloakApi.KeycloakClient{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()

	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}
	plMock.On("GetExternalEndpoint", ctx, namespace, main).Return(url, nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.core.serverBaseURL", url).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.clientId.secured", instance.Name).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.groupsSync.claimName", claimName).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.groupsSync", "true").Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.enabled", "true").Return(errTest)
	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to configure sonar.auth.oidc.enabled!"))
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Integration_SetProjectsDefaultVisibilityErr(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{}, &keycloakApi.KeycloakClient{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()

	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}
	plMock.On("GetExternalEndpoint", ctx, namespace, main).Return(url, nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.core.serverBaseURL", url).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.clientId.secured", instance.Name).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.groupsSync.claimName", claimName).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.groupsSync", "true").Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.enabled", "true").Return(nil)
	clMock.On("SetProjectsDefaultVisibility", "private").Return(errTest)
	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "couldn't set default"))
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Integration(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &v1alpha1.Sonar{}, &keycloakApi.KeycloakRealm{}, &keycloakApi.KeycloakClient{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()

	ctx := context.Background()
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}
	plMock.On("GetExternalEndpoint", ctx, namespace, main).Return(url, nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.core.serverBaseURL", url).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.clientId.secured", instance.Name).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.groupsSync.claimName", claimName).Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.groupsSync", "true").Return(nil)
	clMock.On("ConfigureGeneralSettings", "value", "sonar.auth.oidc.enabled", "true").Return(nil)
	clMock.On("SetProjectsDefaultVisibility", "private").Return(nil)
	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, instance)
	assert.NoError(t, err)
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_IsDeploymentReady_Err(t *testing.T) {
	instance := createSonarInstance()
	plMock := pMock.Service{}
	errTest := errors.New("test")
	plMock.On("GetAvailableDeploymentReplicas", &instance).Return(nil, errTest)
	service := Service{
		platformService: &plMock,
	}
	ready, err := service.IsDeploymentReady(&instance)
	assert.Error(t, err)
	assert.False(t, ready)
	plMock.AssertExpectations(t)
}

func TestService_IsDeploymentReady_False(t *testing.T) {
	instance := createSonarInstance()
	plMock := pMock.Service{}
	val := 0
	plMock.On("GetAvailableDeploymentReplicas", &instance).Return(&val, nil)
	service := Service{
		platformService: &plMock,
	}
	ready, err := service.IsDeploymentReady(&instance)
	assert.NoError(t, err)
	assert.False(t, ready)
	plMock.AssertExpectations(t)
}

func TestService_IsDeploymentReady_True(t *testing.T) {
	instance := createSonarInstance()
	plMock := pMock.Service{}
	val := 1
	plMock.On("GetAvailableDeploymentReplicas", &instance).Return(&val, nil)
	service := Service{
		platformService: &plMock,
	}
	ready, err := service.IsDeploymentReady(&instance)
	assert.NoError(t, err)
	assert.True(t, ready)
	plMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_BadBuilder(t *testing.T) {
	ctx := context.Background()
	instance := createSonarInstance()

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return nil, errors.New("test")
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to initialize Sonar Client!"))
}

func TestService_ExposeConfiguration_CreateUserErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	clMock.On("GetUser", ctx, jenkinsLogin).Return(nil, sonarClient.ErrNotFound("test"))
	clMock.On("CreateUser", ctx, tMock.MatchedBy(func(sonar *sonarClient.User) bool {
		return sonar.Name == jenkinsUsername && sonar.Login == jenkinsLogin
	})).Return(errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to create user"))
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_GetUserErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	clMock.On("GetUser", ctx, jenkinsLogin).Return(nil, errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unexpected error during get user"))
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_AddUserToGroupErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	clMock.On("GetUser", ctx, jenkinsLogin).Return(nil, nil)
	clMock.On("AddUserToGroup", nonInteractiveGroupName, jenkinsLogin).Return(errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to add"))
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_AddPermissionsToUserErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	clMock.On("GetUser", ctx, jenkinsLogin).Return(nil, nil)
	clMock.On("AddUserToGroup", nonInteractiveGroupName, jenkinsLogin).Return(nil)
	clMock.On("AddPermissionsToUser", jenkinsLogin, admin).Return(errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to add admin permissions to"))
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_GetUserTokenErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	clMock.On("GetUser", ctx, jenkinsLogin).Return(nil, nil)
	clMock.On("AddUserToGroup", nonInteractiveGroupName, jenkinsLogin).Return(nil)
	clMock.On("AddPermissionsToUser", jenkinsLogin, admin).Return(nil)
	clMock.On("GetUserToken", ctx, jenkinsLogin, strings.Title(jenkinsLogin)).Return(nil, errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unexpected error during get user token for user"))
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_GenerateUserTokenErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	clMock.On("GetUser", ctx, jenkinsLogin).Return(nil, nil)
	clMock.On("AddUserToGroup", nonInteractiveGroupName, jenkinsLogin).Return(nil)
	clMock.On("AddPermissionsToUser", jenkinsLogin, admin).Return(nil)
	clMock.On("GetUserToken", ctx, jenkinsLogin, strings.Title(jenkinsLogin)).Return(nil, sonarClient.ErrNotFound("test"))
	clMock.On("GenerateUserToken", jenkinsLogin).Return(nil, errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to generate token for"))
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_CreateSecretErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	ciUserName := fmt.Sprintf("%v-ciuser-token", main)
	ciToken := name
	ciSecret := map[string][]byte{
		"username": []byte(jenkinsLogin),
		"secret":   []byte(ciToken),
	}
	instance := createSonarInstance()
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	clMock.On("GetUser", ctx, jenkinsLogin).Return(nil, nil)
	clMock.On("AddUserToGroup", nonInteractiveGroupName, jenkinsLogin).Return(nil)
	clMock.On("AddPermissionsToUser", jenkinsLogin, admin).Return(nil)
	clMock.On("GetUserToken", ctx, jenkinsLogin, strings.Title(jenkinsLogin)).Return(nil, sonarClient.ErrNotFound("test"))
	clMock.On("GenerateUserToken", jenkinsLogin).Return(&ciToken, nil)
	plMock.On("CreateSecret", main, namespace, ciUserName, ciSecret).Return(nil, errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to create secret for"))
	clMock.AssertExpectations(t)
	plMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_SetOwnerReferenceErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	ciUserName := fmt.Sprintf("%v-ciuser-token", main)
	ciToken := name
	ciSecret := map[string][]byte{
		"username": []byte(jenkinsLogin),
		"secret":   []byte(ciToken),
	}
	secret := coreV1Api.Secret{}
	instance := createSonarInstance()
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	clMock.On("GetUser", ctx, jenkinsLogin).Return(nil, nil)
	clMock.On("AddUserToGroup", nonInteractiveGroupName, jenkinsLogin).Return(nil)
	clMock.On("AddPermissionsToUser", jenkinsLogin, admin).Return(nil)
	clMock.On("GetUserToken", ctx, jenkinsLogin, strings.Title(jenkinsLogin)).Return(nil, sonarClient.ErrNotFound("test"))
	clMock.On("GenerateUserToken", jenkinsLogin).Return(&ciToken, nil)
	plMock.On("CreateSecret", main, namespace, ciUserName, ciSecret).Return(&secret, nil)
	plMock.On("SetOwnerReference", &instance, &secret).Return(errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to set owner reference for secret"))
	clMock.AssertExpectations(t)
	plMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_CreateJenkinsServiceAccountErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	ciUserName := fmt.Sprintf("%v-ciuser-token", main)
	instance := createSonarInstance()

	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}

	clMock.On("GetUser", ctx, jenkinsLogin).Return(nil, nil)
	clMock.On("AddUserToGroup", nonInteractiveGroupName, jenkinsLogin).Return(nil)
	clMock.On("AddPermissionsToUser", jenkinsLogin, admin).Return(nil)
	clMock.On("GetUserToken", ctx, jenkinsLogin, strings.Title(jenkinsLogin)).Return(nil, nil)
	plMock.On("CreateJenkinsServiceAccount", instance.Namespace, ciUserName, tokenType).Return(errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
		platformService: &plMock,
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to create Jenkins Service Account for"))
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_ParseDefaultTemplateErr(t *testing.T) {
	ctx := context.Background()
	ciUserName := fmt.Sprintf("%v-ciuser-token", main)
	instance := createSonarInstance()

	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}

	clMock.On("GetUser", ctx, jenkinsLogin).Return(nil, nil)
	clMock.On("AddUserToGroup", nonInteractiveGroupName, jenkinsLogin).Return(nil)
	clMock.On("AddPermissionsToUser", jenkinsLogin, admin).Return(nil)
	clMock.On("GetUserToken", ctx, jenkinsLogin, strings.Title(jenkinsLogin)).Return(nil, nil)
	plMock.On("CreateJenkinsServiceAccount", instance.Namespace, ciUserName, tokenType).Return(nil)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
		platformService: &plMock,
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to parse default Jenkins plugin template"))
	clMock.AssertExpectations(t)
	plMock.AssertExpectations(t)
}

func TestService_Configure_configurePasswordErr(t *testing.T) {
	ctx := context.Background()
	instance := createSonarInstance()
	errTest := errors.New("test")
	plMock := pMock.Service{}

	plMock.On("CreateSecret", instance.Name, instance.Namespace,
		fmt.Sprintf("%s-admin-password", instance.Name),
		tMock.AnythingOfType("map[string][]uint8")).Return(nil, errTest)

	service := Service{platformService: &plMock}
	err := service.Configure(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unable to setup password for sonar"))
	plMock.AssertExpectations(t)
}

func TestService_Configure_BadBuilder(t *testing.T) {
	secret := coreV1Api.Secret{}
	ctx := context.Background()
	instance := createSonarInstance()
	plMock := pMock.Service{}

	plMock.On("CreateSecret", instance.Name, instance.Namespace,
		fmt.Sprintf("%s-admin-password", instance.Name),
		tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	plMock.On("SetOwnerReference", &instance, &secret).Return(nil)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return nil, errors.New("test")
		},
	}
	err := service.Configure(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to initialize Sonar Client!"))
	plMock.AssertExpectations(t)
}

func TestService_Configure_installPluginsErr(t *testing.T) {
	data := map[string][]byte{"password": []byte(defaultPassword)}
	secret := coreV1Api.Secret{Data: data}
	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	plMock.On("CreateSecret", instance.Name, instance.Namespace,
		fmt.Sprintf("%s-admin-password", instance.Name),
		tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	plMock.On("SetOwnerReference", &instance, &secret).Return(nil)
	clMock.On("ChangePassword", ctx, admin, defaultPassword, defaultPassword).Return(nil)
	clMock.On("InstallPlugins", tMock.AnythingOfType("[]string")).Return(errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.Configure(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unable to install plugins"))
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Configure_uploadProfileErr(t *testing.T) {
	data := map[string][]byte{"password": []byte(defaultPassword)}
	secret := coreV1Api.Secret{Data: data}
	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	plMock.On("CreateSecret", instance.Name, instance.Namespace,
		fmt.Sprintf("%s-admin-password", instance.Name),
		tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	plMock.On("SetOwnerReference", &instance, &secret).Return(nil)
	clMock.On("ChangePassword", ctx, admin, defaultPassword, defaultPassword).Return(nil)
	clMock.On("InstallPlugins", tMock.AnythingOfType("[]string")).Return(nil)
	clMock.On("UploadProfile", "EDP way").Return("", errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.Configure(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unable to upload profile"))
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Configure_createQualityGateErr(t *testing.T) {
	data := map[string][]byte{"password": []byte(defaultPassword)}
	secret := coreV1Api.Secret{Data: data}
	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	plMock.On("CreateSecret", instance.Name, instance.Namespace,
		fmt.Sprintf("%s-admin-password", instance.Name),
		tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	plMock.On("SetOwnerReference", &instance, &secret).Return(nil)
	clMock.On("ChangePassword", ctx, admin, defaultPassword, defaultPassword).Return(nil)
	clMock.On("InstallPlugins", tMock.AnythingOfType("[]string")).Return(nil)
	clMock.On("UploadProfile", edpWay).Return("", nil)
	clMock.On("CreateQualityGate", "EDP way").Return("", errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.Configure(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "Failed to configure EDP way quality gate!"))
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Configure_setupWebhookErr(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	jns := jenkinsV1Api.Jenkins{Spec: jenkinsV1Api.JenkinsSpec{BasePath: basePath}, ObjectMeta: metav1.ObjectMeta{
		Name: name, Namespace: namespace,
	}}
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &jenkinsV1Api.JenkinsList{}, &jenkinsV1Api.Jenkins{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&jns).Build()
	data := map[string][]byte{"password": []byte(defaultPassword)}
	secret := coreV1Api.Secret{Data: data}
	errTest := errors.New("test")
	instance := createSonarInstance()
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	plMock.On("CreateSecret", instance.Name, instance.Namespace,
		fmt.Sprintf("%s-admin-password", instance.Name),
		tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	plMock.On("SetOwnerReference", &instance, &secret).Return(nil)
	clMock.On("ChangePassword", ctx, admin, defaultPassword, defaultPassword).Return(nil)
	clMock.On("InstallPlugins", tMock.AnythingOfType("[]string")).Return(nil)
	clMock.On("UploadProfile", edpWay).Return("", nil)
	clMock.On("CreateQualityGate", "EDP way").Return("", nil)
	clMock.On("GetGroup", ctx, nonInteractiveGroupName).Return(nil, nil)
	clMock.On("GetGroup", ctx, sonarDevelopersGroupName).Return(nil, nil)
	clMock.On("AddPermissionsToGroup", nonInteractiveGroupName, "scan").Return(nil)
	clMock.On("AddWebhook", jenkinsLogin,
		"http://jenkins.ns:8080/"+basePath+"/sonarqube-webhook/").Return(errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
		k8sClient: client,
		k8sScheme: scheme,
	}
	err := service.Configure(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unable to setup webhook"))
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Configure_configureGeneralSettingsErr(t *testing.T) {
	scheme := runtime.NewScheme()
	jns := jenkinsV1Api.Jenkins{Spec: jenkinsV1Api.JenkinsSpec{BasePath: basePath}, ObjectMeta: metav1.ObjectMeta{
		Name: name, Namespace: namespace,
	}}
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &jenkinsV1Api.JenkinsList{}, &jenkinsV1Api.Jenkins{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&jns).Build()
	data := map[string][]byte{"password": []byte(defaultPassword)}
	secret := coreV1Api.Secret{Data: data}
	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	plMock.On("CreateSecret", instance.Name, instance.Namespace,
		fmt.Sprintf("%s-admin-password", instance.Name),
		tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	plMock.On("SetOwnerReference", &instance, &secret).Return(nil)
	clMock.On("ChangePassword", ctx, admin, defaultPassword, defaultPassword).Return(nil)
	clMock.On("InstallPlugins", tMock.AnythingOfType("[]string")).Return(nil)
	clMock.On("UploadProfile", edpWay).Return("", nil)
	clMock.On("CreateQualityGate", "EDP way").Return("", nil)
	clMock.On("GetGroup", ctx, nonInteractiveGroupName).Return(nil, nil)
	clMock.On("GetGroup", ctx, sonarDevelopersGroupName).Return(nil, nil)
	clMock.On("AddPermissionsToGroup", nonInteractiveGroupName, "scan").Return(nil)
	clMock.On("AddWebhook", jenkinsLogin,
		"http://jenkins.ns:8080/"+basePath+"/sonarqube-webhook/").Return(nil)
	clMock.On("ConfigureGeneralSettings", "values", "sonar.typescript.lcov.reportPaths",
		"coverage/lcov.info").Return(errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
		k8sClient: client,
		k8sScheme: scheme,
	}
	err := service.Configure(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unable to configure general settings"))
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Configure_setDefaultPermissionTemplateErr(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	jns := jenkinsV1Api.Jenkins{Spec: jenkinsV1Api.JenkinsSpec{BasePath: basePath}, ObjectMeta: metav1.ObjectMeta{
		Name: name, Namespace: namespace,
	}}
	scheme.AddKnownTypes(v1.SchemeGroupVersion, &jenkinsV1Api.JenkinsList{}, &jenkinsV1Api.Jenkins{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&jns).Build()
	data := map[string][]byte{"password": []byte(defaultPassword)}
	secret := coreV1Api.Secret{Data: data}
	errTest := errors.New("test")
	instance := createSonarInstance()
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	plMock.On("CreateSecret", instance.Name, instance.Namespace,
		fmt.Sprintf("%s-admin-password", instance.Name),
		tMock.AnythingOfType("map[string][]uint8")).Return(&secret, nil)
	plMock.On("SetOwnerReference", &instance, &secret).Return(nil)
	clMock.On("ChangePassword", ctx, admin, defaultPassword, defaultPassword).Return(nil)
	clMock.On("InstallPlugins", tMock.AnythingOfType("[]string")).Return(nil)
	clMock.On("UploadProfile", edpWay).Return("", nil)
	clMock.On("CreateQualityGate", "EDP way").Return("", nil)
	clMock.On("GetGroup", ctx, nonInteractiveGroupName).Return(nil, nil)
	clMock.On("GetGroup", ctx, sonarDevelopersGroupName).Return(nil, nil)
	clMock.On("AddPermissionsToGroup", nonInteractiveGroupName, "scan").Return(nil)
	clMock.On("AddWebhook", jenkinsLogin,
		"http://jenkins.ns:8080/"+basePath+"/sonarqube-webhook/").Return(nil)
	clMock.On("ConfigureGeneralSettings", "values", "sonar.typescript.lcov.reportPaths",
		"coverage/lcov.info").Return(nil)
	clMock.On("ConfigureGeneralSettings", "values", "sonar.coverage.jacoco.xmlReportPaths",
		"target/site/jacoco/jacoco.xml").Return(nil)
	clMock.On("SetDefaultPermissionTemplate", ctx, instance.Spec.DefaultPermissionTemplate).Return(errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
			return &clMock, nil
		},
		k8sClient: client,
		k8sScheme: scheme,
	}
	err := service.Configure(ctx, &instance)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unable to set default permission template"))
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}
