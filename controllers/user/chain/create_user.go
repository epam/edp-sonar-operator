package chain

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

// CreateUser is handler for creating sonar user.
type CreateUser struct {
	sonarApiClient sonar.UserInterface
	client         client.Client
}

// NewCreateUser returns CreateUser handler.
func NewCreateUser(sonarApiClient sonar.UserInterface, client client.Client) *CreateUser {
	return &CreateUser{sonarApiClient: sonarApiClient, client: client}
}

// ServeRequest handles sonar user creation.
func (h CreateUser) ServeRequest(ctx context.Context, user *sonarApi.SonarUser) error {
	log := ctrl.LoggerFrom(ctx).WithValues("userlogin", user.Spec.Login)
	log.Info("Creating user in sonar")

	userSecret := &corev1.Secret{}
	if err := h.client.Get(ctx, client.ObjectKey{
		Namespace: user.Namespace,
		Name:      user.Spec.Secret,
	}, userSecret); err != nil {
		return fmt.Errorf("failed to get user secret: %w", err)
	}

	sonarUser := &sonar.User{
		Login:    user.Spec.Login,
		Name:     user.Spec.Name,
		Password: string(userSecret.Data["password"]),
		Email:    user.Spec.Email,
	}

	existingUser, err := h.sonarApiClient.GetUserByLogin(ctx, user.Spec.Login)
	if err != nil {
		if !sonar.IsErrNotFound(err) {
			return fmt.Errorf("failed to get user: %w", err)
		}

		log.Info("User not found, creating new one")

		if err = h.sonarApiClient.CreateUser(ctx, sonarUser); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		log.Info("User has been created")
		return nil
	}

	log.Info("User already exists, updating")

	// to check if user needs to be updated we need to clear password as it is not returned by sonar
	sonarUser.Password = ""
	if reflect.DeepEqual(sonarUser, existingUser) {
		log.Info("User already up to date")
		return nil
	}

	if err = h.sonarApiClient.UpdateUser(ctx, sonarUser); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	log.Info("User has been updated")
	return nil
}
