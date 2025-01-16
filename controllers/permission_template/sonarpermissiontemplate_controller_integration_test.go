package permission_template

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/epam/edp-sonar-operator/api/common"
	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
)

var _ = Describe("PermissionTemplate controller", func() {
	permissionTemplateCRName := "sonar-permission-template"
	It("Should create PermissionTemplate object", func() {
		By("By creating a new PermissionTemplate object")
		newPermissionTemplate := &sonarApi.SonarPermissionTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      permissionTemplateCRName,
				Namespace: namespace,
			},
			Spec: sonarApi.SonarPermissionTemplateSpec{
				Name:              "test permission template",
				Description:       "test description",
				ProjectKeyPattern: ".*.finance",
				Default:           true,
				GroupsPermissions: map[string][]string{
					"sonar-users": {"scan", "codeviewer"},
				},
				SonarRef: common.SonarRef{
					Name: sonarName,
				},
			},
		}
		Expect(k8sClient.Create(ctx, newPermissionTemplate)).Should(Succeed())
		Eventually(func(g Gomega) {
			createdPermissionTemplate := &sonarApi.SonarPermissionTemplate{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: permissionTemplateCRName, Namespace: namespace}, createdPermissionTemplate)

			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(createdPermissionTemplate.Status.Error).Should(Equal(""), "Error should be empty")
			g.Expect(createdPermissionTemplate.Status.Value).Should(Equal(common.StatusCreated), "Status should be created")
		}).WithTimeout(timeout).WithPolling(interval).Should(Succeed())
	})
	It("Should delete PermissionTemplate object", func() {
		By("By creating not default PermissionTemplate object")
		newPermissionTemplate := &sonarApi.SonarPermissionTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "permission-template-to-delete",
				Namespace: namespace,
			},
			Spec: sonarApi.SonarPermissionTemplateSpec{
				Name: "permission-template-to-delete",
				SonarRef: common.SonarRef{
					Name: sonarName,
				},
			},
		}
		Expect(k8sClient.Create(ctx, newPermissionTemplate)).Should(Succeed())
		Eventually(func() bool {
			createdPermissionTemplate := &sonarApi.SonarPermissionTemplate{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: newPermissionTemplate.Name, Namespace: namespace}, createdPermissionTemplate)
			if err != nil {
				return false
			}

			return createdPermissionTemplate.Status.Value == common.StatusCreated && createdPermissionTemplate.Status.Error == ""
		}, timeout, interval).Should(BeTrue())
		By("Getting PermissionTemplate object")
		permissionTemplateToDelete := &sonarApi.SonarPermissionTemplate{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: newPermissionTemplate.Name, Namespace: namespace}, permissionTemplateToDelete)).
			Should(Succeed())
		By("Deleting PermissionTemplate object")
		Expect(k8sClient.Delete(ctx, permissionTemplateToDelete)).Should(Succeed())
		Eventually(func() bool {
			deletedPermissionTemplate := &sonarApi.SonarPermissionTemplate{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: newPermissionTemplate.Name, Namespace: namespace}, deletedPermissionTemplate)
			return k8sErrors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue())
	})
})
