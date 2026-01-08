package group

import (
	"context"
	"fmt"
	"time"

	"github.com/epam/edp-sonar-operator/internal/controller/group/chain"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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

type SonarGroupReconciler struct {
	client            client.Client
	scheme            *runtime.Scheme
	apiClientProvider apiClientProvider
}

func NewSonarGroupReconciler(
	k8sClient client.Client,
	scheme *runtime.Scheme,
	apiClientProvider apiClientProvider,
) *SonarGroupReconciler {
	return &SonarGroupReconciler{
		client:            k8sClient,
		scheme:            scheme,
		apiClientProvider: apiClientProvider,
	}
}

func (r *SonarGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sonarApi.SonarGroup{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonargroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonargroups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonargroups/finalizers,verbs=update
// +kubebuilder:rbac:groups="",namespace=placeholder,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SonarGroupReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling SonarGroup")

	group := &sonarApi.SonarGroup{}

	err := r.client.Get(ctx, req.NamespacedName, group)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	sonarApiClient, err := r.apiClientProvider.GetSonarApiClientFromSonarRef(ctx, req.Namespace, group)
	if err != nil {
		log.Error(err, "An error has occurred while getting sonar api client")

		return ctrl.Result{
			RequeueAfter: errorRequeueTime,
		}, nil
	}

	if group.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(group, sonarOperatorFinalizer) {
			if err = chain.NewRemoveGroup(sonarApiClient).ServeRequest(ctx, group); err != nil {
				log.Error(err, "An error has occurred while deleting SonarGroup")

				return ctrl.Result{
					RequeueAfter: errorRequeueTime,
				}, nil
			}

			controllerutil.RemoveFinalizer(group, sonarOperatorFinalizer)

			if err = r.client.Update(ctx, group); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if controllerutil.AddFinalizer(group, sonarOperatorFinalizer) {
		err = r.client.Update(ctx, group)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	oldStatus := group.Status

	if err = chain.MakeChain(sonarApiClient).ServeRequest(ctx, group); err != nil {
		log.Error(err, "An error has occurred while handling SonarGroup")

		group.Status.Value = "error"
		group.Status.Error = err.Error()

		if err = r.updateSonarGroupStatus(ctx, group, oldStatus); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{
			RequeueAfter: errorRequeueTime,
		}, nil
	}

	group.Status.Value = common.StatusCreated
	group.Status.Error = ""

	if err = r.updateSonarGroupStatus(ctx, group, oldStatus); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SonarGroupReconciler) updateSonarGroupStatus(
	ctx context.Context,
	group *sonarApi.SonarGroup,
	oldStatus sonarApi.SonarGroupStatus,
) error {
	if group.Status == oldStatus {
		return nil
	}

	if err := r.client.Status().Update(ctx, group); err != nil {
		return fmt.Errorf("failed to update SonarGroup status: %w", err)
	}

	return nil
}
