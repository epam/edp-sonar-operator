package platform

import (
	"github.com/epmd-edp/sonar-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/sonar-operator/v2/pkg/service/platform/openshift"
	coreV1Api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
)

type PlatformService interface {
	CreateSecret(sonar v1alpha1.Sonar, name string, data map[string][]byte) error
	GetConfigmap(namespace string, name string) (map[string]string, error)
	GetExternalEndpoint(namespace string, name string) (*string, error)
	CreateServiceAccount(sonar v1alpha1.Sonar) (*coreV1Api.ServiceAccount, error)
	CreateSecurityContext(sonar v1alpha1.Sonar, sa *coreV1Api.ServiceAccount) error
	CreateExternalEndpoint(sonar v1alpha1.Sonar) error
	CreateService(sonar v1alpha1.Sonar) error
	CreateVolume(sonar v1alpha1.Sonar) error
	CreateDbDeployConf(sonar v1alpha1.Sonar) error
	CreateDeployConf(sonar v1alpha1.Sonar) error
	CreateConfigMap(instance v1alpha1.Sonar, configMapName string, configMapData map[string]string) error
	GetAvailiableDeploymentReplicas(instance v1alpha1.Sonar) (*int, error)
	GetSecretData(namespace string, name string) (map[string][]byte, error)
	CreateJenkinsServiceAccount(namespace string, secretName string, serviceAccountType string) error
	CreateJenkinsScript(namespace string, configMap string) error
}

func NewPlatformService(scheme *runtime.Scheme) (PlatformService, error) {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}

	platform := openshift.OpenshiftService{}

	err = platform.Init(restConfig, scheme)
	if err != nil {
		return nil, err
	}
	return platform, nil
}
