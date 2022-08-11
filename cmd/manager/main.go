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
	//+kubebuilder:scaffold:imports
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	buildInfo "github.com/epam/edp-common/pkg/config"
	edpCompApi "github.com/epam/edp-component-operator/pkg/apis/v1/v1"
	jenkinsApi "github.com/epam/edp-jenkins-operator/v2/pkg/apis/v2/v1"
	keycloakApi "github.com/epam/edp-keycloak-operator/pkg/apis/v1/v1"

	sonarApiv1 "github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1"
	sonarApiv1alpha1 "github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epam/edp-sonar-operator/v2/pkg/controller/group"
	"github.com/epam/edp-sonar-operator/v2/pkg/controller/permission_template"
	"github.com/epam/edp-sonar-operator/v2/pkg/controller/sonar"
	"github.com/epam/edp-sonar-operator/v2/pkg/helper"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	sonarOperatorLock = "edp-sonar-operator-lock"
	DefaultPort       = 9443
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(sonarApiv1.AddToScheme(scheme))

	utilruntime.Must(sonarApiv1alpha1.AddToScheme(scheme))

	utilruntime.Must(edpCompApi.AddToScheme(scheme))

	utilruntime.Must(jenkinsApi.AddToScheme(scheme))

	utilruntime.Must(keycloakApi.AddToScheme(scheme))
}

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		probeAddr            string
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", helper.RunningInCluster(),
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	mode, err := helper.GetDebugMode()
	if err != nil {
		setupLog.Error(err, "unable to get debug mode value")
		os.Exit(1)
	}

	opts := zap.Options{
		Development: mode,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

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
		setupLog.Error(err, "unable to get watch namespace")
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
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	ctrlLog := ctrl.Log.WithName("controllers")

	sonarCtrl, err := sonar.NewReconcileSonar(mgr.GetClient(), mgr.GetScheme(), ctrlLog,
		helper.GetPlatformTypeEnv())
	if err != nil {
		setupLog.Error(err, "failed to create sonar reconcile")
		os.Exit(1)
	}

	if err = sonarCtrl.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "failed to setup sonar reconcile")
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

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
