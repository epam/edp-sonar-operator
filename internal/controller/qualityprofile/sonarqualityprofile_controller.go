package qualityprofile

import (
	"context"
	"fmt"
	"time"

	"github.com/epam/edp-sonar-operator/internal/controller/qualityprofile/chain"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
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

// SonarQualityProfileReconciler reconciles a SonarQualityProfile object
type SonarQualityProfileReconciler struct {
	client            client.Client
	scheme            *runtime.Scheme
	apiClientProvider apiClientProvider
}

func NewSonarQualityProfileReconciler(
	client client.Client,
	scheme *runtime.Scheme,
	apiClientProvider apiClientProvider,
) *SonarQualityProfileReconciler {
	return &SonarQualityProfileReconciler{client: client, scheme: scheme, apiClientProvider: apiClientProvider}
}

// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarqualityprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarqualityprofiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarqualityprofiles/finalizers,verbs=update
// +kubebuilder:rbac:groups="",namespace=placeholder,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SonarQualityProfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling SonarQualityProfile")

	profile := &sonarApi.SonarQualityProfile{}

	err := r.client.Get(ctx, req.NamespacedName, profile)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	sonarApiClient, err := r.apiClientProvider.GetSonarApiClientFromSonarRef(ctx, req.Namespace, profile)
	if err != nil {
		log.Error(err, "An error has occurred while getting sonar api client")

		return ctrl.Result{
			RequeueAfter: errorRequeueTime,
		}, nil
	}

	if profile.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(profile, sonarOperatorFinalizer) {
			if err = chain.NewRemoveQualityProfile(sonarApiClient).ServeRequest(ctx, profile); err != nil {
				log.Error(err, "An error has occurred while deleting QualityProfile")

				return ctrl.Result{
					RequeueAfter: errorRequeueTime,
				}, nil
			}

			controllerutil.RemoveFinalizer(profile, sonarOperatorFinalizer)

			if err = r.client.Update(ctx, profile); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if controllerutil.AddFinalizer(profile, sonarOperatorFinalizer) {
		err = r.client.Update(ctx, profile)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	oldStatus := profile.Status

	if err = chain.MakeChain(sonarApiClient).ServeRequest(ctx, profile); err != nil {
		log.Error(err, "An error has occurred while handling SonarQualityProfile")

		profile.Status.Value = "error"
		profile.Status.Error = err.Error()

		if err = r.updateSonarQualityProfileStatus(ctx, profile, oldStatus); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{
			RequeueAfter: errorRequeueTime,
		}, nil
	}

	profile.Status.Value = common.StatusCreated
	profile.Status.Error = ""

	if err = r.updateSonarQualityProfileStatus(ctx, profile, oldStatus); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SonarQualityProfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sonarApi.SonarQualityProfile{}).
		Complete(r)
}

func (r *SonarQualityProfileReconciler) updateSonarQualityProfileStatus(
	ctx context.Context,
	profile *sonarApi.SonarQualityProfile,
	oldStatus sonarApi.SonarQualityProfileStatus,
) error {
	if profile.Status == oldStatus {
		return nil
	}

	if err := r.client.Status().Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to update SonarQualityProfile status: %w", err)
	}

	return nil
}
