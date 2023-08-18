package user

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/epam/edp-sonar-operator/api/common"
	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
)

var _ = Describe("SonarUser controller", func() {
	sonarUserCRName := "sonar-user"
	It("Should create SonarUser object with user secret", func() {
		sonarUserSecretCRName := "sonar-user-secret"
		By("By creating a secret")
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sonarUserSecretCRName,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"password": []byte("password"),
			},
		}
		Expect(k8sClient.Create(ctx, secret)).Should(Succeed())
		By("By creating a new SonarUser object")
		newSonarUser := &sonarApi.SonarUser{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sonarUserCRName,
				Namespace: namespace,
			},
			Spec: sonarApi.SonarUserSpec{
				Email:       "test@mail.com",
				Login:       "test-user",
				Name:        "test user",
				Groups:      []string{"sonar-administrators"},
				Permissions: []string{"scan"},
				Secret:      sonarUserSecretCRName,
				SonarRef: common.SonarRef{
					Name: sonarName,
				},
			},
		}
		Expect(k8sClient.Create(ctx, newSonarUser)).Should(Succeed())
		Eventually(func() bool {
			createdSonarUser := &sonarApi.SonarUser{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarUserCRName, Namespace: namespace}, createdSonarUser)
			if err != nil {
				return false
			}

			return createdSonarUser.Status.Value == "created" && createdSonarUser.Status.Error == ""
		}, timeout, interval).Should(BeTrue())
	})
	It("Should delete SonarUser object", func() {
		By("Getting SonarUser object")
		createdSonarUser := &sonarApi.SonarUser{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: sonarUserCRName, Namespace: namespace}, createdSonarUser)).
			Should(Succeed())

		By("Deleting SonarUser object")
		Expect(k8sClient.Delete(ctx, createdSonarUser)).Should(Succeed())
		Eventually(func() bool {
			createdSonarUser := &sonarApi.SonarUser{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: sonarUserCRName, Namespace: namespace}, createdSonarUser)
			return k8sErrors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue())
	})
})
