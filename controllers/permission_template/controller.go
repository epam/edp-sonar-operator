package permission_template

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	sonarClient "github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/helper"
	"github.com/epam/edp-sonar-operator/pkg/service/platform"
	"github.com/epam/edp-sonar-operator/pkg/service/sonar"
)

const finalizer = "sonar.permission_template.operator"

type Reconcile struct {
	service sonar.ServiceInterface
	client  client.Client
	log     logr.Logger
}

func NewReconcile(client client.Client, scheme *runtime.Scheme, log logr.Logger, platformType string) (*Reconcile, error) {
	ps, err := platform.NewService(platformType, scheme, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create platform service: %w", err)
	}

	return &Reconcile{
		service: sonar.NewService(ps, client),
		client:  client,
		log:     log.WithName("permission-template"),
	}, nil
}

func (r *Reconcile) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sonarApi.SonarPermissionTemplate{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: isSpecUpdated,
		})).
		Complete(r)
}

func isSpecUpdated(e event.UpdateEvent) bool {
	oo := e.ObjectOld.(*sonarApi.SonarPermissionTemplate)
	no := e.ObjectNew.(*sonarApi.SonarPermissionTemplate)

	return !reflect.DeepEqual(oo.Spec, no.Spec) ||
		(oo.GetDeletionTimestamp().IsZero() && !no.GetDeletionTimestamp().IsZero())
}

//+kubebuilder:rbac:groups=v1.edp.epam.com,namespace=placeholder,resources=sonarpermissiontemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=v1.edp.epam.com,namespace=placeholder,resources=sonarpermissiontemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=v1.edp.epam.com,namespace=placeholder,resources=sonarpermissiontemplates/finalizers,verbs=update

// Reconcile is a loop for reconciling
func (r *Reconcile) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := r.log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	log.Info("Reconciling SonarPermissionTemplate")

	var instance sonarApi.SonarPermissionTemplate
	if err := r.client.Get(ctx, request.NamespacedName, &instance); err != nil {
		if k8sErrors.IsNotFound(err) {
			log.Info("instance not found")

			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, fmt.Errorf("failed to get sonar permission tpl from k8s: %w", err)
	}

	var result reconcile.Result

	if err := r.tryReconcile(ctx, &instance); err != nil {
		instance.Status.Value = err.Error()
		result.RequeueAfter = helper.SetFailureCount(&instance)
		log.Error(err, "an error has occurred while handling sonar permission template", "name",
			request.Name)
	} else {
		helper.SetSuccessStatus(&instance)
	}

	if err := r.client.Status().Update(ctx, &instance); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to update status: %w", err)
	}

	log.Info("Reconciling done")

	return result, nil
}

func (r *Reconcile) tryReconcile(ctx context.Context, instance *sonarApi.SonarPermissionTemplate) error {
	sClient, err := r.service.ClientForChild(ctx, instance)
	if err != nil {
		return fmt.Errorf("failed to init sonar rest client: %w", err)
	}

	_, err = sClient.GetPermissionTemplate(ctx, instance.Spec.Name)
	if err != nil {
		if !sonarClient.IsErrNotFound(err) {
			return fmt.Errorf("failed to get sonar permission template: %w", err)
		}

		templateID, createErr := createPermissionTemplate(ctx, instance, sClient, r.service.K8sClient(), r.log)
		if createErr != nil {
			return fmt.Errorf("failed to create sonar permission template: %w", err)
		}

		instance.Status.ID = templateID
	} else {
		if instance.Status.ID == "" {
			return errors.New("permission template already exists in sonar")
		}

		tpl := specToClientTemplate(&instance.Spec, instance.Status.ID)

		if err = sClient.UpdatePermissionTemplate(ctx, tpl); err != nil {
			return fmt.Errorf("failed to update permission template: %w", err)
		}
	}

	if err = syncPermissionTemplateGroups(ctx, instance, sClient); err != nil {
		return fmt.Errorf("failed to sync permission template groups: %w", err)
	}

	if _, err = r.service.DeleteResource(
		ctx, instance, finalizer,
		func() error {
			if err = sClient.DeletePermissionTemplate(ctx, instance.Status.ID); err != nil {
				return fmt.Errorf("failed to delete permission template: %w", err)
			}

			return nil
		},
	); err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
}

func specToClientTemplate(spec *sonarApi.SonarPermissionTemplateSpec, id string) *sonarClient.PermissionTemplate {
	templateData := specToClientTemplateData(spec)

	return &sonarClient.PermissionTemplate{
		ID:                     id,
		PermissionTemplateData: *templateData,
	}
}

func specToClientTemplateData(spec *sonarApi.SonarPermissionTemplateSpec) *sonarClient.PermissionTemplateData {
	return &sonarClient.PermissionTemplateData{
		Name:              spec.Name,
		Description:       spec.Description,
		ProjectKeyPattern: spec.ProjectKeyPattern,
	}
}

func syncPermissionTemplateGroups(ctx context.Context, instance *sonarApi.SonarPermissionTemplate,
	sClient sonarClient.ClientInterface,
) error {
	groups, err := sClient.GetPermissionTemplateGroups(ctx, instance.Status.ID)
	if err != nil {
		return fmt.Errorf("failed to get permission template groups: %w", err)
	}

	for _, g := range groups {
		if err := sClient.RemoveGroupFromPermissionTemplate(ctx, instance.Status.ID, &g); err != nil {
			return fmt.Errorf("failed to remove group from permission template: %w", err)
		}
	}

	for _, g := range instance.Spec.GroupPermissions {
		if err := sClient.AddGroupToPermissionTemplate(ctx, instance.Status.ID, &sonarClient.PermissionTemplateGroup{
			GroupName:   g.GroupName,
			Permissions: g.Permissions,
		}); err != nil {
			return fmt.Errorf("failed to add group to permission template: %w", err)
		}
	}

	return nil
}

func createPermissionTemplate(ctx context.Context, sonarPermissionTemplate *sonarApi.SonarPermissionTemplate,
	sonarClient sonarClient.ClientInterface, k8sClient client.Client, logger logr.Logger,
) (string, error) {
	sonarPermTpl := specToClientTemplateData(&sonarPermissionTemplate.Spec)
	templateID, err := sonarClient.CreatePermissionTemplate(ctx, sonarPermTpl)
	if err != nil {
		return "", fmt.Errorf("failed to create sonar permission template: %w", err)
	}

	logger = logger.
		WithValues("template_id", templateID).
		WithValues("permission_template", sonarPermissionTemplate.Spec.Name)

	logger.Info("created permission template in sonar")

	sonarPermissionTemplate.Status.ID = templateID
	if err = k8sClient.Status().Update(ctx, sonarPermissionTemplate); err != nil {
		return "", fmt.Errorf("failed to update deletable object: %w", err)
	}

	logger.Info("updated cr status in k8s")

	return templateID, nil
}
