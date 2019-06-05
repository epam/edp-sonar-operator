package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/dchest/uniuri"
	"gopkg.in/resty.v1"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sonar-operator/pkg/apis/edp/v1alpha1"
	sonarClient "sonar-operator/pkg/client"
	"time"
)

const (
	StatusInstall           = "installing"
	StatusFailed            = "failed"
	StatusCreated           = "created"
	StatusConfiguring       = "configuring"
	StatusConfigured        = "configured"
	StatusReady             = "ready"
	StatuseExposeConf       = "exposing config"
	JenkinsLogin            = "jenkins"
	JenkinsUsername         = "Jenkins"
	ReaduserLogin           = "read"
	ReaduserUsername        = "Read-only user"
	NonInteractiveGroupName = "non-interactive-users"
	WebhookUrl              = "http://jenkins:8080/sonarqube-webhook/"
	ProfilePath             = "/usr/local/configs/quality-profile.xml"
)

type Client struct {
	client resty.Client
}

type SonarService interface {
	// This is an entry point for service package. Invoked in err = r.service.Install(*instance) sonar_controller.go, Reconcile method.
	Install(instance *v1alpha1.Sonar) error
	Configure(instance *v1alpha1.Sonar) error
	ExposeConfiguration(instance *v1alpha1.Sonar) error
}

func NewSonarService(platformService PlatformService, k8sClient client.Client) SonarService {
	return SonarServiceImpl{platformService: platformService, k8sClient: k8sClient}
}

type SonarServiceImpl struct {
	// Providing sonar service implementation through the interface (platform abstract)
	platformService PlatformService
	k8sClient       client.Client
}

func (s SonarServiceImpl) ExposeConfiguration(instance *v1alpha1.Sonar) error {
	if instance.Status.Status == StatusConfigured || instance.Status.Status == "" {
		log.Println("Sonar expose configuration has been started")
		err := s.updateStatus(instance, StatuseExposeConf)
		if err != nil {
			return logErrorAndReturn(err)
		}
	}

	sonarApiUrl := fmt.Sprintf("http://%v.%v:9000/api", instance.Name, instance.Namespace)
	externalConfig := v1alpha1.SonarExternalConfiguration{nil, nil, nil, nil}

	credentials, err := s.platformService.GetSecret(instance.Namespace, instance.Name+"-admin-password")
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	password := string(credentials["password"])

	sc := sonarClient.SonarClient{}
	err = sc.InitNewRestClient(sonarApiUrl, "admin", password)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	jenkinsPassword := uniuri.New()
	err = sc.CreateUser(JenkinsLogin, JenkinsUsername, jenkinsPassword)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	err = sc.AddUserToGroup(NonInteractiveGroupName, JenkinsLogin)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	err = sc.AddPermissionsToUser(JenkinsLogin, "admin")
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	ciToken, err := sc.GenerateUserToken(JenkinsLogin)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	if ciToken != nil {
		ciSecret := map[string][]byte{
			"username": []byte(JenkinsLogin),
			"token":    []byte(*ciToken),
		}

		err = s.platformService.CreateSecret(*instance, instance.Name+"-ciuser-token", ciSecret)
		if err != nil {
			return s.resourceActionFailed(instance, err)
		}
	}

	readPassword := uniuri.New()
	err = sc.CreateUser(ReaduserLogin, ReaduserUsername, readPassword)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	readToken, err := sc.GenerateUserToken(ReaduserLogin)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	if readToken != nil {
		readSecret := map[string][]byte{
			"username": []byte(ReaduserLogin),
			"token":    []byte(*readToken),
		}

		err = s.platformService.CreateSecret(*instance, instance.Name+"-readuser-token", readSecret)
		if err != nil {
			return s.resourceActionFailed(instance, err)
		}
	}

	err = sc.AddUserToGroup(NonInteractiveGroupName, ReaduserLogin)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	identityServiceClientSecret := uniuri.New()
	identityServiceClientCredenrials := map[string][]byte{
		"client_id":     []byte(instance.Name),
		"client_secret": []byte(identityServiceClientSecret),
	}

	err = s.platformService.CreateSecret(*instance, instance.Name+"-is-credentials", identityServiceClientCredenrials)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	externalConfig.CiUser = &v1alpha1.SonarExternalConfigurationItem{instance.Name + "-ciuser-token", "Secret", "Token for CI tool user"}
	externalConfig.AdminUser = &v1alpha1.SonarExternalConfigurationItem{instance.Name + "-admin-password", "Secret", "Password for Sonar admin user"}
	externalConfig.ReadUser = &v1alpha1.SonarExternalConfigurationItem{instance.Name + "-readuser-token", "Secret", "Token for read-only user"}
	externalConfig.IsCredentials = &v1alpha1.SonarExternalConfigurationItem{instance.Name + "-is-credentials", "Secret", "Credentials for Identity Server integration"}

	err = s.updateExternalConfig(instance, externalConfig)
	if err != nil {
		return logErrorAndReturn(errors.New(fmt.Sprintf("Sonar expose configuration failed with error - %v", err)))
	}

	if instance.Status.Status == StatuseExposeConf {
		log.Println("Sonar configuration expose has been finished")
		err = s.updateStatus(instance, StatusReady)
		if err != nil {
			return logErrorAndReturn(err)
		}

	}

	err = s.updateAvailableStatus(instance, true)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	return nil
}

func (s SonarServiceImpl) Configure(instance *v1alpha1.Sonar) error {
	if instance.Status.Status == StatusCreated || instance.Status.Status == "" {
		log.Println("Sonar component configuration has been started")
		err := s.updateStatus(instance, StatusConfiguring)
		if err != nil {
			return logErrorAndReturn(err)
		}
	}

	sonarApiUrl := fmt.Sprintf("http://%v.%v:9000/api", instance.Name, instance.Namespace)

	sc := sonarClient.SonarClient{}
	err := sc.InitNewRestClient(sonarApiUrl, "admin", "admin")
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	// TODO(Serhii Shydlovskyi): Error handling here ?
	sc.WaitForStatusIsUp(60, 10)

	credentials, err := s.platformService.GetSecret(instance.Namespace, instance.Name+"-admin-password")
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}
	password := string(credentials["password"])
	// TODO(Serhii Shydlovskyi): Add check for password presence. Breaks status update.
	sc.ChangePassword("admin", "admin", password)
	//if err != nil {
	//	return err
	//}

	err = sc.InitNewRestClient(sonarApiUrl, "admin", password)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	plugins := []string{"authoidc", "checkstyle", "findbugs", "pmd"}
	err = sc.InstallPlugins(plugins)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	_, err = sc.UploadProfile("EDP way", ProfilePath)
	if err != nil {
		return s.resourceActionFailed(instance, err)
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
		return s.resourceActionFailed(instance, err)
	}

	err = sc.CreateGroup(NonInteractiveGroupName)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	err = sc.AddPermissionsToGroup(NonInteractiveGroupName, "scan")
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	err = sc.AddWebhook(JenkinsLogin, WebhookUrl)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	err = sc.ConfigureGeneralSettings()
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	if instance.Status.Status == StatusConfiguring {
		log.Println("Sonar component configuration has been finished")
		err = s.updateStatus(instance, StatusConfigured)
		if err != nil {
			return logErrorAndReturn(err)
		}
	}

	return nil
}

// Invoking install method against SonarServiceImpl object should trigger list of methods, stored in client edp.PlatformService
func (s SonarServiceImpl) Install(instance *v1alpha1.Sonar) error {

	if instance.Status.Status == "" || instance.Status.Status == StatusFailed {
		log.Printf("Installing Sonar component has been started")
		err := s.updateStatus(instance, StatusInstall)
		if err != nil {
			return logErrorAndReturn(err)
		}
	}

	dbSecret := map[string][]byte{
		"database-user":     []byte("admin"),
		"database-password": []byte(uniuri.New()),
	}

	err := s.platformService.CreateSecret(*instance, instance.Name+"-db", dbSecret)
	if err != nil {
		return err
	}

	adminSecret := map[string][]byte{
		"user":     []byte("admin"),
		"password": []byte(uniuri.New()),
	}

	err = s.platformService.CreateSecret(*instance, instance.Name+"-admin-password", adminSecret)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	sa, err := s.platformService.CreateServiceAccount(*instance)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	err = s.platformService.CreateSecurityContext(*instance, sa)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	err = s.platformService.CreateDeployConf(*instance)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	err = s.platformService.CreateExternalEndpoint(*instance)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	err = s.platformService.CreateVolume(*instance)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	err = s.platformService.CreateService(*instance)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	err = s.platformService.CreateDbDeployConf(*instance)
	if err != nil {
		return s.resourceActionFailed(instance, err)
	}

	if instance.Status.Status == StatusInstall {
		log.Printf("Installing Sonar component has been finished")
		err = s.updateStatus(instance, StatusCreated)
		if err != nil {
			return logErrorAndReturn(err)
		}
	}

	return nil
}

func (s SonarServiceImpl) updateAvailableStatus(instance *v1alpha1.Sonar, value bool) error {
	if instance.Status.Available != value {
		instance.Status.Available = value
		instance.Status.LastTimeUpdated = time.Now()
		err := s.k8sClient.Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s SonarServiceImpl) updateStatus(instance *v1alpha1.Sonar, status string) error {
	instance.Status.Status = status
	instance.Status.LastTimeUpdated = time.Now()
	err := s.k8sClient.Update(context.TODO(), instance)
	if err != nil {
		return err
	}

	log.Printf("Status for Sonar %v has been updated to '%v' at %v.", instance.Name, status, instance.Status.LastTimeUpdated)
	return nil
}

func (s SonarServiceImpl) updateExternalConfig(instance *v1alpha1.Sonar, config v1alpha1.SonarExternalConfiguration) error {
	instance.Spec.SonarExternalConfiguration = config

	err := s.k8sClient.Update(context.TODO(), instance)
	if err != nil {
		return err
	}
	return nil
}

func (s SonarServiceImpl) resourceActionFailed(instance *v1alpha1.Sonar, err error) error {
	if s.updateStatus(instance, StatusFailed) != nil {
		return err
	}
	return err
}
