package sonar

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	coreV1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/service/platform"
	"github.com/epam/edp-sonar-operator/pkg/service/sonar"
)

const (
	StatusConfiguring      = "configuring"
	StatusConfigured       = "configured"
	StatusExposeStart      = "exposing configs"
	StatusExposeFinish     = "configs exposed"
	StatusIntegrationStart = "integration started"
	StatusReady            = "ready"
	DefaultRequeueTime     = 30
	ShortRequeueTime       = 10
)

func NewReconcileSonar(
	client client.Client,
	scheme *runtime.Scheme,
	log logr.Logger,
	platformType string,
) (*ReconcileSonar, error) {
	ps, err := platform.NewService(platformType, scheme, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create platform service: %w", err)
	}

	return &ReconcileSonar{
		client:   client,
		scheme:   scheme,
		service:  sonar.NewService(ps, client),
		log:      log.WithName("sonar"),
		platform: ps,
	}, nil
}

type ReconcileSonar struct {
	client   client.Client
	scheme   *runtime.Scheme
	service  sonar.ServiceInterface
	log      logr.Logger
	platform platform.Service
}

func (r *ReconcileSonar) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sonarApi.Sonar{}).
		Complete(r)
}

func (r *ReconcileSonar) getSonarSecret(ctx context.Context, secretName, namespace string) (*coreV1.Secret, error) {
	var secret coreV1.Secret

	if err := r.client.Get(ctx, types.NamespacedName{
		Name:      secretName,
		Namespace: namespace,
	}, &secret); err != nil {
		return nil, fmt.Errorf("failed to get sonar secret: %w", err)
	}

	if secret.Data["token"] == nil {
		return nil, fmt.Errorf("one or more fields are missing in sonar secret")
	}

	return &secret, nil
}

//+kubebuilder:rbac:groups=v1.edp.epam.com,namespace=placeholder,resources=sonars,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=v1.edp.epam.com,namespace=placeholder,resources=sonars/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=v1.edp.epam.com,namespace=placeholder,resources=sonars/finalizers,verbs=update

func (r *ReconcileSonar) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := r.log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	log.Info("Reconciling Sonar")

	instance := &sonarApi.Sonar{}
	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		if k8sErrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	secret, err := r.getSonarSecret(ctx, instance.Spec.Secret, instance.Namespace)
	if err != nil {
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, err
	}

	if err = r.platform.SetOwnerReference(instance, secret); err != nil {
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, err
	}

	if instance.Status.Status == "" {
		log.Info("Configuration has started")
		if err = r.updateStatus(ctx, instance, StatusConfiguring); err != nil {
			return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, err
		}
	}

	if err = r.service.Configure(ctx, instance); err != nil {
		log.Error(err, "Configuration has failed")
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second},
			fmt.Errorf("failed to configure: %w", err)
	}

	if instance.Status.Status == StatusConfiguring {
		log.Info("Configuration has finished")
		if err = r.updateStatus(ctx, instance, StatusConfigured); err != nil {
			return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, err
		}
	}

	if instance.Status.Status == StatusConfigured {
		log.Info("Exposing configuration has started")
		if err = r.updateStatus(ctx, instance, StatusExposeStart); err != nil {
			return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, err
		}
	}

	err = r.service.ExposeConfiguration(ctx, instance)
	if err != nil {
		return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, fmt.Errorf("failed to expose configuration: %w", err)
	}

	if instance.Status.Status == StatusExposeStart {
		log.Info("Exposing configuration has finished")
		if err = r.updateStatus(ctx, instance, StatusExposeFinish); err != nil {
			return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, err
		}
	}

	if instance.Status.Status == StatusExposeFinish {
		log.Info("Integration has started")
		if err = r.updateStatus(ctx, instance, StatusIntegrationStart); err != nil {
			return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, err
		}
	}

	instance, err = r.service.Integration(ctx, instance)
	if err != nil {
		return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, fmt.Errorf("failed to integrate: %w", err)
	}

	if instance.Status.Status == StatusIntegrationStart {
		log.Info("Integration has finished")

		if err = r.updateStatus(ctx, instance, StatusReady); err != nil {
			return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, err
		}
	}

	if err = r.updateAvailableStatus(ctx, instance, true); err != nil {
		log.Info("failed to update availability status")

		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, err
	}

	log.Info("Reconciling Sonar component has been finished", "namespace", request.Namespace, "name", request.Name)

	return reconcile.Result{Requeue: false}, nil
}

func (r *ReconcileSonar) updateStatus(ctx context.Context, instance *sonarApi.Sonar, newStatus string) error {
	log := r.log.WithValues("Request.Namespace", instance.Namespace, "Request.Name", instance.Name).
		WithName("status_update")
	currentStatus := instance.Status.Status
	instance.Status.Status = newStatus
	instance.Status.LastTimeUpdated = metaV1.Now()

	if err := r.client.Status().Update(ctx, instance); err != nil {
		if updErr := r.client.Update(ctx, instance); updErr != nil {
			return fmt.Errorf("failed to update status from %s to %s: %w", currentStatus, newStatus, err)
		}
	}

	log.Info(fmt.Sprintf("Status has been updated to '%v'", newStatus))

	return nil
}

func (r *ReconcileSonar) updateAvailableStatus(ctx context.Context, instance *sonarApi.Sonar, value bool) error {
	log := r.log.
		WithValues("Request.Namespace", instance.Namespace, "Request.Name", instance.Name).
		WithName("status_update")

	if instance.Status.Available != value {
		instance.Status.Available = value
		instance.Status.LastTimeUpdated = metaV1.Now()

		if err := r.client.Status().Update(ctx, instance); err != nil {
			if updErr := r.client.Update(ctx, instance); updErr != nil {
				return fmt.Errorf("failed to update availability status to %t: %w", value, err)
			}
		}

		log.Info(fmt.Sprintf("Availability status has been updated to '%v'", value))
	}

	return nil
}
