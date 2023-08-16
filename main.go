package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	buildInfo "github.com/epam/edp-common/pkg/config"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/controllers/group"
	"github.com/epam/edp-sonar-operator/controllers/permission_template"
	"github.com/epam/edp-sonar-operator/controllers/sonar"
	sonaruser "github.com/epam/edp-sonar-operator/controllers/user"
	sonarclient "github.com/epam/edp-sonar-operator/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/pkg/helper"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	sonarOperatorLock = "edp-sonar-operator-lock"
	DefaultPort       = 9443
)

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		probeAddr            string
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", helper.RunningInCluster(),
		"Enable leader election for controllers manager. "+
			"Enabling this will ensure there is only one active controllers manager.")

	mode, err := helper.GetDebugMode()
	if err != nil {
		setupLog.Error(err, "failed to get debug mode value")
		os.Exit(1)
	}

	opts := zap.Options{
		Development: mode,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(sonarApi.AddToScheme(scheme))
	utilruntime.Must(sonarApi.AddToScheme(scheme))

	v := buildInfo.Get()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("Starting the Sonar Operator",
		"version", v.Version,
		"git-commit", v.GitCommit,
		"git-tag", v.GitTag,
		"build-date", v.BuildDate,
		"go-version", v.Go,
		"go-client", v.KubectlVersion,
		"platform", v.Platform,
	)

	ns, err := helper.GetWatchNamespace()
	if err != nil {
		setupLog.Error(err, "failed to get watch namespace")
		os.Exit(1)
	}

	cfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		HealthProbeBindAddress: probeAddr,
		Port:                   DefaultPort,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       sonarOperatorLock,
		MapperProvider: func(c *rest.Config) (meta.RESTMapper, error) {
			return apiutil.NewDynamicRESTMapper(cfg)
		},
		Namespace: ns,
	})
	if err != nil {
		setupLog.Error(err, "failed to start manager")
		os.Exit(1)
	}

	ctrlLog := ctrl.Log.WithName("controllers")

	apiClientProvider := sonarclient.NewApiClientProvider(mgr.GetClient())

	if err = sonar.NewReconcileSonar(
		mgr.GetClient(),
		mgr.GetScheme(),
		apiClientProvider,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "failed to setup sonar reconcile")
		os.Exit(1)
	}

	if err = sonaruser.NewSonarUserReconciler(
		mgr.GetClient(),
		mgr.GetScheme(),
		apiClientProvider,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "failed to setup sonar user reconcile")
		os.Exit(1)
	}

	permTplCtrl, err := permission_template.NewReconcile(mgr.GetClient(), mgr.GetScheme(), ctrlLog,
		helper.GetPlatformTypeEnv())
	if err != nil {
		setupLog.Error(err, "failed to create permission template reconcile")
		os.Exit(1)
	}

	if err = permTplCtrl.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "failed to setup permission template reconcile")
		os.Exit(1)
	}

	groupCtrl, err := group.NewReconcile(mgr.GetClient(), mgr.GetScheme(), ctrlLog, helper.GetPlatformTypeEnv())
	if err != nil {
		setupLog.Error(err, "failed to create sonar group reconcile")
		os.Exit(1)
	}

	if err = groupCtrl.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "failed to setup sonar group reconcile")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "failed to set up health check")
		os.Exit(1)
	}

	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "failed to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
