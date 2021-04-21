package kubernetes

import (
	"context"
	"fmt"
	edpCompApi "github.com/epam/edp-component-operator/pkg/apis/v1/v1alpha1"
	edpCompClient "github.com/epam/edp-component-operator/pkg/client"
	jenkinsV1Api "github.com/epam/edp-jenkins-operator/v2/pkg/apis/v2/v1alpha1"
	jenkinsScriptV1Client "github.com/epam/edp-jenkins-operator/v2/pkg/controller/jenkinsscript/client"
	jenkinsSAV1Client "github.com/epam/edp-jenkins-operator/v2/pkg/controller/jenkinsserviceaccount/client"
	"github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform/helper"
	platformHelper "github.com/epam/edp-sonar-operator/v2/pkg/service/platform/helper"
	"github.com/pkg/errors"
	coreV1Api "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	appsV1Client "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreV1Client "k8s.io/client-go/kubernetes/typed/core/v1"
	extensionsV1Client "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

var log = ctrl.Log.WithName("platform")

type K8SService struct {
	Scheme                      *runtime.Scheme
	coreClient                  coreV1Client.CoreV1Client
	AppsClient                  appsV1Client.AppsV1Client
	ExtensionsV1Client          extensionsV1Client.ExtensionsV1beta1Client
	JenkinsScriptClient         jenkinsScriptV1Client.EdpV1Client
	JenkinsServiceAccountClient jenkinsSAV1Client.EdpV1Client
	edpCompClient               edpCompClient.EDPComponentV1Client
}

func (service *K8SService) Init(config *rest.Config, scheme *runtime.Scheme) error {
	coreClient, err := coreV1Client.NewForConfig(config)
	if err != nil {
		return err
	}

	jenkinsScriptClient, err := jenkinsScriptV1Client.NewForConfig(config)
	if err != nil {
		return err
	}

	JenkinsServiceAccountClient, err := jenkinsSAV1Client.NewForConfig(config)
	if err != nil {
		return err
	}

	acl, err := appsV1Client.NewForConfig(config)
	if err != nil {
		return errors.New("appsV1 client initialization failed!")
	}

	ecl, err := extensionsV1Client.NewForConfig(config)
	if err != nil {
		return errors.New("extensionsV1 client initialization failed!")
	}

	compCl, err := edpCompClient.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to init edp component client")
	}

	service.coreClient = *coreClient
	service.ExtensionsV1Client = *ecl
	service.AppsClient = *acl
	service.JenkinsScriptClient = *jenkinsScriptClient
	service.JenkinsServiceAccountClient = *JenkinsServiceAccountClient
	service.edpCompClient = *compCl
	service.Scheme = scheme
	return nil
}

func (service K8SService) GetSecretData(namespace string, name string) (map[string][]byte, error) {
	sonarSecret, err := service.coreClient.Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil && k8serr.IsNotFound(err) {
		log.Info("Secret in namespace not found", "secret name", name, "namespace", namespace)
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return sonarSecret.Data, nil
}

func (service K8SService) CreateSecret(sonar v1alpha1.Sonar, name string, data map[string][]byte) error {

	labels := helper.GenerateLabels(sonar.Name)

	sonarSecretObject := &coreV1Api.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: sonar.Namespace,
			Labels:    labels,
		},
		Data: data,
		Type: "Opaque",
	}

	if err := controllerutil.SetControllerReference(&sonar, sonarSecretObject, service.Scheme); err != nil {
		return err
	}

	sonarSecret, err := service.coreClient.Secrets(sonarSecretObject.Namespace).Get(context.TODO(), sonarSecretObject.Name, metav1.GetOptions{})

	if err != nil && k8serr.IsNotFound(err) {
		log.V(1).Info("Creating a new Secret for Sonar", "namespace", sonarSecretObject.Namespace, "secret name", sonarSecretObject.Name, "sonar name", sonar.Name)

		sonarSecret, err = service.coreClient.Secrets(sonarSecretObject.Namespace).Create(context.TODO(), sonarSecretObject, metav1.CreateOptions{})

		if err != nil {
			return err
		}
		log.Info("Secret has been created", "namespace", sonarSecret.Namespace, "secret name", sonarSecret.Name)

	} else if err != nil {
		return err
	}

	return nil
}

func (service K8SService) GetExternalEndpoint(namespace string, name string) (*string, error) {
	r, err := service.ExtensionsV1Client.Ingresses(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	rs := "https"
	u := fmt.Sprintf("%v://%v%v", rs, r.Spec.Rules[0].Host,
		strings.TrimRight(r.Spec.Rules[0].HTTP.Paths[0].Path, platformHelper.UrlCutset))

	return &u, nil
}

func (service K8SService) CreateConfigMap(instance v1alpha1.Sonar, configMapName string, configMapData map[string]string) error {
	labels := platformHelper.GenerateLabels(instance.Name)
	configMapObject := &coreV1Api.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Data: configMapData,
	}

	if err := controllerutil.SetControllerReference(&instance, configMapObject, service.Scheme); err != nil {
		return errors.Wrapf(err, "Couldn't set reference for Config Map %v object", configMapObject.Name)
	}

	cm, err := service.coreClient.ConfigMaps(instance.Namespace).Get(context.TODO(), configMapObject.Name, metav1.GetOptions{})
	if err != nil {
		if k8serr.IsNotFound(err) {
			cm, err = service.coreClient.ConfigMaps(configMapObject.Namespace).Create(context.TODO(), configMapObject, metav1.CreateOptions{})
			if err != nil {
				return errors.Wrapf(err, "Couldn't create Config Map %v object", configMapObject.Name)
			}
			log.Info("ConfigMap has been created", "namespace", cm.Namespace, "config map name", cm.Name)
		}
		return errors.Wrapf(err, "Couldn't get ConfigMap %v object", configMapObject.Name)
	}
	return nil
}

func (service K8SService) CreateJenkinsScript(namespace string, configMap string) error {
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

	_, err := service.JenkinsScriptClient.Get(context.TODO(), configMap, namespace, metav1.GetOptions{})
	if err != nil {
		if k8serr.IsNotFound(err) {
			_, err = service.JenkinsScriptClient.Create(context.TODO(), js, namespace)
			if err != nil {
				return err
			}
		}
		return err
	}
	return nil

}

func (service K8SService) CreateJenkinsServiceAccount(namespace string, secretName string, serviceAccountType string) error {

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

	_, err := service.JenkinsServiceAccountClient.Get(context.TODO(), secretName, namespace, metav1.GetOptions{})
	if err != nil {
		if k8serr.IsNotFound(err) {
			_, err = service.JenkinsServiceAccountClient.Create(context.TODO(), jsa, namespace)
			if err != nil {
				return err
			}
		}
		return err
	}

	return nil
}

func (service K8SService) GetAvailiableDeploymentReplicas(instance v1alpha1.Sonar) (*int, error) {
	c, err := service.AppsClient.Deployments(instance.Namespace).Get(context.TODO(), instance.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	r := int(c.Status.AvailableReplicas)

	return &r, nil
}

func (service K8SService) CreateEDPComponentIfNotExist(sonar v1alpha1.Sonar, url string, icon string) error {
	_, err := service.edpCompClient.
		EDPComponents(sonar.Namespace).
		Get(sonar.Name, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if k8serr.IsNotFound(err) {
		return service.createEDPComponent(sonar, url, icon)
	}
	return errors.Wrapf(err, "failed to get edp component: %v", sonar.Name)
}

func (service K8SService) createEDPComponent(sonar v1alpha1.Sonar, url string, icon string) error {
	obj := &edpCompApi.EDPComponent{
		ObjectMeta: metav1.ObjectMeta{
			Name: sonar.Name,
		},
		Spec: edpCompApi.EDPComponentSpec{
			Type:    "sonar",
			Url:     url,
			Icon:    icon,
			Visible: true,
		},
	}
	if err := controllerutil.SetControllerReference(&sonar, obj, service.Scheme); err != nil {
		return err
	}
	_, err := service.edpCompClient.
		EDPComponents(sonar.Namespace).
		Create(obj)
	return err
}
