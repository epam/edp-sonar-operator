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

type Client struct {
	client resty.Client
}

type SonarService interface {
	// This is an entry point for service package. Invoked in err = r.service.Install(*instance) sonar_controller.go, Reconcile method.
	Install(instance v1alpha1.Sonar) error
	Configure(instance v1alpha1.Sonar) error
}

func NewSonarService(platformService PlatformService, k8sClient client.Client) SonarService {
	return SonarServiceImpl{platformService: platformService, k8sClient: k8sClient}
}

type SonarServiceImpl struct {
	// Providing sonar service implementation through the interface (platform abstract)
	platformService PlatformService
	k8sClient       client.Client
}

func (s SonarServiceImpl) Configure(instance v1alpha1.Sonar) error {
	log.Println("Sonar component configuration has been started")

	sonarApiUrl := fmt.Sprintf("http://sonar.%v:9000/api", instance.Namespace)

	// Retrieve password from secret
	credentials := s.platformService.GetSecret(instance.Namespace, instance.Name)
	if credentials == nil {
		log.Printf("Sonar secret not found. Configuration failed")
		return errors.New("sonar secret not found")
	}

	sc := sonarClient.SonarClient{}

	err := sc.InitNewRestClient(sonarApiUrl, "admin", "admin")
	if err != nil {
		return err
	}

	_, err = sc.UploadProfile()
	if err != nil {
		return err
	}

	log.Println("Sonar component configuration has been finished")
	return nil
}

// Invoking install method against SonarServiceImpl object should trigger list of methods, stored in client edp.PlatformService
func (s SonarServiceImpl) Install(instance v1alpha1.Sonar) error {
	log.Println("Installing Sonar component has been started")
	err := s.updateStatus(instance, true)
	if err != nil {
		return err
	}

	dbSecret := map[string][]byte{
		"database-user":     []byte("admin"),
		"database-password": []byte(uniuri.New()),
	}

	err = s.platformService.CreateSecret(instance, instance.Name+"-db", dbSecret)
	if err != nil {
		return err
	}

	adminSecret := map[string][]byte{
		"user":     []byte("admin"),
		"password": []byte(uniuri.New()),
	}

	err = s.platformService.CreateSecret(instance, instance.Name+"-admin-password", adminSecret)
	if err != nil {
		return err
	}

	sa, err := s.platformService.CreateServiceAccount(instance)
	if err != nil {
		return err
	}

	err = s.platformService.CreateSecurityContext(instance, sa)
	if err != nil {
		return err
	}

	err = s.platformService.CreateDeployConf(instance)
	if err != nil {
		return err
	}

	err = s.platformService.CreateExternalEndpoint(instance)
	if err != nil {
		return err
	}

	err = s.platformService.CreateVolume(instance)
	if err != nil {
		return err
	}

	err = s.platformService.CreateService(instance)
	if err != nil {
		return err
	}

	err = s.platformService.CreateDbDeployConf(instance)
	if err != nil {
		return err
	}

	log.Println("Installing Sonar component has been finished")
	return nil
}

func (s SonarServiceImpl) updateStatus(instance v1alpha1.Sonar, value bool) error {
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
