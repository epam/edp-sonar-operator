package user

import (
	"context"
	"fmt"
	"time"

	"github.com/epam/edp-sonar-operator/internal/controller/user/chain"

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

// SonarUserReconciler reconciles a SonarUser object
type SonarUserReconciler struct {
	client            client.Client
	scheme            *runtime.Scheme
	apiClientProvider apiClientProvider
}

// NewSonarUserReconciler returns a new SonarUserReconciler instance.
func NewSonarUserReconciler(
	client client.Client,
	scheme *runtime.Scheme,
	apiClientProvider apiClientProvider,
) *SonarUserReconciler {
	return &SonarUserReconciler{client: client, scheme: scheme, apiClientProvider: apiClientProvider}
}

// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarusers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarusers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarusers/finalizers,verbs=update
// +kubebuilder:rbac:groups="",namespace=placeholder,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SonarUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling SonarUser")

	user := &sonarApi.SonarUser{}

	err := r.client.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	sonarApiClient, err := r.apiClientProvider.GetSonarApiClientFromSonarRef(ctx, req.Namespace, user)
	if err != nil {
		log.Error(err, "An error has occurred while getting sonar api client")

		return ctrl.Result{
			RequeueAfter: errorRequeueTime,
		}, nil
	}

	if user.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(user, sonarOperatorFinalizer) {
			if err = chain.NewRemoveUser(sonarApiClient).ServeRequest(ctx, user); err != nil {
				log.Error(err, "An error has occurred while deleting SonarUser")

				return ctrl.Result{
					RequeueAfter: errorRequeueTime,
				}, nil
			}

			controllerutil.RemoveFinalizer(user, sonarOperatorFinalizer)

			if err = r.client.Update(ctx, user); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if controllerutil.AddFinalizer(user, sonarOperatorFinalizer) {
		err = r.client.Update(ctx, user)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	oldStatus := user.Status

	if err = chain.MakeChain(sonarApiClient, r.client).ServeRequest(ctx, user); err != nil {
		log.Error(err, "An error has occurred while handling SonarUser")

		user.Status.Value = "error"
		user.Status.Error = err.Error()

		if err = r.updateSonarUserStatus(ctx, user, oldStatus); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{
			RequeueAfter: errorRequeueTime,
		}, nil
	}

	user.Status.Value = common.StatusCreated
	user.Status.Error = ""

	if err = r.updateSonarUserStatus(ctx, user, oldStatus); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SonarUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sonarApi.SonarUser{}).
		Complete(r)
}

func (r *SonarUserReconciler) updateSonarUserStatus(ctx context.Context, sonarUser *sonarApi.SonarUser, oldStatus sonarApi.SonarUserStatus) error {
	if sonarUser.Status == oldStatus {
		return nil
	}

	if err := r.client.Status().Update(ctx, sonarUser); err != nil {
		return fmt.Errorf("failed to update SonarUser status: %w", err)
	}

	return nil
}
