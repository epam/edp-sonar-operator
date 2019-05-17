package service

import (
	"context"
	"fmt"
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
	sonarApiUrl := fmt.Sprintf("http://sonar.%s:9000/api", instance.Namespace)

	//credentials := s.platformService.GetSecret(instance)
	//if credentials == nil {
	//	log.Println("Sonar secret not found. Configuration failed")
	//	return errors.New("sonar secret not found")
	//}
	//
	//password := string(credentials["password"])
	sc := sonarClient.SonarClient{}
	err := sc.InitNewRestClient(sonarApiUrl, "admin", "admin")
	if err != nil {
		return logErrorAndReturn(err)
	}

	//// Install plugins
	//// Do we have to pass this as a parameter&
	//plugins := [4]string{"authoidc", "checkstyle", "findbugs", "pmd"}
	//
	//resp, err = restClient.R().
	//	Get(sonarApiUrl + "/plugins/installed")
	//
	//err = json.Unmarshal([]byte(resp.String()), &raw)
	//if err != nil {
	//	return err
	//}
	//
	//for _, plugin := range plugins {
	//	if !strings.Contains(resp.String(), plugin) {
	//		resp, err := resty.R().
	//			SetBody("key="+plugin).
	//			SetHeader("Content-Type", "application/x-www-form-urlencoded").
	//			SetBasicAuth("admin", password).
	//			Post(sonarApiUrl + "/plugins/install")
	//
	//		if err != nil || resp.IsError() {
	//			log.Println("Plugin " + plugin + " installation failed - " + resp.String())
	//			return err
	//		}
	//		log.Println("Plugin " + plugin + "has been installed")
	//	}
	//}
	//
	//// Reboot Sonar
	//resp, err = resty.R().
	//	SetBasicAuth("admin", password).
	//	Post(sonarApiUrl + "/system/restart")
	//
	//if err != nil || resp.IsError() {
	//	log.Println("Sonar restart failed - " + resp.String())
	//	return errors.New("Sonar restart failed - " + resp.String())
	//}
	//
	//// Wait for Sonar Up
	//resp, err = resty.
	//	SetRetryCount(60).
	//	SetRetryWaitTime(10*time.Second).
	//	AddRetryCondition(
	//		func(response *resty.Response) (bool, error) {
	//			if response.IsError() {
	//				return response.IsError(), nil
	//			}
	//			json.Unmarshal([]byte(response.String()), &raw)
	//			log.Println("Current Sonar status - " + raw["status"].(string))
	//			if raw["status"].(string) == "UP" {
	//				return false, nil
	//			} else {
	//				return true, nil
	//			}
	//		},
	//	).
	//	R().
	//	SetBasicAuth("admin", password).
	//	Get(sonarApiUrl + "/system/status")
	//
	//log.Println("Sonar restarted")

	//_, err = sc.UploadProfile()
	//if err != nil {
	//	return err
	//}

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

	err = s.platformService.CreateSecret(instance)
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
