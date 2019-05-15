package service

import (
	"sonar-operator/pkg/apis/edp/v1alpha1"

	coreV1Api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
)

type PlatformService interface {
	CreateSecret(sonar v1alpha1.Sonar) error
	GetSecret(sonar v1alpha1.Sonar) map[string][]byte
	CreateServiceAccount(sonar v1alpha1.Sonar) (*coreV1Api.ServiceAccount, error)
	CreateSecurityContext(sonar v1alpha1.Sonar, sa *coreV1Api.ServiceAccount) error
	CreateExternalEndpoint(sonar v1alpha1.Sonar) error
	CreateService(sonar v1alpha1.Sonar) error
	CreateVolume(sonar v1alpha1.Sonar) error
	CreateDbDeployConf(sonar v1alpha1.Sonar) error
	CreateDeployConf(sonar v1alpha1.Sonar) error
}

func NewPlatformService(scheme *runtime.Scheme) (PlatformService, error) {

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, logErrorAndReturn(err)
	}

	platform := OpenshiftService{}

	err = platform.Init(restConfig, scheme)
	if err != nil {
		return nil, logErrorAndReturn(err)
	}
	return platform, nil
}
