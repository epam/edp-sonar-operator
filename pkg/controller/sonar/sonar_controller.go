package sonar

import (
	"context"
	"fmt"
	"github.com/epmd-edp/sonar-operator/v2/pkg/helper"
	"github.com/epmd-edp/sonar-operator/v2/pkg/service/platform"
	"github.com/epmd-edp/sonar-operator/v2/pkg/service/sonar"
	"os"
	"time"

	v2v1alpha1 "github.com/epmd-edp/sonar-operator/v2/pkg/apis/edp/v1alpha1"
	errorsf "github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	StatusInstall          = "installing"
	StatusFailed           = "failed"
	StatusCreated          = "created"
	StatusConfiguring      = "configuring"
	StatusConfigured       = "configured"
	StatusExposeStart      = "exposing configs"
	StatusExposeFinish     = "configs exposed"
	StatusIntegrationStart = "integration started"
	StatusReady            = "ready"
	DefaultRequeueTime     = 30
)

var log = logf.Log.WithName("controller_sonar")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Sonar Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	scheme := mgr.GetScheme()
	client := mgr.GetClient()
	platformType := helper.GetPlatformTypeEnv()
	platformService, err := platform.NewPlatformService(platformType, scheme, &client)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	sonarService := sonar.NewSonarService(platformService, client, scheme)
	return &ReconcileSonar{
		client:  client,
		scheme:  scheme,
		service: sonarService,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("sonar-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Sonar
	err = c.Watch(&source.Kind{Type: &v2v1alpha1.Sonar{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileSonar{}

// ReconcileSonar reconciles a Sonar object
type ReconcileSonar struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client  client.Client
	scheme  *runtime.Scheme
	service sonar.SonarService
}

// Reconcile reads that state of the cluster for a Sonar object and makes changes based on the state read
// and what is in the Sonar.Spec
func (r *ReconcileSonar) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Sonar")

	// Fetch the Sonar instance
	instance := &v2v1alpha1.Sonar{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Status.Status == "" || instance.Status.Status == StatusFailed {
		reqLogger.Info("Installation has been started")
		err = r.updateStatus(instance, StatusInstall)
		if err != nil {
			return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, err
		}
	}

	instance, err = r.service.Install(*instance)
	if err != nil {
		r.updateStatus(instance, StatusFailed)
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, errorsf.Wrapf(err, "Installation has been failed")
	}

	if instance.Status.Status == StatusInstall {
		log.Info("Installation has finished")
		err = r.updateStatus(instance, StatusCreated)
		if err != nil {
			return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, err
		}
	}

	if dcIsReady, err := r.service.IsDeploymentReady(*instance); err != nil {
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, errorsf.Wrapf(err, "Checking if Deployment configs is ready has been failed")
	} else if !dcIsReady {
		reqLogger.Info("Deployment config is not ready for configuration yet")
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, nil
	}

	if instance.Status.Status == StatusCreated || instance.Status.Status == "" {
		reqLogger.Info("Configuration has started")
		err := r.updateStatus(instance, StatusConfiguring)
		if err != nil {
			return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, err
		}
	}

	instance, err, isFinished := r.service.Configure(*instance)
	if err != nil {
		reqLogger.Error(err, "Configuration has failed")
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, errorsf.Wrapf(err, "Configuration failed")
	} else if !isFinished {
		return reconcile.Result{RequeueAfter: DefaultRequeueTime * time.Second}, nil
	}

	if instance.Status.Status == StatusConfiguring {
		reqLogger.Info("Configuration has finished")
		err = r.updateStatus(instance, StatusConfigured)
		if err != nil {
			return reconcile.Result{RequeueAfter: 10 * time.Second}, err
		}
	}

	if instance.Status.Status == StatusConfigured {
		reqLogger.Info("Exposing configuration has started")
		err = r.updateStatus(instance, StatusExposeStart)
		if err != nil {
			return reconcile.Result{RequeueAfter: 10 * time.Second}, err
		}
	}

	instance, err = r.service.ExposeConfiguration(*instance)
	if err != nil {
		return reconcile.Result{RequeueAfter: 10 * time.Second}, errorsf.Wrapf(err, "Exposing configuration failed")
	}

	if instance.Status.Status == StatusExposeStart {
		reqLogger.Info("Exposing configuration has finished")
		err = r.updateStatus(instance, StatusExposeFinish)
		if err != nil {
			return reconcile.Result{RequeueAfter: 10 * time.Second}, err
		}
	}

	if instance.Status.Status == StatusExposeFinish {
		reqLogger.Info("Integration has started")
		err = r.updateStatus(instance, StatusIntegrationStart)
		if err != nil {
			return reconcile.Result{RequeueAfter: 10 * time.Second}, err
		}
	}

	instance, err = r.service.Integration(*instance)
	if err != nil {
		return reconcile.Result{RequeueAfter: 10 * time.Second}, errorsf.Wrapf(err, "Integration failed")
	}

	if instance.Status.Status == StatusIntegrationStart {
		reqLogger.Info("Integration has finished")
		err = r.updateStatus(instance, StatusReady)
		if err != nil {
			return reconcile.Result{RequeueAfter: 10 * time.Second}, err
		}
	}

	err = r.updateAvailableStatus(instance, true)
	if err != nil {
		reqLogger.Info("Failed to update availability status")
		return reconcile.Result{RequeueAfter: 30 * time.Second}, err
	}

	reqLogger.Info("Reconciling Sonar component has been finished", "namespace", request.Namespace, "name", request.Name)
	return reconcile.Result{Requeue: false}, nil
}

func (r *ReconcileSonar) updateStatus(instance *v2v1alpha1.Sonar, newStatus string) error {
	reqLogger := log.WithValues("Request.Namespace", instance.Namespace, "Request.Name", instance.Name).WithName("status_update")
	currentStatus := instance.Status.Status
	instance.Status.Status = newStatus
	instance.Status.LastTimeUpdated = time.Now()
	err := r.client.Status().Update(context.TODO(), instance)
	if err != nil {
		err = r.client.Update(context.TODO(), instance)
		if err != nil {
			return errorsf.Wrapf(err, "Couldn't update status from '%v' to '%v'", currentStatus, newStatus)
		}
	}

	reqLogger.Info(fmt.Sprintf("Status has been updated to '%v'", newStatus))
	return nil
}

func (r ReconcileSonar) updateAvailableStatus(instance *v2v1alpha1.Sonar, value bool) error {
	reqLogger := log.WithValues("Request.Namespace", instance.Namespace, "Request.Name", instance.Name).WithName("status_update")
	if instance.Status.Available != value {
		instance.Status.Available = value
		instance.Status.LastTimeUpdated = time.Now()
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			err := r.client.Update(context.TODO(), instance)
			if err != nil {
				return errorsf.Wrapf(err, "Couldn't update availability status to %v", value)
			}
		}
		reqLogger.Info(fmt.Sprintf("Availability status has been updated to '%v'", value))
	}
	return nil
}
