package kubernetes

import (
	"fmt"
	jenkinsV1Api "github.com/epmd-edp/jenkins-operator/v2/pkg/apis/v2/v1alpha1"
	jenkinsScriptV1Client "github.com/epmd-edp/jenkins-operator/v2/pkg/controller/jenkinsscript/client"
	jenkinsSAV1Client "github.com/epmd-edp/jenkins-operator/v2/pkg/controller/jenkinsserviceaccount/client"
	"github.com/epmd-edp/sonar-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/sonar-operator/v2/pkg/service/platform/helper"
	platformHelper "github.com/epmd-edp/sonar-operator/v2/pkg/service/platform/helper"
	sonarSpec "github.com/epmd-edp/sonar-operator/v2/pkg/service/sonar/spec"
	"github.com/pkg/errors"
	coreV1Api "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	coreV1Client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var logV2 = logf.Log.WithName("platform")

type K8SService struct {
	Scheme                      *runtime.Scheme
	coreClient                  coreV1Client.CoreV1Client
	JenkinsScriptClient         jenkinsScriptV1Client.EdpV1Client
	JenkinsServiceAccountClient jenkinsSAV1Client.EdpV1Client
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

	service.coreClient = *coreClient
	service.JenkinsScriptClient = *jenkinsScriptClient
	service.JenkinsServiceAccountClient = *JenkinsServiceAccountClient
	service.Scheme = scheme
	return nil
}

func (service K8SService) GetConfigmap(namespace string, name string) (map[string]string, error) {
	configmap, err := service.coreClient.ConfigMaps(namespace).Get(name, metav1.GetOptions{})

	if err != nil && k8serr.IsNotFound(err) {
		logV2.Info(fmt.Sprintf("Config map %v in namespace %v not found", name, namespace))
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return configmap.Data, nil
}

func (service K8SService) GetSecretData(namespace string, name string) (map[string][]byte, error) {
	sonarSecret, err := service.coreClient.Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil && k8serr.IsNotFound(err) {
		logV2.Info(fmt.Sprintf("Secret %v in namespace %v not found", name, namespace))
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

	sonarSecret, err := service.coreClient.Secrets(sonarSecretObject.Namespace).Get(sonarSecretObject.Name, metav1.GetOptions{})

	if err != nil && k8serr.IsNotFound(err) {
		logV2.V(1).Info(fmt.Sprintf("Creating a new Secret %s/%s for Sonar %s", sonarSecretObject.Namespace, sonarSecretObject.Name, sonar.Name))

		sonarSecret, err = service.coreClient.Secrets(sonarSecretObject.Namespace).Create(sonarSecretObject)

		if err != nil {
			return err
		}
		logV2.Info(fmt.Sprintf("Secret %s/%s has been created", sonarSecret.Namespace, sonarSecret.Name))

	} else if err != nil {
		return err
	}

	return nil
}

func (service K8SService) CreateVolume(sonar v1alpha1.Sonar) error {

	labels := helper.GenerateLabels(sonar.Name)

	for _, volume := range sonar.Spec.Volumes {

		sonarVolumeObject := &coreV1Api.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sonar.Name + "-" + volume.Name,
				Namespace: sonar.Namespace,
				Labels:    labels,
			},
			Spec: coreV1Api.PersistentVolumeClaimSpec{
				AccessModes: []coreV1Api.PersistentVolumeAccessMode{
					coreV1Api.ReadWriteOnce,
				},
				StorageClassName: &volume.StorageClass,
				Resources: coreV1Api.ResourceRequirements{
					Requests: map[coreV1Api.ResourceName]resource.Quantity{
						coreV1Api.ResourceStorage: resource.MustParse(volume.Capacity),
					},
				},
			},
		}

		if err := controllerutil.SetControllerReference(&sonar, sonarVolumeObject, service.Scheme); err != nil {
			return err
		}

		sonarVolume, err := service.coreClient.PersistentVolumeClaims(sonarVolumeObject.Namespace).Get(sonarVolumeObject.Name, metav1.GetOptions{})

		if err != nil && k8serr.IsNotFound(err) {
			logV2.V(1).Info(fmt.Sprintf("Creating a new PersistantVolumeClaim %s/%s for %s", sonarVolumeObject.Namespace, sonarVolumeObject.Name, sonar.Name))

			sonarVolume, err = service.coreClient.PersistentVolumeClaims(sonarVolumeObject.Namespace).Create(sonarVolumeObject)

			if err != nil {
				return err
			}

			logV2.Info(fmt.Sprintf("PersistantVolumeClaim %s/%s has been created", sonarVolume.Namespace, sonarVolume.Name))
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (service K8SService) CreateServiceAccount(sonar v1alpha1.Sonar) (*coreV1Api.ServiceAccount, error) {

	labels := helper.GenerateLabels(sonar.Name)

	sonarServiceAccountObject := &coreV1Api.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sonar.Name,
			Namespace: sonar.Namespace,
			Labels:    labels,
		},
	}

	if err := controllerutil.SetControllerReference(&sonar, sonarServiceAccountObject, service.Scheme); err != nil {
		return nil, err
	}

	sonarServiceAccount, err := service.coreClient.ServiceAccounts(sonarServiceAccountObject.Namespace).Get(sonarServiceAccountObject.Name, metav1.GetOptions{})

	if err != nil && k8serr.IsNotFound(err) {
		logV2.V(1).Info(fmt.Sprintf("Creating a new ServiceAccount %s/%s for Sonar %s", sonarServiceAccountObject.Namespace, sonarServiceAccountObject.Name, sonar.Name))

		sonarServiceAccount, err = service.coreClient.ServiceAccounts(sonarServiceAccountObject.Namespace).Create(sonarServiceAccountObject)

		if err != nil {
			return nil, err
		}

		logV2.Info(fmt.Sprintf("ServiceAccount %s/%s has been created", sonarServiceAccount.Namespace, sonarServiceAccount.Name))
	} else if err != nil {
		return nil, err
	}

	return sonarServiceAccount, nil
}

func (service K8SService) CreateExternalEndpoint(sonar v1alpha1.Sonar) error {
	logV2.Info("No implementation for K8s yet.")
	return nil
}

func (service K8SService) CreateService(sonar v1alpha1.Sonar) error {
	portMap := map[string]int32{
		sonar.Name:         sonarSpec.Port,
		sonar.Name + "-db": sonarSpec.DBPort,
	}
	for _, serviceName := range []string{sonar.Name, sonar.Name + "-db"} {
		labels := helper.GenerateLabels(serviceName)

		sonarServiceObject, err := newSonarInternalBalancingService(serviceName, sonar.Namespace, labels, portMap[serviceName])

		if err := controllerutil.SetControllerReference(&sonar, sonarServiceObject, service.Scheme); err != nil {
			return err
		}

		sonarService, err := service.coreClient.Services(sonar.Namespace).Get(serviceName, metav1.GetOptions{})

		if err != nil && k8serr.IsNotFound(err) {
			logV2.V(1).Info(fmt.Sprintf("Creating a new service %s/%s for sonar %s", sonarServiceObject.Namespace, sonarServiceObject.Name, sonar.Name))

			sonarService, err = service.coreClient.Services(sonarServiceObject.Namespace).Create(sonarServiceObject)

			if err != nil {
				return err
			}

			logV2.Info(fmt.Sprintf("Service %s/%s has been created", sonarService.Namespace, sonarService.Name))
		} else if err != nil {
			return err
		}
	}

	return nil
}

func newSonarInternalBalancingService(serviceName string, namespace string, labels map[string]string, port int32) (*coreV1Api.Service, error) {
	return &coreV1Api.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: coreV1Api.ServiceSpec{
			Selector: labels,
			Ports: []coreV1Api.ServicePort{
				{
					TargetPort: intstr.IntOrString{StrVal: serviceName},
					Port:       port,
				},
			},
		},
	}, nil
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

	cm, err := service.coreClient.ConfigMaps(instance.Namespace).Get(configMapObject.Name, metav1.GetOptions{})
	if err != nil {
		if k8serr.IsNotFound(err) {
			cm, err = service.coreClient.ConfigMaps(configMapObject.Namespace).Create(configMapObject)
			if err != nil {
				return errors.Wrapf(err, "Couldn't create Config Map %v object", configMapObject.Name)
			}
			logV2.Info(fmt.Sprintf("ConfigMap %s/%s has been created", cm.Namespace, cm.Name))
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

	_, err := service.JenkinsScriptClient.Get(configMap, namespace, metav1.GetOptions{})
	if err != nil {
		if k8serr.IsNotFound(err) {
			_, err = service.JenkinsScriptClient.Create(js, namespace)
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

	_, err := service.JenkinsServiceAccountClient.Get(secretName, namespace, metav1.GetOptions{})
	if err != nil {
		if k8serr.IsNotFound(err) {
			_, err = service.JenkinsServiceAccountClient.Create(jsa, namespace)
			if err != nil {
				return err
			}
		}
		return err
	}

	return nil
}
