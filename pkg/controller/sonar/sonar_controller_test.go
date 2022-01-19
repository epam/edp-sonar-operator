package sonar

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/epam/edp-common/pkg/mock"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/epam/edp-sonar-operator/v2/pkg/apis/edp/v1alpha1"
)

func TestReconcileSonar_Reconcile_ShouldFailToCreateDBSecret(t *testing.T) {
	sn := v1alpha1.Sonar{
		ObjectMeta: metav1.ObjectMeta{Name: "sonar", Namespace: "fake"},
		TypeMeta:   metav1.TypeMeta{Kind: "Sonar", APIVersion: "v2.edp.epam.com/v1alpha1"},
		Spec:       v1alpha1.SonarSpec{BasePath: "path"},
	}
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	rq := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: sn.Namespace, Name: sn.Name}}
	fakeCl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(&sn).Build()
	l := mock.Logger{}
	rec, err := NewReconcileSonar(fakeCl, scheme, &l, "kubernetes")
	if err != nil {
		t.Fatal(err)
	}

	rs, err := rec.Reconcile(context.Background(), rq)
	assert.Error(t, err)
	if !strings.Contains(err.Error(), "Failed to create secret for sonar-db") {
		t.Fatalf("wrong error returned: %s", err.Error())
	}
	assert.Equal(t, reconcile.Result{RequeueAfter: 30 * time.Second}, rs)
}
