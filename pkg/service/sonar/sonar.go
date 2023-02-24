package sonar

import (
	"context"
	"fmt"
	"strings"

	"github.com/dchest/uniuri"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/epam/edp-keycloak-operator/controllers/helper"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
	sonarClient "github.com/epam/edp-sonar-operator/pkg/client/sonar"
	pkgHelper "github.com/epam/edp-sonar-operator/pkg/helper"
	"github.com/epam/edp-sonar-operator/pkg/service/platform"
)

const (
	defaultConfigFilesAbsolutePath = "/usr/local/"
	localConfigsRelativePath       = "configs"
	defaultProfileAbsolutePath     = defaultConfigFilesAbsolutePath + localConfigsRelativePath + "/" + defaultQualityProfilesFileName
	defaultQualityProfilesFileName = "quality-profile.xml"
	main                           = "main"
	admin                          = "admin"
	failMsgTemplate                = "failed to set owner reference for secret"
	failInitSonarMsg               = "failed to initialize Sonar Client"
)

var log = ctrl.Log.WithName("sonar_service")

type ServiceInterface interface {
	Configure(ctx context.Context, instance *sonarApi.Sonar) error
	ExposeConfiguration(ctx context.Context, instance *sonarApi.Sonar) error
	Integration(ctx context.Context, instance *sonarApi.Sonar) (*sonarApi.Sonar, error)
	IsDeploymentReady(instance *sonarApi.Sonar) (bool, error)
	ClientForChild(ctx context.Context, instance ChildInstance) (sonarClient.ClientInterface, error)
	DeleteResource(ctx context.Context, instance Deletable, finalizer string,
		deleteFunc func() error) (bool, error)
	K8sClient() client.Client
}

type ChildInstance interface {
	SonarOwner() string
	GetNamespace() string
}

type Deletable interface {
	v1.Object
	runtime.Object
}

func NewService(platformService platform.Service, k8sClient client.Client) *Service {
	svc := Service{
		platformService:      platformService,
		k8sClient:            k8sClient,
		runningInClusterFunc: pkgHelper.RunningInCluster,
	}

	svc.sonarClientBuilder = svc.initSonarClient

	return &svc
}

type Service struct {
	// Providing sonar service implementation through the interface (platform abstract)
	platformService      platform.Service
	k8sClient            client.Client
	sonarClientBuilder   func(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error)
	runningInClusterFunc func() bool
}

func (s *Service) K8sClient() client.Client {
	return s.k8sClient
}

func (s *Service) ClientForChild(ctx context.Context, instance ChildInstance) (sonarClient.ClientInterface, error) {
	var rootSonar sonarApi.Sonar
	if err := s.k8sClient.Get(
		ctx,
		types.NamespacedName{
			Namespace: instance.GetNamespace(),
			Name:      instance.SonarOwner(),
		},
		&rootSonar,
	); err != nil {
		return nil, fmt.Errorf("failed to get root sonar instance: %w", err)
	}

	sClient, err := s.sonarClientBuilder(ctx, &rootSonar)
	if err != nil {
		return nil, fmt.Errorf("failed to init sonar rest client: %w", err)
	}

	return sClient, nil
}

func (s *Service) DeleteResource(ctx context.Context, instance Deletable, finalizer string,
	deleteFunc func() error,
) (bool, error) {
	finalizers := instance.GetFinalizers()

	if instance.GetDeletionTimestamp().IsZero() {
		if !helper.ContainsString(finalizers, finalizer) {
			finalizers = append(finalizers, finalizer)
			instance.SetFinalizers(finalizers)

			if err := s.k8sClient.Update(ctx, instance); err != nil {
				return false, fmt.Errorf("failed to update deletable object: %w", err)
			}
		}

		return false, nil
	}

	if err := deleteFunc(); err != nil {
		return false, fmt.Errorf("failed to delete resource: %w", err)
	}

	if helper.ContainsString(finalizers, finalizer) {
		finalizers = helper.RemoveString(finalizers, finalizer)
		instance.SetFinalizers(finalizers)

		if err := s.k8sClient.Update(ctx, instance); err != nil {
			return false, fmt.Errorf("failed to update realm role cr: %w", err)
		}
	}

	return true, nil
}

func (s *Service) initSonarClient(ctx context.Context, instance *sonarApi.Sonar) (sonarClient.ClientInterface, error) {
	data, err := s.platformService.GetSecretData(instance.Namespace, instance.Spec.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to get sonar secret: %w", err)
	}

	rawToken, ok := data["token"]
	if !ok {
		log.Info("No token found in secret", "secret", instance.Spec.Secret)

		return nil, fmt.Errorf("failed to find token in secret %v", instance.Spec.Secret)
	}

	token := strings.TrimSuffix(string(rawToken), "\n")

	url := instance.Spec.Url

	sonarCl, err := sonarClient.NewClientFromToken(ctx, url+"/api", token)
	if err != nil {
		return nil, fmt.Errorf("failed to init Sonar client: %w", err)
	}

	return sonarCl, nil
}

func (s *Service) Integration(ctx context.Context, instance *sonarApi.Sonar) (*sonarApi.Sonar, error) {
	valueType := "value"

	sc, err := s.sonarClientBuilder(ctx, instance)
	if err != nil {
		return instance, errors.Wrap(err, failInitSonarMsg)
	}

	url := instance.Spec.Url

	if err = sc.ConfigureGeneralSettings(sonarClient.SettingRequest{
		ValueType: valueType,
		Key:       "sonar.core.serverBaseURL",
		Value:     url,
	}); err != nil {
		return instance, fmt.Errorf("failed to configure sonar.core.serverBaseURL: %w", err)
	}

	if err = sc.ConfigureGeneralSettings(sonarClient.SettingRequest{
		ValueType: valueType,
		Key:       "sonar.auth.oidc.clientId.secured",
		Value:     instance.Name,
	}); err != nil {
		return instance, fmt.Errorf("failed to configure sonar.auth.oidc.clientId.secured: %w", err)
	}

	dv := "private"

	log.Info(fmt.Sprintf("trying to set %v visibility for projects as default", dv))

	if err = sc.SetProjectsDefaultVisibility(dv); err != nil {
		return nil, fmt.Errorf("failed to set default %v visibility for projects: %w", dv, err)
	}

	return instance, nil
}

func (s *Service) ExposeConfiguration(ctx context.Context, instance *sonarApi.Sonar) error {
	sc, err := s.sonarClientBuilder(ctx, instance)
	if err != nil {
		return fmt.Errorf("%s: %w", failInitSonarMsg, err)
	}

	for _, user := range instance.Spec.Users {
		if err = s.createSonarUser(ctx, user, instance, sc); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) createSonarUser(ctx context.Context, user sonarApi.User, sonar *sonarApi.Sonar, sc sonarClient.ClientInterface) error {
	if _, err := sc.GetUser(ctx, user.Login); err != nil {
		if !sonarClient.IsErrNotFound(err) {
			return fmt.Errorf("failed to get user %s: %w", user.Username, err)
		}

		sonarUser := sonarClient.User{
			Login:    user.Login,
			Name:     user.Username,
			Password: uniuri.New(),
		}

		if err = sc.CreateUser(ctx, &sonarUser); err != nil {
			return fmt.Errorf("failed to create user %v in Sonar: %w", user.Username, err)
		}
	}

	tokenSecretName := fmt.Sprintf("%s-%s-token", sonar.Name, user.Login)

	if _, err := sc.GetUserToken(
		ctx,
		user.Login,
		cases.Title(language.English).
			String(user.Login),
	); err != nil {
		if !sonarClient.IsErrNotFound(err) {
			return fmt.Errorf("failed to get user token for user %s: %w", user.Login, err)
		}

		token, errGen := sc.GenerateUserToken(user.Login)
		if errGen != nil {
			return fmt.Errorf("failed to generate token for %v user: %w", user.Login, err)
		}

		if token != nil {
			ciSecret := map[string][]byte{
				"username": []byte(user.Login),
				"secret":   []byte(*token),
			}

			secret, errCreateSecret := s.platformService.CreateSecret(sonar.Name, sonar.Namespace, tokenSecretName, ciSecret)
			if errCreateSecret != nil {
				return fmt.Errorf("failed to create secret for  %v user: %w", tokenSecretName, err)
			}

			if errCreateSecret = s.platformService.SetOwnerReference(sonar, secret); errCreateSecret != nil {
				return fmt.Errorf("%s %v: %w", failMsgTemplate, secret, err)
			}
		}
	}

	if user.Group != "" {
		if err := sc.AddUserToGroup(user.Group, user.Login); err != nil {
			return fmt.Errorf("failed to add %v user in %v group: %w", user.Login, user.Group, err)
		}
	}

	for _, permission := range user.Permissions {
		if err := sc.AddPermissionToUser(user.Login, permission); err != nil {
			return fmt.Errorf("failed to add permission %s to  %v user: %w", permission, user.Login, err)
		}
	}

	return nil
}

func (s *Service) Configure(ctx context.Context, instance *sonarApi.Sonar) error {
	if s.runningInClusterFunc == nil {
		return errors.New("missing runningInClusterFunc")
	}

	sc, err := s.sonarClientBuilder(ctx, instance)
	if err != nil {
		return fmt.Errorf("%s: %w", failInitSonarMsg, err)
	}

	if err = installPlugins(instance.Spec.Plugins, sc); err != nil {
		return fmt.Errorf("failed to install plugins: %w", err)
	}

	if err = createQualityGate(instance.Spec.QualityGates, sc); err != nil {
		return fmt.Errorf("failed to configure EDP way quality gate: %w", err)
	}

	if err = setupGroups(ctx, &instance.Spec, sc); err != nil {
		return fmt.Errorf("failed to setup groups: %w", err)
	}

	if err = configureGeneralSettings(instance.Spec.Settings, sc); err != nil {
		return fmt.Errorf("failed to configure general settings: %w", err)
	}

	if err = setDefaultPermissionTemplate(ctx, instance.Spec.DefaultPermissionTemplate, sc); err != nil {
		return fmt.Errorf("failed to set default permission template: %w", err)
	}

	return nil
}

func setDefaultPermissionTemplate(ctx context.Context, templateName string, sc sonarClient.ClientInterface) error {
	if templateName != "" {
		if err := sc.SetDefaultPermissionTemplate(ctx, templateName); err != nil {
			return errors.Wrap(err, "unable to set default permission template")
		}
	}

	return nil
}

func setupGroups(ctx context.Context, spec *sonarApi.SonarSpec, sc sonarClient.ClientInterface) error {
	for _, group := range spec.Groups {
		if err := setupGroup(ctx, group, sc); err != nil {
			return fmt.Errorf("failed to setup group %s: %w", group.Name, err)
		}
	}

	return nil
}

func setupGroup(ctx context.Context, group sonarApi.Group, sc sonarClient.ClientInterface) error {
	if _, err := sc.GetGroup(ctx, group.Name); err != nil {
		if !sonar.IsErrNotFound(err) {
			return fmt.Errorf("unexpected error during group check: %w", err)
		}

		if err = sc.CreateGroup(ctx, &sonar.Group{Name: group.Name}); err != nil {
			return fmt.Errorf("failed to create group %s: %w", group, err)
		}
	}

	for _, permission := range group.Permissions {
		if err := sc.AddPermissionsToGroup(group.Name, permission); err != nil {
			return fmt.Errorf("failed to add scan permission for group %s: %w", group.Name, err)
		}
	}

	return nil
}

func parseQualityGates(qualityGates []sonarApi.QualityGate) sonarClient.QualityGates {
	parsedGates := make(map[string]sonarClient.QualityGateSettings)

	for _, gate := range qualityGates {
		parsedConditions := parseQualityGateConditions(gate.Conditions)

		if existingSettings, ok := parsedGates[gate.Name]; ok {
			parsedGates[gate.Name] = sonarClient.QualityGateSettings{
				MakeDefault: gate.SetAsDefault || existingSettings.MakeDefault,
				Conditions:  append(existingSettings.Conditions, parsedConditions...),
			}

			continue
		}

		parsedGates[gate.Name] = sonarClient.QualityGateSettings{
			MakeDefault: gate.SetAsDefault,
			Conditions:  parsedConditions,
		}
	}

	return parsedGates
}

func parseQualityGateConditions(conditions []sonarApi.QualityGateCondition) []sonarClient.QualityGateCondition {
	parsedConditions := make([]sonarClient.QualityGateCondition, 0, len(conditions))

	for _, condition := range conditions {
		parsedConditions = append(parsedConditions, sonarClient.QualityGateCondition{
			Error:  condition.Error,
			Metric: condition.Metric,
			OP:     condition.OP,
			Period: condition.Period,
		})
	}

	return parsedConditions
}

func createQualityGate(qualityGates []sonarApi.QualityGate, sc sonarClient.ClientInterface) error {
	if qualityGates == nil {
		return nil
	}

	parsedGates := parseQualityGates(qualityGates)

	if err := sc.CreateQualityGates(parsedGates); err != nil {
		return fmt.Errorf("failed to configure given quality gates: %w", err)
	}

	return nil
}

func installPlugins(plugins []string, sc sonarClient.ClientInterface) error {
	if plugins == nil {
		return nil
	}

	if err := sc.InstallPlugins(plugins); err != nil {
		return fmt.Errorf("failed to install specified plugins: %w", err)
	}

	return nil
}

func configureGeneralSettings(settings []sonarApi.SonarSetting, sc sonarClient.ClientInterface) error {
	if settings == nil {
		return nil
	}

	clientSettings := make([]sonarClient.SettingRequest, 0, len(settings))

	for _, setting := range settings {
		clientSettings = append(clientSettings, sonarClient.SettingRequest{
			Key:       setting.Key,
			Value:     setting.Value,
			ValueType: setting.ValueType,
		})
	}

	if err := sc.ConfigureGeneralSettings(clientSettings...); err != nil {
		return err
	}

	return nil
}

func (s *Service) IsDeploymentReady(instance *sonarApi.Sonar) (bool, error) {
	r, err := s.platformService.GetAvailableDeploymentReplicas(instance)
	if err != nil {
		return false, err
	}

	if *r == 1 {
		return true, nil
	}

	return false, nil
}
