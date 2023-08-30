package group

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/epam/edp-sonar-operator/api/common"
	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
)

var _ = Describe("SonarGroup controller", func() {
	sonarGroupCRName := "sonar-group"
	It("Should create SonarGroup object", func() {
		By("By creating a new SonarGroup object")
		newSonarGroup := &sonarApi.SonarGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sonarGroupCRName,
				Namespace: namespace,
			},
			Spec: sonarApi.SonarGroupSpec{
				Name:        "test group",
				Description: "test description",
				Permissions: []string{"scan"},
				SonarRef: common.SonarRef{
					Name: sonarName,
				},
			},
		}
		Expect(k8sClient.Create(ctx, newSonarGroup)).Should(Succeed())
		Eventually(func() bool {
			createdSonarGroup := &sonarApi.SonarGroup{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarGroupCRName, Namespace: namespace}, createdSonarGroup)
			if err != nil {
				return false
			}

			return createdSonarGroup.Status.Value == "created" && createdSonarGroup.Status.Error == ""
		}, timeout, interval).Should(BeTrue())
	})
	It("Should delete SonarGroup object", func() {
		By("Getting SonarGroup object")
		createdSonarGroup := &sonarApi.SonarGroup{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: sonarGroupCRName, Namespace: namespace}, createdSonarGroup)).
			Should(Succeed())

		By("Deleting SonarGroup object")
		Expect(k8sClient.Delete(ctx, createdSonarGroup)).Should(Succeed())
		Eventually(func() bool {
			createdSonarGroup := &sonarApi.SonarGroup{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarGroupCRName, Namespace: namespace}, createdSonarGroup)
			return k8sErrors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue())
	})
})
