package helper

import (
	sonarSpec "github.com/epmd-edp/sonar-operator/v2/pkg/service/sonar/spec"
	coreV1Api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GenerateLabels(name string) map[string]string {
	return map[string]string{
		"app": name,
	}
}

func GenerateProbe(delay int32, probePath string) *coreV1Api.Probe {
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
				Path: probePath,
			},
		},
		TimeoutSeconds: 5,
	}
}

func GenerateDbProbe(delay int32) *coreV1Api.Probe {
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
