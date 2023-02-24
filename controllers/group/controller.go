package group

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

const finalizer = "sonar.group.operator"

type Reconcile struct {
	service sonar.ServiceInterface
	client  client.Client
	log     logr.Logger
}

func NewReconcile(client client.Client, scheme *runtime.Scheme, log logr.Logger, platformType string) (*Reconcile, error) {
	ps, err := platform.NewService(platformType, scheme, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create platform: %w", err)
	}

	return &Reconcile{
		service: sonar.NewService(ps, client),
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

//+kubebuilder:rbac:groups=v1.edp.epam.com,namespace=placeholder,resources=sonargroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=v1.edp.epam.com,namespace=placeholder,resources=sonargroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=v1.edp.epam.com,namespace=placeholder,resources=sonargroups/finalizers,verbs=update

func (r *Reconcile) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := r.log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	log.Info("Reconciling SonarGroup")

	var instance sonarApi.SonarGroup
	if err := r.client.Get(ctx, request.NamespacedName, &instance); err != nil {
		if k8sErrors.IsNotFound(err) {
			log.Info("instance not found")

			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, fmt.Errorf("failed to get sonar group from k8s: %w", err)
	}

	var result reconcile.Result

	if err := r.tryReconcile(ctx, &instance); err != nil {
		instance.Status.Value = err.Error()
		result.RequeueAfter = helper.SetFailureCount(&instance)

		log.Error(err, "an error has occurred while handling sonar group", "name",
			request.Name)
	} else {
		helper.SetSuccessStatus(&instance)
	}

	if err := r.client.Status().Update(ctx, &instance); err != nil {
		return result, fmt.Errorf("failed to update status: %w", err)
	}

	log.Info("Reconciling done")

	return result, nil
}

func (r *Reconcile) tryReconcile(ctx context.Context, instance *sonarApi.SonarGroup) error {
	sClient, err := r.service.ClientForChild(ctx, instance)
	if err != nil {
		return fmt.Errorf("failed to init sonar rest client: %w", err)
	}

	_, err = sClient.GetGroup(ctx, instance.Spec.Name)

	switch {
	case sonarClient.IsErrNotFound(err):
		sonarGroup := sonarClient.Group{Name: instance.Spec.Name, Description: instance.Spec.Description}
		if err = sClient.CreateGroup(ctx, &sonarGroup); err != nil {
			return fmt.Errorf("failed to create sonar group: %w", err)
		}

		instance.Status.ID = sonarGroup.ID
	case err != nil:
		return fmt.Errorf("failed to get group: %w", err)
	default:
		if instance.Status.ID == "" {
			return errors.New("group already exists in sonar")
		}

		if err = sClient.UpdateGroup(ctx, instance.Spec.Name, &sonarClient.Group{
			Name:        instance.Spec.Name,
			Description: instance.Spec.Description,
		}); err != nil {
			return fmt.Errorf("failed to update group: %w", err)
		}
	}

	if _, err = r.service.DeleteResource(ctx, instance, finalizer, func() error {
		if err = sClient.DeleteGroup(ctx, instance.Spec.Name); err != nil {
			return fmt.Errorf("failed to delete group: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
}
