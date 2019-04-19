package service

import (
	"k8s.io/client-go/tools/clientcmd"
	"sonar-operator/pkg/apis/edp/v1alpha1"
)

type PlatformService interface {
	CreateSecret(sonar v1alpha1.Sonar) error
	CreateServiceAccount(sonar v1alpha1.Sonar) error
}

func NewPlatformService() (PlatformService, error) {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, logErrorAndReturn(err)
	}
	platform := OpenshiftService{}
	err = platform.Init(restConfig)
	if err != nil {
		return nil, logErrorAndReturn(err)
	}
	return platform, nil
}
