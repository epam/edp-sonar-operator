package service

import (
	"sonar-operator/pkg/apis/edp/v1alpha1"
)

type SonarService interface {
	// This is an entry point for service package. Invoked in err = r.service.Install(*instance) staticanalysistool_controller.go, Reconcile method.
	Install(instance v1alpha1.Sonar) error
}

func NewSonarService(platformService PlatformService) SonarService {
	return SonarServiceImpl{platformService: platformService}
}

type SonarServiceImpl struct {
	// Providing sonar service implementation through the interface (platform abstract)
	platformService PlatformService
}

// Invoking install method against SonarServiceImpl object should trigger list of methods, stored in client edp.PlatformService
func (s SonarServiceImpl) Install(instance v1alpha1.Sonar) error {
	_ = s.platformService.CreateSecret(instance)
	_ = s.platformService.CreateServiceAccount(instance)
	return nil
}
