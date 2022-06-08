package permission_template

import (
	"context"
	"testing"
	"time"

	"github.com/epam/edp-common/pkg/mock"
	k8sMockClient "github.com/epam/edp-common/pkg/mock/controller-runtime/client"
	"github.com/pkg/errors"
	tMock "github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cMock "github.com/epam/edp-sonar-operator/v2/mocks/client"
	sMock "github.com/epam/edp-sonar-operator/v2/mocks/service"
	sonarApi "github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1"
	sonarClient "github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform"
)

func TestNewReconcile_NotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := sonarApi.AddToScheme(scheme); err != nil {
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
	k8sMock.On("Get", rq.NamespacedName, &sonarApi.SonarPermissionTemplate{}).
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

const (
	objectMetaName      = "tpl1"
	objectMetaNamespace = "ns"
)

func ObjectMeta() metav1.ObjectMeta {
	now := time.Now().UTC()
	return metav1.ObjectMeta{
		Name:              objectMetaName,
		Namespace:         objectMetaNamespace,
		DeletionTimestamp: &metav1.Time{Time: now},
	}
}

func TestNewReconcile(t *testing.T) {
	ctx := context.Background()
	sampleTemplate := sonarApi.SonarPermissionTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SonarPermissionTemplate",
			APIVersion: "v2.edp.epam.com/v1",
		},
		ObjectMeta: ObjectMeta(),
		Spec: sonarApi.SonarPermissionTemplateSpec{
			Name:              objectMetaName,
			ProjectKeyPattern: ".+",
			SonarOwner:        "sonar",
			Description:       "desc",
			GroupPermissions: []sonarApi.GroupPermission{
				{
					GroupName:   "gr1",
					Permissions: []string{"admin", "user"},
				},
			},
		},
		Status: sonarApi.SonarPermissionTemplateStatus{
			Value:        "",
			FailureCount: 0,
			ID:           "",
		},
	}
	sn := sonarApi.Sonar{
		ObjectMeta: metav1.ObjectMeta{Name: "sonar", Namespace: objectMetaNamespace},
		TypeMeta:   metav1.TypeMeta{Kind: "Sonar", APIVersion: "v2.edp.epam.com/v1"},
		Spec:       sonarApi.SonarSpec{BasePath: "path"},
	}

	scheme := runtime.NewScheme()
	if err := sonarApi.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	permissionTemplate1 := sampleTemplate.DeepCopy()
	rq := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: objectMetaNamespace, Name: objectMetaName}}
	fakeCl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(permissionTemplate1, &sn).Build()
	l := mock.Logger{}
	rec, err := NewReconcile(fakeCl, scheme, &l, platform.Kubernetes)
	if err != nil {
		t.Fatal(err)
	}

	serviceMock := sMock.ServiceInterface{}
	rec.service = &serviceMock
	clientMock := &cMock.ClientInterface{}

	serviceMock.On("ClientForChild", ctx, tMock.AnythingOfType("*v1.SonarPermissionTemplate")).Return(clientMock, nil)
	serviceMock.On("K8sClient").Return(fakeCl)
	clientMock.On("GetPermissionTemplate", ctx, permissionTemplate1.Spec.Name).Return(nil,
		sonarClient.ErrNotFound("not found")).Once()
	permissionTemplateID := "uniq_tpl_id_1"
	clientMock.On("CreatePermissionTemplate", ctx, specToClientTemplateData(&permissionTemplate1.Spec)).
		Return(permissionTemplateID, nil)

	tplGroup := sonarClient.PermissionTemplateGroup{GroupName: "baz", Permissions: []string{"scan"}}
	clientMock.On("GetPermissionTemplateGroups", ctx, permissionTemplateID).
		Return([]sonarClient.PermissionTemplateGroup{tplGroup}, nil)
	clientMock.On("RemoveGroupFromPermissionTemplate", ctx, permissionTemplateID, &tplGroup).Return(nil)
	clientMock.On("AddGroupToPermissionTemplate", ctx, permissionTemplateID,
		&sonarClient.PermissionTemplateGroup{
			GroupName:   permissionTemplate1.Spec.GroupPermissions[0].GroupName,
			Permissions: permissionTemplate1.Spec.GroupPermissions[0].Permissions,
		}).Return(nil)
	serviceMock.
		On("DeleteResource",
			ctx,
			tMock.AnythingOfType("*v1.SonarPermissionTemplate"),
			finalizer,
			tMock.AnythingOfType("func() error"),
		).
		Return(true, nil)
	clientMock.On("DeletePermissionTemplate", permissionTemplateID).Return(nil)
	if _, err = rec.Reconcile(ctx, rq); err != nil {
		t.Fatal(err)
	}

	if err = l.LastError(); err != nil {
		t.Fatalf("%+v", err)
	}

	//ptpl.Status.ID = "id11"
	permissionTemplateID2 := "id11"
	permissionTemplate2 := sampleTemplate.DeepCopy()
	permissionTemplate2.Status.ID = permissionTemplateID2
	clientMock.On("GetPermissionTemplate", ctx, permissionTemplate2.Spec.Name).
		Return(specToClientTemplate(&permissionTemplate2.Spec, permissionTemplateID2), nil).Once()
	tpl := specToClientTemplate(&permissionTemplate2.Spec, permissionTemplateID2)
	//tpl.ID = ptpl.Status.ID
	clientMock.On("UpdatePermissionTemplate", ctx, tpl).Return(nil)
	clientMock.On("GetPermissionTemplateGroups", ctx, tpl.ID).
		Return(nil, errors.New("get perm groups fatal"))
	rec.client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(permissionTemplate2, &sn).Build()

	if _, err = rec.Reconcile(ctx, rq); err != nil {
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
	if isSpecUpdated(event.UpdateEvent{ObjectNew: &sonarApi.SonarPermissionTemplate{},
		ObjectOld: &sonarApi.SonarPermissionTemplate{}}) {
		t.Fatal("spec is updated")
	}
}
