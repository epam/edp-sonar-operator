package platform

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewService_NonValidPlatform(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	scheme := runtime.NewScheme()
	platformType := "test"
	service, err := NewService(platformType, scheme, client)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "is not supported"))
	assert.Nil(t, service)
}

func TestNewService_K8SPlatform(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	scheme := runtime.NewScheme()
	platformType := Kubernetes
	service, err := NewService(platformType, scheme, client)
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewService_OpenshiftPlatform(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	scheme := runtime.NewScheme()
	platformType := Openshift
	service, err := NewService(platformType, scheme, client)
	assert.NoError(t, err)
	assert.NotNil(t, service)
}
