package qualitygate

import (
	"context"
	"fmt"
	"time"

	"github.com/epam/edp-sonar-operator/internal/controller/qualitygate/chain"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/epam/edp-sonar-operator/api/common"
	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	sonarclient "github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

const (
	sonarOperatorFinalizer = "edp.epam.com/finalizer"
	errorRequeueTime       = time.Second * 30
)

type apiClientProvider interface {
	GetSonarApiClientFromSonarRef(ctx context.Context, namespace string, sonarRef common.HasSonarRef) (*sonarclient.Client, error)
}

// SonarQualityGateReconciler reconciles a SonarQualityGate object
type SonarQualityGateReconciler struct {
	client            client.Client
	scheme            *runtime.Scheme
	apiClientProvider apiClientProvider
}

func NewSonarQualityGateReconciler(
	client client.Client,
	scheme *runtime.Scheme,
	apiClientProvider apiClientProvider,
) *SonarQualityGateReconciler {
	return &SonarQualityGateReconciler{client: client, scheme: scheme, apiClientProvider: apiClientProvider}
}

// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarqualitygates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarqualitygates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarqualitygates/finalizers,verbs=update
// +kubebuilder:rbac:groups="",namespace=placeholder,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SonarQualityGateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling SonarQualityGate")

	gate := &sonarApi.SonarQualityGate{}

	err := r.client.Get(ctx, req.NamespacedName, gate)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	sonarApiClient, err := r.apiClientProvider.GetSonarApiClientFromSonarRef(ctx, req.Namespace, gate)
	if err != nil {
		log.Error(err, "An error has occurred while getting sonar api client")

		return ctrl.Result{
			RequeueAfter: errorRequeueTime,
		}, nil
	}

	if gate.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(gate, sonarOperatorFinalizer) {
			if err = chain.NewRemoveQualityGate(sonarApiClient).ServeRequest(ctx, gate); err != nil {
				log.Error(err, "An error has occurred while deleting QualityGate")

				return ctrl.Result{
					RequeueAfter: errorRequeueTime,
				}, nil
			}

			controllerutil.RemoveFinalizer(gate, sonarOperatorFinalizer)

			if err = r.client.Update(ctx, gate); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if controllerutil.AddFinalizer(gate, sonarOperatorFinalizer) {
		err = r.client.Update(ctx, gate)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	oldStatus := gate.Status

	if err = chain.MakeChain(sonarApiClient).ServeRequest(ctx, gate); err != nil {
		log.Error(err, "An error has occurred while handling SonarQualityGate")

		gate.Status.Value = "error"
		gate.Status.Error = err.Error()

		if err = r.updateSonarQualityGateStatus(ctx, gate, oldStatus); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{
			RequeueAfter: errorRequeueTime,
		}, nil
	}

	gate.Status.Value = common.StatusCreated
	gate.Status.Error = ""

	if err = r.updateSonarQualityGateStatus(ctx, gate, oldStatus); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SonarQualityGateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sonarApi.SonarQualityGate{}).
		Complete(r)
}

func (r *SonarQualityGateReconciler) updateSonarQualityGateStatus(
	ctx context.Context,
	gate *sonarApi.SonarQualityGate,
	oldStatus sonarApi.SonarQualityGateStatus,
) error {
	if gate.Status == oldStatus {
		return nil
	}

	if err := r.client.Status().Update(ctx, gate); err != nil {
		return fmt.Errorf("failed to update SonarQualityGate status: %w", err)
	}

	return nil
}
