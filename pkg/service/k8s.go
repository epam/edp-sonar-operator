package service

import (
	"fmt"
	"log"
	"sonar-operator/pkg/apis/edp/v1alpha1"

	coreV1Api "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	coreV1Client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type K8SService struct {
	scheme     *runtime.Scheme
	coreClient coreV1Client.CoreV1Client
}

func (service *K8SService) Init(config *rest.Config, scheme *runtime.Scheme) error {

	coreClient, err := coreV1Client.NewForConfig(config)
	if err != nil {
		return logErrorAndReturn(err)
	}
	service.coreClient = *coreClient
	service.scheme = scheme
	return nil
}

func (service K8SService) GetSecret(namespace string, name string) map[string][]byte {
	sonarSecret, err := service.coreClient.Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil
	}
	return sonarSecret.Data
}

func (service K8SService) CreateSecret(sonar v1alpha1.Sonar, name string, data map[string][]byte) error {

	labels := generateLabels(sonar.Name)

	sonarSecretObject := &coreV1Api.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: sonar.Namespace,
			Labels:    labels,
		},
		Data: data,
		Type: "Opaque",
	}

	if err := controllerutil.SetControllerReference(&sonar, sonarSecretObject, service.scheme); err != nil {
		return logErrorAndReturn(err)
	}

	sonarSecret, err := service.coreClient.Secrets(sonarSecretObject.Namespace).Get(sonarSecretObject.Name, metav1.GetOptions{})

	if err != nil && k8serr.IsNotFound(err) {
		log.Printf("Creating a new Secret %s/%s for Sonar %s", sonarSecretObject.Namespace, sonarSecretObject.Name, sonar.Name)

		sonarSecret, err = service.coreClient.Secrets(sonarSecretObject.Namespace).Create(sonarSecretObject)

		if err != nil {
			return logErrorAndReturn(err)
		}
		log.Printf("Secret %s/%s has been created", sonarSecret.Namespace, sonarSecret.Name)

	} else if err != nil {
		return logErrorAndReturn(err)
	}

	return nil
}

func (service K8SService) CreateVolume(sonar v1alpha1.Sonar) error {

	labels := generateLabels(sonar.Name)

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

		if err := controllerutil.SetControllerReference(&sonar, sonarVolumeObject, service.scheme); err != nil {
			return logErrorAndReturn(err)
		}

		sonarVolume, err := service.coreClient.PersistentVolumeClaims(sonarVolumeObject.Namespace).Get(sonarVolumeObject.Name, metav1.GetOptions{})

		if err != nil && k8serr.IsNotFound(err) {
			log.Printf("Creating a new PersistantVolumeClaim %s/%s for %s", sonarVolumeObject.Namespace, sonarVolumeObject.Name, sonar.Name)

			sonarVolume, err = service.coreClient.PersistentVolumeClaims(sonarVolumeObject.Namespace).Create(sonarVolumeObject)

			if err != nil {
				return logErrorAndReturn(err)
			}

			log.Printf("PersistantVolumeClaim %s/%s has been created", sonarVolume.Namespace, sonarVolume.Name)
		} else if err != nil {
			return logErrorAndReturn(err)
		}
	}
	return nil
}

func (service K8SService) CreateServiceAccount(sonar v1alpha1.Sonar) (*coreV1Api.ServiceAccount, error) {

	labels := generateLabels(sonar.Name)

	sonarServiceAccountObject := &coreV1Api.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sonar.Name,
			Namespace: sonar.Namespace,
			Labels:    labels,
		},
	}

	if err := controllerutil.SetControllerReference(&sonar, sonarServiceAccountObject, service.scheme); err != nil {
		return nil, logErrorAndReturn(err)
	}

	sonarServiceAccount, err := service.coreClient.ServiceAccounts(sonarServiceAccountObject.Namespace).Get(sonarServiceAccountObject.Name, metav1.GetOptions{})

	if err != nil && k8serr.IsNotFound(err) {
		log.Printf("Creating a new ServiceAccount %s/%s for Sonar %s", sonarServiceAccountObject.Namespace, sonarServiceAccountObject.Name, sonar.Name)

		sonarServiceAccount, err = service.coreClient.ServiceAccounts(sonarServiceAccountObject.Namespace).Create(sonarServiceAccountObject)

		if err != nil {
			return nil, logErrorAndReturn(err)
		}

		log.Printf("ServiceAccount %s/%s has been created", sonarServiceAccount.Namespace, sonarServiceAccount.Name)
	} else if err != nil {
		return nil, logErrorAndReturn(err)
	}

	return sonarServiceAccount, nil
}

func (service K8SService) CreateExternalEndpoint(sonar v1alpha1.Sonar) error {
	fmt.Printf("No implementation for K8s yet.")
	return nil
}

func (service K8SService) CreateService(sonar v1alpha1.Sonar) error {
	portMap := map[string]int32{
		sonar.Name:         Port,
		sonar.Name + "-db": DBPort,
	}
	for _, serviceName := range []string{sonar.Name, sonar.Name + "-db"} {
		log.Printf("Start creating sonar service %v in namespace %v", serviceName, sonar.Namespace)
		labels := generateLabels(serviceName)

		sonarServiceObject, err := newSonarInternalBalancingService(serviceName, sonar.Namespace, labels, portMap[serviceName])

		if err := controllerutil.SetControllerReference(&sonar, sonarServiceObject, service.scheme); err != nil {
			return logErrorAndReturn(err)
		}

		sonarService, err := service.coreClient.Services(sonarServiceObject.Namespace).Get(sonarServiceObject.Name, metav1.GetOptions{})

		if err != nil && k8serr.IsNotFound(err) {
			log.Printf("Creating a new service %s/%s for sonar %s", sonarService.Namespace, sonarService.Name, sonar.Name)

			sonarService, err = service.coreClient.Services(sonarServiceObject.Namespace).Create(sonarServiceObject)

			if err != nil {
				return logErrorAndReturn(err)
			}

			log.Printf("service %s/%s has been created", sonarService.Namespace, sonarService.Name)
		} else if err != nil {
			return logErrorAndReturn(err)
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

func logErrorAndReturn(err error) error {
	log.Printf("[ERROR] %v", err)
	return err
}

func generateLabels(name string) map[string]string {
	return map[string]string{
		"app": name,
	}
}
