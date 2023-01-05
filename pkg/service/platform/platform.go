package platform

import (
	"context"
	"fmt"
	"strings"

	coreV1Api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sonarApi "github.com/epam/edp-sonar-operator/v2/api/edp/v1"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform/kubernetes"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform/openshift"
)

const (
	Openshift  string = "openshift"
	Kubernetes string = "kubernetes"
)

type Service interface {
	CreateSecret(sonarName, namespace, secretName string, data map[string][]byte) (*coreV1Api.Secret, error)
	GetExternalEndpoint(ctx context.Context, namespace string, name string) (string, error)
	CreateConfigMap(instance *sonarApi.Sonar, configMapName string, configMapData map[string]string) error
	GetAvailableDeploymentReplicas(instance *sonarApi.Sonar) (*int, error)
	GetSecretData(namespace string, name string) (map[string][]byte, error)
	CreateJenkinsServiceAccount(namespace string, secretName string, serviceAccountType string) error
	CreateJenkinsScript(namespace string, configMap string) error
	CreateEDPComponentIfNotExist(sonar *sonarApi.Sonar, url string, icon string) error
	SetOwnerReference(sonar *sonarApi.Sonar, object client.Object) error
}

func NewService(platformType string, scheme *runtime.Scheme, client client.Client) (Service, error) {
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
		if err = s.Init(restConfig, scheme, client); err != nil {
			return nil, fmt.Errorf("failed to initialize Kubernetes platform service: %w", err)
		}

		return s, nil
	case Openshift:
		s := openshift.OpenshiftService{}
		if err = s.Init(restConfig, scheme, client); err != nil {
			return nil, fmt.Errorf("failed to initialize OpenShift platform service: %w", err)
		}

		return &s, nil
	default:
		return nil, fmt.Errorf("platform %s is not supported", platformType)
	}
}
