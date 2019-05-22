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
	StatusInstall   = "installing"
	StatusFailed    = "failed"
	StatusCreated   = "created"
	JenkinsUsername = "jenkins"
	GroupName       = "non-interactive-users"
	WebhookUrl      = "http://jenkins:8080/sonarqube-webhook/"
	ProfilePath     = "../configs/quality-profile.xml"
)

type Client struct {
	client resty.Client
}

type SonarService interface {
	// This is an entry point for service package. Invoked in err = r.service.Install(*instance) sonar_controller.go, Reconcile method.
	Install(instance v1alpha1.Sonar) error
	Configure(instance v1alpha1.Sonar) error
	ExposeConfiguration(instance v1alpha1.Sonar) error
}

func NewSonarService(platformService PlatformService, k8sClient client.Client) SonarService {
	return SonarServiceImpl{platformService: platformService, k8sClient: k8sClient}
}

type SonarServiceImpl struct {
	// Providing sonar service implementation through the interface (platform abstract)
	platformService PlatformService
	k8sClient       client.Client
}

func (s SonarServiceImpl) ExposeConfiguration(instance v1alpha1.Sonar) error {
	log.Println("Sonar expose configuration has been started")
	sonarApiUrl := fmt.Sprintf("http://%v.%v:9000/api", instance.Name, instance.Namespace)

	credentials := s.platformService.GetSecret(instance.Namespace, instance.Name+"-admin-password")
	if credentials == nil {
		logErrorAndReturn(errors.New("Sonar secret not found. Configuration failed"))
	}
	password := string(credentials["password"])

	sc := sonarClient.SonarClient{}
	err := sc.InitNewRestClient(sonarApiUrl, "admin", password)
	if err != nil {
		return logErrorAndReturn(err)
	}

	jenkinsPassword := uniuri.New()
	err = sc.CreateUser(JenkinsUsername, "Jenkins", jenkinsPassword)
	if err != nil {
		return err
	}

	err = sc.AddUserToGroup(GroupName, JenkinsUsername)
	if err != nil {
		return err
	}

	err = sc.AddPermissionsToUser(JenkinsUsername, "admin")
	if err != nil {
		return err
	}

	ciToken, err := sc.GenerateUserToken(JenkinsUsername)
	if err != nil {
		return err
	}

	ciSecret := map[string][]byte{
		"password": []byte(*ciToken),
	}

	if s.platformService.GetSecret(instance.Namespace, "sonar-ciuser-token") == nil {
		err = s.platformService.CreateSecret(instance, "sonar-ciuser-token", ciSecret)
		if err != nil {
			return resourceActionFailed(&instance, err)
		}
	}

	perf := s.platformService.GetConfigmap(instance.Namespace, "user-settings")

	if perfIntergation, ok := perf["perf_integration_enabled"]; ok && perfIntergation == "true" {
		perfPassword := uniuri.New()

		err = sc.CreateUser("perf", "Perf", perfPassword)
		if err != nil {
			return err
		}

		if s.platformService.GetSecret(instance.Namespace, "sonar-perfuser-password") == nil {
			perfSecret := map[string][]byte{
				"password": []byte(perfPassword),
			}

			err = s.platformService.CreateSecret(instance, "sonar-perfuser-password", perfSecret)
			if err != nil {
				return resourceActionFailed(&instance, err)
			}
		}

	} else {
		log.Println("Perf integration disabled or can't be determined")
	}

	log.Println("Sonar expose configuration has been finished")
	return nil
}

func (s SonarServiceImpl) Configure(instance v1alpha1.Sonar) error {
	log.Println("Sonar component configuration has been started")
	sonarApiUrl := fmt.Sprintf("http://%v.%v:9000/api", instance.Name, instance.Namespace)
	//sonarApiUrl := "https://example-sonar-am-sonar-operator-test.delivery.aws.main.edp.projects.epam.com/api"

	sc := sonarClient.SonarClient{}
	err := sc.InitNewRestClient(sonarApiUrl, "admin", "admin")
	if err != nil {
		return logErrorAndReturn(err)
	}
	// TODO(Serhii Shydlovskyi): Error handling here ?
	sc.WaitForStatusIsUp(60, 10)

	credentials := s.platformService.GetSecret(instance.Namespace, instance.Name+"-admin-password")
	if credentials == nil {
		return logErrorAndReturn(errors.New("Sonar secret not found. Configuration failed"))
	}
	password := string(credentials["password"])
	// TODO(Serhii Shydlovskyi): Error handling here ?
	sc.ChangePassword("admin", "admin", password)
	if err != nil {
		return err
	}

	err = sc.InitNewRestClient(sonarApiUrl, "admin", password)
	if err != nil {
		return err
	}

	plugins := []string{"authoidc", "checkstyle", "findbugs", "pmd"}
	err = sc.InstallPlugins(plugins)
	if err != nil {
		return err
	}

	_, err = sc.UploadProfile("EDP way", ProfilePath)
	if err != nil {
		return err
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
		return err
	}

	err = sc.CreateGroup(GroupName)
	if err != nil {
		return err
	}

	err = sc.AddPermissionsToGroup(GroupName, "scan")
	if err != nil {
		return err
	}

	err = sc.AddWebhook(JenkinsUsername, WebhookUrl)
	if err != nil {
		return err
	}

	err = sc.ConfigureGeneralSettings()
	if err != nil {
		return logErrorAndReturn(err)
	}

	log.Println("Sonar component configuration has been finished")
	return nil
}

// Invoking install method against SonarServiceImpl object should trigger list of methods, stored in client edp.PlatformService
func (s SonarServiceImpl) Install(instance v1alpha1.Sonar) error {

	if instance.Status.Status != StatusCreated {
		log.Printf("Installing Sonar component has been started")
		updateStatus(&instance, StatusInstall, time.Now())
	}

	dbSecret := map[string][]byte{
		"database-user":     []byte("admin"),
		"database-password": []byte(uniuri.New()),
	}

	err := s.platformService.CreateSecret(instance, instance.Name+"-db", dbSecret)
	if err != nil {
		return err
	}

	adminSecret := map[string][]byte{
		"user":     []byte("admin"),
		"password": []byte(uniuri.New()),
	}

	err = s.platformService.CreateSecret(instance, instance.Name+"-admin-password", adminSecret)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	sa, err := s.platformService.CreateServiceAccount(instance)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	err = s.platformService.CreateSecurityContext(instance, sa)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	err = s.platformService.CreateDeployConf(instance)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	err = s.platformService.CreateExternalEndpoint(instance)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	err = s.platformService.CreateVolume(instance)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	err = s.platformService.CreateService(instance)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	err = s.platformService.CreateDbDeployConf(instance)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	if instance.Status.Status != StatusCreated {
		log.Printf("Installing Sonar component has been finished")
		updateStatus(&instance, StatusCreated, time.Now())
	}

	err = s.updateAvailableStatus(instance, true)
	if err != nil {
		return resourceActionFailed(&instance, err)
	}

	_ = s.k8sClient.Update(context.TODO(), &instance)

	return nil
}

func (s SonarServiceImpl) updateAvailableStatus(instance v1alpha1.Sonar, value bool) error {
	if instance.Status.Available != value {
		instance.Status.Available = value
		instance.Status.LastTimeUpdated = time.Now()
		err := s.k8sClient.Update(context.TODO(), &instance)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateStatus(s *v1alpha1.Sonar, status string, time time.Time) {
	s.Status.Status = status
	s.Status.LastTimeUpdated = time
	log.Printf("Status for Sonar %v has been updated to '%v' at %v.", s.Name, status, time)
}

func resourceActionFailed(instance *v1alpha1.Sonar, err error) error {
	updateStatus(instance, StatusFailed, time.Now())
	log.Printf("[ERROR] %v", err)
	return err
}
