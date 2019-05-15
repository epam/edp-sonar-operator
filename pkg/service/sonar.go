package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
    "sigs.k8s.io/controller-runtime/pkg/client"
	"sonar-operator/pkg/apis/edp/v1alpha1"
	"strings"
	"time"

	"gopkg.in/resty.v1"
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
	log.Printf("Sonar component configuration has been started")

	var raw map[string]interface{}
	sonarApiUrl := "http://sonar." + instance.Namespace + ":9000/api"

	// Retrive password from secret
	credentials := s.platformService.GetSecret(instance)
	if credentials == nil {
		log.Printf("Sonar secret not found. Configuration failed")
		return errors.New("Sonar secret not found")
	}

	password := string(credentials["password"])

	// Change admin password
	resp, err := resty.R().
		SetBody("login=admin&password="+password+"&previousPassword=admin").
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBasicAuth("admin", "admin").
		Post(sonarApiUrl + "/users/change_password")

	if err != nil || resp.IsError() {
		log.Printf("Password change unsuccessful - " + resp.String())
		// Here we should return real error
		return errors.New("Password change unsuccessful - " + resp.String())
	}

	log.Printf("Password changed successfuly")

	// Install plugins
	// Do we have to pass this as a parameter&
	plugins := [4]string{"authoidc", "checkstyle", "findbugs", "pmd"}

	resp, err = resty.R().
		SetBasicAuth("admin", password).
		Get(sonarApiUrl + "/plugins/installed")

	json.Unmarshal([]byte(resp.String()), &raw)

	for _, plugin := range plugins {
		if !strings.Contains(resp.String(), plugin) {
			resp, err := resty.R().
				SetBody("key="+plugin).
				SetHeader("Content-Type", "application/x-www-form-urlencoded").
				SetBasicAuth("admin", password).
				Post(sonarApiUrl + "/plugins/install")

			if err != nil || resp.IsError() {
				log.Printf("Plugin " + plugin + " installation failed - " + resp.String())
				return err
			}
			log.Printf("Plugin " + plugin + "has been installed")
		}
	}

	// Reboot Sonar
	resp, err = resty.R().
		SetBasicAuth("admin", password).
		Post(sonarApiUrl + "/system/restart")

	if err != nil || resp.IsError() {
		log.Printf("Sonar restart failed - " + resp.String())
		return errors.New("Sonar restart failed - " + resp.String())
	}

	// Wait for Sonar Up
	resp, err = resty.
		SetRetryCount(60).
		SetRetryWaitTime(10*time.Second).
		AddRetryCondition(
			func(response *resty.Response) (bool, error) {
				if response.IsError() {
					return response.IsError(), nil
				}
				json.Unmarshal([]byte(response.String()), &raw)
				log.Printf("Current Sonar status - " + raw["status"].(string))
				if raw["status"].(string) == "UP" {
					return false, nil
				} else {
					return true, nil
				}
			},
		).
		R().
		SetBasicAuth("admin", password).
		Get(sonarApiUrl + "/system/status")

	log.Printf("Sonar restarted")

	log.Printf("Sonar component configuration has been finished")

	return nil
}

// Invoking install method against SonarServiceImpl object should trigger list of methods, stored in client edp.PlatformService
func (s SonarServiceImpl) Install(instance v1alpha1.Sonar) error {
	log.Printf("Installing Sonar component has been started")
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

	log.Printf("Installing Sonar component has been finished")
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
