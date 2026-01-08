package project

import (
	"context"
	"fmt"
	"time"

	"github.com/epam/edp-sonar-operator/internal/controller/project/chain"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/epam/edp-sonar-operator/api/common"
	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	sonarclient "github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/helper"
)

const (
	errorRequeueTime = time.Second * 30
)

type apiClientProvider interface {
	GetSonarApiClientFromSonarRef(ctx context.Context, namespace string, sonarRef common.HasSonarRef) (*sonarclient.Client, error)
}

// SonarProjectReconciler reconciles a SonarProject object
type SonarProjectReconciler struct {
	client            client.Client
	scheme            *runtime.Scheme
	apiClientProvider apiClientProvider
}

// NewSonarProjectReconciler returns a new SonarProjectReconciler instance.
func NewSonarProjectReconciler(
	k8sClient client.Client,
	scheme *runtime.Scheme,
	apiClientProvider apiClientProvider,
) *SonarProjectReconciler {
	return &SonarProjectReconciler{client: k8sClient, scheme: scheme, apiClientProvider: apiClientProvider}
}

// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarprojects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarprojects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarprojects/finalizers,verbs=update
// +kubebuilder:rbac:groups="",namespace=placeholder,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SonarProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling SonarProject")

	project := &sonarApi.SonarProject{}

	err := r.client.Get(ctx, req.NamespacedName, project)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	sonarApiClient, err := r.apiClientProvider.GetSonarApiClientFromSonarRef(ctx, req.Namespace, project)
	if err != nil {
		log.Error(err, "An error has occurred while getting sonar api client")

		return ctrl.Result{
			RequeueAfter: errorRequeueTime,
		}, nil
	}

	if project.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(project, helper.FinalizerName) {
			if err = chain.NewRemoveProject(sonarApiClient).ServeRequest(ctx, project); err != nil {
				log.Error(err, "An error has occurred while deleting SonarProject")

				return ctrl.Result{
					RequeueAfter: errorRequeueTime,
				}, nil
			}

			controllerutil.RemoveFinalizer(project, helper.FinalizerName)

			if err = r.client.Update(ctx, project); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if controllerutil.AddFinalizer(project, helper.FinalizerName) {
		err = r.client.Update(ctx, project)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	oldStatus := project.Status.DeepCopy()

	if err = chain.MakeChain(sonarApiClient, r.client).ServeRequest(ctx, project); err != nil {
		log.Error(err, "An error has occurred while handling SonarProject")

		project.Status.Value = "error"
		project.Status.Error = err.Error()

		if err = r.updateSonarProjectStatus(ctx, project, oldStatus); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{
			RequeueAfter: errorRequeueTime,
		}, nil
	}

	project.Status.Value = common.StatusCreated
	project.Status.Error = ""
	project.Status.ProjectKey = project.Spec.Key

	if err = r.updateSonarProjectStatus(ctx, project, oldStatus); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SonarProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sonarApi.SonarProject{}).
		Complete(r)
}

func (r *SonarProjectReconciler) updateSonarProjectStatus(ctx context.Context, sonarProject *sonarApi.SonarProject, oldStatus *sonarApi.SonarProjectStatus) error {
	if equality.Semantic.DeepEqual(&sonarProject.Status, oldStatus) {
		return nil
	}

	if err := r.client.Status().Update(ctx, sonarProject); err != nil {
		return fmt.Errorf("failed to update SonarProject status: %w", err)
	}

	return nil
}
