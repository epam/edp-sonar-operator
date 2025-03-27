package chain

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

// SyncQualityProfileRules is a handler for syncing quality profile rules.
type SyncQualityProfileRules struct {
	sonarApiClient sonarApiClient
}

// NewSyncQualityProfileRules returns an instance of SyncQualityProfileRules handler.
func NewSyncQualityProfileRules(sonarApiClient sonarApiClient) *SyncQualityProfileRules {
	return &SyncQualityProfileRules{sonarApiClient: sonarApiClient}
}

// ServeRequest implements the logic of syncing quality profile rules.
func (h SyncQualityProfileRules) ServeRequest(ctx context.Context, profile *sonarApi.SonarQualityProfile) error {
	log := ctrl.LoggerFrom(ctx).WithValues("name", profile.Spec.Name)
	log.Info("Start syncing quality profile rules")

	sonarProfile, err := h.sonarApiClient.GetQualityProfile(ctx, profile.Spec.Name)
	if err != nil {
		return fmt.Errorf("failed to get quality profile: %w", err)
	}

	activeRules, err := h.sonarApiClient.GetQualityProfileActiveRules(ctx, sonarProfile.Key)
	if err != nil {
		return fmt.Errorf("failed to get quality profile active rules: %w", err)
	}

	existingRulesMap := rulesToMap(activeRules)

	for ruleKey, rule := range profile.Spec.Rules {
		_, ok := existingRulesMap[ruleKey]
		if ok {
			delete(existingRulesMap, ruleKey)
			continue
		}

		log.Info("Activating quality profile rule", "rule", ruleKey)

		if err = h.sonarApiClient.ActivateQualityProfileRule(
			ctx,
			sonarProfile.Key,
			sonar.Rule{
				Rule:     ruleKey,
				Severity: rule.Severity,
				Params:   rule.Params,
			},
		); err != nil {
			return fmt.Errorf("failed to acticate rule: %w", err)
		}
	}

	for ruleKey := range existingRulesMap {
		log.Info("Deactivating quality profile rule", "rule", ruleKey)

		if err = h.sonarApiClient.DeactivateQualityProfileRule(ctx, sonarProfile.Key, ruleKey); err != nil {
			return fmt.Errorf("failed to deactivate quality profile rule: %w", err)
		}
	}

	log.Info("Quality profile rules have been synced")

	return nil
}

func rulesToMap(rules []sonar.Rule) map[string]sonar.Rule {
	res := make(map[string]sonar.Rule, len(rules))
	for _, r := range rules {
		res[r.Key] = r
	}

	return res
}
