package sonar

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/epam/edp-sonar-operator/api/common"
	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
)

// ApiClientProvider is a struct for providing sonar api client.
type ApiClientProvider struct {
	k8sClient client.Client
}

// NewApiClientProvider returns a new instance of ApiClientProvider.
func NewApiClientProvider(k8sClient client.Client) *ApiClientProvider {
	return &ApiClientProvider{k8sClient: k8sClient}
}

// GetSonarApiClientFromSonar returns sonar api client from sonar CR.
func (p *ApiClientProvider) GetSonarApiClientFromSonar(ctx context.Context, sonar *sonarApi.Sonar) (*Client, error) {
	secret := corev1.Secret{}
	if err := p.k8sClient.Get(ctx, types.NamespacedName{
		Name:      sonar.Spec.Secret,
		Namespace: sonar.Namespace,
	}, &secret); err != nil {
		return nil, fmt.Errorf("failed to get sonar secret: %w", err)
	}

	if secret.Data["user"] == nil {
		return nil, fmt.Errorf("sonar secret doesn't contain user")
	}

	password := ""
	if secret.Data["password"] != nil {
		password = string(secret.Data["password"])
	}

	return NewClient(sonar.Spec.Url, string(secret.Data["user"]), password), nil
}

// GetSonarApiClientFromSonarRef returns sonar api client from sonar ref.
func (p *ApiClientProvider) GetSonarApiClientFromSonarRef(
	ctx context.Context,
	namespace string,
	sonarRef common.HasSonarRef,
) (*Client, error) {
	sonar := &sonarApi.Sonar{}
	if err := p.k8sClient.Get(ctx, types.NamespacedName{
		Name:      sonarRef.GetSonarRef().Name,
		Namespace: namespace,
	}, sonar); err != nil {
		return nil, fmt.Errorf("failed to get sonar: %w", err)
	}

	if !sonar.Status.Connected {
		return nil, errors.New("sonar is not connected")
	}

	return p.GetSonarApiClientFromSonar(ctx, sonar)
}
