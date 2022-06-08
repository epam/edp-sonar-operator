package openshift

import (
	"context"
	"fmt"
	"os"
	"strings"

	appsV1client "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	projectV1Client "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	routeV1Client "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityV1Client "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	templateV1Client "github.com/openshift/client-go/template/clientset/versioned/typed/template/v1"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sonarApi "github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1"
	platformHelper "github.com/epam/edp-sonar-operator/v2/pkg/service/platform/helper"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform/kubernetes"
)

type OpenshiftPodsStateClient interface {
	appsV1client.AppsV1Interface
}

type RouteClient interface {
	routeV1Client.RouteV1Interface
}

type SecurityClient interface {
	securityV1Client.SecurityV1Interface
}

type ProjectClient interface {
	projectV1Client.ProjectV1Interface
}

type TemplateClient interface {
	templateV1Client.TemplateV1Interface
}

type OpenshiftService struct {
	kubernetes.K8SService

	templateClient TemplateClient
	projectClient  ProjectClient
	securityClient SecurityClient
	appClient      OpenshiftPodsStateClient
	routeClient    RouteClient
}

const (
	deploymentTypeEnvName           = "DEPLOYMENT_TYPE"
	deploymentConfigsDeploymentType = "deploymentConfigs"
)

func (service *OpenshiftService) Init(config *rest.Config, scheme *runtime.Scheme, client client.Client) error {

	err := service.K8SService.Init(config, scheme, client)
	if err != nil {
		return err
	}

	templateClient, err := templateV1Client.NewForConfig(config)
	if err != nil {
		return err
	}

	service.templateClient = templateClient
	projectClient, err := projectV1Client.NewForConfig(config)
	if err != nil {
		return err
	}

	service.projectClient = projectClient
	securityClient, err := securityV1Client.NewForConfig(config)
	if err != nil {
		return err
	}

	service.securityClient = securityClient
	appClient, err := appsV1client.NewForConfig(config)
	if err != nil {
		return err
	}

	service.appClient = appClient
	routeClient, err := routeV1Client.NewForConfig(config)
	if err != nil {
		return err
	}
	service.routeClient = routeClient

	return nil
}

// GetExternalEndpoint returns scheme and host name from Openshift
func (service *OpenshiftService) GetExternalEndpoint(ctx context.Context, namespace string, name string) (string, error) {
	r, err := service.routeClient.Routes(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil && k8serrors.IsNotFound(err) {
		return "", errors.Wrapf(err, "Route %v in namespace %v not found", name, namespace)
	} else if err != nil {
		return "", err
	}

	var routeScheme = "http"
	if r.Spec.TLS.Termination != "" {
		routeScheme = "https"
	}
	return fmt.Sprintf("%s://%s%s",
		routeScheme, r.Spec.Host, strings.TrimRight(r.Spec.Path, platformHelper.UrlCutset)), nil
}

func (service *OpenshiftService) GetAvailableDeploymentReplicas(instance *sonarApi.Sonar) (*int, error) {
	if os.Getenv(deploymentTypeEnvName) == deploymentConfigsDeploymentType {
		c, err := service.appClient.DeploymentConfigs(instance.Namespace).Get(context.Background(), instance.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		r := int(c.Status.AvailableReplicas)

		return &r, nil
	}
	return service.K8SService.GetAvailableDeploymentReplicas(instance)
}
