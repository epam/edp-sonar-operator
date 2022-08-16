package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	coreV1Api "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	appsV1Client "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreV1Client "k8s.io/client-go/kubernetes/typed/core/v1"
	networkingV1Client "k8s.io/client-go/kubernetes/typed/networking/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	edpCompApi "github.com/epam/edp-component-operator/pkg/apis/v1/v1"
	jenkinsV1Api "github.com/epam/edp-jenkins-operator/v2/pkg/apis/v2/v1"

	sonarApi "github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1"
	platformHelper "github.com/epam/edp-sonar-operator/v2/pkg/service/platform/helper"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform/helper"
)

var log = ctrl.Log.WithName("platform")

const namespaceField = "namespace"

// K8SClusterClient is client for k8s cluster
type K8SClusterClient interface {
	coreV1Client.CoreV1Interface
}

// PodsStateClient is client to control, update, snapshot pods state
type PodsStateClient interface {
	appsV1Client.AppsV1Interface
}

type NetworkingClient interface {
	networkingV1Client.NetworkingV1Interface
}

type K8SService struct {
	Scheme             *runtime.Scheme
	client             client.Client
	k8sClusterClient   K8SClusterClient
	AppsClient         PodsStateClient
	NetworkingV1Client NetworkingClient
}

func (s *K8SService) Init(config *rest.Config, scheme *runtime.Scheme, client client.Client) error {
	coreClient, err := coreV1Client.NewForConfig(config)
	if err != nil {
		return err
	}

	acl, err := appsV1Client.NewForConfig(config)
	if err != nil {
		return errors.New("appsV1 client initialization failed!")
	}

	ecl, err := networkingV1Client.NewForConfig(config)
	if err != nil {
		return errors.New("networkingV1 client initialization failed!")
	}

	s.client = client
	s.k8sClusterClient = coreClient
	s.NetworkingV1Client = ecl
	s.AppsClient = acl
	s.Scheme = scheme
	return nil
}

func (s K8SService) GetSecretData(namespace string, name string) (map[string][]byte, error) {
	sonarSecret, err := s.k8sClusterClient.Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil && k8serr.IsNotFound(err) {
		log.Info("Secret in namespace not found", "secret name", name, namespaceField, namespace)
		var emptyMap map[string][]byte
		return emptyMap, nil
	} else if err != nil {
		return nil, err
	}
	return sonarSecret.Data, nil
}

func (s K8SService) CreateSecret(sonarName, namespace, secretName string, data map[string][]byte) (*coreV1Api.Secret, error) {
	labels := helper.GenerateLabels(sonarName)

	sonarSecretObject := &coreV1Api.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: data,
		Type: "Opaque",
	}

	existingSecret, err := s.k8sClusterClient.Secrets(sonarSecretObject.Namespace).
		Get(context.TODO(), sonarSecretObject.Name, metav1.GetOptions{})

	if err != nil {
		if k8serr.IsNotFound(err) {
			log.V(1).Info("Creating a new Secret for Sonar", namespace, sonarSecretObject.Namespace, "secret name", sonarSecretObject.Name, "sonar name", sonarName)
			sonarSecret, errCreateSecret := s.k8sClusterClient.Secrets(sonarSecretObject.Namespace).Create(context.Background(), sonarSecretObject, metav1.CreateOptions{})
			if errCreateSecret != nil {
				return nil, errCreateSecret
			}
			log.Info("Secret has been created", namespace, sonarSecret.Namespace, "secret name", sonarSecret.Name)
			return sonarSecretObject, nil
		}
		return nil, err
	}

	return existingSecret, nil
}

func (s K8SService) GetExternalEndpoint(ctx context.Context, namespace string, name string) (string, error) {
	r, err := s.NetworkingV1Client.Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s%s", r.Spec.Rules[0].Host,
		strings.TrimRight(r.Spec.Rules[0].HTTP.Paths[0].Path, platformHelper.UrlCutset)), nil
}

func (s K8SService) CreateConfigMap(instance *sonarApi.Sonar, configMapName string, configMapData map[string]string) error {
	labels := platformHelper.GenerateLabels(instance.Name)
	configMapObject := &coreV1Api.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Data: configMapData,
	}

	if err := controllerutil.SetControllerReference(instance, configMapObject, s.Scheme); err != nil {
		return errors.Wrapf(err, "Couldn't set reference for Config Map %v object", configMapObject.Name)
	}
	_, err := s.k8sClusterClient.ConfigMaps(instance.Namespace).Get(context.Background(), configMapObject.Name, metav1.GetOptions{})
	if err != nil {
		if k8serr.IsNotFound(err) {
			cm, errCreateConfigMap := s.k8sClusterClient.ConfigMaps(configMapObject.Namespace).Create(context.Background(), configMapObject, metav1.CreateOptions{})
			if errCreateConfigMap != nil {
				return errors.Wrapf(errCreateConfigMap, "Couldn't create Config Map %v object", configMapObject.Name)
			}
			log.Info("ConfigMap has been created", namespaceField, cm.Namespace, "config map name", cm.Name)
			return nil
		}
		return errors.Wrapf(err, "Couldn't get ConfigMap %v object", configMapObject.Name)
	}
	return nil
}

func (s K8SService) CreateJenkinsScript(namespace string, configMap string) error {
	_, err := s.getJenkinsScript(configMap, namespace)
	if err != nil {
		if k8serr.IsNotFound(err) {
			js := &jenkinsV1Api.JenkinsScript{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      configMap,
					Namespace: namespace,
				},
				Spec: jenkinsV1Api.JenkinsScriptSpec{
					SourceCmName: configMap,
				},
			}

			if err = s.client.Create(context.TODO(), js); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil

}

func (s K8SService) getJenkinsScript(name, namespace string) (*jenkinsV1Api.JenkinsScript, error) {
	js := &jenkinsV1Api.JenkinsScript{}
	err := s.client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, js)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func (s K8SService) CreateJenkinsServiceAccount(namespace string, secretName string, serviceAccountType string) error {
	_, err := s.getJenkinsServiceAccount(secretName, namespace)
	if err != nil {
		if k8serr.IsNotFound(err) {
			jsa := &jenkinsV1Api.JenkinsServiceAccount{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: namespace,
				},
				Spec: jenkinsV1Api.JenkinsServiceAccountSpec{
					Type:        serviceAccountType,
					Credentials: secretName,
				},
			}

			if err = s.client.Create(context.TODO(), jsa); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

func (s K8SService) getJenkinsServiceAccount(name, namespace string) (*jenkinsV1Api.JenkinsServiceAccount, error) {
	jsa := &jenkinsV1Api.JenkinsServiceAccount{}
	err := s.client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, jsa)
	if err != nil {
		return nil, err
	}
	return jsa, nil
}

func (s K8SService) GetAvailableDeploymentReplicas(instance *sonarApi.Sonar) (*int, error) {
	c, err := s.AppsClient.Deployments(instance.Namespace).Get(context.Background(), instance.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	r := int(c.Status.AvailableReplicas)

	return &r, nil
}

func (s K8SService) CreateEDPComponentIfNotExist(sonar *sonarApi.Sonar, url string, icon string) error {
	_, err := s.getEDPComponent(sonar.Name, sonar.Namespace)
	if err != nil {
		if k8serr.IsNotFound(err) {
			return s.createEDPComponent(sonar, url, icon)
		}
		return errors.Wrapf(err, "failed to get edp component: %v", sonar.Name)
	}
	return nil
}

func (s K8SService) getEDPComponent(name, namespace string) (*edpCompApi.EDPComponent, error) {
	c := &edpCompApi.EDPComponent{}
	err := s.client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s K8SService) createEDPComponent(sonar *sonarApi.Sonar, url string, icon string) error {
	obj := &edpCompApi.EDPComponent{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sonar.Namespace,
			Name:      sonar.Name,
		},
		Spec: edpCompApi.EDPComponentSpec{
			Type:    "sonar",
			Url:     url,
			Icon:    icon,
			Visible: true,
		},
	}
	if err := controllerutil.SetControllerReference(sonar, obj, s.Scheme); err != nil {
		return err
	}
	return s.client.Create(context.TODO(), obj)
}

func (s K8SService) SetOwnerReference(sonar *sonarApi.Sonar, object client.Object) error {
	err := controllerutil.SetControllerReference(sonar, object, s.Scheme)
	return err
}
