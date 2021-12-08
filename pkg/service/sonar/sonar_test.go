package sonar

import (
	"context"
	"testing"
	"time"

	coreV1Api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSonarServiceImpl_DeleteResource(t *testing.T) {
	secret := coreV1Api.Secret{ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "ns"}}
	s := SonarServiceImpl{
		k8sClient: fake.NewClientBuilder().WithRuntimeObjects(&secret).Build(),
	}

	if _, err := s.DeleteResource(context.Background(), &secret, "fin", func() error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	secret.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	secret.Finalizers = []string{"fin"}
	s.k8sClient = fake.NewClientBuilder().WithRuntimeObjects(&secret).Build()

	if _, err := s.DeleteResource(context.Background(), &secret, "fin", func() error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}

}
