package qualityprofile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"
	"time"

	"github.com/epam/edp-sonar-operator/internal/controller/sonar"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	sonarclient "github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

const (
	sonarName = "test-sonar"
	namespace = "test-sonar-qualityprofile"

	timeout  = time.Second * 10
	interval = time.Millisecond * 250
)

var (
	cfg           *rest.Config
	k8sClient     client.Client
	testEnv       *envtest.Environment
	ctx           context.Context
	cancel        context.CancelFunc
	sonarUrl      string
	sonarUser     string
	sonarPassword string
)

func TestSonarQualityProfile(t *testing.T) {
	RegisterFailHandler(Fail)

	if os.Getenv("TEST_SONAR_URL") == "" {
		t.Skip("TEST_SONAR_URL is not set")
	}

	RunSpecs(t, "SonarQualityProfile Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.Background())
	sonarUrl = os.Getenv("TEST_SONAR_URL")
	sonarUser = os.Getenv("TEST_SONAR_USER")
	sonarPassword = os.Getenv("TEST_SONAR_PASSWORD")

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
		BinaryAssetsDirectory: filepath.Join("..", "..", "..", "bin", "k8s",
			fmt.Sprintf("1.31.0-%s-%s", goruntime.GOOS, goruntime.GOARCH)),
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	scheme := runtime.NewScheme()
	Expect(sonarApi.AddToScheme(scheme)).NotTo(HaveOccurred())
	Expect(corev1.AddToScheme(scheme)).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	Expect(err).ToNot(HaveOccurred())

	err = sonar.NewReconcileSonar(
		k8sManager.GetClient(),
		k8sManager.GetScheme(),
		sonarclient.NewApiClientProvider(k8sManager.GetClient()),
	).
		SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = NewSonarQualityProfileReconciler(
		k8sManager.GetClient(),
		k8sManager.GetScheme(),
		sonarclient.NewApiClientProvider(k8sManager.GetClient()),
	).
		SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

	By("By creating namespace")
	Expect(k8sClient.Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	})).Should(Succeed())
	By("By creating a secret")
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sonar-auth-secret",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"user":     []byte(sonarUser),
			"password": []byte(sonarPassword),
		},
	}
	Expect(k8sClient.Create(ctx, secret)).Should(Succeed())
	By("By creating a new Sonar object")
	newSonar := &sonarApi.Sonar{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sonarName,
			Namespace: namespace,
		},
		Spec: sonarApi.SonarSpec{
			Url:    sonarUrl,
			Secret: secret.Name,
		},
	}
	Expect(k8sClient.Create(ctx, newSonar)).Should(Succeed())
	Eventually(func() bool {
		createdSonar := &sonarApi.Sonar{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarName, Namespace: namespace}, createdSonar)
		if err != nil {
			return false
		}
		return createdSonar.Status.Connected

	}, timeout, interval).Should(BeTrue())
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
