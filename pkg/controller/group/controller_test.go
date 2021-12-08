package group

import (
	"context"
	"github.com/epam/edp-common/pkg/mock"
	"testing"
	"time"

	"github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1alpha1"
	sonarClient "github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/sonar"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestNewReconcile(t *testing.T) {
	tm := metav1.Time{Time: time.Now()}

	sg := v1alpha1.SonarGroup{
		Spec: v1alpha1.SonarGroupSpec{
			SonarOwner:  "sonar",
			Name:        "sc-name",
			Description: "sc-desc",
		},
		Status: v1alpha1.SonarGroupStatus{ID: "id1"},
		TypeMeta: metav1.TypeMeta{
			Kind:       "SonarGroup",
			APIVersion: "v2.edp.epam.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "sg1", Namespace: "ns1",
			DeletionTimestamp: &tm},
	}

	sn := v1alpha1.Sonar{
		ObjectMeta: metav1.ObjectMeta{Name: "sonar", Namespace: sg.Namespace},
		TypeMeta:   metav1.TypeMeta{Kind: "Sonar", APIVersion: "v2.edp.epam.com/v1alpha1"},
		Spec:       v1alpha1.SonarSpec{BasePath: "path"},
	}
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	rq := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: sg.Namespace, Name: sg.Name}}
	fakeCl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&sg, &sn).Build()
	l := mock.Logger{}
	rec, err := NewReconcile(fakeCl, scheme, &l, "kubernetes")
	if err != nil {
		t.Fatal(err)
	}

	serviceMock := sonar.ServiceMock{}
	rec.service = &serviceMock
	clientMock := sonar.ClientMock{}

	serviceMock.On("ClientForChild").Return(&clientMock, nil)
	serviceMock.On("DeleteResource").Return(true, nil)
	clientMock.On("GetGroup", sg.Spec.Name).Return(&sonarClient.Group{}, nil).Once()
	clientMock.On("UpdateGroup", sg.Spec.Name,
		&sonarClient.Group{Name: sg.Spec.Name, Description: sg.Spec.Description}).Return(nil).Once()
	clientMock.On("DeleteGroup", sg.Spec.Name).Return(nil)
	if _, err := rec.Reconcile(context.Background(), rq); err != nil {
		t.Fatal(err)
	}

	if err := l.LastError(); err != nil {
		t.Fatal(err)
	}

	clientMock.On("GetGroup", sg.Spec.Name).
		Return(nil, sonarClient.ErrNotFound("not found")).Once()
	clientMock.On("CreateGroup",
		&sonarClient.Group{Name: sg.Spec.Name, Description: sg.Spec.Description}).Return(nil)
	if _, err := rec.Reconcile(context.Background(), rq); err != nil {
		t.Fatal(err)
	}

}
