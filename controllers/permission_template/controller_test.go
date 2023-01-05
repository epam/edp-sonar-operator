package permission_template

import (
	"context"
	"errors"
	"testing"

	tMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/epam/edp-common/pkg/mock"
	k8sMockClient "github.com/epam/edp-common/pkg/mock/controller-runtime/client"
	v1 "github.com/epam/edp-sonar-operator/v2/api/edp/v1"
	cMock "github.com/epam/edp-sonar-operator/v2/mocks/client"
	sMock "github.com/epam/edp-sonar-operator/v2/mocks/service"
	sonarClient "github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform"
)

func TestNewReconcile_NotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	rq := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "ns", Name: "name"}}
	fakeCl := fake.NewClientBuilder().WithScheme(scheme).Build()
	l := mock.NewLogr()
	rec, err := NewReconcile(fakeCl, scheme, l, "kubernetes")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = rec.Reconcile(context.Background(), rq); err != nil {
		t.Fatal(err)
	}

	loggerSink, ok := l.GetSink().(*mock.Logger)
	require.True(t, ok)

	if _, ok = loggerSink.InfoMessages()["instance not found"]; !ok {
		t.Fatal("no warning message logged")
	}

	k8sMock := k8sMockClient.Client{}
	k8sMock.On("Get", rq.NamespacedName, &v1.SonarPermissionTemplate{}).
		Return(errors.New("get fatal"))
	rec.client = &k8sMock

	_, err = rec.Reconcile(context.Background(), rq)
	if err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "failed to get sonar permission tpl from k8s: get fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

const (
	objectMetaName      = "tpl1"
	objectMetaNamespace = "ns"
)

func ObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      objectMetaName,
		Namespace: objectMetaNamespace,
	}
}

func TestNewReconcile(t *testing.T) {
	ctx := context.Background()
	sampleTemplate := v1.SonarPermissionTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SonarPermissionTemplate",
			APIVersion: "v2.edp.epam.com/v1",
		},
		ObjectMeta: ObjectMeta(),
		Spec: v1.SonarPermissionTemplateSpec{
			Name:              objectMetaName,
			ProjectKeyPattern: ".+",
			SonarOwner:        "sonar",
			Description:       "desc",
			GroupPermissions: []v1.GroupPermission{
				{
					GroupName:   "gr1",
					Permissions: []string{"admin", "user"},
				},
			},
		},
		Status: v1.SonarPermissionTemplateStatus{
			Value:        "",
			FailureCount: 0,
			ID:           "",
		},
	}
	sn := v1.Sonar{
		ObjectMeta: metav1.ObjectMeta{Name: "sonar", Namespace: objectMetaNamespace},
		TypeMeta:   metav1.TypeMeta{Kind: "Sonar", APIVersion: "v2.edp.epam.com/v1"},
		Spec:       v1.SonarSpec{BasePath: "path"},
	}

	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	permissionTemplate1 := sampleTemplate.DeepCopy()
	fakeCl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(permissionTemplate1, &sn).Build()
	l := mock.NewLogr()

	rec, err := NewReconcile(fakeCl, scheme, l, platform.Kubernetes)
	if err != nil {
		t.Fatal(err)
	}

	serviceMock := sMock.ServiceInterface{}
	rec.service = &serviceMock
	clientMock := &cMock.ClientInterface{}

	serviceMock.On("ClientForChild", ctx, tMock.AnythingOfType("*v1.SonarPermissionTemplate")).Return(clientMock, nil)
	serviceMock.On("K8sClient").Return(fakeCl)
	clientMock.
		On("GetPermissionTemplate", ctx, permissionTemplate1.Spec.Name).
		Return(nil, sonarClient.NotFoundError("not found")).
		Once()

	permissionTemplateID := "uniq_tpl_id_1"
	clientMock.
		On("CreatePermissionTemplate", ctx, specToClientTemplateData(&permissionTemplate1.Spec)).
		Return(permissionTemplateID, nil)

	tplGroup := sonarClient.PermissionTemplateGroup{GroupName: "baz", Permissions: []string{"scan"}}
	clientMock.
		On("GetPermissionTemplateGroups", ctx, permissionTemplateID).
		Return([]sonarClient.PermissionTemplateGroup{tplGroup}, nil)
	clientMock.On("RemoveGroupFromPermissionTemplate", ctx, permissionTemplateID, &tplGroup).Return(nil)
	clientMock.
		On("AddGroupToPermissionTemplate", ctx, permissionTemplateID,
			&sonarClient.PermissionTemplateGroup{
				GroupName:   permissionTemplate1.Spec.GroupPermissions[0].GroupName,
				Permissions: permissionTemplate1.Spec.GroupPermissions[0].Permissions,
			}).
		Return(nil)
	serviceMock.
		On("DeleteResource",
			ctx,
			tMock.AnythingOfType("*v1.SonarPermissionTemplate"),
			finalizer,
			tMock.AnythingOfType("func() error"),
		).
		Return(true, nil)
	clientMock.On("DeletePermissionTemplate", permissionTemplateID).Return(nil)

	rq := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: objectMetaNamespace, Name: objectMetaName}}
	if _, err = rec.Reconcile(ctx, rq); err != nil {
		t.Fatal(err)
	}

	loggerSink, ok := l.GetSink().(*mock.Logger)
	require.True(t, ok)

	if err = loggerSink.LastError(); err != nil {
		t.Fatalf("%+v", err)
	}

	permissionTemplateID2 := "id11"
	permissionTemplate2 := sampleTemplate.DeepCopy()
	permissionTemplate2.Status.ID = permissionTemplateID2
	clientMock.On("GetPermissionTemplate", ctx, permissionTemplate2.Spec.Name).
		Return(specToClientTemplate(&permissionTemplate2.Spec, permissionTemplateID2), nil).Once()
	tpl := specToClientTemplate(&permissionTemplate2.Spec, permissionTemplateID2)

	clientMock.On("UpdatePermissionTemplate", ctx, tpl).Return(nil)
	clientMock.On("GetPermissionTemplateGroups", ctx, tpl.ID).
		Return(nil, errors.New("get perm groups fatal"))
	rec.client = fake.NewClientBuilder().WithScheme(scheme).WithObjects(permissionTemplate2, &sn).Build()

	if _, err = rec.Reconcile(ctx, rq); err != nil {
		t.Fatal(err)
	}

	if err = loggerSink.LastError(); err == nil {
		t.Fatal("no error returned")
	}

	if err.Error() != "unable to sync permission template groups: unable to get permission template groups: get perm groups fatal" {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
}

func TestSpecIsUpdated(t *testing.T) {
	if isSpecUpdated(event.UpdateEvent{ObjectNew: &v1.SonarPermissionTemplate{},
		ObjectOld: &v1.SonarPermissionTemplate{}}) {
		t.Fatal("spec is updated")
	}
}
