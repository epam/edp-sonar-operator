package permission_template

import (
	"context"
	"fmt"
	"time"

	"github.com/epam/edp-sonar-operator/internal/controller/permission_template/chain"

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

type SonarPermissionTemplateReconciler struct {
	client            client.Client
	scheme            *runtime.Scheme
	apiClientProvider apiClientProvider
}

func NewSonarPermissionTemplateReconciler(
	client client.Client,
	scheme *runtime.Scheme,
	apiClientProvider apiClientProvider,
) *SonarPermissionTemplateReconciler {
	return &SonarPermissionTemplateReconciler{
		client:            client,
		scheme:            scheme,
		apiClientProvider: apiClientProvider,
	}
}

func (r *SonarPermissionTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sonarApi.SonarPermissionTemplate{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarpermissiontemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarpermissiontemplates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=edp.epam.com,namespace=placeholder,resources=sonarpermissiontemplates/finalizers,verbs=update
// +kubebuilder:rbac:groups="",namespace=placeholder,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SonarPermissionTemplateReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling SonarPermissionTemplate")

	template := &sonarApi.SonarPermissionTemplate{}

	err := r.client.Get(ctx, req.NamespacedName, template)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	sonarApiClient, err := r.apiClientProvider.GetSonarApiClientFromSonarRef(ctx, req.Namespace, template)
	if err != nil {
		log.Error(err, "An error has occurred while getting sonar api client")

		return ctrl.Result{
			RequeueAfter: errorRequeueTime,
		}, nil
	}

	if template.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(template, sonarOperatorFinalizer) {
			if err = chain.NewRemovePermissionTemplate(sonarApiClient).ServeRequest(ctx, template); err != nil {
				log.Error(err, "An error has occurred while deleting SonarPermissionTemplate")

				return ctrl.Result{
					RequeueAfter: errorRequeueTime,
				}, nil
			}

			controllerutil.RemoveFinalizer(template, sonarOperatorFinalizer)

			if err = r.client.Update(ctx, template); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if controllerutil.AddFinalizer(template, sonarOperatorFinalizer) {
		err = r.client.Update(ctx, template)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	oldStatus := template.Status

	if err = chain.MakeChain(sonarApiClient).ServeRequest(ctx, template); err != nil {
		log.Error(err, "An error has occurred while handling SonarPermissionTemplate")

		template.Status.Value = "error"
		template.Status.Error = err.Error()

		if err = r.updateSonarPermissionTemplateStatus(ctx, template, oldStatus); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{
			RequeueAfter: errorRequeueTime,
		}, nil
	}

	template.Status.Value = common.StatusCreated
	template.Status.Error = ""

	if err = r.updateSonarPermissionTemplateStatus(ctx, template, oldStatus); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SonarPermissionTemplateReconciler) updateSonarPermissionTemplateStatus(
	ctx context.Context,
	template *sonarApi.SonarPermissionTemplate,
	oldStatus sonarApi.SonarPermissionTemplateStatus,
) error {
	if template.Status == oldStatus {
		return nil
	}

	if err := r.client.Status().Update(ctx, template); err != nil {
		return fmt.Errorf("failed to update SonarPermissionTemplate status: %w", err)
	}

	return nil
}
