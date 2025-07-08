package project

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/epam/edp-sonar-operator/api/common"
	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

var _ = Describe("SonarProject controller", func() {
	sonarProjectCRName := "sonar-project"
	It("Should create SonarProject object", func() {
		By("By creating a new SonarProject object")
		newSonarProject := &sonarApi.SonarProject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sonarProjectCRName,
				Namespace: namespace,
			},
			Spec: sonarApi.SonarProjectSpec{
				Key:        "test-project-key",
				Name:       "Test Project",
				Visibility: "private",
				SonarRef: common.SonarRef{
					Name: sonarName,
				},
			},
		}
		Expect(k8sClient.Create(ctx, newSonarProject)).Should(Succeed())
		Eventually(func() bool {
			createdSonarProject := &sonarApi.SonarProject{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarProjectCRName, Namespace: namespace}, createdSonarProject)
			if err != nil {
				return false
			}

			return createdSonarProject.Status.Value == common.StatusCreated && createdSonarProject.Status.Error == ""
		}, timeout, interval).Should(BeTrue())
	})
	It("Should delete SonarProject object", func() {
		By("Getting SonarProject object")
		createdSonarProject := &sonarApi.SonarProject{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: sonarProjectCRName, Namespace: namespace}, createdSonarProject)).
			Should(Succeed())

		By("Deleting SonarProject object")
		Expect(k8sClient.Delete(ctx, createdSonarProject)).Should(Succeed())
		Eventually(func() bool {
			createdSonarProject := &sonarApi.SonarProject{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarProjectCRName, Namespace: namespace}, createdSonarProject)
			return k8sErrors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue())
	})
})
