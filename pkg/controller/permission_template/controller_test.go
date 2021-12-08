package permission_template

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/epam/edp-common/pkg/mock"
	k8sMockClient "github.com/epam/edp-common/pkg/mock/controller-runtime/client"
	"github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1alpha1"
	sonarClient "github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/sonar"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestNewReconcile_NotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	rq := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "ns", Name: "name"}}
	fakeCl := fake.NewClientBuilder().WithScheme(scheme).Build()
	l := mock.Logger{}
	rec, err := NewReconcile(fakeCl, scheme, &l, "kubernetes")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := rec.Reconcile(context.Background(), rq); err != nil {
		t.Fatal(err)
	}

	if _, ok := l.InfoMessages["instance not found"]; !ok {
		t.Fatal("no warning message logged")
	}

	k8sMock := k8sMockClient.Client{}
	k8sMock.On("Get", rq.NamespacedName, &v1alpha1.SonarPermissionTemplate{}).
		Return(errors.New("get fatal"))
	rec.client = &k8sMock

	_, err = rec.Reconcile(context.Background(), rq)
	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to get sonar permission tpl from k8s: get fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestNewReconcile(t *testing.T) {
	ptpl := v1alpha1.SonarPermissionTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "tpl1",
			Namespace:         "ns",
			DeletionTimestamp: &metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.SonarPermissionTemplateSpec{
			Name:              "tpl1",
			ProjectKeyPattern: ".+",
			SonarOwner:        "sonar",
			Description:       "desc",
			GroupPermissions: []v1alpha1.GroupPermission{
				{
					GroupName:   "gr1",
					Permissions: []string{"admin", "user"},
				},
			},
		},
	}
	sn := v1alpha1.Sonar{
		ObjectMeta: metav1.ObjectMeta{Name: "sonar", Namespace: ptpl.Namespace},
		TypeMeta:   metav1.TypeMeta{Kind: "Sonar", APIVersion: "v2.edp.epam.com/v1alpha1"},
		Spec:       v1alpha1.SonarSpec{BasePath: "path"},
	}

	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	rq := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: ptpl.Namespace, Name: ptpl.Name}}
	fakeCl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&ptpl, &sn).Build()
	l := mock.Logger{}
	rec, err := NewReconcile(fakeCl, scheme, &l, "kubernetes")
	if err != nil {
		t.Fatal(err)
	}

	serviceMock := sonar.ServiceMock{}
	rec.service = &serviceMock
	clientMock := sonar.ClientMock{}

	serviceMock.On("ClientForChild").Return(&clientMock, nil)
	clientMock.On("GetPermissionTemplate", ptpl.Spec.Name).Return(nil,
		sonarClient.ErrNotFound("not found")).Once()
	clientMock.On("CreatePermissionTemplate", specToClientTemplate(&ptpl.Spec)).Return(nil)

	tplGroup := sonarClient.PermissionTemplateGroup{GroupName: "baz", Permissions: []string{"scan"}}
	clientMock.On("GetPermissionTemplateGroups", "").
		Return([]sonarClient.PermissionTemplateGroup{tplGroup}, nil)
	clientMock.On("RemoveGroupFromPermissionTemplate", &tplGroup).Return(nil)
	clientMock.On("AddGroupToPermissionTemplate",
		&sonarClient.PermissionTemplateGroup{
			GroupName:   ptpl.Spec.GroupPermissions[0].GroupName,
			Permissions: ptpl.Spec.GroupPermissions[0].Permissions,
		}).Return(nil)
	serviceMock.On("DeleteResource").Return(true, nil)
	clientMock.On("DeletePermissionTemplate", "").Return(nil)
	if _, err := rec.Reconcile(context.Background(), rq); err != nil {
		t.Fatal(err)
	}

	if err := l.LastError(); err != nil {
		t.Fatalf("%+v", err)
	}

	ptpl.Status.ID = "id11"
	clientMock.On("GetPermissionTemplate", ptpl.Spec.Name).
		Return(specToClientTemplate(&ptpl.Spec), nil).Once()
	tpl := specToClientTemplate(&ptpl.Spec)
	tpl.ID = ptpl.Status.ID
	clientMock.On("UpdatePermissionTemplate", tpl).Return(nil)
	clientMock.On("GetPermissionTemplateGroups", tpl.ID).
		Return(nil, errors.New("get perm groups fatal"))
	rec.client = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&ptpl, &sn).Build()

	if _, err := rec.Reconcile(context.Background(), rq); err != nil {
		t.Fatal(err)
	}

	err = l.LastError()
	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to sync permission template groups: unable to get permission template groups: get perm groups fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSpecIsUpdated(t *testing.T) {
	if isSpecUpdated(event.UpdateEvent{ObjectNew: &v1alpha1.SonarPermissionTemplate{},
		ObjectOld: &v1alpha1.SonarPermissionTemplate{}}) {
		t.Fatal("spec is updated")
	}
}
