package sonar

import (
	"time"

	"github.com/epam/edp-sonar-operator/api/common"

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
				"user":      []byte(sonarUser),
				"password":  []byte(sonarPassword),
				"smtp-pass": []byte("smtp-password"),
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
					{
						Key: "email.smtp_password.secured",
						ValueRef: &common.SourceRef{
							SecretKeyRef: &common.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: secret.Name,
								},
								Key: "smtp-pass",
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, newSonar)).Should(Succeed())
		Eventually(func(g Gomega) {
			createdSonar := &sonarApi.Sonar{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarName, Namespace: namespace}, createdSonar)
			g.Expect(err).ShouldNot(HaveOccurred())

			g.Expect(createdSonar.Status.Connected).Should(BeTrue(), "Sonar should be connected")
			g.Expect(createdSonar.Status.Error).Should(BeEmpty(), "Error should be empty")
			g.Expect(createdSonar.Status.Value).ShouldNot(BeEmpty(), "Value should not be empty")
			g.Expect(createdSonar.Status.ProcessedSettings).
				Should(
					Equal("email.smtp_password.secured,sonar.dbcleaner.hoursBeforeKeepingOnlyOneSnapshotByDay,sonar.global.exclusions,sonar.issue.ignore.block"),
					"Processed settings should be equal",
				)
		}).WithTimeout(timeout).WithPolling(interval).Should(Succeed())
	})
})
