package helper

import (
	coreV1Api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	sonarSpec "github.com/epam/edp-sonar-operator/v2/pkg/service/sonar/spec"
)

const (
	UrlCutset        = "!\"#$%&'()*+,-./@:;<=>[\\]^_`{|}~"
	failureThreshold = 5
	periodSeconds    = 20
	timeOutSeconds   = 5
)

func GenerateLabels(name string) map[string]string {
	return map[string]string{
		"app": name,
	}
}

func GenerateProbe(delay int32, probePath string) *coreV1Api.Probe {
	return &coreV1Api.Probe{
		FailureThreshold:    failureThreshold,
		InitialDelaySeconds: delay,
		PeriodSeconds:       periodSeconds,
		SuccessThreshold:    1,
		Handler: coreV1Api.Handler{
			HTTPGet: &coreV1Api.HTTPGetAction{
				Port: intstr.IntOrString{
					IntVal: sonarSpec.Port,
				},
				Path: probePath,
			},
		},
		TimeoutSeconds: timeOutSeconds,
	}
}

func GenerateDbProbe(delay int32) *coreV1Api.Probe {
	return &coreV1Api.Probe{
		FailureThreshold:    failureThreshold,
		InitialDelaySeconds: delay,
		PeriodSeconds:       periodSeconds,
		SuccessThreshold:    1,
		Handler: coreV1Api.Handler{
			Exec: &coreV1Api.ExecAction{
				Command: []string{"sh", "-c", "exec pg_isready --host $POD_IP"},
			},
		},
		TimeoutSeconds: timeOutSeconds,
	}
}
