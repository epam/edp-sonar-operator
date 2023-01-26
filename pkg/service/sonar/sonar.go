package sonar

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dchest/uniuri"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	k8sErr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	jenkinsApi "github.com/epam/edp-jenkins-operator/v2/pkg/apis/v2/v1"
	platformHelper "github.com/epam/edp-jenkins-operator/v2/pkg/service/platform/helper"
	keycloakApi "github.com/epam/edp-keycloak-operator/api/v1"
	"github.com/epam/edp-keycloak-operator/controllers/helper"

	sonarApi "github.com/epam/edp-sonar-operator/v2/api/v1"
	"github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	sonarClient "github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	pkgHelper "github.com/epam/edp-sonar-operator/v2/pkg/helper"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform"
	sonarHelper "github.com/epam/edp-sonar-operator/v2/pkg/service/sonar/helper"
	sonarSpec "github.com/epam/edp-sonar-operator/v2/pkg/service/sonar/spec"
)

const (
	ciUserLogin                      = "ci-user"
	ciUsername                       = "EDP CI User"
	readUserLogin                    = "read"
	readUserUsername                 = "Read-only user"
	nonInteractiveGroupName          = "non-interactive-users"
	sonarDevelopersGroupName         = "sonar-developers"
	webhookUrl                       = "sonarqube-webhook/"
	defaultPassword                  = "admin"
	claimName                        = "roles"
	defaultConfigFilesAbsolutePath   = "/usr/local/"
	localConfigsRelativePath         = "configs"
	defaultProfileAbsolutePath       = defaultConfigFilesAbsolutePath + localConfigsRelativePath + "/" + defaultQualityProfilesFileName
	defaultQualityProfilesFileName   = "quality-profile.xml"
	imgFolder                        = "img"
	sonarIcon                        = "sonar.svg"
	jenkinsDefaultScriptConfigMapKey = "context"
	main                             = "main"
	annotation                       = "openid-configuration"
	admin                            = "admin"
	tokenType                        = "token"
	failMsgTemplate                  = "Failed to set owner reference for secret %v"
	failInitSonarMsg                 = "Failed to initialize Sonar Client!"
)

var log = ctrl.Log.WithName("sonar_service")

type ServiceInterface interface {
	Configure(ctx context.Context, instance *sonarApi.Sonar) error
	ExposeConfiguration(ctx context.Context, instance *sonarApi.Sonar) error
	Integration(ctx context.Context, instance *sonarApi.Sonar) (*sonarApi.Sonar, error)
	IsDeploymentReady(instance *sonarApi.Sonar) (bool, error)
	ClientForChild(ctx context.Context, instance ChildInstance) (ClientInterface, error)
	DeleteResource(ctx context.Context, instance Deletable, finalizer string,
		deleteFunc func() error) (bool, error)
	K8sClient() client.Client
}

type ChildInstance interface {
	SonarOwner() string
	GetNamespace() string
}

type Deletable interface {
	v1.Object
	runtime.Object
}

func NewService(platformService platform.Service, k8sClient client.Client) *Service {
	svc := Service{
		platformService:      platformService,
		k8sClient:            k8sClient,
		runningInClusterFunc: pkgHelper.RunningInCluster,
	}

	svc.sonarClientBuilder = svc.initSonarClient

	return &svc
}

type Service struct {
	// Providing sonar service implementation through the interface (platform abstract)
	platformService      platform.Service
	k8sClient            client.Client
	sonarClientBuilder   func(ctx context.Context, instance *sonarApi.Sonar, useDefaultPassword bool) (ClientInterface, error)
	runningInClusterFunc func() bool
}

func (s Service) K8sClient() client.Client {
	return s.k8sClient
}

func (s Service) ClientForChild(ctx context.Context, instance ChildInstance) (ClientInterface, error) {
	var rootSonar sonarApi.Sonar
	if err := s.k8sClient.Get(ctx, types.NamespacedName{Namespace: instance.GetNamespace(), Name: instance.SonarOwner()},
		&rootSonar); err != nil {
		return nil, errors.Wrap(err, "unable to get root sonar instance")
	}

	sClient, err := s.sonarClientBuilder(ctx, &rootSonar, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init sonar rest client")
	}

	return sClient, nil
}

func (s Service) DeleteResource(ctx context.Context, instance Deletable, finalizer string,
	deleteFunc func() error) (bool, error) {
	finalizers := instance.GetFinalizers()

	if instance.GetDeletionTimestamp().IsZero() {
		if !helper.ContainsString(finalizers, finalizer) {
			finalizers = append(finalizers, finalizer)
			instance.SetFinalizers(finalizers)

			if err := s.k8sClient.Update(ctx, instance); err != nil {
				return false, errors.Wrap(err, "unable to update deletable object")
			}
		}

		return false, nil
	}

	if err := deleteFunc(); err != nil {
		return false, errors.Wrap(err, "unable to delete resource")
	}

	if helper.ContainsString(finalizers, finalizer) {
		finalizers = helper.RemoveString(finalizers, finalizer)
		instance.SetFinalizers(finalizers)

		if err := s.k8sClient.Update(ctx, instance); err != nil {
			return false, errors.Wrap(err, "unable to update realm role cr")
		}
	}

	return true, nil
}

func (s Service) initSonarClient(ctx context.Context, instance *sonarApi.Sonar, useDefaultPassword bool) (ClientInterface, error) {
	password := defaultPassword
	if !useDefaultPassword {
		adminSecretName := fmt.Sprintf("%v-admin-password", instance.Name)
		credentials, err := s.platformService.GetSecretData(instance.Namespace, adminSecretName)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to get %v secret for Sonar client!", adminSecretName)
		}

		if newPassword, ok := credentials["password"]; ok {
			password = string(newPassword)
		} else {
			log.Info("No password found in secret", "secret", adminSecretName)
		}
	}

	u, err := s.platformService.GetExternalEndpoint(ctx, instance.Namespace, instance.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get route for %v", instance.Name)
	}

	return sonarClient.InitNewRestClient(fmt.Sprintf("%s/api", u), admin, password), nil
}

func (s Service) Integration(ctx context.Context, instance *sonarApi.Sonar) (*sonarApi.Sonar, error) {
	valueType := "value"
	sc, err := s.sonarClientBuilder(ctx, instance, false)
	if err != nil {
		return instance, errors.Wrap(err, failInitSonarMsg)
	}
	realm, err := s.getKeycloakRealm(instance)
	if err != nil {
		return instance, err
	}
	if realm != nil {
		if realm.Annotations == nil {
			return instance, errors.New("realm main does not have required annotations")
		}
		openIdConfiguration := realm.Annotations[annotation]
		var c map[string]interface{}
		err = json.Unmarshal([]byte(openIdConfiguration), &c)
		if err != nil {
			return instance, errors.Wrap(err, "failed to unmarshal OpenID configuration")
		}
		if len(c["issuer"].(string)) > 0 {
			err = sc.ConfigureGeneralSettings(valueType, "sonar.auth.oidc.issuerUri", c["issuer"].(string))
			if err != nil {
				return instance, errors.Wrap(err, "failed to to configure sonar.auth.oidc.issuerUri")
			}
		} else {
			return instance, errors.New("issuer field in oidc configuration is empty or configuration is invalid")
		}
	}

	url, err := s.platformService.GetExternalEndpoint(ctx, instance.Namespace, instance.Name)
	if err != nil {
		return nil, err
	}
	err = sc.ConfigureGeneralSettings(valueType, "sonar.core.serverBaseURL", url)
	if err != nil {
		return instance, errors.Wrap(err, "Failed to configure sonar.core.serverBaseURL!")
	}
	cl, err := s.getKeycloakClient(instance)
	if err != nil {
		return instance, err
	}

	if cl == nil {
		err = s.createKeycloakClient(instance, url)
	}

	if err != nil {
		return instance, err
	}

	err = sc.ConfigureGeneralSettings(valueType, "sonar.auth.oidc.clientId.secured", instance.Name)
	if err != nil {
		return instance, errors.Wrap(err, "Failed to configure sonar.auth.oidc.clientId.secured!")
	}

	err = sc.ConfigureGeneralSettings(valueType, "sonar.auth.oidc.groupsSync.claimName", claimName)
	if err != nil {
		return instance, errors.Wrap(err, "Failed to configure sonar.auth.oidc.groupsSync.claimName!")
	}

	err = sc.ConfigureGeneralSettings(valueType, "sonar.auth.oidc.groupsSync", "true")
	if err != nil {
		return instance, errors.Wrap(err, "Failed to configure sonar.auth.oidc.groupsSync!")
	}

	err = sc.ConfigureGeneralSettings(valueType, "sonar.auth.oidc.enabled", "true")
	if err != nil {
		return instance, errors.Wrap(err, "Failed to configure sonar.auth.oidc.enabled!")
	}

	dv := "private"
	log.Info(fmt.Sprintf("trying to set %v visibility for projects as default", dv))
	if err = sc.SetProjectsDefaultVisibility(dv); err != nil {
		return nil, errors.Wrapf(err, "couldn't set default %v visibility for projects", dv)
	}

	return instance, nil
}

func (s *Service) getKeycloakRealm(instance *sonarApi.Sonar) (*keycloakApi.KeycloakRealm, error) {
	realm := &keycloakApi.KeycloakRealm{}
	err := s.k8sClient.Get(context.TODO(), types.NamespacedName{
		Name:      main,
		Namespace: instance.Namespace,
	}, realm)
	if err != nil {
		if k8sErr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return realm, nil
}

func (s *Service) getKeycloakClient(instance *sonarApi.Sonar) (*keycloakApi.KeycloakClient, error) {
	cl := &keycloakApi.KeycloakClient{}
	err := s.k8sClient.Get(context.TODO(), types.NamespacedName{
		Name:      instance.Name,
		Namespace: instance.Namespace,
	}, cl)
	if err != nil {
		if k8sErr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return cl, nil
}

func (s *Service) createKeycloakClient(instance *sonarApi.Sonar, baseUrl string) error {
	trueStr := "true"
	cl := &keycloakApi.KeycloakClient{
		ObjectMeta: v1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: keycloakApi.KeycloakClientSpec{
			ClientId:                instance.Name,
			Public:                  true,
			WebUrl:                  baseUrl,
			AdvancedProtocolMappers: true,
			RealmRoles: &[]keycloakApi.RealmRole{
				{
					Name:      "sonar-administrators",
					Composite: "administrator",
				},
				{
					Name:      sonarDevelopersGroupName,
					Composite: "developer",
				},
			},
			ProtocolMappers: &[]keycloakApi.ProtocolMapper{
				{
					Name:           "realm roles",
					Protocol:       "openid-connect",
					ProtocolMapper: "oidc-usermodel-realm-role-mapper",
					Config: map[string]string{
						"access.token.claim":   "false",
						"claim.name":           "roles",
						"id.token.claim":       trueStr,
						"jsonType.label":       "String",
						"multivalued":          trueStr,
						"userinfo.token.claim": trueStr,
					},
				},
			},
		},
	}
	return s.k8sClient.Create(context.TODO(), cl)
}

func (s Service) ExposeConfiguration(ctx context.Context, instance *sonarApi.Sonar) error {
	sc, err := s.sonarClientBuilder(ctx, instance, false)
	if err != nil {
		return errors.Wrap(err, failInitSonarMsg)
	}

	_, err = sc.GetUser(ctx, ciUserLogin)
	if sonarClient.IsErrNotFound(err) {
		sonarUser := sonarClient.User{
			Login: ciUserLogin, Name: ciUsername, Password: uniuri.New()}
		if err = sc.CreateUser(ctx, &sonarUser); err != nil {
			return errors.Wrapf(err, "Failed to create user %v in Sonar", ciUsername)
		}
	} else if err != nil {
		return errors.Wrapf(err, "unexpected error during get user %s", ciUsername)
	}

	err = sc.AddUserToGroup(nonInteractiveGroupName, ciUserLogin)
	if err != nil {
		return errors.Wrapf(err, "Failed to add %v user in %v group!", ciUserLogin, nonInteractiveGroupName)
	}

	err = sc.AddPermissionsToUser(ciUserLogin, admin)
	if err != nil {
		return errors.Wrapf(err, "Failed to add admin permissions to  %v user", ciUserLogin)
	}

	ciUserName := fmt.Sprintf("%v-ciuser-token", instance.Name)
	_, err = sc.GetUserToken(ctx, ciUserLogin, cases.Title(language.English).String(ciUserLogin))
	if sonarClient.IsErrNotFound(err) {
		ciToken, errGen := sc.GenerateUserToken(ciUserLogin)
		if errGen != nil {
			return errors.Wrapf(errGen, "Failed to generate token for %v user", ciUserLogin)
		}

		if ciToken != nil {
			ciSecret := map[string][]byte{
				"username": []byte(ciUserLogin),
				"secret":   []byte(*ciToken),
			}

			secret, errCreateSecret := s.platformService.CreateSecret(instance.Name, instance.Namespace, ciUserName, ciSecret)
			if errCreateSecret != nil {
				return errors.Wrapf(errCreateSecret, "Failed to create secret for  %v user", ciUserName)
			}
			if errCreateSecret = s.platformService.SetOwnerReference(instance, secret); errCreateSecret != nil {
				return errors.Wrapf(errCreateSecret, failMsgTemplate, secret)
			}
		}
	} else if err != nil {
		return errors.Wrapf(err, "unexpected error during get user token for user %s", ciUserLogin)
	}

	if s.jenkinsEnabled(ctx, instance.Namespace) {
		if err = s.exposeJenkinsConfiguration(instance, ciUserName); err != nil {
			return err
		}
	}

	_, err = sc.GetUser(ctx, readUserLogin)
	if sonarClient.IsErrNotFound(err) {
		sonarUser := sonarClient.User{
			Login: readUserLogin, Name: readUserUsername, Password: uniuri.New()}
		if err = sc.CreateUser(ctx, &sonarUser); err != nil {
			return errors.Wrapf(err, "Failed to create user %v in Sonar", readUserLogin)
		}
	} else if err != nil {
		return errors.Wrapf(err, "unexpected error during get user %s", readUserLogin)
	}
	readUserSecretName := fmt.Sprintf("%v-readuser-token", instance.Name)
	_, err = sc.GetUserToken(ctx, readUserLogin, cases.Title(language.English).String(readUserLogin))
	if sonarClient.IsErrNotFound(err) {
		readToken, errGenToken := sc.GenerateUserToken(readUserLogin)
		if errGenToken != nil {
			return errors.Wrapf(errGenToken, "Failed to generate token for %s user", readUserLogin)
		}

		if readToken != nil {
			readSecret := map[string][]byte{
				"username": []byte(readUserLogin),
				tokenType:  []byte(*readToken),
			}

			secret, errCreateSecret := s.platformService.CreateSecret(instance.Name, instance.Namespace, readUserSecretName, readSecret)
			if errCreateSecret != nil {
				return errors.Wrapf(errCreateSecret, "Failed to create secret for  %v user", readUserSecretName)
			}
			if err = s.platformService.SetOwnerReference(instance, secret); err != nil {
				return errors.Wrapf(err, failMsgTemplate, secret)
			}
		}
	} else if err != nil {
		return errors.Wrapf(err, "unexpected error during get user token for user %s", readUserLogin)
	}

	err = sc.AddUserToGroup(nonInteractiveGroupName, readUserLogin)
	if err != nil {
		return errors.Wrapf(err, "Failed to add %v user in %v group!", readUserLogin, nonInteractiveGroupName)
	}

	identityServerSecretName := fmt.Sprintf("%v-is-credentials", instance.Name)
	identityServiceClientSecret := uniuri.New()
	identityServiceClientCredentials := map[string][]byte{
		"client_id":     []byte(instance.Name),
		"client_secret": []byte(identityServiceClientSecret),
	}

	secret, err := s.platformService.CreateSecret(instance.Name, instance.Namespace, identityServerSecretName, identityServiceClientCredentials)
	if err != nil {
		return errors.Wrapf(err, "Failed to create secret for  %v Keycloak client!", readUserSecretName)
	}
	if err = s.platformService.SetOwnerReference(instance, secret); err != nil {
		return errors.Wrapf(err, failMsgTemplate, secret)
	}

	err = s.createEDPComponent(ctx, instance)

	return err
}

func (s *Service) exposeJenkinsConfiguration(instance *sonarApi.Sonar, ciUserName string) error {
	err := s.platformService.CreateJenkinsServiceAccount(instance.Namespace, ciUserName, tokenType)
	if err != nil {
		return fmt.Errorf("failed to create Jenkins Service Account for %v: %w", ciUserName, err)
	}

	data := sonarHelper.InitNewJenkinsPluginInfo(true)
	data.ServerName = instance.Name
	data.SecretName = ciUserName
	data.ServerPath = ""
	if len(instance.Spec.BasePath) != 0 {
		data.ServerPath = fmt.Sprintf("/%v", instance.Spec.BasePath)
	}

	jenkinsScriptContext, err := sonarHelper.ParseDefaultTemplate(data)
	if err != nil {
		return fmt.Errorf("failed to parse default Jenkins plugin template: %w", err)
	}

	configMapName := fmt.Sprintf("%s-%s", instance.Name, sonarSpec.JenkinsPluginConfigPostfix)
	configMapData := map[string]string{
		jenkinsDefaultScriptConfigMapKey: jenkinsScriptContext.String(),
	}

	err = s.platformService.CreateConfigMap(instance, configMapName, configMapData)
	if err != nil {
		return fmt.Errorf("failed to create Config Map %v: %w", configMapName, err)
	}

	err = s.platformService.CreateJenkinsScript(instance.Namespace, configMapName)
	if err != nil {
		return fmt.Errorf("failed to create Jenkins Script for %v: %w", ciUserName, err)
	}

	return nil
}

func (s Service) createEDPComponent(ctx context.Context, sonar *sonarApi.Sonar) error {
	url, err := s.platformService.GetExternalEndpoint(ctx, sonar.Namespace, sonar.Name)
	if err != nil {
		return err
	}
	icon, err := getIcon()
	if err != nil {
		return err
	}
	return s.platformService.CreateEDPComponentIfNotExist(sonar, url, *icon)
}

func getIcon() (*string, error) {
	p, err := platformHelper.CreatePathToTemplateDirectory(imgFolder)
	if err != nil {
		return nil, err
	}
	fp := fmt.Sprintf("%v/%v", p, sonarIcon)
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(content)
	return &encoded, nil
}

func (s Service) Configure(ctx context.Context, instance *sonarApi.Sonar) error {
	if s.runningInClusterFunc == nil {
		return errors.New("missing runningInClusterFunc")
	}
	if err := s.configurePassword(ctx, instance); err != nil {
		return errors.Wrap(err, "unable to setup password for sonar")
	}

	sc, err := s.sonarClientBuilder(ctx, instance, false)
	if err != nil {
		return errors.Wrap(err, failInitSonarMsg)
	}

	if err = installPlugins(sc); err != nil {
		return errors.Wrap(err, "unable to install plugins")
	}

	if err = uploadProfile(sc, s.runningInClusterFunc); err != nil {
		return errors.Wrap(err, "unable to upload profile")
	}

	if err = createQualityGate(sc); err != nil {
		return errors.Wrap(err, "Failed to configure EDP way quality gate!")
	}

	if err = setupGroups(ctx, sc); err != nil {
		return errors.Wrap(err, "unable to setup groups")
	}

	if s.jenkinsEnabled(ctx, instance.Namespace) {
		if err = s.setupWebhook(ctx, sc, instance.Namespace); err != nil {
			return fmt.Errorf("unable to setup webhook: %w", err)
		}
	}

	if err = configureGeneralSettings(sc); err != nil {
		return errors.Wrap(err, "unable to configure general settings")
	}

	if err = setDefaultPermissionTemplate(ctx, sc, instance.Spec.DefaultPermissionTemplate); err != nil {
		return errors.Wrap(err, "unable to set default permission template")
	}

	return nil
}

func (s *Service) setupWebhook(ctx context.Context, sc ClientInterface, instanceNamespace string) error {
	jenkinsUrl, err := s.getInternalJenkinsUrl(ctx, instanceNamespace)
	if err != nil {
		return errors.Wrap(err, "unable to get internal jenkins url")
	}
	if err = sc.AddWebhook(ciUserLogin, fmt.Sprintf("%v/%v", jenkinsUrl, webhookUrl)); err != nil {
		return errors.Wrap(err, "Failed to add Jenkins webhook!")
	}

	return nil
}

func (s *Service) configurePassword(ctx context.Context, instance *sonarApi.Sonar) error {
	credentials, err := s.createAdminSecret(instance)
	if err != nil {
		return errors.Wrap(err, "unable to create admin secret")
	}

	sc, err := s.sonarClientBuilder(ctx, instance, true)
	if err != nil {
		return errors.Wrap(err, failInitSonarMsg)
	}

	err = sc.ChangePassword(ctx, admin, defaultPassword, string(credentials["password"]))
	if sonarClient.IsHTTPErrorCode(err, http.StatusUnauthorized) ||
		sonarClient.IsHTTPErrorCode(err, http.StatusForbidden) {
		log.Error(err, "Failed to change default password for SonarQube")
		return nil
	} else if err != nil {
		return errors.Wrap(err, "Failed to change password!")
	}

	return nil
}

func setDefaultPermissionTemplate(ctx context.Context, sc ClientInterface, templateName string) error {
	if templateName != "" {
		if err := sc.SetDefaultPermissionTemplate(ctx, templateName); err != nil {
			return errors.Wrap(err, "unable to set default permission template")
		}
	}

	return nil
}

func setupGroups(ctx context.Context, sc ClientInterface) error {
	groups := []string{nonInteractiveGroupName, sonarDevelopersGroupName}
	for _, g := range groups {
		if _, err := sc.GetGroup(ctx, g); sonar.IsErrNotFound(err) {
			if err = sc.CreateGroup(ctx, &sonar.Group{Name: g}); err != nil {
				return errors.Wrapf(err, "Failed to create %s group!", g)
			}
		} else if err != nil {
			return errors.Wrap(err, "unexpected error during group check")
		}
	}

	if err := sc.AddPermissionsToGroup(nonInteractiveGroupName, "scan"); err != nil {
		return errors.Wrapf(err, "Failed to add scan permission for %s group!", nonInteractiveGroupName)
	}

	return nil
}

func uploadProfile(sc ClientInterface, isRunningInCluster func() bool) error {
	profilePath := defaultProfileAbsolutePath
	if !isRunningInCluster() {
		profilePath = fmt.Sprintf("%v\\..\\%v\\%v", pkgHelper.GetExecutableFilePath(), localConfigsRelativePath,
			defaultQualityProfilesFileName)
	}

	if _, err := sc.UploadProfile("EDP way", profilePath); err != nil {
		return errors.Wrap(err, "Failed to upload EDP way profile!")
	}

	return nil
}

func createQualityGate(sc ClientInterface) error {
	gt := "GT"
	errorStr := "error"
	metric := "metric"
	zero := "0"
	op := "op"
	if _, err := sc.CreateQualityGate("EDP way", []map[string]string{
		{errorStr: "80", metric: "new_coverage", op: "LT", "period": "1"},
		{errorStr: zero, metric: "test_errors", op: gt},
		{errorStr: "3", metric: "new_duplicated_lines_density", op: gt, "period": "1"},
		{errorStr: zero, metric: "test_failures", op: gt},
		{errorStr: zero, metric: "blocker_violations", op: gt},
		{errorStr: zero, metric: "critical_violations", op: gt},
	}); err != nil {
		return errors.Wrap(err, "Failed to configure EDP way quality gate!")
	}

	return nil
}

func installPlugins(sc ClientInterface) error {
	plugins := []string{"authoidc", "checkstyle", "findbugs", "pmd", "jacoco", "xml", "javascript", "go", "ansible",
		"yaml", "python", "csharp", "groovy"}

	if err := sc.InstallPlugins(plugins); err != nil {
		return errors.Wrap(err, "Failed to install plugins for Sonar!")
	}

	return nil
}

func configureGeneralSettings(sc ClientInterface) error {
	if err := sc.ConfigureGeneralSettings("values", "sonar.typescript.lcov.reportPaths",
		"coverage/lcov.info"); err != nil {
		return errors.Wrap(err, "Failed to configure sonar.typescript.lcov.reportPaths!")
	}

	if err := sc.ConfigureGeneralSettings("values", "sonar.coverage.jacoco.xmlReportPaths",
		"target/site/jacoco/jacoco.xml"); err != nil {
		return errors.Wrap(err, "Failed to configure sonar.coverage.jacoco.xmlReportPaths!")
	}

	return nil
}

func (s Service) createAdminSecret(instance *sonarApi.Sonar) (map[string][]byte, error) {
	secret, err := s.platformService.CreateSecret(instance.Name, instance.Namespace,
		fmt.Sprintf("%s-admin-password", instance.Name), map[string][]byte{
			"user":     []byte(admin),
			"password": []byte(uniuri.New()),
		})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create password for Admin in %s Sonar!", instance.Name)
	}

	if err = s.platformService.SetOwnerReference(instance, secret); err != nil {
		return nil, errors.Wrapf(err, failMsgTemplate, secret)
	}

	return secret.Data, nil
}

func (s Service) IsDeploymentReady(instance *sonarApi.Sonar) (bool, error) {
	r, err := s.platformService.GetAvailableDeploymentReplicas(instance)
	if err != nil {
		return false, err
	}

	if *r == 1 {
		return true, nil
	}

	return false, nil
}

func (s Service) getInternalJenkinsUrl(ctx context.Context, namespace string) (string, error) {
	jenkinsList := &jenkinsApi.JenkinsList{}

	if err := s.k8sClient.List(ctx, jenkinsList, &client.ListOptions{Namespace: namespace}); err != nil {
		return "", errors.Wrapf(err, "Unable to get Jenkins CRs in namespace %s", namespace)
	}

	if len(jenkinsList.Items) == 0 {
		return "", errors.Errorf("Jenkins installation is not found in namespace %s", namespace)
	}

	jenkins := jenkinsList.Items[0]
	basePath := ""
	if len(jenkins.Spec.BasePath) > 0 {
		basePath = fmt.Sprintf("/%v", jenkins.Spec.BasePath)
	}

	return fmt.Sprintf("http://jenkins.%s:8080%v", namespace, basePath), nil
}

func (s Service) jenkinsEnabled(ctx context.Context, namespace string) bool {
	jenkinsList := &jenkinsApi.JenkinsList{}

	if err := s.k8sClient.List(ctx, jenkinsList, &client.ListOptions{Namespace: namespace}); err != nil {
		return false
	}

	return len(jenkinsList.Items) != 0
}
