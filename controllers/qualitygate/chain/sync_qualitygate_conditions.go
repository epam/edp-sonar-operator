package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

// SyncQualityGateConditions is a handler for syncing quality gate conditions.
type SyncQualityGateConditions struct {
	sonarApiClient sonar.QualityGateClient
}

// NewSyncQualityGateConditions creates an instance of SyncQualityGateConditions handler.
func NewSyncQualityGateConditions(sonarApiClient sonar.QualityGateClient) *SyncQualityGateConditions {
	return &SyncQualityGateConditions{sonarApiClient: sonarApiClient}
}

// ServeRequest implements the logic of syncing quality gate conditions.
func (h SyncQualityGateConditions) ServeRequest(ctx context.Context, gate *sonarApi.SonarQualityGate) error {
	log := ctrl.LoggerFrom(ctx).WithValues("name", gate.Spec.Name)
	log.Info("Start syncing quality gate conditions")

	sonarGate, err := h.sonarApiClient.GetQualityGate(ctx, gate.Spec.Name)
	if err != nil {
		return fmt.Errorf("failed to get quality gate: %w", err)
	}

	existingCondMap := conditionsToMap(sonarGate.Conditions)

	for metric, cond := range gate.Spec.Conditions {
		existingCond, ok := existingCondMap[metric]
		if !ok {
			log.Info("Creating quality gate condition", "metric", metric)

			if err = h.sonarApiClient.CreateQualityGateCondition(
				ctx,
				gate.Spec.Name,
				sonar.QualityGateCondition{
					Error:  cond.Error,
					Metric: metric,
					OP:     cond.Op,
				},
			); err != nil {
				return fmt.Errorf("failed to create quality gate condition: %w", err)
			}

			continue
		}

		if existingCond.Error != cond.Error || existingCond.OP != cond.Op {
			log.Info("Updating quality gate condition", "metric", metric)

			existingCond.Error = cond.Error
			existingCond.OP = cond.Op
			if err = h.sonarApiClient.UpdateQualityGateCondition(ctx, existingCond); err != nil {
				return fmt.Errorf("failed to update quality gate condition: %w", err)
			}
		}

		delete(existingCondMap, metric)
	}

	for _, cond := range existingCondMap {
		log.Info("Deleting quality gate condition", "metric", cond.Metric)

		if err = h.sonarApiClient.DeleteQualityGateCondition(ctx, cond.ID); err != nil {
			return fmt.Errorf("failed to delete quality gate condition: %w", err)
		}
	}

	log.Info("Quality gate conditions have been synced")

	return nil
}

func conditionsToMap(conditions []sonar.QualityGateCondition) map[string]sonar.QualityGateCondition {
	res := make(map[string]sonar.QualityGateCondition, len(conditions))
	for _, c := range conditions {
		res[c.Metric] = c
	}

	return res
}
