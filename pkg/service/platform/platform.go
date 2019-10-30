package platform

import (
	"fmt"
	"github.com/epmd-edp/sonar-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/sonar-operator/v2/pkg/service/platform/kubernetes"
	"github.com/epmd-edp/sonar-operator/v2/pkg/service/platform/openshift"
	"github.com/pkg/errors"
	coreV1Api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	Openshift  string = "openshift"
	Kubernetes string = "kubernetes"
)

type PlatformService interface {
	CreateSecret(sonar v1alpha1.Sonar, name string, data map[string][]byte) error
	GetConfigmap(namespace string, name string) (map[string]string, error)
	GetExternalEndpoint(namespace string, name string) (*string, error)
	CreateServiceAccount(sonar v1alpha1.Sonar) (*coreV1Api.ServiceAccount, error)
	CreateSecurityContext(sonar v1alpha1.Sonar) error
	CreateExternalEndpoint(sonar v1alpha1.Sonar) error
	CreateService(sonar v1alpha1.Sonar) error
	CreateVolume(sonar v1alpha1.Sonar) error
	CreateDbDeployment(sonar v1alpha1.Sonar) error
	CreateDeployment(sonar v1alpha1.Sonar) error
	CreateConfigMap(instance v1alpha1.Sonar, configMapName string, configMapData map[string]string) error
	GetAvailiableDeploymentReplicas(instance v1alpha1.Sonar) (*int, error)
	GetSecretData(namespace string, name string) (map[string][]byte, error)
	CreateJenkinsServiceAccount(namespace string, secretName string, serviceAccountType string) error
	CreateJenkinsScript(namespace string, configMap string) error
	CreateEDPComponentIfNotExist(sonar v1alpha1.Sonar, url string, icon string) error
}

func NewPlatformService(platformType string, scheme *runtime.Scheme, client *client.Client) (PlatformService, error) {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(platformType) {
	case Kubernetes:
		s := kubernetes.K8SService{}
		err = s.Init(restConfig, scheme)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to initialize Kubernetes platform service!")
		}
		return s, nil
	case Openshift:
		s := openshift.OpenshiftService{}
		err = s.Init(restConfig, scheme)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to initialize OpenShift platform service!")
		}
		return s, nil
	default:
		err := errors.New(fmt.Sprintf("Platform %s is not supported!", platformType))
		return nil, err
	}
}
