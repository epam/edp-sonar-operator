package permission_template

import (
	"context"
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

	sonarApi "github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1alpha1"
	sonarClient "github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/v2/pkg/helper"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/sonar"
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
		return nil, errors.Wrap(err, "unable to create platform service")
	}

	return &Reconcile{
		service: sonar.NewService(ps, client, scheme),
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

func (r *Reconcile) Reconcile(ctx context.Context, request reconcile.Request) (result reconcile.Result, resultErr error) {
	log := r.log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	log.Info("Reconciling SonarPermissionTemplate")

	var instance sonarApi.SonarPermissionTemplate
	if err := r.client.Get(ctx, request.NamespacedName, &instance); err != nil {
		if k8sErrors.IsNotFound(err) {
			log.Info("instance not found")
			return
		}

		resultErr = errors.Wrap(err, "unable to get sonar permission tpl from k8s")
		return
	}

	if err := r.tryReconcile(ctx, &instance); err != nil {
		instance.Status.Value = err.Error()
		result.RequeueAfter = helper.SetFailureCount(&instance)
		log.Error(err, "an error has occurred while handling sonar permission template", "name",
			request.Name)
	} else {
		helper.SetSuccessStatus(&instance)
	}

	if err := r.client.Status().Update(ctx, &instance); err != nil {
		resultErr = errors.Wrap(err, "unable to update status")
	}

	log.Info("Reconciling done")
	return
}

func (r *Reconcile) tryReconcile(ctx context.Context, instance *sonarApi.SonarPermissionTemplate) error {
	sClient, err := r.service.ClientForChild(ctx, instance)
	if err != nil {
		return errors.Wrap(err, "unable to init sonar rest client")
	}

	_, err = sClient.GetPermissionTemplate(ctx, instance.Spec.Name)
	if sonarClient.IsErrNotFound(err) {
		templateID, createErr := createPermissionTemplate(ctx, instance, sClient, r.service.K8sClient(), r.log)
		if createErr != nil {
			return errors.Wrap(createErr, "unable to create sonar permission template")
		}
		instance.Status.ID = templateID
	} else if err != nil {
		return errors.Wrap(err, "unexpected error during get sonar permission template")
	} else {
		if instance.Status.ID == "" {
			return errors.New("permission template already exists in sonar")
		}

		tpl := specToClientTemplate(&instance.Spec, instance.Status.ID)
		if err = sClient.UpdatePermissionTemplate(ctx, tpl); err != nil {
			return errors.Wrap(err, "unable to update permission template")
		}
	}

	if err = syncPermissionTemplateGroups(ctx, instance, sClient); err != nil {
		return errors.Wrap(err, "unable to sync permission template groups")
	}

	if _, err = r.service.DeleteResource(ctx, instance, finalizer, func() error {
		if err = sClient.DeletePermissionTemplate(ctx, instance.Status.ID); err != nil {
			return errors.Wrap(err, "unable to delete permission template")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "unable to delete resource")
	}

	return nil
}

func specToClientTemplate(spec *sonarApi.SonarPermissionTemplateSpec, ID string) *sonarClient.PermissionTemplate {
	templateData := specToClientTemplateData(spec)
	return &sonarClient.PermissionTemplate{
		ID:                     ID,
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
	sClient sonar.ClientInterface) error {
	groups, err := sClient.GetPermissionTemplateGroups(ctx, instance.Status.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get permission template groups")
	}

	for _, g := range groups {
		if err := sClient.RemoveGroupFromPermissionTemplate(ctx, instance.Status.ID, &g); err != nil {
			return errors.Wrap(err, "unable to remove group from permission template")
		}
	}

	for _, g := range instance.Spec.GroupPermissions {
		if err := sClient.AddGroupToPermissionTemplate(ctx, instance.Status.ID, &sonarClient.PermissionTemplateGroup{
			GroupName:   g.GroupName,
			Permissions: g.Permissions,
		}); err != nil {
			return errors.Wrap(err, "unable to add group to permission template")
		}
	}

	return nil
}

func createPermissionTemplate(ctx context.Context, sonarPermissionTemplate *sonarApi.SonarPermissionTemplate,
	sonarClient sonar.ClientInterface, k8sClient client.Client, logger logr.Logger,
) (string, error) {
	sonarPermTpl := specToClientTemplateData(&sonarPermissionTemplate.Spec)
	templateID, err := sonarClient.CreatePermissionTemplate(ctx, sonarPermTpl)
	if err != nil {
		return "", errors.Wrap(err, "unable to create sonar permission template")
	}
	logger = logger.
		WithValues("template_id", templateID).
		WithValues("permission_template", sonarPermissionTemplate.Spec.Name)
	logger.Info("created permission template in sonar")

	sonarPermissionTemplate.Status.ID = templateID
	if err = k8sClient.Status().Update(ctx, sonarPermissionTemplate); err != nil {
		return "", errors.Wrap(err, "unable to update deletable object")
	}
	logger.Info("updated cr status in k8s")

	return templateID, nil
}
