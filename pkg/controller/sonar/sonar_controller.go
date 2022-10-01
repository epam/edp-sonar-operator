package sonar

import (
	"context"
	"fmt"
	"time"

	"github.com/dchest/uniuri"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	coreV1Api "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	sonarApi "github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/sonar"
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
		return nil, errors.Wrap(err, "unable to create platform service")
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

func (r *ReconcileSonar) createDBSecret(sonarName, namespace string) (*coreV1Api.Secret, error) {
	dbSecret := map[string][]byte{
		"database-user":     []byte("admin"),
		"database-password": []byte(uniuri.New()),
	}

	sonarDbName := fmt.Sprintf("%v-db", sonarName)
	secret, err := r.platform.CreateSecret(sonarName, namespace, sonarDbName, dbSecret)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create secret for %s", sonarDbName)
	}
	return secret, nil
}

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

	secret, err := r.createDBSecret(instance.Name, instance.Namespace)
	if err != nil {
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, err
	}
	if err = r.platform.SetOwnerReference(instance, secret); err != nil {
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, err
	}

	if dcIsReady, errIsReady := r.service.IsDeploymentReady(instance); errIsReady != nil {
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, errors.Wrapf(errIsReady, "Checking if Deployment configs is ready has been failed")
	} else if !dcIsReady {
		log.Info("Deployment config is not ready for configuration yet")
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, nil
	}

	if instance.Status.Status == "" {
		log.Info("Configuration has started")
		err = r.updateStatus(ctx, instance, StatusConfiguring)
		if err != nil {
			return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, err
		}
	}

	if err = r.service.Configure(ctx, instance); err != nil {
		log.Error(err, "Configuration has failed")
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second},
			errors.Wrapf(err, "Configuration failed")
	}

	if instance.Status.Status == StatusConfiguring {
		log.Info("Configuration has finished")
		err = r.updateStatus(ctx, instance, StatusConfigured)
		if err != nil {
			return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, err
		}
	}

	if instance.Status.Status == StatusConfigured {
		log.Info("Exposing configuration has started")
		err = r.updateStatus(ctx, instance, StatusExposeStart)
		if err != nil {
			return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, err
		}
	}

	err = r.service.ExposeConfiguration(ctx, instance)
	if err != nil {
		return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, errors.Wrapf(err, "Exposing configuration failed")
	}

	if instance.Status.Status == StatusExposeStart {
		log.Info("Exposing configuration has finished")
		err = r.updateStatus(ctx, instance, StatusExposeFinish)
		if err != nil {
			return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, err
		}
	}

	if instance.Status.Status == StatusExposeFinish {
		log.Info("Integration has started")
		err = r.updateStatus(ctx, instance, StatusIntegrationStart)
		if err != nil {
			return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, err
		}
	}

	instance, err = r.service.Integration(ctx, instance)
	if err != nil {
		return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, errors.Wrapf(err, "Integration failed")
	}

	if instance.Status.Status == StatusIntegrationStart {
		log.Info("Integration has finished")
		err = r.updateStatus(ctx, instance, StatusReady)
		if err != nil {
			return reconcile.Result{RequeueAfter: ShortRequeueTime * time.Second}, err
		}
	}

	err = r.updateAvailableStatus(ctx, instance, true)
	if err != nil {
		log.Info("Failed to update availability status")
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
	instance.Status.LastTimeUpdated = metav1.Now()
	if err := r.client.Status().Update(ctx, instance); err != nil {
		if err := r.client.Update(ctx, instance); err != nil {
			return errors.Wrapf(err, "Couldn't update status from '%v' to '%v'", currentStatus, newStatus)
		}
	}

	log.Info(fmt.Sprintf("Status has been updated to '%v'", newStatus))
	return nil
}

func (r ReconcileSonar) updateAvailableStatus(ctx context.Context, instance *sonarApi.Sonar, value bool) error {
	log := r.log.WithValues("Request.Namespace", instance.Namespace, "Request.Name", instance.Name).
		WithName("status_update")
	if instance.Status.Available != value {
		instance.Status.Available = value
		instance.Status.LastTimeUpdated = metav1.Now()
		if err := r.client.Status().Update(ctx, instance); err != nil {
			if err = r.client.Update(ctx, instance); err != nil {
				return errors.Wrapf(err, "Couldn't update availability status to %v", value)
			}
		}
		log.Info(fmt.Sprintf("Availability status has been updated to '%v'", value))
	}
	return nil
}
