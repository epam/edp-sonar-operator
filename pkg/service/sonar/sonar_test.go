package sonar

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	tMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	admissionV1 "k8s.io/api/admission/v1"
	coreV1Api "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	jenkinsV1Api "github.com/epam/edp-jenkins-operator/v2/pkg/apis/v2/v1"
	keycloakApi "github.com/epam/edp-keycloak-operator/api/v1"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	cMock "github.com/epam/edp-sonar-operator/mocks/client"
	pMock "github.com/epam/edp-sonar-operator/mocks/platform"
	sonarClient "github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

const (
	namespace = "ns"
	name      = "name"
	basePath  = "path"
	template  = "EDP default"
	sonarURL  = "http://sonarqube.com"
)

func createSonarInstance() sonarApi.Sonar {
	return sonarApi.Sonar{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      main,
			Namespace: namespace,
		},
		Spec: sonarApi.SonarSpec{
			DefaultPermissionTemplate: template,
		},
	}
}

func plugins() []string {
	return []string{
		"authoidc", "checkstyle", "findbugs", "pmd", "jacoco", "xml", "javascript", "go", "ansible",
		"yaml", "python", "csharp", "groovy",
	}
}

func returnTrue() bool {
	return true
}

func TestSonarServiceImpl_DeleteResource(t *testing.T) {
	secret := coreV1Api.Secret{ObjectMeta: metaV1.ObjectMeta{Name: "name", Namespace: "ns"}}
	s := Service{
		k8sClient: fake.NewClientBuilder().WithRuntimeObjects(&secret).Build(),
	}

	if _, err := s.DeleteResource(context.Background(), &secret, "fin", func() error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	secret.DeletionTimestamp = &metaV1.Time{Time: time.Now()}
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

	err := sonarApi.AddToScheme(sch)
	require.NoError(t, err)

	err = coreV1Api.AddToScheme(sch)
	require.NoError(t, err)

	err = jenkinsV1Api.AddToScheme(sch)
	require.NoError(t, err)

	snr := sonarApi.Sonar{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: "ns",
			Name:      "snr1",
		},
		Spec: sonarApi.SonarSpec{
			DefaultPermissionTemplate: "tpl123",
			Groups: []sonarApi.Group{
				{
					Name:        "non-interactive-users",
					Permissions: []string{"scan"},
				},
				{
					Name: "sonar-developers",
				},
			},
			Plugins: []string{"authoidc", "checkstyle", "findbugs", "pmd"},
			Users: []sonarApi.User{
				{
					Login:       "ci-user",
					Username:    "EDP CI User",
					Group:       "non-interactive-users",
					Permissions: []string{"admin"},
				},
			},
			QualityGates: []sonarApi.QualityGate{
				{
					Name:         "gate1",
					SetAsDefault: false,
					Conditions: []sonarApi.QualityGateCondition{
						{
							Error: "80", Metric: "new_coverage", OP: "LT", Period: "1",
						},
					},
				},
			},
			Settings: []sonarApi.SonarSetting{
				{
					Key:       "sonar.typescript.lcov.reportPaths",
					Value:     "coverage/lcov.info",
					ValueType: "values",
				},
				{
					Key:       "sonar.coverage.jacoco.xmlReportPaths",
					Value:     "target/site/jacoco/jacoco.xml",
					ValueType: "values",
				},
			},
		},
	}

	jns := jenkinsV1Api.Jenkins{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "js1",
			Namespace: snr.Namespace,
		},
		Spec: jenkinsV1Api.JenkinsSpec{
			BasePath: "zabagdo",
		},
	}

	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	s := Service{
		k8sClient:       fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(&jns).Build(),
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
		runningInClusterFunc: returnTrue,
	}

	clMock.On("InstallPlugins", snr.Spec.Plugins).Return(nil)
	clMock.On("CreateQualityGates", sonarClient.QualityGates{
		"gate1": sonarClient.QualityGateSettings{
			MakeDefault: false,
			Conditions: []sonarClient.QualityGateCondition{
				{Error: "80", Metric: "new_coverage", OP: "LT", Period: "1"},
			},
		},
	}).Return(nil)
	clMock.On("GetGroup", ctx, "non-interactive-users").Return(nil, sonarClient.NotFoundError("not found"))
	clMock.On("GetGroup", ctx, "sonar-developers").Return(nil, sonarClient.NotFoundError("not found"))
	clMock.On("CreateGroup", ctx, &sonarClient.Group{Name: "non-interactive-users"}).Return(nil)
	clMock.On("CreateGroup", ctx, &sonarClient.Group{Name: "sonar-developers"}).Return(nil)
	clMock.On("AddPermissionsToGroup", "non-interactive-users", "scan").Return(nil)
	clMock.On("ConfigureGeneralSettings",
		sonarClient.SettingRequest{Key: "sonar.typescript.lcov.reportPaths", Value: "coverage/lcov.info", ValueType: "values"},
		sonarClient.SettingRequest{Key: "sonar.coverage.jacoco.xmlReportPaths", Value: "target/site/jacoco/jacoco.xml", ValueType: "values"},
	).Return(nil)
	clMock.On("SetDefaultPermissionTemplate", ctx, snr.Spec.DefaultPermissionTemplate).Return(nil)

	err = s.Configure(ctx, &snr)
	require.NoError(t, err)

	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestServiceMock_Configure_FailGetGroupForCreation(t *testing.T) {
	ctx := context.Background()
	sch := runtime.NewScheme()
	if err := sonarApi.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}
	if err := coreV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}
	if err := jenkinsV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}

	snr := sonarApi.Sonar{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: "ns",
			Name:      "snr1",
		},
		Spec: sonarApi.SonarSpec{
			DefaultPermissionTemplate: "tpl123",
			Groups: []sonarApi.Group{
				{
					Name: "sonar-developers",
				},
			},
		},
	}

	jns := jenkinsV1Api.Jenkins{
		Spec: jenkinsV1Api.JenkinsSpec{
			BasePath: "zabagdo",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "js1",
			Namespace: snr.Namespace,
		},
	}

	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	s := Service{
		k8sClient:       fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(&jns).Build(),
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar,
		) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
		runningInClusterFunc: returnTrue,
	}

	adminSecret := coreV1Api.Secret{
		Data: map[string][]byte{
			"password": []byte("pwd123"),
		},
	}

	plMock.On("CreateSecret", snr.Name, snr.Namespace,
		fmt.Sprintf("%s-admin-password", snr.Name), tMock.AnythingOfType("map[string][]uint8")).Return(&adminSecret, nil)
	plMock.On("SetOwnerReference", &snr, &adminSecret).Return(nil)
	plMock.On("GetExternalEndpoint", snr.Namespace, snr.Name).Return("url", nil)
	clMock.On("ChangePassword", ctx, "admin", "admin", "pwd123").Return(nil)
	clMock.On("InstallPlugins", plugins()).Return(nil)
	clMock.On("UploadProfile", "EDP way", defaultProfileAbsolutePath).
		Return("profile123", nil)
	clMock.On("CreateQualityGates", sonarClient.QualityGates{}).Return(nil)
	clMock.On("GetGroup", ctx, "sonar-developers").Return(nil, errors.New("FATAL:GETGROUPS"))
	clMock.On("CreateGroup", ctx, &sonarClient.Group{Name: "non-interactive-users"}).Return(nil)

	err := s.Configure(ctx, &snr)
	assert.Error(t, err)
	require.Contains(t, err.Error(), "unexpected error during group check: FATAL:GETGROUPS")
}

func TestServiceMock_Configure_FailCreateGroup(t *testing.T) {
	ctx := context.Background()
	sch := runtime.NewScheme()

	if err := sonarApi.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}

	if err := coreV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}

	if err := jenkinsV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}

	snr := sonarApi.Sonar{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: "ns", Name: "snr1",
		},
		Spec: sonarApi.SonarSpec{
			DefaultPermissionTemplate: "tpl123",
			Groups: []sonarApi.Group{
				{
					Name: "sonar-developers",
				},
			},
		},
	}

	jns := jenkinsV1Api.Jenkins{Spec: jenkinsV1Api.JenkinsSpec{BasePath: "zabagdo"}, ObjectMeta: metaV1.ObjectMeta{
		Name: "js1", Namespace: snr.Namespace,
	}}
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	s := Service{
		k8sClient:       fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(&jns).Build(),
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar,
		) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
		runningInClusterFunc: returnTrue,
	}

	adminSecret := coreV1Api.Secret{Data: map[string][]byte{
		"password": []byte("pwd123"),
	}}

	plMock.On("CreateSecret", snr.Name, snr.Namespace,
		fmt.Sprintf("%s-admin-password", snr.Name), tMock.AnythingOfType("map[string][]uint8")).Return(&adminSecret, nil)
	plMock.On("SetOwnerReference", &snr, &adminSecret).Return(nil)
	plMock.On("GetExternalEndpoint", snr.Namespace, snr.Name).Return("url", nil)
	clMock.On("ChangePassword", ctx, "admin", "admin", "pwd123").Return(nil)
	clMock.On("InstallPlugins", plugins()).Return(nil)
	clMock.On("UploadProfile", "EDP way", defaultProfileAbsolutePath).
		Return("profile123", nil)
	clMock.On("CreateQualityGates", sonarClient.QualityGates{}).Return(nil)
	clMock.On("GetGroup", ctx, "sonar-developers").Return(nil, sonarClient.NotFoundError("not found"))
	clMock.On("CreateGroup", ctx, &sonarClient.Group{Name: "sonar-developers"}).Return(errors.New("FATAL:CREATE"))

	err := s.Configure(ctx, &snr)
	assert.Error(t, err)
}

func TestServiceMock_Configure_FailAddPermissions(t *testing.T) {
	ctx := context.Background()
	sch := runtime.NewScheme()
	if err := sonarApi.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}
	if err := coreV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}
	if err := jenkinsV1Api.AddToScheme(sch); err != nil {
		t.Fatal(err)
	}

	snr := sonarApi.Sonar{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: "ns",
			Name:      "snr1",
		},
		Spec: sonarApi.SonarSpec{
			DefaultPermissionTemplate: "tpl123",
			Groups: []sonarApi.Group{
				{
					Name:        "non-interactive-users",
					Permissions: []string{"scan"},
				},
				{
					Name: "sonar-developers",
				},
			},
		},
	}

	jns := jenkinsV1Api.Jenkins{Spec: jenkinsV1Api.JenkinsSpec{BasePath: "zabagdo"}, ObjectMeta: metaV1.ObjectMeta{
		Name: "js1", Namespace: snr.Namespace,
	}}
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	s := Service{
		k8sClient:       fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(&jns).Build(),
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
		runningInClusterFunc: returnTrue,
	}

	adminSecret := coreV1Api.Secret{Data: map[string][]byte{
		"password": []byte("pwd123"),
	}}

	plMock.On("CreateSecret", snr.Name, snr.Namespace,
		fmt.Sprintf("%s-admin-password", snr.Name), tMock.AnythingOfType("map[string][]uint8")).Return(&adminSecret, nil)
	plMock.On("SetOwnerReference", &snr, &adminSecret).Return(nil)
	plMock.On("GetExternalEndpoint", snr.Namespace, snr.Name).Return("url", nil)
	clMock.On("ChangePassword", ctx, "admin", "admin", "pwd123").Return(nil)
	clMock.On("InstallPlugins", plugins()).Return(nil)
	clMock.On("UploadProfile", "EDP way", defaultProfileAbsolutePath).
		Return("profile123", nil)
	clMock.On("CreateQualityGates", sonarClient.QualityGates{}).Return(nil)
	clMock.On("GetGroup", ctx, "non-interactive-users").Return(nil, sonarClient.NotFoundError("not found"))
	clMock.On("GetGroup", ctx, "sonar-developers").Return(nil, sonarClient.NotFoundError("not found"))
	clMock.On("CreateGroup", ctx, &sonarClient.Group{Name: "non-interactive-users"}).Return(nil)
	clMock.On("CreateGroup", ctx, &sonarClient.Group{Name: "sonar-developers"}).Return(nil)
	clMock.On("AddPermissionsToGroup", "non-interactive-users", "scan").Return(errors.New("FATAL:ADDPERM"))

	err := s.Configure(ctx, &snr)
	assert.Error(t, err)
	require.Contains(t, err.Error(), "failed to add scan permission for group non-interactive-users: FATAL:ADDPERM")
}

func TestService_Integration_BadBuilder(t *testing.T) {
	ctx := context.Background()
	instance := sonarApi.Sonar{}
	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return nil, errors.New("test")
		},
	}
	_, err := service.Integration(ctx, &instance)
	assert.Error(t, err)
}

func TestService_Integration_ConfigureGeneralSettingsErr(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(admissionV1.SchemeGroupVersion, &sonarApi.Sonar{})

	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	ctx := context.Background()
	instance := createSonarInstance()
	instance.Spec.Url = sonarURL
	clMock := cMock.ClientInterface{}
	errTest := errors.New("test")
	clMock.On("ConfigureGeneralSettings", sonarClient.SettingRequest{
		Key: "sonar.core.serverBaseURL", Value: sonarURL, ValueType: "value",
	}).
		Return(errTest)

	service := Service{
		k8sClient: client,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}
	_, err := service.Integration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to configure sonar.core.serverBaseURL")
	clMock.AssertExpectations(t)
}

func TestService_Integration_ConfigureGeneralSettingsErr2(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(admissionV1.SchemeGroupVersion, &sonarApi.Sonar{}, &keycloakApi.KeycloakRealm{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()
	errTest := errors.New("test")

	ctx := context.Background()
	instance := createSonarInstance()
	instance.Spec.Url = sonarURL
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}
	clMock.On("ConfigureGeneralSettings", sonarClient.SettingRequest{
		Key: "sonar.core.serverBaseURL", Value: sonarURL, ValueType: "value",
	}).
		Return(errTest)

	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}

	_, err := service.Integration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to configure sonar.core.serverBaseURL")
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Integration_ConfigureGeneralSettingsErr3(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(admissionV1.SchemeGroupVersion, &sonarApi.Sonar{}, &keycloakApi.KeycloakRealm{}, &keycloakApi.KeycloakClient{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()

	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	instance.Spec.Url = sonarURL
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}
	clMock.On("ConfigureGeneralSettings", sonarClient.SettingRequest{
		Key: "sonar.core.serverBaseURL", Value: sonarURL, ValueType: "value",
	}).
		Return(nil)
	clMock.On("ConfigureGeneralSettings", sonarClient.SettingRequest{
		Key: "sonar.auth.oidc.clientId.secured", Value: "main", ValueType: "value",
	}).Return(errTest)

	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}

	_, err := service.Integration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to configure sonar.auth.oidc.clientId.secured")
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Integration_SetProjectsDefaultVisibilityErr(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(admissionV1.SchemeGroupVersion, &sonarApi.Sonar{}, &keycloakApi.KeycloakRealm{}, &keycloakApi.KeycloakClient{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()

	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	instance.Spec.Url = sonarURL
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}

	clMock.On("ConfigureGeneralSettings", sonarClient.SettingRequest{
		Key: "sonar.core.serverBaseURL", Value: sonarURL, ValueType: "value",
	}).
		Return(nil)
	clMock.On("ConfigureGeneralSettings", sonarClient.SettingRequest{
		Key: "sonar.auth.oidc.clientId.secured", Value: "main", ValueType: "value",
	}).Return(nil)
	clMock.On("SetProjectsDefaultVisibility", "private").Return(errTest)

	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}

	_, err := service.Integration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set default")
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Integration(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(admissionV1.SchemeGroupVersion, &sonarApi.Sonar{}, &keycloakApi.KeycloakRealm{}, &keycloakApi.KeycloakClient{})

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()

	ctx := context.Background()
	instance := createSonarInstance()
	instance.Spec.Url = sonarURL
	clMock := cMock.ClientInterface{}
	plMock := pMock.Service{}

	clMock.On("ConfigureGeneralSettings", sonarClient.SettingRequest{
		Key: "sonar.core.serverBaseURL", Value: sonarURL, ValueType: "value",
	}).
		Return(nil)
	clMock.On("ConfigureGeneralSettings", sonarClient.SettingRequest{
		Key: "sonar.auth.oidc.clientId.secured", Value: "main", ValueType: "value",
	}).Return(nil)
	clMock.On("SetProjectsDefaultVisibility", "private").Return(nil)

	service := Service{
		k8sClient:       client,
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}

	_, err := service.Integration(ctx, &instance)
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
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return nil, errors.New("test")
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize Sonar Client")
}

func TestService_ExposeConfiguration_CreateUserErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	instance := createSonarInstance()
	instance.Spec.Users = []sonarApi.User{{Login: "ci-user", Username: "EDP CI User"}}

	clMock := cMock.ClientInterface{}
	clMock.On("GetUser", ctx, "ci-user").Return(nil, sonarClient.NotFoundError("test"))
	clMock.On("CreateUser", ctx, tMock.MatchedBy(func(sonar *sonarClient.User) bool {
		return sonar.Name == "EDP CI User" && sonar.Login == "ci-user"
	})).Return(errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user")
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_GetUserErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	instance := createSonarInstance()
	instance.Spec.Users = []sonarApi.User{{Login: "ci-user", Username: "EDP CI User"}}

	clMock := cMock.ClientInterface{}
	clMock.On("GetUser", ctx, "ci-user").Return(nil, errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user")
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_AddUserToGroupErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	instance := createSonarInstance()
	instance.Spec.Users = []sonarApi.User{{Login: "ci-user", Username: "EDP CI User", Group: "non-interactive-users"}}
	instance.Spec.Groups = []sonarApi.Group{{Name: "non-interactive-users", Permissions: []string{"scan"}}}

	clMock := cMock.ClientInterface{}
	clMock.On("GetUser", ctx, "ci-user").Return(nil, nil)
	clMock.On("GetUserToken", ctx, "ci-user", "Ci-User").Return(nil, nil)
	clMock.On("AddUserToGroup", "non-interactive-users", "ci-user").Return(errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add")
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_AddPermissionsToUserErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	instance := createSonarInstance()
	clMock := cMock.ClientInterface{}
	instance.Spec.Users = []sonarApi.User{{Login: "ci-user", Username: "EDP CI User", Permissions: []string{"admin"}, Group: "non-interactive-users"}}
	instance.Spec.Groups = []sonarApi.Group{{Name: "non-interactive-users", Permissions: []string{"scan"}}}

	clMock.On("GetUser", ctx, "ci-user").Return(nil, nil)
	clMock.On("GetUserToken", ctx, "ci-user", "Ci-User").Return(nil, nil)
	clMock.On("AddUserToGroup", "non-interactive-users", "ci-user").Return(nil)
	clMock.On("AddPermissionToUser", "ci-user", admin).Return(errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add permission admin to")
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_GetUserTokenErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	instance := createSonarInstance()
	instance.Spec.Users = []sonarApi.User{{Login: "ci-user", Username: "EDP CI User", Permissions: []string{"admin"}, Group: "non-interactive-users"}}
	instance.Spec.Groups = []sonarApi.Group{{Name: "non-interactive-users", Permissions: []string{"scan"}}}

	clMock := cMock.ClientInterface{}
	clMock.On("GetUser", ctx, "ci-user").Return(nil, nil)
	clMock.On("GetUserToken", ctx, "ci-user", cases.Title(language.English).String("ci-user")).Return(nil, errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user token for user")
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_GenerateUserTokenErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	instance := createSonarInstance()
	instance.Spec.Users = []sonarApi.User{{Login: "ci-user", Username: "EDP CI User", Permissions: []string{"admin"}, Group: "non-interactive-users"}}
	instance.Spec.Groups = []sonarApi.Group{{Name: "non-interactive-users", Permissions: []string{"scan"}}}

	clMock := cMock.ClientInterface{}
	clMock.On("GetUser", ctx, "ci-user").Return(nil, nil)
	clMock.On("GetUserToken", ctx, "ci-user", cases.Title(language.English).String("ci-user")).Return(nil, sonarClient.NotFoundError("test"))
	clMock.On("GenerateUserToken", "ci-user").Return(nil, errTest)

	service := Service{
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate token for")
	clMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_CreateSecretErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	ciUserName := fmt.Sprintf("%v-ci-user-token", main)
	ciToken := name
	ciSecret := map[string][]byte{
		"username": []byte("ci-user"),
		"secret":   []byte(ciToken),
	}
	instance := createSonarInstance()
	instance.Spec.Users = []sonarApi.User{{Login: "ci-user", Username: "EDP CI User", Permissions: []string{"admin"}, Group: "non-interactive-users"}}
	instance.Spec.Groups = []sonarApi.Group{{Name: "non-interactive-users", Permissions: []string{"scan"}}}

	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	clMock.On("GetUser", ctx, "ci-user").Return(nil, nil)
	clMock.On("GetUserToken", ctx, "ci-user", cases.Title(language.English).String("ci-user")).Return(nil, sonarClient.NotFoundError("test"))
	clMock.On("GenerateUserToken", "ci-user").Return(&ciToken, nil)
	plMock.On("CreateSecret", main, namespace, ciUserName, ciSecret).Return(nil, errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create secret for")
	clMock.AssertExpectations(t)
	plMock.AssertExpectations(t)
}

func TestService_ExposeConfiguration_SetOwnerReferenceErr(t *testing.T) {
	ctx := context.Background()
	errTest := errors.New("test")
	ciUserName := fmt.Sprintf("%v-ci-user-token", main)
	ciToken := name
	ciSecret := map[string][]byte{
		"username": []byte("ci-user"),
		"secret":   []byte(ciToken),
	}
	secret := coreV1Api.Secret{}
	instance := createSonarInstance()
	instance.Spec.Users = []sonarApi.User{{Login: "ci-user", Username: "EDP CI User", Permissions: []string{"admin"}, Group: "non-interactive-users"}}
	instance.Spec.Groups = []sonarApi.Group{{Name: "non-interactive-users", Permissions: []string{"scan"}}}

	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	clMock.On("GetUser", ctx, "ci-user").Return(nil, nil)
	clMock.On("GetUserToken", ctx, "ci-user", cases.Title(language.English).String("ci-user")).Return(nil, sonarClient.NotFoundError("test"))
	clMock.On("GenerateUserToken", "ci-user").Return(&ciToken, nil)
	plMock.On("CreateSecret", main, namespace, ciUserName, ciSecret).Return(&secret, nil)
	plMock.On("SetOwnerReference", &instance, &secret).Return(errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
	}
	err := service.ExposeConfiguration(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set owner reference for secret")
	clMock.AssertExpectations(t)
	plMock.AssertExpectations(t)
}

func TestService_Configure_installPluginsErr(t *testing.T) {
	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	instance.Spec.Plugins = plugins()
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	clMock.On("InstallPlugins", plugins()).Return(errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
		runningInClusterFunc: returnTrue,
	}
	err := service.Configure(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install plugins")
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Configure_createQualityGateErr(t *testing.T) {
	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	instance.Spec.Plugins = plugins()
	instance.Spec.QualityGates = []sonarApi.QualityGate{
		{
			Name: "EDP way",
			Conditions: []sonarApi.QualityGateCondition{
				{
					Metric: "coverage",
					OP:     "LT",
					Error:  "70",
					Period: "1",
				},
			},
		},
	}
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	clMock.On("InstallPlugins", plugins()).Return(nil)
	clMock.On("CreateQualityGates", sonarClient.QualityGates{
		"EDP way": sonarClient.QualityGateSettings{
			MakeDefault: false, Conditions: []sonarClient.QualityGateCondition{
				{ID: "", Error: "70", Metric: "coverage", OP: "LT", Period: "1"},
			},
		},
	},
	).Return(errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
		runningInClusterFunc: returnTrue,
	}

	err := service.Configure(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to configure EDP way quality gate")
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Configure_configureGeneralSettingsErr(t *testing.T) {
	scheme := runtime.NewScheme()
	jns := jenkinsV1Api.Jenkins{Spec: jenkinsV1Api.JenkinsSpec{BasePath: basePath}, ObjectMeta: metaV1.ObjectMeta{
		Name: name, Namespace: namespace,
	}}
	scheme.AddKnownTypes(admissionV1.SchemeGroupVersion, &jenkinsV1Api.JenkinsList{}, &jenkinsV1Api.Jenkins{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&jns).Build()
	errTest := errors.New("test")
	ctx := context.Background()
	instance := createSonarInstance()
	instance.Spec.Plugins = plugins()
	instance.Spec.QualityGates = []sonarApi.QualityGate{
		{
			Name: "EDP way",
			Conditions: []sonarApi.QualityGateCondition{
				{
					Metric: "coverage",
					OP:     "LT",
					Error:  "70",
					Period: "1",
				},
			},
		},
	}
	instance.Spec.Users = []sonarApi.User{{Login: "ci-user", Username: "EDP CI User"}}
	instance.Spec.Groups = []sonarApi.Group{{Name: "non-interactive-users", Permissions: []string{"scan"}}, {Name: "sonar-developers"}}
	instance.Spec.Settings = []sonarApi.SonarSetting{
		{
			Key:       "sonar.typescript.lcov.reportPaths",
			Value:     "coverage/lcov.info",
			ValueType: "values",
		},
	}
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	clMock.On("InstallPlugins", plugins()).Return(nil)
	clMock.On("CreateQualityGates", sonarClient.QualityGates{
		"EDP way": sonarClient.QualityGateSettings{
			MakeDefault: false, Conditions: []sonarClient.QualityGateCondition{
				{ID: "", Error: "70", Metric: "coverage", OP: "LT", Period: "1"},
			},
		},
	},
	).Return(nil)
	clMock.On("GetGroup", ctx, "non-interactive-users").Return(nil, nil)
	clMock.On("GetGroup", ctx, "sonar-developers").Return(nil, nil)
	clMock.On("AddPermissionsToGroup", "non-interactive-users", "scan").Return(nil)
	clMock.On("ConfigureGeneralSettings", sonarClient.SettingRequest{Key: "sonar.typescript.lcov.reportPaths", Value: "coverage/lcov.info", ValueType: "values"}).Return(errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
		k8sClient:            client,
		runningInClusterFunc: returnTrue,
	}
	err := service.Configure(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to configure general settings")
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func TestService_Configure_setDefaultPermissionTemplateErr(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	jns := jenkinsV1Api.Jenkins{Spec: jenkinsV1Api.JenkinsSpec{BasePath: basePath}, ObjectMeta: metaV1.ObjectMeta{
		Name: name, Namespace: namespace,
	}}
	scheme.AddKnownTypes(admissionV1.SchemeGroupVersion, &jenkinsV1Api.JenkinsList{}, &jenkinsV1Api.Jenkins{})
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&jns).Build()
	errTest := errors.New("test")
	instance := createSonarInstance()
	instance.Spec.Plugins = plugins()
	instance.Spec.QualityGates = []sonarApi.QualityGate{
		{
			Name: "EDP way",
			Conditions: []sonarApi.QualityGateCondition{
				{
					Metric: "coverage",
					OP:     "LT",
					Error:  "70",
					Period: "1",
				},
			},
		},
	}
	instance.Spec.Users = []sonarApi.User{{Login: "ci-user", Username: "EDP CI User"}}
	instance.Spec.Groups = []sonarApi.Group{{Name: "non-interactive-users", Permissions: []string{"scan"}}, {Name: "sonar-developers"}}

	instance.Spec.Settings = []sonarApi.SonarSetting{
		{
			Key:       "sonar.typescript.lcov.reportPaths",
			Value:     "coverage/lcov.info",
			ValueType: "values",
		},
		{
			Key:       "sonar.coverage.jacoco.xmlReportPaths",
			Value:     "target/site/jacoco/jacoco.xml",
			ValueType: "values",
		},
	}
	plMock := pMock.Service{}
	clMock := cMock.ClientInterface{}

	clMock.On("InstallPlugins", plugins()).Return(nil)
	clMock.On("CreateQualityGates", sonarClient.QualityGates{
		"EDP way": sonarClient.QualityGateSettings{
			MakeDefault: false, Conditions: []sonarClient.QualityGateCondition{
				{ID: "", Error: "70", Metric: "coverage", OP: "LT", Period: "1"},
			},
		},
	},
	).Return(nil)
	clMock.On("GetGroup", ctx, "non-interactive-users").Return(nil, nil)
	clMock.On("GetGroup", ctx, "sonar-developers").Return(nil, nil)
	clMock.On("AddPermissionsToGroup", "non-interactive-users", "scan").Return(nil)
	clMock.On("ConfigureGeneralSettings",
		sonarClient.SettingRequest{Key: "sonar.typescript.lcov.reportPaths", Value: "coverage/lcov.info", ValueType: "values"},
		sonarClient.SettingRequest{Key: "sonar.coverage.jacoco.xmlReportPaths", Value: "target/site/jacoco/jacoco.xml", ValueType: "values"},
	).Return(nil)
	clMock.On("SetDefaultPermissionTemplate", ctx, instance.Spec.DefaultPermissionTemplate).Return(errTest)

	service := Service{
		platformService: &plMock,
		sonarClientBuilder: func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
			return &clMock, nil
		},
		k8sClient:            client,
		runningInClusterFunc: returnTrue,
	}

	err := service.Configure(ctx, &instance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set default permission template")
	plMock.AssertExpectations(t)
	clMock.AssertExpectations(t)
}

func Test_parseQualityGates(t *testing.T) {
	t.Parallel()

	type args struct {
		qualityGates []sonarApi.QualityGate
	}

	tests := []struct {
		name string
		args args
		want sonarClient.QualityGates
	}{
		{
			name: "should parse quality gates",
			args: args{
				qualityGates: []sonarApi.QualityGate{
					{
						Name:         "EDP way",
						SetAsDefault: true,
						Conditions: []sonarApi.QualityGateCondition{
							{
								Error:  "80",
								Metric: "new_coverage",
								OP:     "LT",
								Period: "1",
							},
							{
								Error:  "0",
								Metric: "test_errors",
								OP:     "GT",
							},
						},
					},
				},
			},
			want: sonarClient.QualityGates{
				"EDP way": sonarClient.QualityGateSettings{
					MakeDefault: true,
					Conditions: []sonarClient.QualityGateCondition{
						{
							Metric: "new_coverage",
							OP:     "LT",
							Error:  "80",
							Period: "1",
						},
						{
							Metric: "test_errors",
							OP:     "GT",
							Error:  "0",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equalf(t, tt.want, parseQualityGates(tt.args.qualityGates), "parseQualityGates(%v)", tt.args.qualityGates)
		})
	}
}
