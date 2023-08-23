package sonar

import (
	"context"
	"fmt"
	"net/http"
)

type QualityProfile struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	IsDefault bool   `json:"isDefault"`
}

// CreateQualityProfile creates a new quality profile.
func (sc *Client) CreateQualityProfile(ctx context.Context, name, language string) (*QualityProfile, error) {
	profile := struct {
		Profile QualityProfile `json:"profile"`
	}{}

	resp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			nameField:  name,
			"language": language,
		}).
		SetResult(&profile).
		Post("/qualityprofiles/create")

	if err = sc.checkError(resp, err); err != nil {
		return nil, fmt.Errorf("failed to create quality profile: %w", err)
	}

	return &profile.Profile, nil
}

// GetQualityProfile returns the quality profile with the given name.
func (sc *Client) GetQualityProfile(ctx context.Context, name string) (*QualityProfile, error) {
	profiles := struct {
		Profiles []QualityProfile `json:"profiles"`
	}{}
	resp, err := sc.startRequest(ctx).
		SetQueryParam("qualityProfile", name).
		SetResult(&profiles).
		Get("/qualityprofiles/search")

	if err = sc.checkError(resp, err); err != nil {
		return nil, fmt.Errorf("failed to get quality profile: %w", err)
	}

	for _, profile := range profiles.Profiles {
		if profile.Name == name {
			return &profile, nil
		}
	}

	return nil, NewHTTPError(http.StatusNotFound, fmt.Sprintf("quality profile %s not found", name))
}

// DeleteQualityProfile deletes the quality profile with the given name and language.
func (sc *Client) DeleteQualityProfile(ctx context.Context, name, language string) error {
	resp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			nameField:  name,
			"language": language,
		}).
		Post("/qualityprofiles/delete")

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to delete quality profile: %w", err)
	}

	return nil
}

// SetAsDefaultQualityProfile sets the quality profile with the given name and language as default.
func (sc *Client) SetAsDefaultQualityProfile(ctx context.Context, name, language string) error {
	resp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"qualityProfile": name,
			"language":       language,
		}).
		Post("/qualityprofiles/set_default")

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to set default quality profile: %w", err)
	}

	return nil
}

// ActivateQualityProfileRule activates the rule in the quality profile.
func (sc *Client) ActivateQualityProfileRule(ctx context.Context, profileKey string, rule Rule) error {
	resp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"key":      profileKey,
			"rule":     rule.Rule,
			"severity": rule.Severity,
			"params":   rule.Params,
		}).
		Post("/qualityprofiles/activate_rule")

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to activate rule %s: %w", rule.Key, err)
	}

	return nil
}

// DeactivateQualityProfileRule deactivates the rule in the quality profile.
func (sc *Client) DeactivateQualityProfileRule(ctx context.Context, profileKey, ruleKey string) error {
	resp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"key":  profileKey,
			"rule": ruleKey,
		}).
		Post("/qualityprofiles/deactivate_rule")

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to deactivate rule %s: %w", ruleKey, err)
	}

	return nil
}
