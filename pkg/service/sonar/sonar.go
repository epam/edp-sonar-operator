package sonar

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/dchest/uniuri"
	jenkinsApi "github.com/epam/edp-jenkins-operator/v2/pkg/apis/v2/v1alpha1"
	jenkinsHelper "github.com/epam/edp-jenkins-operator/v2/pkg/controller/jenkinsscript/helper"
	platformHelper "github.com/epam/edp-jenkins-operator/v2/pkg/service/platform/helper"
	keycloakApi "github.com/epam/edp-keycloak-operator/pkg/apis/v1/v1alpha1"
	"github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	pkgHelper "github.com/epam/edp-sonar-operator/v2/pkg/helper"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform"
	sonarHelper "github.com/epam/edp-sonar-operator/v2/pkg/service/sonar/helper"
	sonarSpec "github.com/epam/edp-sonar-operator/v2/pkg/service/sonar/spec"
	"github.com/pkg/errors"
	"gopkg.in/resty.v1"
	"io/ioutil"
	k8sErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	JenkinsLogin            = "jenkins"
	JenkinsUsername         = "Jenkins"
	ReaduserLogin           = "read"
	ReaduserUsername        = "Read-only user"
	NonInteractiveGroupName = "non-interactive-users"
	WebhookUrl              = "sonarqube-webhook/"
	DefaultPassword         = "admin"
	ClaimName               = "roles"

	defaultConfigFilesAbsolutePath = "/usr/local/"
	localConfigsRelativePath       = "configs"

	defaultProfileAbsolutePath = defaultConfigFilesAbsolutePath + localConfigsRelativePath + "/" + defaultQualityProfilesFileName

	defaultQualityProfilesFileName = "quality-profile.xml"

	imgFolder = "img"
	sonarIcon = "sonar.svg"
)

type Client struct {
	client resty.Client
}

type SonarService interface {
	Configure(instance v1alpha1.Sonar) (*v1alpha1.Sonar, error, bool)
	ExposeConfiguration(instance v1alpha1.Sonar) (*v1alpha1.Sonar, error)
	Integration(instance v1alpha1.Sonar) (*v1alpha1.Sonar, error)
	IsDeploymentReady(instance v1alpha1.Sonar) (bool, error)
}

func NewSonarService(platformService platform.PlatformService, k8sClient client.Client, k8sScheme *runtime.Scheme) SonarService {
	return SonarServiceImpl{platformService: platformService, k8sClient: k8sClient, k8sScheme: k8sScheme}
}

type SonarServiceImpl struct {
	// Providing sonar service implementation through the interface (platform abstract)
	platformService platform.PlatformService
	k8sClient       client.Client
	k8sScheme       *runtime.Scheme
}

func (s SonarServiceImpl) initSonarClient(instance *v1alpha1.Sonar, defaultPassword bool) (*sonar.SonarClient, error) {
	sc := &sonar.SonarClient{}

	password := DefaultPassword
	if !defaultPassword {
		adminSecretName := fmt.Sprintf("%v-admin-password", instance.Name)
		credentials, err := s.platformService.GetSecretData(instance.Namespace, adminSecretName)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to get %v secret for Sonar client!", adminSecretName)
		}
		password = string(credentials["password"])
	}

	u, err := s.platformService.GetExternalEndpoint(instance.Namespace, instance.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get route for %v", instance.Name)
	}

	err = sc.InitNewRestClient(fmt.Sprintf("%v/api", *u), "admin", password)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to init new Sonar client!")
	}

	return sc, nil
}

func (s SonarServiceImpl) Integration(instance v1alpha1.Sonar) (*v1alpha1.Sonar, error) {
	sc, err := s.initSonarClient(&instance, false)
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

	url, err := s.platformService.GetExternalEndpoint(instance.Namespace, instance.Name)
	if err != nil {
		return nil, err
	}
	err = sc.ConfigureGeneralSettings("value", "sonar.core.serverBaseURL", *url)
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to configure sonar.core.serverBaseURL!")
	}
	cl, err := s.getKeycloakClient(instance)
	if err != nil {
		return &instance, err
	}

	if cl == nil {
		err = s.createKeycloakClient(instance, *url)
	}

	if err != nil {
		return &instance, err
	}

	err = sc.ConfigureGeneralSettings("value", "sonar.auth.oidc.clientId.secured", instance.Name)
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to configure sonar.auth.oidc.clientId.secured!")
	}

	err = sc.ConfigureGeneralSettings("value", "sonar.auth.oidc.groupsSync.claimName", ClaimName)
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

func (s SonarServiceImpl) getKeycloakRealm(instance v1alpha1.Sonar) (*keycloakApi.KeycloakRealm, error) {
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

func (s SonarServiceImpl) getKeycloakClient(instance v1alpha1.Sonar) (*keycloakApi.KeycloakClient, error) {
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

func (s SonarServiceImpl) createKeycloakClient(instance v1alpha1.Sonar, baseUrl string) error {
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
		},
	}
	return s.k8sClient.Create(context.TODO(), cl)
}

func (s SonarServiceImpl) ExposeConfiguration(instance v1alpha1.Sonar) (*v1alpha1.Sonar, error) {

	sc, err := s.initSonarClient(&instance, false)
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to initialize Sonar Client!")
	}

	jenkinsPassword := uniuri.New()
	err = sc.CreateUser(JenkinsLogin, JenkinsUsername, jenkinsPassword)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to create user %v in Sonar!", JenkinsUsername)
	}

	err = sc.AddUserToGroup(NonInteractiveGroupName, JenkinsLogin)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to add %v user in %v group!", JenkinsLogin, NonInteractiveGroupName)
	}

	err = sc.AddPermissionsToUser(JenkinsLogin, "admin")
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to add admin persmissions to  %v user", JenkinsLogin)
	}

	ciToken, err := sc.GenerateUserToken(JenkinsLogin)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to generate token for %v user", JenkinsLogin)
	}

	ciUserName := fmt.Sprintf("%v-ciuser-token", instance.Name)
	if ciToken != nil {
		ciSecret := map[string][]byte{
			"username": []byte(JenkinsLogin),
			"secret":   []byte(*ciToken),
		}

		err = s.platformService.CreateSecret(instance, ciUserName, ciSecret)
		if err != nil {
			return &instance, errors.Wrapf(err, "Failed to create secret for  %v user", ciUserName)
		}
	}

	err = s.platformService.CreateJenkinsServiceAccount(instance.Namespace, ciUserName, "token")
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to create Jenkins Service Account for %v", ciUserName)
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
		return &instance, errors.Wrapf(err, "Failed to parse default Jenkins plugin template!")
	}

	configMapName := fmt.Sprintf("%s-%s", instance.Name, sonarSpec.JenkinsPluginConfigPostfix)
	configMapData := map[string]string{
		jenkinsHelper.JenkinsDefaultScriptConfigMapKey: jenkinsScriptContext.String(),
	}

	err = s.platformService.CreateConfigMap(instance, configMapName, configMapData)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to create Config Map %v", configMapName)
	}

	err = s.platformService.CreateJenkinsScript(instance.Namespace, configMapName)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to create Jenkins Script for %v", ciUserName)
	}

	readPassword := uniuri.New()
	err = sc.CreateUser(ReaduserLogin, ReaduserUsername, readPassword)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to create %v user in Sonar!", ReaduserUsername)
	}

	readToken, err := sc.GenerateUserToken(ReaduserLogin)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to generate token for %v user", ReaduserLogin)
	}

	readUserSecretName := fmt.Sprintf("%v-readuser-token", instance.Name)
	if readToken != nil {
		readSecret := map[string][]byte{
			"username": []byte(ReaduserLogin),
			"token":    []byte(*readToken),
		}

		err = s.platformService.CreateSecret(instance, readUserSecretName, readSecret)
		if err != nil {
			return &instance, errors.Wrapf(err, "Failed to create secret for  %v user", readUserSecretName)
		}
	}

	err = sc.AddUserToGroup(NonInteractiveGroupName, ReaduserLogin)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to add %v user in %v group!", ReaduserLogin, NonInteractiveGroupName)
	}

	identityServerSecretName := fmt.Sprintf("%v-is-credentials", instance.Name)
	identityServiceClientSecret := uniuri.New()
	identityServiceClientCredentials := map[string][]byte{
		"client_id":     []byte(instance.Name),
		"client_secret": []byte(identityServiceClientSecret),
	}

	err = s.platformService.CreateSecret(instance, identityServerSecretName, identityServiceClientCredentials)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to create secret for  %v Keycloak client!", readUserSecretName)
	}

	err = s.createEDPComponent(instance)

	return &instance, err
}

func (s SonarServiceImpl) createEDPComponent(sonar v1alpha1.Sonar) error {
	url, err := s.platformService.GetExternalEndpoint(sonar.Namespace, sonar.Name)
	if err != nil {
		return err
	}
	icon, err := getIcon()
	if err != nil {
		return err
	}
	return s.platformService.CreateEDPComponentIfNotExist(sonar, *url, *icon)
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

func (s SonarServiceImpl) Configure(instance v1alpha1.Sonar) (*v1alpha1.Sonar, error, bool) {
	dbSecret := map[string][]byte{
		"database-user":     []byte("admin"),
		"database-password": []byte(uniuri.New()),
	}

	sonarDbName := fmt.Sprintf("%v-db", instance.Name)
	err := s.platformService.CreateSecret(instance, sonarDbName, dbSecret)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to create secret for %s", sonarDbName), false
	}

	adminSecret := map[string][]byte{
		"user":     []byte("admin"),
		"password": []byte(uniuri.New()),
	}

	adminSecretName := fmt.Sprintf("%v-admin-password", instance.Name)
	err = s.platformService.CreateSecret(instance, adminSecretName, adminSecret)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to create password for Admin in %s Sonar!", instance.Name), false
	}

	sc, err := s.initSonarClient(&instance, true)
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to initialize Sonar Client!"), false
	}

	sc.WaitForStatusIsUp(60, 10)

	credentials, err := s.platformService.GetSecretData(instance.Namespace, adminSecretName)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to get secret data from %v!", adminSecretName), false
	}
	password := string(credentials["password"])
	sc.ChangePassword("admin", DefaultPassword, password)

	sc, err = s.initSonarClient(&instance, false)
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to initialize Sonar Client!"), false
	}

	plugins := []string{"authoidc", "checkstyle", "findbugs", "pmd", "jacoco", "xml", "javascript", "go", "ansible", "yaml", "python", "csharp"}
	err = sc.InstallPlugins(plugins)
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to install plugins for Sonar!"), false
	}

	executableFilePath := pkgHelper.GetExecutableFilePath()
	profilePath := defaultProfileAbsolutePath

	if !pkgHelper.RunningInCluster() {
		profilePath = fmt.Sprintf("%v\\..\\%v\\%v", executableFilePath, localConfigsRelativePath, defaultQualityProfilesFileName)
	}
	_, err = sc.UploadProfile("EDP way", profilePath)
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to upload EDP way profile!"), false
	}

	qgContidions := []map[string]string{
		{"error": "80", "metric": "new_coverage", "op": "LT", "period": "1"},
		{"error": "0", "metric": "test_errors", "op": "GT"},
		{"error": "3", "metric": "new_duplicated_lines_density", "op": "GT", "period": "1"},
		{"error": "0", "metric": "test_failures", "op": "GT"},
		{"error": "0", "metric": "blocker_violations", "op": "GT"},
		{"error": "0", "metric": "critical_violations", "op": "GT"},
	}
	_, err = sc.CreateQualityGate("EDP way", qgContidions)
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to configure EDP way quality gate!"), false
	}

	err = sc.CreateGroup(NonInteractiveGroupName)
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to create %v group!", NonInteractiveGroupName), false
	}

	err = sc.AddPermissionsToGroup(NonInteractiveGroupName, "scan")
	if err != nil {
		return &instance, errors.Wrapf(err, "Failed to add scan permission for %v group!", NonInteractiveGroupName), false
	}

	jenkinsUrl := s.getInternalJenkinsUrl(instance.Namespace)
	if jenkinsUrl != nil {
		err = sc.AddWebhook(JenkinsLogin, fmt.Sprintf("%v/%v", *jenkinsUrl, WebhookUrl))
		if err != nil {
			return &instance, errors.Wrap(err, "Failed to add Jenkins webhook!"), false
		}
	}

	err = sc.ConfigureGeneralSettings("values", "sonar.typescript.lcov.reportPaths", "coverage/lcov.info")
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to configure sonar.typescript.lcov.reportPaths!"), false
	}

	err = sc.ConfigureGeneralSettings("values", "sonar.coverage.jacoco.xmlReportPaths", "target/site/jacoco/jacoco.xml")
	if err != nil {
		return &instance, errors.Wrap(err, "Failed to configure sonar.coverage.jacoco.xmlReportPaths!"), false
	}

	return &instance, nil, true
}

func (s SonarServiceImpl) IsDeploymentReady(instance v1alpha1.Sonar) (bool, error) {
	isReady := false

	r, err := s.platformService.GetAvailiableDeploymentReplicas(instance)
	if err != nil {
		return isReady, err
	}

	if *r == 1 {
		isReady = true
	}

	return isReady, nil
}

func (s SonarServiceImpl) getInternalJenkinsUrl(namespace string) *string {
	options := client.ListOptions{Namespace: namespace}
	jenkinsList := &jenkinsApi.JenkinsList{}

	err := s.k8sClient.List(context.TODO(), jenkinsList, &options)
	if err != nil {
		log.Printf("Unable to get Jenkins CRs in namespace %v", namespace)
		return nil
	}

	if len(jenkinsList.Items) == 0 {
		log.Printf("Jenkins installation is not found in namespace %v", namespace)
		return nil
	}

	jenkins := jenkinsList.Items[0]
	basePath := ""
	if len(jenkins.Spec.BasePath) > 0 {
		basePath = fmt.Sprintf("/%v", jenkins.Spec.BasePath)
	}
	jenkinsInternalUrl := fmt.Sprintf("http://jenkins.%s:8080%v", namespace, basePath)
	return &jenkinsInternalUrl
}
