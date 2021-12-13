package sonar

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dchest/uniuri"
	jenkinsApi "github.com/epam/edp-jenkins-operator/v2/pkg/apis/v2/v1alpha1"
	platformHelper "github.com/epam/edp-jenkins-operator/v2/pkg/service/platform/helper"
	keycloakApi "github.com/epam/edp-keycloak-operator/pkg/apis/v1/v1alpha1"
	"github.com/epam/edp-keycloak-operator/pkg/controller/helper"
	"github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	sonarClient "github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	pkgHelper "github.com/epam/edp-sonar-operator/v2/pkg/helper"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform"
	sonarHelper "github.com/epam/edp-sonar-operator/v2/pkg/service/sonar/helper"
	sonarSpec "github.com/epam/edp-sonar-operator/v2/pkg/service/sonar/spec"
	"github.com/pkg/errors"
	k8sErr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	jenkinsLogin                     = "jenkins"
	jenkinsUsername                  = "Jenkins"
	readUserLogin                    = "read"
	readUserUsername                 = "Read-only user"
	nonInteractiveGroupName          = "non-interactive-users"
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
)

type ServiceInterface interface {
	Configure(ctx context.Context, instance *v1alpha1.Sonar) error
	ExposeConfiguration(ctx context.Context, instance *v1alpha1.Sonar) error
	Integration(ctx context.Context, instance v1alpha1.Sonar) (*v1alpha1.Sonar, error)
	IsDeploymentReady(instance *v1alpha1.Sonar) (bool, error)
	ClientForChild(ctx context.Context, instance ChildInstance) (ClientInterface, error)
	DeleteResource(ctx context.Context, instance Deletable, finalizer string,
		deleteFunc func() error) (bool, error)
}

type ChildInstance interface {
	SonarOwner() string
	GetNamespace() string
}

type Deletable interface {
	v1.Object
	runtime.Object
}

func NewService(platformService platform.Service, k8sClient client.Client, k8sScheme *runtime.Scheme) *Service {
	svc := Service{
		platformService: platformService,
		k8sClient:       k8sClient,
		k8sScheme:       k8sScheme,
	}

	svc.sonarClientBuilder = svc.initSonarClient

	return &svc
}

type Service struct {
	// Providing sonar service implementation through the interface (platform abstract)
	platformService    platform.Service
	k8sClient          client.Client
	k8sScheme          *runtime.Scheme
	sonarClientBuilder func(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error)
}

func (s Service) ClientForChild(ctx context.Context, instance ChildInstance) (ClientInterface, error) {
	var rootSonar v1alpha1.Sonar
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

func (s Service) initSonarClient(ctx context.Context, instance *v1alpha1.Sonar, useDefaultPassword bool) (ClientInterface, error) {
	password := defaultPassword
	if !useDefaultPassword {
		adminSecretName := fmt.Sprintf("%v-admin-password", instance.Name)
		credentials, err := s.platformService.GetSecretData(instance.Namespace, adminSecretName)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to get %v secret for Sonar client!", adminSecretName)
		}
		password = string(credentials["password"])
	}

	u, err := s.platformService.GetExternalEndpoint(ctx, instance.Namespace, instance.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get route for %v", instance.Name)
	}

	return sonarClient.InitNewRestClient(fmt.Sprintf("%s/api", u), "admin", password), nil
}

func (s Service) Integration(ctx context.Context, instance v1alpha1.Sonar) (*v1alpha1.Sonar, error) {
	sc, err := s.sonarClientBuilder(ctx, &instance, false)
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to initialize Sonar Client!")
	}
	realm, err := s.getKeycloakRealm(instance)
	if err != nil {
		return &instance, err
	}
	if realm != nil {
		if realm.Annotations == nil {
			return &instance, errors.New("realm main does not have required annotations")
		}
		openIdConfiguration := realm.Annotations["openid-configuration"]
		var c map[string]interface{}
		err := json.Unmarshal([]byte(openIdConfiguration), &c)
		if err != nil {
			return &instance, errors.Wrap(err, "failed to unmarshal OpenID configuration")
		}
		if len(c["issuer"].(string)) > 0 {
			err = sc.ConfigureGeneralSettings("value", "sonar.auth.oidc.issuerUri", c["issuer"].(string))
			if err != nil {
				return &instance, errors.Wrap(err, "failed to to configure sonar.auth.oidc.issuerUri")
			}
		} else {
			return &instance, errors.New("issuer field in oidc configuration is empty or configuration is invalid")
		}
	}

	url, err := s.platformService.GetExternalEndpoint(ctx, instance.Namespace, instance.Name)
	if err != nil {
		return nil, err
	}
	err = sc.ConfigureGeneralSettings("value", "sonar.core.serverBaseURL", url)
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to configure sonar.core.serverBaseURL!")
	}
	cl, err := s.getKeycloakClient(instance)
	if err != nil {
		return &instance, err
	}

	if cl == nil {
		err = s.createKeycloakClient(instance, url)
	}

	if err != nil {
		return &instance, err
	}

	err = sc.ConfigureGeneralSettings("value", "sonar.auth.oidc.clientId.secured", instance.Name)
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to configure sonar.auth.oidc.clientId.secured!")
	}

	err = sc.ConfigureGeneralSettings("value", "sonar.auth.oidc.groupsSync.claimName", claimName)
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to configure sonar.auth.oidc.groupsSync.claimName!")
	}

	err = sc.ConfigureGeneralSettings("value", "sonar.auth.oidc.groupsSync", "true")
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to configure sonar.auth.oidc.groupsSync!")
	}

	err = sc.ConfigureGeneralSettings("value", "sonar.auth.oidc.enabled", "true")
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to configure sonar.auth.oidc.enabled!")
	}

	dv := "private"
	log.Printf("trying to set %v visibility for projects as default", dv)
	if err := sc.SetProjectsDefaultVisibility(dv); err != nil {
		return nil, errors.Wrapf(err, "couldn't set default %v visibility for projects", dv)
	}

	return &instance, nil
}

func (s Service) getKeycloakRealm(instance v1alpha1.Sonar) (*keycloakApi.KeycloakRealm, error) {
	realm := &keycloakApi.KeycloakRealm{}
	err := s.k8sClient.Get(context.TODO(), types.NamespacedName{
		Name:      "main",
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

func (s Service) getKeycloakClient(instance v1alpha1.Sonar) (*keycloakApi.KeycloakClient, error) {
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

func (s Service) createKeycloakClient(instance v1alpha1.Sonar, baseUrl string) error {
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
					Name:      "sonar-users",
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
						"id.token.claim":       "true",
						"jsonType.label":       "String",
						"multivalued":          "true",
						"userinfo.token.claim": "true",
					},
				},
			},
		},
	}
	return s.k8sClient.Create(context.TODO(), cl)
}

func (s Service) ExposeConfiguration(ctx context.Context, instance *v1alpha1.Sonar) error {
	sc, err := s.sonarClientBuilder(ctx, instance, false)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize Sonar Client!")
	}

	_, err = sc.GetUser(ctx, jenkinsLogin)
	if sonarClient.IsErrNotFound(err) {
		sonarUser := sonarClient.User{
			Login: jenkinsLogin, Name: jenkinsUsername, Password: uniuri.New()}
		if err := sc.CreateUser(ctx, &sonarUser); err != nil {
			return errors.Wrapf(err, "Failed to create user %v in Sonar", jenkinsUsername)
		}
	} else if err != nil {
		return errors.Wrapf(err, "unexpected error during get user %s", jenkinsUsername)
	}

	err = sc.AddUserToGroup(nonInteractiveGroupName, jenkinsLogin)
	if err != nil {
		return errors.Wrapf(err, "Failed to add %v user in %v group!", jenkinsLogin, nonInteractiveGroupName)
	}

	err = sc.AddPermissionsToUser(jenkinsLogin, "admin")
	if err != nil {
		return errors.Wrapf(err, "Failed to add admin permissions to  %v user", jenkinsLogin)
	}

	ciUserName := fmt.Sprintf("%v-ciuser-token", instance.Name)
	_, err = sc.GetUserToken(ctx, jenkinsLogin, strings.Title(jenkinsLogin))
	if sonarClient.IsErrNotFound(err) {
		ciToken, err := sc.GenerateUserToken(jenkinsLogin)
		if err != nil {
			return errors.Wrapf(err, "Failed to generate token for %v user", jenkinsLogin)
		}

		if ciToken != nil {
			ciSecret := map[string][]byte{
				"username": []byte(jenkinsLogin),
				"secret":   []byte(*ciToken),
			}

			secret, err := s.platformService.CreateSecret(instance.Name, instance.Namespace, ciUserName, ciSecret)
			if err != nil {
				return errors.Wrapf(err, "Failed to create secret for  %v user", ciUserName)
			}
			if err := s.platformService.SetOwnerReference(instance, secret); err != nil {
				return errors.Wrapf(err, "Failed to set owner reference for secret %v", secret)
			}
		}
	} else if err != nil {
		return errors.Wrapf(err, "unexpected error during get user token for user %s", jenkinsLogin)
	}

	err = s.platformService.CreateJenkinsServiceAccount(instance.Namespace, ciUserName, "token")
	if err != nil {
		return errors.Wrapf(err, "Failed to create Jenkins Service Account for %v", ciUserName)
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
		return errors.Wrapf(err, "Failed to parse default Jenkins plugin template!")
	}

	configMapName := fmt.Sprintf("%s-%s", instance.Name, sonarSpec.JenkinsPluginConfigPostfix)
	configMapData := map[string]string{
		jenkinsDefaultScriptConfigMapKey: jenkinsScriptContext.String(),
	}

	err = s.platformService.CreateConfigMap(instance, configMapName, configMapData)
	if err != nil {
		return errors.Wrapf(err, "Failed to create Config Map %v", configMapName)
	}

	err = s.platformService.CreateJenkinsScript(instance.Namespace, configMapName)
	if err != nil {
		return errors.Wrapf(err, "Failed to create Jenkins Script for %v", ciUserName)
	}

	_, err = sc.GetUser(ctx, readUserLogin)
	if sonarClient.IsErrNotFound(err) {
		sonarUser := sonarClient.User{
			Login: readUserLogin, Name: readUserUsername, Password: uniuri.New()}
		if err := sc.CreateUser(ctx, &sonarUser); err != nil {
			return errors.Wrapf(err, "Failed to create user %v in Sonar", readUserLogin)
		}
	} else if err != nil {
		return errors.Wrapf(err, "unexpected error during get user %s", readUserLogin)
	}
	readUserSecretName := fmt.Sprintf("%v-readuser-token", instance.Name)
	_, err = sc.GetUserToken(ctx, readUserLogin, strings.Title(readUserLogin))
	if sonarClient.IsErrNotFound(err) {
		readToken, err := sc.GenerateUserToken(readUserLogin)
		if err != nil {
			return errors.Wrapf(err, "Failed to generate token for %s user", readUserLogin)
		}

		if readToken != nil {
			readSecret := map[string][]byte{
				"username": []byte(readUserLogin),
				"token":    []byte(*readToken),
			}

			secret, err := s.platformService.CreateSecret(instance.Name, instance.Namespace, readUserSecretName, readSecret)
			if err != nil {
				return errors.Wrapf(err, "Failed to create secret for  %v user", readUserSecretName)
			}
			if err := s.platformService.SetOwnerReference(instance, secret); err != nil {
				return errors.Wrapf(err, "Failed to set owner reference for secret %v", secret)
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
	if err := s.platformService.SetOwnerReference(instance, secret); err != nil {
		return errors.Wrapf(err, "Failed to set owner reference for secret %v", secret)
	}

	err = s.createEDPComponent(ctx, instance)

	return err
}

func (s Service) createEDPComponent(ctx context.Context, sonar *v1alpha1.Sonar) error {
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
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(content)
	return &encoded, nil
}

func (s Service) Configure(ctx context.Context, instance *v1alpha1.Sonar) error {
	if err := s.configurePassword(ctx, instance); err != nil {
		return errors.Wrap(err, "unable to setup password for sonar")
	}

	sc, err := s.sonarClientBuilder(ctx, instance, false)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize Sonar Client!")
	}

	if err := installPlugins(sc); err != nil {
		return errors.Wrap(err, "unable to install plugins")
	}

	if err := uploadProfile(sc); err != nil {
		return errors.Wrap(err, "unable to upload profile")
	}

	if err := createQualityGate(sc); err != nil {
		return errors.Wrap(err, "Failed to configure EDP way quality gate!")
	}

	if err := setupGroup(ctx, sc); err != nil {
		return errors.Wrap(err, "unable to setup group")
	}

	if err := s.setupWebhook(ctx, sc, instance.Namespace); err != nil {
		return errors.Wrap(err, "unable to setup webhook")
	}

	if err := configureGeneralSettings(sc); err != nil {
		return errors.Wrap(err, "unable to configure general settings")
	}

	if err := setDefaultPermissionTemplate(ctx, sc, instance.Spec.DefaultPermissionTemplate); err != nil {
		return errors.Wrap(err, "unable to set default permission template")
	}

	return nil
}

func (s *Service) setupWebhook(ctx context.Context, sc ClientInterface, instanceNamespace string) error {
	jenkinsUrl, err := s.getInternalJenkinsUrl(ctx, instanceNamespace)
	if err != nil {
		return errors.Wrap(err, "unable to get internal jenkins url")
	}
	if err := sc.AddWebhook(jenkinsLogin, fmt.Sprintf("%v/%v", jenkinsUrl, webhookUrl)); err != nil {
		return errors.Wrap(err, "Failed to add Jenkins webhook!")
	}

	return nil
}

func (s *Service) configurePassword(ctx context.Context, instance *v1alpha1.Sonar) error {
	credentials, err := s.createAdminSecret(instance)
	if err != nil {
		return errors.Wrap(err, "unable to create admin secret")
	}

	sc, err := s.sonarClientBuilder(ctx, instance, true)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize Sonar Client!")
	}

	err = sc.ChangePassword(ctx, "admin", defaultPassword, string(credentials["password"]))
	if sonarClient.IsHTTPErrorCode(err, http.StatusUnauthorized) {
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

func setupGroup(ctx context.Context, sc ClientInterface) error {
	if _, err := sc.GetGroup(ctx, nonInteractiveGroupName); sonar.IsErrNotFound(err) {
		if err := sc.CreateGroup(ctx, &sonar.Group{Name: nonInteractiveGroupName}); err != nil {
			return errors.Wrapf(err, "Failed to create %s group!", nonInteractiveGroupName)
		}
	} else if err != nil {
		return errors.Wrap(err, "unexpected error during group check")
	}

	if err := sc.AddPermissionsToGroup(nonInteractiveGroupName, "scan"); err != nil {
		return errors.Wrapf(err, "Failed to add scan permission for %s group!", nonInteractiveGroupName)
	}

	return nil
}

func uploadProfile(sc ClientInterface) error {
	profilePath := defaultProfileAbsolutePath
	if !pkgHelper.RunningInCluster() {
		profilePath = fmt.Sprintf("%v\\..\\%v\\%v", pkgHelper.GetExecutableFilePath(), localConfigsRelativePath,
			defaultQualityProfilesFileName)
	}

	if _, err := sc.UploadProfile("EDP way", profilePath); err != nil {
		return errors.Wrap(err, "Failed to upload EDP way profile!")
	}

	return nil
}

func createQualityGate(sc ClientInterface) error {
	if _, err := sc.CreateQualityGate("EDP way", []map[string]string{
		{"error": "80", "metric": "new_coverage", "op": "LT", "period": "1"},
		{"error": "0", "metric": "test_errors", "op": "GT"},
		{"error": "3", "metric": "new_duplicated_lines_density", "op": "GT", "period": "1"},
		{"error": "0", "metric": "test_failures", "op": "GT"},
		{"error": "0", "metric": "blocker_violations", "op": "GT"},
		{"error": "0", "metric": "critical_violations", "op": "GT"},
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

func (s Service) createAdminSecret(instance *v1alpha1.Sonar) (map[string][]byte, error) {
	secret, err := s.platformService.CreateSecret(instance.Name, instance.Namespace,
		fmt.Sprintf("%s-admin-password", instance.Name), map[string][]byte{
			"user":     []byte("admin"),
			"password": []byte(uniuri.New()),
		})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create password for Admin in %s Sonar!", instance.Name)
	}

	if err := s.platformService.SetOwnerReference(instance, secret); err != nil {
		return nil, errors.Wrapf(err, "Failed to set owner reference for secret %v", secret)
	}

	return secret.Data, nil
}

func (s Service) IsDeploymentReady(instance *v1alpha1.Sonar) (bool, error) {
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
