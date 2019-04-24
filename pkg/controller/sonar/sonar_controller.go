package sonar

import (
	"context"
	"sonar-operator/pkg/service"

	edpv1alpha1 "sonar-operator/pkg/apis/edp/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	logPrint "log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
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
	platformService, _ := service.NewPlatformService(*mgr.GetScheme())
	sonarService := service.NewSonarService(platformService)
	return &ReconcileSonar{client: mgr.GetClient(),
		scheme:  mgr.GetScheme(),
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
	err = c.Watch(&source.Kind{Type: &edpv1alpha1.Sonar{}}, &handler.EnqueueRequestForObject{})
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
	service service.SonarService
}

// Reconcile reads that state of the cluster for a Sonar object and makes changes based on the state read
// and what is in the Sonar.Spec
func (r *ReconcileSonar) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Sonar")

	// Fetch the Sonar instance
	instance := &edpv1alpha1.Sonar{}
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

	_ = r.service.Install(*instance)
	logPrint.Printf("Reconciling StaticAnalysisTool %s/%s has been finished", request.Namespace, request.Name)
	reqLogger.Info("Reconciling Sonar component %s/%s has been finished", request.Namespace, request.Name)
	return reconcile.Result{}, nil
}
