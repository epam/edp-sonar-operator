package service

import (
	"fmt"
	"k8s.io/client-go/rest"
	"sonar-operator/pkg/apis/edp/v1alpha1"

	appsV1client "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	projectV1Client "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	routeV1Client "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityTypedClient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	templateV1Client "github.com/openshift/client-go/template/clientset/versioned/typed/template/v1"
)

type OpenshiftService struct {
	K8SService

	templateClient templateV1Client.TemplateV1Client
	projectClient  projectV1Client.ProjectV1Client
	securityClient securityTypedClient.SecurityV1Client
	appClient      appsV1client.AppsV1Client
	routeClient    routeV1Client.RouteV1Client
}

func (service OpenshiftService) Init(config *rest.Config) error {
	err := service.K8SService.Init(config)
	if err != nil {
		return logErrorAndReturn(err)
	}
	templateClient, err := templateV1Client.NewForConfig(config)
	if err != nil {
		return logErrorAndReturn(err)
	}
	service.templateClient = *templateClient
	projectClient, err := projectV1Client.NewForConfig(config)
	if err != nil {
		return logErrorAndReturn(err)
	}
	service.projectClient = *projectClient
	securityClient, err := securityTypedClient.NewForConfig(config)
	if err != nil {
		return logErrorAndReturn(err)
	}
	service.securityClient = *securityClient
	appClient, err := appsV1client.NewForConfig(config)
	if err != nil {
		return logErrorAndReturn(err)
	}
	service.appClient = *appClient
	routeClient, err := routeV1Client.NewForConfig(config)
	if err != nil {
		return logErrorAndReturn(err)
	}
	service.routeClient = *routeClient
	return nil
}

func (service OpenshiftService) CreateServiceAccount(sonar v1alpha1.Sonar) error {
	fmt.Printf("Create service account for sonar: %v", sonar.Name)
	return nil
}
