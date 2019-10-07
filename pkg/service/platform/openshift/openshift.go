package openshift

import (
	"errors"
	"fmt"
	"github.com/epmd-edp/sonar-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/sonar-operator/v2/pkg/service/platform/helper"
	"github.com/epmd-edp/sonar-operator/v2/pkg/service/platform/kubernetes"
	sonarSpec "github.com/epmd-edp/sonar-operator/v2/pkg/service/sonar/spec"
	appsV1Api "github.com/openshift/api/apps/v1"
	routeV1Api "github.com/openshift/api/route/v1"
	securityV1Api "github.com/openshift/api/security/v1"
	appsV1client "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	projectV1Client "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	routeV1Client "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityV1Client "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	templateV1Client "github.com/openshift/client-go/template/clientset/versioned/typed/template/v1"
	coreV1Api "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
)

var log = logf.Log.WithName("platform")

type OpenshiftService struct {
	kubernetes.K8SService

	templateClient templateV1Client.TemplateV1Client
	projectClient  projectV1Client.ProjectV1Client
	securityClient securityV1Client.SecurityV1Client
	appClient      appsV1client.AppsV1Client
	routeClient    routeV1Client.RouteV1Client
}

func (service *OpenshiftService) Init(config *rest.Config, scheme *runtime.Scheme) error {

	err := service.K8SService.Init(config, scheme)
	if err != nil {
		return err
	}

	templateClient, err := templateV1Client.NewForConfig(config)
	if err != nil {
		return err
	}

	service.templateClient = *templateClient
	projectClient, err := projectV1Client.NewForConfig(config)
	if err != nil {
		return err
	}

	service.projectClient = *projectClient
	securityClient, err := securityV1Client.NewForConfig(config)
	if err != nil {
		return err
	}

	service.securityClient = *securityClient
	appClient, err := appsV1client.NewForConfig(config)
	if err != nil {
		return err
	}

	service.appClient = *appClient
	routeClient, err := routeV1Client.NewForConfig(config)
	if err != nil {
		return err
	}
	service.routeClient = *routeClient

	return nil
}

// GetRoute returns Route object and connection protocol from Openshift
func (service OpenshiftService) GetRoute(namespace string, name string) (*routeV1Api.Route, string, error) {
	route, err := service.routeClient.Routes(namespace).Get(name, metav1.GetOptions{})
	if err != nil && k8serrors.IsNotFound(err) {
		return nil, "", errors.New(fmt.Sprintf("Route %v in namespace %v not found", name, namespace))
	} else if err != nil {
		return nil, "", err
	}

	var routeScheme = "http"
	if route.Spec.TLS.Termination != "" {
		routeScheme = "https"
	}
	return route, routeScheme, nil
}
func (service OpenshiftService) CreateSecurityContext(sonar v1alpha1.Sonar, sa *coreV1Api.ServiceAccount) error {

	labels := helper.GenerateLabels(sonar.Name)
	priority := int32(1)

	project, err := service.projectClient.Projects().Get(sonar.Namespace, metav1.GetOptions{})
	if err != nil && k8serrors.IsNotFound(err) {
		return errors.New(fmt.Sprintf("Unable to retrieve project %s", sonar.Namespace))
	}

	displayName := project.GetObjectMeta().GetAnnotations()["openshift.io/display-name"]
	if displayName == "" {
		return errors.New(fmt.Sprintf("Project display name does not set"))
	}

	sonarSccObject := &securityV1Api.SecurityContextConstraints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", sonar.Name, displayName),
			Namespace: sonar.Namespace,
			Labels:    labels,
		},
		Volumes: []securityV1Api.FSType{
			securityV1Api.FSTypeSecret,
			securityV1Api.FSTypeDownwardAPI,
			securityV1Api.FSTypeEmptyDir,
			securityV1Api.FSTypePersistentVolumeClaim,
			securityV1Api.FSProjected,
			securityV1Api.FSTypeConfigMap,
		},
		AllowHostDirVolumePlugin: false,
		AllowHostIPC:             true,
		AllowHostNetwork:         false,
		AllowHostPID:             false,
		AllowHostPorts:           false,
		AllowPrivilegedContainer: false,
		AllowedCapabilities:      []coreV1Api.Capability{},
		AllowedFlexVolumes:       []securityV1Api.AllowedFlexVolume{},
		DefaultAddCapabilities:   []coreV1Api.Capability{},
		FSGroup: securityV1Api.FSGroupStrategyOptions{
			Type:   securityV1Api.FSGroupStrategyRunAsAny,
			Ranges: []securityV1Api.IDRange{},
		},
		Groups:                 []string{},
		Priority:               &priority,
		ReadOnlyRootFilesystem: false,
		RunAsUser: securityV1Api.RunAsUserStrategyOptions{
			Type:        securityV1Api.RunAsUserStrategyRunAsAny,
			UID:         nil,
			UIDRangeMin: nil,
			UIDRangeMax: nil,
		},
		SELinuxContext: securityV1Api.SELinuxContextStrategyOptions{
			Type:           securityV1Api.SELinuxStrategyMustRunAs,
			SELinuxOptions: nil,
		},
		SupplementalGroups: securityV1Api.SupplementalGroupsStrategyOptions{
			Type:   securityV1Api.SupplementalGroupsStrategyRunAsAny,
			Ranges: nil,
		},
		Users: []string{
			"system:serviceaccount:" + sonar.Namespace + ":" + sonar.Name,
		},
	}

	if err := controllerutil.SetControllerReference(&sonar, sonarSccObject, service.Scheme); err != nil {
		return err
	}

	sonarSCC, err := service.securityClient.SecurityContextConstraints().Get(sonarSccObject.Name, metav1.GetOptions{})
	if err != nil && k8serrors.IsNotFound(err) {
		msg := fmt.Sprintf("Creating a new Security Context Constraint %s for Sonar %s", sonarSccObject.Name, sonar.Name)
		log.V(1).Info(msg)

		sonarSCC, err = service.securityClient.SecurityContextConstraints().Create(sonarSccObject)

		if err != nil {
			return err
		}

		log.Info(fmt.Sprintf("Security Context Constraint %s has been created", sonarSCC.Name))
	} else if err != nil {
		return err

	} else {
		// TODO(Serhii Shydlovskyi): Reflect reports that present users and currently stored in object are different for some reason.
		if !reflect.DeepEqual(sonarSCC.Users, sonarSccObject.Users) {

			sonarSCC, err = service.securityClient.SecurityContextConstraints().Update(sonarSccObject)

			if err != nil {
				return err
			}

			log.Info(fmt.Sprintf("Security Context Constraint %s has been updated", sonarSCC.Name))
		}
	}

	return nil
}

func (service OpenshiftService) CreateExternalEndpoint(sonar v1alpha1.Sonar) error {

	labels := helper.GenerateLabels(sonar.Name)

	sonarRouteObject := &routeV1Api.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sonar.Name,
			Namespace: sonar.Namespace,
			Labels:    labels,
		},
		Spec: routeV1Api.RouteSpec{
			TLS: &routeV1Api.TLSConfig{
				Termination:                   routeV1Api.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routeV1Api.InsecureEdgeTerminationPolicyRedirect,
			},
			To: routeV1Api.RouteTargetReference{
				Name: sonar.Name,
				Kind: "Service",
			},
		},
	}

	if err := controllerutil.SetControllerReference(&sonar, sonarRouteObject, service.Scheme); err != nil {
		return err
	}

	sonarRoute, err := service.routeClient.Routes(sonarRouteObject.Namespace).Get(sonarRouteObject.Name, metav1.GetOptions{})

	if err != nil && k8serrors.IsNotFound(err) {
		log.V(1).Info(fmt.Sprintf("Creating a new Route for Sonar %s", sonar.Name))
		sonarRoute, err = service.routeClient.Routes(sonarRouteObject.Namespace).Create(sonarRouteObject)

		if err != nil {
			return err
		}

		log.Info("Route %s/%s has been created", sonarRoute.Namespace, sonarRoute.Name)
	} else if err != nil {
		return err
	}

	return nil
}

func (service OpenshiftService) CreateDbDeployConf(sonar v1alpha1.Sonar) error {
	labels := helper.GenerateLabels(sonar.Name)
	name := sonar.Name + "-db"

	sonarDbDcObject := newSonarDatabaseDeploymentConfig(name, sonar.Name, sonar.Namespace, labels)

	if err := controllerutil.SetControllerReference(&sonar, sonarDbDcObject, service.Scheme); err != nil {
		return err
	}

	sonarDbDc, err := service.appClient.DeploymentConfigs(sonarDbDcObject.Namespace).Get(sonarDbDcObject.Name, metav1.GetOptions{})

	if err != nil && k8serrors.IsNotFound(err) {
		log.V(1).Info(fmt.Sprintf("Creating a new DeploymentConfig for Sonar %s", sonar.Name))

		sonarDbDc, err = service.appClient.DeploymentConfigs(sonarDbDcObject.Namespace).Create(sonarDbDcObject)

		if err != nil {
			return err
		}

		log.Info("DeploymentConfig %s/%s has been created", sonarDbDc.Namespace, sonarDbDc.Name)
	} else if err != nil {
		return err
	}

	return nil
}

func (service OpenshiftService) CreateDeployConf(sonar v1alpha1.Sonar) error {
	labels := helper.GenerateLabels(sonar.Name)

	sonarDcObject := newSonarDeploymentConfig(sonar.Name, sonar.Namespace, sonar.Spec.Version, labels)
	if err := controllerutil.SetControllerReference(&sonar, sonarDcObject, service.Scheme); err != nil {
		return err
	}

	sonarDc, err := service.appClient.DeploymentConfigs(sonarDcObject.Namespace).Get(sonarDcObject.Name, metav1.GetOptions{})
	if err != nil && k8serrors.IsNotFound(err) {
		log.V(1).Info(fmt.Sprintf("Creating a new DeploymentConfig for Sonar %s", sonar.Name))

		sonarDc, err = service.appClient.DeploymentConfigs(sonarDcObject.Namespace).Create(sonarDcObject)
		if err != nil {
			return err
		}

		log.Info(fmt.Sprintf("DeploymentConfig %s/%s has been created", sonarDc.Namespace, sonarDc.Name))
	} else if err != nil {
		return err
	}

	return nil
}

func generateProbe(delay int32) *coreV1Api.Probe {
	return &coreV1Api.Probe{
		FailureThreshold:    5,
		InitialDelaySeconds: delay,
		PeriodSeconds:       20,
		SuccessThreshold:    1,
		Handler: coreV1Api.Handler{
			HTTPGet: &coreV1Api.HTTPGetAction{
				Port: intstr.IntOrString{
					IntVal: sonarSpec.Port,
				},
				Path: "/",
			},
		},
		TimeoutSeconds: 5,
	}
}

func generateDbProbe(delay int32) *coreV1Api.Probe {
	return &coreV1Api.Probe{
		FailureThreshold:    5,
		InitialDelaySeconds: delay,
		PeriodSeconds:       20,
		SuccessThreshold:    1,
		Handler: coreV1Api.Handler{
			Exec: &coreV1Api.ExecAction{
				Command: []string{"sh", "-c", "exec pg_isready --host $POD_IP"},
			},
		},
		TimeoutSeconds: 5,
	}
}

func newSonarDeploymentConfig(name string, namespace string, version string, labels map[string]string) *appsV1Api.DeploymentConfig {
	fsGroup, _ := strconv.ParseInt("999", 10, 64)
	return &appsV1Api.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsV1Api.DeploymentConfigSpec{
			Replicas: 1,
			Triggers: []appsV1Api.DeploymentTriggerPolicy{
				{
					Type: appsV1Api.DeploymentTriggerOnConfigChange,
				},
			},
			Strategy: appsV1Api.DeploymentStrategy{
				Type: appsV1Api.DeploymentStrategyTypeRolling,
			},
			Selector: labels,
			Template: &coreV1Api.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: coreV1Api.PodSpec{
					InitContainers: []coreV1Api.Container{
						{
							Name:    name + "init",
							Image:   "busybox",
							Command: []string{"sh", "-c", "while ! nc -w 1 " + name + "-db " + strconv.Itoa(sonarSpec.DBPort) + " </dev/null; do echo waiting for " + name + "-db; sleep 10; done;"},
						},
					},
					Containers: []coreV1Api.Container{
						{
							Name:            name,
							Image:           sonarSpec.Image + ":" + version,
							ImagePullPolicy: coreV1Api.PullIfNotPresent,
							Env: []coreV1Api.EnvVar{
								{
									Name: "SONARQUBE_JDBC_USERNAME",
									ValueFrom: &coreV1Api.EnvVarSource{
										SecretKeyRef: &coreV1Api.SecretKeySelector{
											LocalObjectReference: coreV1Api.LocalObjectReference{
												Name: name + "-db",
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
												Name: name + "-db",
											},
											Key: "database-password",
										},
									},
								},
								{
									Name:  "SONARQUBE_JDBC_URL",
									Value: "jdbc:postgresql://" + name + "-db/sonar",
								},
							},
							Ports: []coreV1Api.ContainerPort{
								{
									Name:          name,
									ContainerPort: sonarSpec.Port,
								},
							},
							LivenessProbe:          generateProbe(sonarSpec.LivenessProbeDelay),
							ReadinessProbe:         generateProbe(sonarSpec.ReadinessProbeDelay),
							TerminationMessagePath: "/dev/termination-log",
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
						FSGroup: &fsGroup,
					},
					ServiceAccountName: name,
					Volumes: []coreV1Api.Volume{
						{
							Name: "data",
							VolumeSource: coreV1Api.VolumeSource{
								PersistentVolumeClaim: &coreV1Api.PersistentVolumeClaimVolumeSource{
									ClaimName: name + "-data",
								},
							},
						},
					},
				},
			},
		},
	}
}

func newSonarDatabaseDeploymentConfig(name string, sa string, namespace string, labels map[string]string) *appsV1Api.DeploymentConfig {
	return &appsV1Api.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsV1Api.DeploymentConfigSpec{
			Replicas: 1,
			Triggers: []appsV1Api.DeploymentTriggerPolicy{
				{
					Type: appsV1Api.DeploymentTriggerOnConfigChange,
				},
			},
			Strategy: appsV1Api.DeploymentStrategy{
				Type: appsV1Api.DeploymentStrategyTypeRolling,
			},
			Selector: helper.GenerateLabels(name),
			Template: &coreV1Api.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: helper.GenerateLabels(name),
				},
				Spec: coreV1Api.PodSpec{
					Containers: []coreV1Api.Container{
						{
							Name:            name,
							Image:           sonarSpec.DbImage,
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
							LivenessProbe:          generateDbProbe(sonarSpec.DbLivenessProbeDelay),
							ReadinessProbe:         generateDbProbe(sonarSpec.DbReadinessProbeDelay),
							TerminationMessagePath: "/dev/termination-log",
							Resources: coreV1Api.ResourceRequirements{
								Requests: map[coreV1Api.ResourceName]resource.Quantity{
									coreV1Api.ResourceMemory: resource.MustParse(sonarSpec.MemoryRequest),
								},
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

func (service OpenshiftService) GetAvailiableDeploymentReplicas(instance v1alpha1.Sonar) (*int, error) {
	c, err := service.appClient.DeploymentConfigs(instance.Namespace).Get(instance.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	r := int(c.Status.AvailableReplicas)

	return &r, nil
}