package sonar

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
)

var _ = Describe("Sonar controller", func() {
	const (
		sonarName = "test-sonar"
		namespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	It("Should create Sonar object with secret auth", func() {
		By("By creating a secret")
		secret := &v1.Secret{
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
				Settings: []sonarApi.SonarSetting{
					{
						Key:   "sonar.dbcleaner.hoursBeforeKeepingOnlyOneSnapshotByDay",
						Value: "20",
					},
					{
						Key: "sonar.issue.ignore.block",
						FieldValues: map[string]string{
							"beginBlockRegexp": ".*",
							"endBlockRegexp":   ".*",
						},
					},
					{
						Key:    "sonar.global.exclusions",
						Values: []string{"**/*.js", "**/*.ts", "**/*.tsx", "**/*.jsx"},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, newSonar)).Should(Succeed())
		Eventually(func() bool {
			createdSonar := &sonarApi.Sonar{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarName, Namespace: namespace}, createdSonar)
			if err != nil {
				return false
			}

			processedSettings := "sonar.dbcleaner.hoursBeforeKeepingOnlyOneSnapshotByDay,sonar.global.exclusions,sonar.issue.ignore.block"

			return createdSonar.Status.Connected &&
				createdSonar.Status.Error == "" &&
				createdSonar.Status.Value != "" &&
				createdSonar.Status.ProcessedSettings == processedSettings

		}, timeout, interval).Should(BeTrue())
	})
})
