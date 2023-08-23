package sonar

import (
	"context"
	"fmt"
)

type Rule struct {
	//
	Key      string `json:"key"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Params   string `json:"-"`
}

// GetQualityProfileActiveRules returns the active rules of the quality profile with the given key.
func (sc *Client) GetQualityProfileActiveRules(ctx context.Context, profileKey string) ([]Rule, error) {
	rulesResp := struct {
		Rules []Rule `json:"rules"`
	}{}

	resp, err := sc.startRequest(ctx).
		SetQueryParams(map[string]string{
			"activation": "true",
			"qprofile":   profileKey,
			"ps":         "500",
		}).
		SetResult(&rulesResp).
		Get("/rules/search")

	if err = sc.checkError(resp, err); err != nil {
		return nil, fmt.Errorf("failed to get quality profile active rules: %w", err)
	}

	return rulesResp.Rules, nil
}
