package qualityprofile

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/epam/edp-sonar-operator/api/common"
	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
)

var _ = Describe("SonarQualityProfile controller", func() {
	sonarQualityProfileCRName := "sonar-profile"
	It("Should create SonarQualityProfile object", func() {
		By("By creating a new SonarQualityProfile object")
		newSonarQualityProfile := &sonarApi.SonarQualityProfile{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sonarQualityProfileCRName,
				Namespace: namespace,
			},
			Spec: sonarApi.SonarQualityProfileSpec{
				Name:     "test-profile",
				Default:  true,
				Language: "go",
				Rules: map[string]sonarApi.Rule{
					"go:S126": {
						Severity: "CRITICAL",
					},
					"go:S1151": {
						Severity: "MAJOR",
						Params:   `max="6"`,
					},
				},
				SonarRef: common.SonarRef{
					Name: sonarName,
				},
			},
		}
		Expect(k8sClient.Create(ctx, newSonarQualityProfile)).Should(Succeed())
		Eventually(func() bool {
			createdSonarQualityprofile := &sonarApi.SonarQualityProfile{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarQualityProfileCRName, Namespace: namespace}, createdSonarQualityprofile)
			if err != nil {
				return false
			}

			return createdSonarQualityprofile.Status.Value == "created" && createdSonarQualityprofile.Status.Error == ""
		}, timeout, interval).Should(BeTrue())
	})
	It("Should update SonarQualityProfile object", func() {
		By("Getting SonarQualityProfile object")
		createdSonarQualityProfile := &sonarApi.SonarQualityProfile{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarQualityProfileCRName, Namespace: namespace}, createdSonarQualityProfile)
		Expect(err).Should(Succeed())

		By("Updating SonarQualityProfile object")
		delete(createdSonarQualityProfile.Spec.Rules, "go:S1151")

		Expect(k8sClient.Update(ctx, createdSonarQualityProfile)).Should(Succeed())
	})
	It("Should delete SonarQualityProfile object", func() {
		By("Creating not default SonarQualityProfile object")
		profile := &sonarApi.SonarQualityProfile{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "not-default-profile",
				Namespace: namespace,
			},
			Spec: sonarApi.SonarQualityProfileSpec{
				Name:    "not-default-profile",
				Default: false,
			},
		}
		Expect(k8sClient.Create(ctx, profile)).Should(Succeed())

		By("Getting SonarQualityProfile object")
		profileToDelete := &sonarApi.SonarQualityProfile{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: profile.Name, Namespace: profile.Namespace}, profileToDelete)
		Expect(err).Should(Succeed())

		By("Deleting SonarQualityProfile object")
		Expect(k8sClient.Delete(ctx, profileToDelete)).Should(Succeed())
		Eventually(func() bool {
			deletedProfile := &sonarApi.SonarQualityProfile{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: profile.Name, Namespace: profile.Namespace}, deletedProfile)
			return k8sErrors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue())
	})
})
