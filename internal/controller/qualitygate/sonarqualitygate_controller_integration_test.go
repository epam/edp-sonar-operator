package qualitygate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/epam/edp-sonar-operator/api/common"
	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
)

var _ = Describe("SonarQualityGate controller", func() {
	sonarQualityGateCRName := "sonar-gate"
	It("Should create SonarQualityGate object", func() {
		By("By creating a new SonarQualityGate object")
		newSonarQualityGate := &sonarApi.SonarQualityGate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sonarQualityGateCRName,
				Namespace: namespace,
			},
			Spec: sonarApi.SonarQualityGateSpec{
				Name:    "test-gate",
				Default: true,
				Conditions: map[string]sonarApi.Condition{
					"new_security_hotspots_reviewed": {
						Error: "75",
						Op:    "LT",
					},
					"new_duplicated_lines_density": {
						Error: "10",
						Op:    "GT",
					},
				},
				SonarRef: common.SonarRef{
					Name: sonarName,
				},
			},
		}
		Expect(k8sClient.Create(ctx, newSonarQualityGate)).Should(Succeed())
		Eventually(func() bool {
			createdSonarQualityGate := &sonarApi.SonarQualityGate{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarQualityGateCRName, Namespace: namespace}, createdSonarQualityGate)
			if err != nil {
				return false
			}

			return createdSonarQualityGate.Status.Value == common.StatusCreated && createdSonarQualityGate.Status.Error == ""
		}, timeout, interval).Should(BeTrue())
	})
	It("Should update SonarQualityGate object", func() {
		By("Getting SonarQualityGate object")
		createdSonarQualityGate := &sonarApi.SonarQualityGate{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarQualityGateCRName, Namespace: namespace}, createdSonarQualityGate)
		Expect(err).Should(Succeed())

		By("Updating SonarQualityGate object")
		createdSonarQualityGate.Spec.Conditions["new_security_rating"] = sonarApi.Condition{
			Error: "1",
			Op:    "GT",
		}
		delete(createdSonarQualityGate.Spec.Conditions, "new_duplicated_lines_density")

		Expect(k8sClient.Update(ctx, createdSonarQualityGate)).Should(Succeed())
	})
	It("Should delete SonarQualityGate object", func() {
		By("Creating not default SonarQualityGate object")
		gate := &sonarApi.SonarQualityGate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "not-default-gate",
				Namespace: namespace,
			},
			Spec: sonarApi.SonarQualityGateSpec{
				Name:    "not-default-gate",
				Default: false,
			},
		}
		Expect(k8sClient.Create(ctx, gate)).Should(Succeed())

		By("Getting SonarQualityGate object")
		gateToDelete := &sonarApi.SonarQualityGate{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: gate.Name, Namespace: gate.Namespace}, gateToDelete)
		Expect(err).Should(Succeed())

		By("Deleting SonarQualityGate object")
		Expect(k8sClient.Delete(ctx, gateToDelete)).Should(Succeed())
		Eventually(func() bool {
			deletedGate := &sonarApi.SonarQualityGate{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: gate.Name, Namespace: gate.Namespace}, deletedGate)
			return k8sErrors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue())
	})
})
