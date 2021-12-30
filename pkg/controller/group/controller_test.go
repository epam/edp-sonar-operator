package group

import (
	"context"
	"testing"
	"time"

	"github.com/epam/edp-common/pkg/mock"
	tMock "github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1alpha1"
	sonarClient "github.com/epam/edp-sonar-operator/v2/pkg/client/sonar"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/platform"
	sonarMocks "github.com/epam/edp-sonar-operator/v2/pkg/service/sonar/mocks"
)

func TestNewReconcile(t *testing.T) {
	tm := metav1.Time{Time: time.Now()}
	ctx := context.Background()

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
	rec, err := NewReconcile(fakeCl, scheme, &l, platform.Kubernetes)
	if err != nil {
		t.Fatal(err)
	}

	serviceMock := sonarMocks.ServiceInterface{}
	rec.service = &serviceMock
	clientMock := sonarMocks.ClientInterface{}

	serviceMock.On("ClientForChild", ctx, tMock.AnythingOfType("*v1alpha1.SonarGroup")).Return(&clientMock, nil)
	serviceMock.
		On("DeleteResource",
			ctx,
			tMock.AnythingOfType("*v1alpha1.SonarGroup"),
			finalizer,
			tMock.AnythingOfType("func() error"),
		).
		Return(true, nil)
	clientMock.On("GetGroup", ctx, sg.Spec.Name).Return(&sonarClient.Group{}, nil)
	clientMock.
		On("UpdateGroup",
			ctx,
			sg.Spec.Name,
			&sonarClient.Group{
				Name:        sg.Spec.Name,
				Description: sg.Spec.Description,
			}).
		Return(nil)
	if _, err = rec.Reconcile(ctx, rq); err != nil {
		t.Fatal(err)
	}

	if err = l.LastError(); err != nil {
		t.Fatal(err)
	}

	serviceMock.AssertExpectations(t)
	clientMock.AssertExpectations(t)
}
