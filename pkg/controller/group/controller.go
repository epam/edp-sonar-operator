package group

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

const finalizer = "sonar.group.operator"

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
		log:     log.WithName("sonar-group"),
	}, nil
}

func (r *Reconcile) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sonarApi.SonarGroup{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: isSpecUpdated,
		})).
		Complete(r)
}

func isSpecUpdated(e event.UpdateEvent) bool {
	oo := e.ObjectOld.(*sonarApi.SonarGroup)
	no := e.ObjectNew.(*sonarApi.SonarGroup)

	return !reflect.DeepEqual(oo.Spec, no.Spec) ||
		(oo.GetDeletionTimestamp().IsZero() && !no.GetDeletionTimestamp().IsZero())
}

func (r *Reconcile) Reconcile(ctx context.Context, request reconcile.Request) (result reconcile.Result, resultErr error) {
	log := r.log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	log.Info("Reconciling SonarGroup")

	var instance sonarApi.SonarGroup
	if err := r.client.Get(ctx, request.NamespacedName, &instance); err != nil {
		if k8sErrors.IsNotFound(err) {
			log.Info("instance not found")
			return
		}

		resultErr = errors.Wrap(err, "unable to get sonar group from k8s")
		return
	}

	if err := r.tryReconcile(ctx, &instance); err != nil {
		instance.Status.Value = err.Error()
		result.RequeueAfter = helper.SetFailureCount(&instance)
		log.Error(err, "an error has occurred while handling sonar group", "name",
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

func (r *Reconcile) tryReconcile(ctx context.Context, instance *sonarApi.SonarGroup) error {
	sClient, err := r.service.ClientForChild(ctx, instance)
	if err != nil {
		return errors.Wrap(err, "unable to init sonar rest client")
	}

	_, err = sClient.GetGroup(ctx, instance.Spec.Name)

	switch {
	case sonarClient.IsErrNotFound(err):
		sonarGroup := sonarClient.Group{Name: instance.Spec.Name, Description: instance.Spec.Description}
		if err = sClient.CreateGroup(ctx, &sonarGroup); err != nil {
			return errors.Wrap(err, "unable to create sonar group")
		}
		instance.Status.ID = sonarGroup.ID
	case err != nil:
		return errors.Wrap(err, "unexpected error during get group")
	default:
		if instance.Status.ID == "" {
			return errors.New("group already exists in sonar")
		}

		if err = sClient.UpdateGroup(ctx, instance.Spec.Name, &sonarClient.Group{Name: instance.Spec.Name,
			Description: instance.Spec.Description}); err != nil {
			return errors.Wrap(err, "unable to update group")
		}
	}

	if _, err = r.service.DeleteResource(ctx, instance, finalizer, func() error {
		if err = sClient.DeleteGroup(ctx, instance.Spec.Name); err != nil {
			return errors.Wrap(err, "unable to delete group")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "unable to delete resource")
	}

	return nil
}
