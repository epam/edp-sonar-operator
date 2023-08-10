package sonar

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/controllers/sonar/chain"
	sonarclient "github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

const (
	defaultRequeueTime = time.Second * 30
	successRequeueTime = time.Minute * 10
)

func NewReconcileSonar(
	client client.Client,
	scheme *runtime.Scheme,
) *ReconcileSonar {
	return &ReconcileSonar{
		client: client,
		scheme: scheme,
	}
}

type ReconcileSonar struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileSonar) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sonarApi.Sonar{}).
		Complete(r)
}

//+kubebuilder:rbac:groups=v1.edp.epam.com,namespace=placeholder,resources=sonars,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=v1.edp.epam.com,namespace=placeholder,resources=sonars/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=v1.edp.epam.com,namespace=placeholder,resources=sonars/finalizers,verbs=update
//+kubebuilder:rbac:groups="",namespace=placeholder,resources=secrets,verbs=get

func (r *ReconcileSonar) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling Sonar")

	sonar := &sonarApi.Sonar{}
	if err := r.client.Get(ctx, request.NamespacedName, sonar); err != nil {
		if k8sErrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	oldStatus := sonar.Status

	sonarApiClient, err := r.getSonarApiClient(ctx, sonar)
	if err != nil {
		sonar.Status.Error = err.Error()
		sonar.Status.Connected = false

		if statusErr := r.updateSonarStatus(ctx, sonar, oldStatus); statusErr != nil {
			return reconcile.Result{}, statusErr
		}

		return reconcile.Result{RequeueAfter: defaultRequeueTime}, err
	}

	if err = chain.MakeChain(sonarApiClient).ServeRequest(ctx, sonar); err != nil {
		sonar.Status.Error = err.Error()

		if statusErr := r.updateSonarStatus(ctx, sonar, oldStatus); statusErr != nil {
			return reconcile.Result{}, statusErr
		}

		return reconcile.Result{RequeueAfter: defaultRequeueTime}, err
	}

	sonar.Status.Connected = true
	sonar.Status.Error = ""

	if err = r.updateSonarStatus(ctx, sonar, oldStatus); err != nil {
		return reconcile.Result{}, err
	}

	log.Info("Reconciling Sonar is finished")

	return reconcile.Result{
		RequeueAfter: successRequeueTime,
	}, nil
}

func (r *ReconcileSonar) getSonarApiClient(ctx context.Context, sonar *sonarApi.Sonar) (sonarclient.ClientInterface, error) {
	secret := corev1.Secret{}
	if err := r.client.Get(ctx, types.NamespacedName{
		Name:      sonar.Spec.Secret,
		Namespace: sonar.Namespace,
	}, &secret); err != nil {
		return nil, fmt.Errorf("failed to get sonar secret: %w", err)
	}

	if secret.Data["user"] == nil {
		return nil, fmt.Errorf("sonar secret doesn't contain user")
	}

	password := ""
	if secret.Data["password"] != nil {
		password = string(secret.Data["password"])
	}

	return sonarclient.NewClient(sonar.Spec.Url, string(secret.Data["user"]), password), nil
}

func (r *ReconcileSonar) updateSonarStatus(ctx context.Context, sonar *sonarApi.Sonar, oldStatus sonarApi.SonarStatus) error {
	if sonar.Status == oldStatus {
		return nil
	}

	if err := r.client.Status().Update(ctx, sonar); err != nil {
		return fmt.Errorf("failed to update sonar oldStatus: %w", err)
	}

	return nil
}
