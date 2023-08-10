package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"

	sonarApi "github.com/epam/edp-sonar-operator/api/v1alpha1"
	"github.com/epam/edp-sonar-operator/pkg/client/sonar"
)

type UpdateSettings struct {
	sonarApiClient sonar.Settings
}

func NewUpdateSettings(sonarApiClient sonar.Settings) *UpdateSettings {
	return &UpdateSettings{sonarApiClient: sonarApiClient}
}

func (h *UpdateSettings) ServeRequest(ctx context.Context, sonar *sonarApi.Sonar) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Start updating settings to sonar")

	// if the user removes a setting from the CR, we need to reset it in Sonar
	settingsToReset := getSettingsKeysMap(sonar.Status)
	// we need to save processed settings to know which settings we need to reset
	processedSettings := make([]string, 0, len(sonar.Spec.Settings))

	for _, s := range sonar.Spec.Settings {
		if err := h.sonarApiClient.SetSetting(ctx, makeSetting(s)); err != nil {
			return fmt.Errorf("failed to set setting %s: %w", s.Key, err)
		}

		processedSettings = append(processedSettings, s.Key)
		delete(settingsToReset, s.Key)
	}

	if len(settingsToReset) != 0 {
		if err := h.sonarApiClient.ResetSettings(ctx, settingsKeysMapToSlice(settingsToReset)); err != nil {
			return fmt.Errorf("failed to reset settings: %w", err)
		}
	}

	setProcessedSettings(&sonar.Status, processedSettings)

	log.Info("Sonar settings have been updated")

	return nil
}

func getSettingsKeysMap(status sonarApi.SonarStatus) map[string]struct{} {
	var processedSettings []string
	if status.ProcessedSettings != "" {
		processedSettings = strings.Split(status.ProcessedSettings, ",")
	}
	m := make(map[string]struct{}, len(processedSettings))

	for _, s := range processedSettings {
		m[s] = struct{}{}
	}

	return m
}

func settingsKeysMapToSlice(m map[string]struct{}) []string {
	s := make([]string, 0, len(m))

	for k := range m {
		s = append(s, k)
	}

	return s
}

func makeSetting(setting sonarApi.SonarSetting) url.Values {
	if setting.FieldValues != nil {
		// nolint:errchkjson //we can skip error for marshal map[string]string
		fv, _ := json.Marshal(setting.FieldValues)

		return url.Values{
			"key":         []string{setting.Key},
			"fieldValues": []string{string(fv)},
		}
	}

	if setting.Values != nil {
		return url.Values{
			"key":    []string{setting.Key},
			"values": setting.Values,
		}
	}

	return url.Values{
		"key":   []string{setting.Key},
		"value": []string{setting.Value},
	}
}

func setProcessedSettings(status *sonarApi.SonarStatus, settingsKeys []string) {
	// we need to sort keys to make sure that we have the same order of keys in the status
	// to not update ProcessedSettings filed every time
	sort.Strings(settingsKeys)
	status.ProcessedSettings = strings.Join(settingsKeys, ",")
}
