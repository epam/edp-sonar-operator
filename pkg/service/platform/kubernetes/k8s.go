package kubernetes

import (
	"fmt"
	edpCompApi "github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	edpCompClient "github.com/epmd-edp/edp-component-operator/pkg/client"
	jenkinsV1Api "github.com/epmd-edp/jenkins-operator/v2/pkg/apis/v2/v1alpha1"
	jenkinsScriptV1Client "github.com/epmd-edp/jenkins-operator/v2/pkg/controller/jenkinsscript/client"
	jenkinsSAV1Client "github.com/epmd-edp/jenkins-operator/v2/pkg/controller/jenkinsserviceaccount/client"
	"github.com/epmd-edp/sonar-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/sonar-operator/v2/pkg/service/platform/helper"
	platformHelper "github.com/epmd-edp/sonar-operator/v2/pkg/service/platform/helper"
	sonarSpec "github.com/epmd-edp/sonar-operator/v2/pkg/service/sonar/spec"
	"github.com/pkg/errors"
	appsV1Api "k8s.io/api/apps/v1"
	coreV1Api "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsV1Client "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreV1Client "k8s.io/client-go/kubernetes/typed/core/v1"
	extensionsV1Client "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
	"strings"
)

var log = logf.Log.WithName("platform")

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

func (service K8SService) GetConfigmap(namespace string, name string) (map[string]string, error) {
	configmap, err := service.coreClient.ConfigMaps(namespace).Get(name, metav1.GetOptions{})

	if err != nil && k8serr.IsNotFound(err) {
		log.Info("Config map in namespace not found", "configmap name", name, "namespace", namespace)
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return configmap.Data, nil
}

func (service K8SService) GetSecretData(namespace string, name string) (map[string][]byte, error) {
	sonarSecret, err := service.coreClient.Secrets(namespace).Get(name, metav1.GetOptions{})
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

	sonarSecret, err := service.coreClient.Secrets(sonarSecretObject.Namespace).Get(sonarSecretObject.Name, metav1.GetOptions{})

	if err != nil && k8serr.IsNotFound(err) {
		log.V(1).Info("Creating a new Secret for Sonar", "namespace", sonarSecretObject.Namespace, "secret name", sonarSecretObject.Name, "sonar name", sonar.Name)

		sonarSecret, err = service.coreClient.Secrets(sonarSecretObject.Namespace).Create(sonarSecretObject)

		if err != nil {
			return err
		}
		log.Info("Secret has been created", "namespace", sonarSecret.Namespace, "secret name", sonarSecret.Name)

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
			log.V(1).Info("Creating a new PersistentVolumeClaim", "namespace", sonarVolumeObject.Namespace, "volume name", sonarVolumeObject.Name, "sonar name", sonar.Name)

			sonarVolume, err = service.coreClient.PersistentVolumeClaims(sonarVolumeObject.Namespace).Create(sonarVolumeObject)

			if err != nil {
				return err
			}

			log.Info("PersistentVolumeClaim has been created", "namespace", sonarVolume.Namespace, "sonar name", sonarVolume.Name)
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (service K8SService) CreateDbDeployment(sonar v1alpha1.Sonar) error {
	l := helper.GenerateLabels(sonar.Name)
	n := sonar.Name + "-db"

	o := newDatabaseDeployment(n, sonar.Name, sonar.Namespace, l, sonar.Spec.DBImage)

	if err := controllerutil.SetControllerReference(&sonar, o, service.Scheme); err != nil {
		return err
	}

	_, err := service.AppsClient.Deployments(o.Namespace).Get(o.Name, metav1.GetOptions{})

	if err == nil || !k8serr.IsNotFound(err) {
		return err
	}
	log.V(1).Info("Creating a new Deployment for Sonar", "sonar name", sonar.Name)

	_, err = service.AppsClient.Deployments(o.Namespace).Create(o)
	return err
}

func (service K8SService) CreateSecurityContext(sonar v1alpha1.Sonar) error {
	return nil
}

func (service K8SService) CreateDeployment(sonar v1alpha1.Sonar) error {
	l := helper.GenerateLabels(sonar.Name)

	o := newSonarDeployment(sonar, l)
	if err := controllerutil.SetControllerReference(&sonar, o, service.Scheme); err != nil {
		return err
	}

	_, err := service.AppsClient.Deployments(o.Namespace).Get(o.Name, metav1.GetOptions{})
	if err == nil || !k8serr.IsNotFound(err) {
		return err
	}

	log.V(1).Info("Creating a new Deployment for Sonar", "sonar name", sonar.Name)
	_, err = service.AppsClient.Deployments(o.Namespace).Create(o)

	return err
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
		log.V(1).Info("Creating a new ServiceAccount for Sonar", "namespace", sonarServiceAccountObject.Namespace, "service account name", sonarServiceAccountObject.Name, "sonar name", sonar.Name)

		sonarServiceAccount, err = service.coreClient.ServiceAccounts(sonarServiceAccountObject.Namespace).Create(sonarServiceAccountObject)

		if err != nil {
			return nil, err
		}

		log.Info("ServiceAccount has been created", "namespace", sonarServiceAccount.Namespace, "service account name", sonarServiceAccount.Name)
	} else if err != nil {
		return nil, err
	}

	return sonarServiceAccount, nil
}

func (service K8SService) CreateExternalEndpoint(sonar v1alpha1.Sonar) error {
	log.V(1).Info("Creating Sonar external endpoint.",
		"Namespace", sonar.Namespace, "Name", sonar.Name)

	l := helper.GenerateLabels(sonar.Name)

	s, err := service.coreClient.Services(sonar.Namespace).Get(sonar.Name, metav1.GetOptions{})
	if err != nil {
		log.Info("Sonar Service has not been found")
		return err
	}

	hostname := fmt.Sprintf("%v-%v.%v", sonar.Name, sonar.Namespace, sonar.Spec.EdpSpec.DnsWildcard)
	path := "/"
	if len(sonar.Spec.BasePath) != 0 {
		hostname = sonar.Spec.EdpSpec.DnsWildcard
		path = fmt.Sprintf("/%v", sonar.Spec.BasePath)
	}

	o := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sonar.Name,
			Namespace: sonar.Namespace,
			Labels:    l,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: hostname,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: path,
									Backend: v1beta1.IngressBackend{
										ServiceName: sonar.Name,
										ServicePort: intstr.IntOrString{
											IntVal: s.Spec.Ports[0].TargetPort.IntVal,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(&sonar, o, service.Scheme); err != nil {
		return err
	}

	_, err = service.ExtensionsV1Client.Ingresses(o.Namespace).Get(o.Name, metav1.GetOptions{})
	if err == nil || !k8serr.IsNotFound(err) {
		return err
	}

	log.V(1).Info("Creating a new Ingress for Sonar", "sonar name", sonar.Name)
	_, err = service.ExtensionsV1Client.Ingresses(o.Namespace).Create(o)

	return err
}

func (service K8SService) GetExternalEndpoint(namespace string, name string) (*string, error) {
	r, err := service.ExtensionsV1Client.Ingresses(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	rs := "https"
	u := fmt.Sprintf("%v://%v%v", rs, r.Spec.Rules[0].Host,
		strings.TrimRight(r.Spec.Rules[0].HTTP.Paths[0].Path, "/"))

	return &u, nil
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
			log.V(1).Info("Creating a new service for sonar", "namespace", sonarServiceObject.Namespace, "service name", sonarServiceObject.Name, "sonar name", sonar.Name)

			sonarService, err = service.coreClient.Services(sonarServiceObject.Namespace).Create(sonarServiceObject)

			if err != nil {
				return err
			}

			log.Info("Service has been created", "namespace", sonarService.Namespace, "sonar name", sonarService.Name)
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

func newSonarDeployment(sonar v1alpha1.Sonar, labels map[string]string) *appsV1Api.Deployment {
	g, _ := strconv.ParseInt("999", 10, 64)
	var rc int32 = 1
	t := true

	sonarWebContextEnv := "/"
	if len(sonar.Spec.BasePath) != 0 {
		sonarWebContextEnv = fmt.Sprintf("%v%v", sonarWebContextEnv, sonar.Spec.BasePath)
	}

	return &appsV1Api.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sonar.Name,
			Namespace: sonar.Namespace,
			Labels:    labels,
		},
		Spec: appsV1Api.DeploymentSpec{
			Replicas: &rc,
			Strategy: appsV1Api.DeploymentStrategy{
				Type: appsV1Api.RollingUpdateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: coreV1Api.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: coreV1Api.PodSpec{
					ImagePullSecrets: sonar.Spec.ImagePullSecrets,
					InitContainers: []coreV1Api.Container{
						{
							Name:    sonar.Name + "init",
							Image:   sonar.Spec.InitImage,
							Command: []string{"sh", "-c", "while ! nc -w 1 " + sonar.Name + "-db " + strconv.Itoa(sonarSpec.DBPort) + " </dev/null; do echo waiting for " + sonar.Name + "-db; sleep 10; done;"},
						},
					},
					Containers: []coreV1Api.Container{
						{
							Name:            sonar.Name,
							Image:           fmt.Sprintf("%s:%s", sonar.Spec.Image, sonar.Spec.Version),
							ImagePullPolicy: coreV1Api.PullIfNotPresent,
							Args:            []string{"-Dsonar.search.javaAdditionalOpts=-Dnode.store.allow_mmapfs=false"},
							Env: []coreV1Api.EnvVar{
								{
									Name: "SONARQUBE_JDBC_USERNAME",
									ValueFrom: &coreV1Api.EnvVarSource{
										SecretKeyRef: &coreV1Api.SecretKeySelector{
											LocalObjectReference: coreV1Api.LocalObjectReference{
												Name: sonar.Name + "-db",
											},
											Key: "database-user",
										},
									},
								},
								{
									Name: "SONARQUBE_JDBC_PASSWORD",
									ValueFrom: &coreV1Api.EnvVarSource{
										SecretKeyRef: &coreV1Api.SecretKeySelector{
											LocalObjectReference: coreV1Api.LocalObjectReference{
												Name: sonar.Name + "-db",
											},
											Key: "database-password",
										},
									},
								},
								{
									Name:  "SONARQUBE_JDBC_URL",
									Value: "jdbc:postgresql://" + sonar.Name + "-db/sonar",
								},
								{
									Name:  "sonar.web.context",
									Value: sonarWebContextEnv,
								},
							},
							Ports: []coreV1Api.ContainerPort{
								{
									Name:          sonar.Name,
									ContainerPort: sonarSpec.Port,
								},
							},
							LivenessProbe:          helper.GenerateProbe(sonarSpec.LivenessProbeDelay, sonarWebContextEnv),
							ReadinessProbe:         helper.GenerateProbe(sonarSpec.ReadinessProbeDelay, sonarWebContextEnv),
							TerminationMessagePath: "/dev/termination-log",
							SecurityContext: &coreV1Api.SecurityContext{
								AllowPrivilegeEscalation: &t,
							},
							Resources: coreV1Api.ResourceRequirements{
								Requests: map[coreV1Api.ResourceName]resource.Quantity{
									coreV1Api.ResourceMemory: resource.MustParse(sonarSpec.MemoryRequest),
								},
							},
							VolumeMounts: []coreV1Api.VolumeMount{
								{
									MountPath: "/opt/sonarqube/extensions/plugins",
									Name:      "data",
								},
							},
						},
					},
					SecurityContext: &coreV1Api.PodSecurityContext{
						FSGroup:      &g,
						RunAsUser:    &g,
						RunAsNonRoot: &t,
					},
					ServiceAccountName: sonar.Name,
					Volumes: []coreV1Api.Volume{
						{
							Name: "data",
							VolumeSource: coreV1Api.VolumeSource{
								PersistentVolumeClaim: &coreV1Api.PersistentVolumeClaimVolumeSource{
									ClaimName: sonar.Name + "-data",
								},
							},
						},
					},
				},
			},
		},
	}
}

func newDatabaseDeployment(name string, sa string, namespace string, labels map[string]string, dbImage string) *appsV1Api.Deployment {
	var rc int32 = 1
	var uid int64 = 999
	f := false
	t := true

	return &appsV1Api.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsV1Api.DeploymentSpec{
			Replicas: &rc,
			Strategy: appsV1Api.DeploymentStrategy{
				Type: appsV1Api.RollingUpdateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: helper.GenerateLabels(name),
			},
			Template: coreV1Api.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: helper.GenerateLabels(name),
				},
				Spec: coreV1Api.PodSpec{
					SecurityContext: &coreV1Api.PodSecurityContext{
						RunAsUser:    &uid,
						FSGroup:      &uid,
						RunAsNonRoot: &t,
					},
					Containers: []coreV1Api.Container{
						{
							Name:            name,
							Image:           dbImage,
							ImagePullPolicy: coreV1Api.PullIfNotPresent,
							Env: []coreV1Api.EnvVar{
								{
									Name:  "PGDATA",
									Value: "/var/lib/postgresql/data/pgdata",
								},
								{
									Name:  "POSTGRES_DB",
									Value: "sonar",
								},
								{
									Name: "POD_IP",
									ValueFrom: &coreV1Api.EnvVarSource{
										FieldRef: &coreV1Api.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									},
								},
								{
									Name: "POSTGRES_USER",
									ValueFrom: &coreV1Api.EnvVarSource{
										SecretKeyRef: &coreV1Api.SecretKeySelector{
											LocalObjectReference: coreV1Api.LocalObjectReference{
												Name: name,
											},
											Key: "database-user",
										},
									},
								},
								{
									Name: "POSTGRES_PASSWORD",
									ValueFrom: &coreV1Api.EnvVarSource{
										SecretKeyRef: &coreV1Api.SecretKeySelector{
											LocalObjectReference: coreV1Api.LocalObjectReference{
												Name: name,
											},
											Key: "database-password",
										},
									},
								},
							},
							Ports: []coreV1Api.ContainerPort{
								{
									ContainerPort: sonarSpec.DBPort,
								},
							},
							LivenessProbe:          helper.GenerateDbProbe(sonarSpec.DbLivenessProbeDelay),
							ReadinessProbe:         helper.GenerateDbProbe(sonarSpec.DbReadinessProbeDelay),
							TerminationMessagePath: "/dev/termination-log",
							Resources: coreV1Api.ResourceRequirements{
								Requests: map[coreV1Api.ResourceName]resource.Quantity{
									coreV1Api.ResourceMemory: resource.MustParse(sonarSpec.MemoryRequest),
								},
							},
							SecurityContext: &coreV1Api.SecurityContext{
								AllowPrivilegeEscalation: &f,
							},
							VolumeMounts: []coreV1Api.VolumeMount{
								{
									MountPath: "/var/lib/postgresql/data",
									Name:      "data",
								},
							},
						},
					},
					ServiceAccountName: sa,
					Volumes: []coreV1Api.Volume{
						{
							Name: "data",
							VolumeSource: coreV1Api.VolumeSource{
								PersistentVolumeClaim: &coreV1Api.PersistentVolumeClaimVolumeSource{
									ClaimName: name,
								},
							},
						},
					},
				},
			},
		},
	}
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

func (service K8SService) GetAvailiableDeploymentReplicas(instance v1alpha1.Sonar) (*int, error) {
	c, err := service.AppsClient.Deployments(instance.Namespace).Get(instance.Name, metav1.GetOptions{})
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
			Type: "sonar",
			Url:  url,
			Icon: icon,
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
