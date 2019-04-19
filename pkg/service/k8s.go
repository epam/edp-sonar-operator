package service

import (
	"fmt"
	coreV1Client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"log"
	"sonar-operator/pkg/apis/edp/v1alpha1"
)

type K8SService struct {
	coreClient coreV1Client.CoreV1Client
}

func (service K8SService) Init(config *rest.Config) error {
	coreClient, err := coreV1Client.NewForConfig(config)
	if err != nil {
		return logErrorAndReturn(err)
	}
	service.coreClient = *coreClient
	return nil
}

func (service K8SService) CreateSecret(sonar v1alpha1.Sonar) error {
	fmt.Printf("Create secret for sonar: %v", sonar.Name)
	return nil
}

func (service K8SService) CreateServiceAccount(sonar v1alpha1.Sonar) error {
	fmt.Printf("Method does not implementation")
	return nil
}

func logErrorAndReturn(err error) error {
	log.Printf("[ERROR] %v", err)
	return err
}
