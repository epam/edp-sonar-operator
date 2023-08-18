package sonar

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type QualityProfile struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

type qProfileSearchResponse struct {
	Profiles []QualityProfile `json:"profiles"`
}

func (sc Client) GetQualityProfile(ctx context.Context, name string) (*QualityProfile, error) {
	var searchRsp qProfileSearchResponse
	rsp, err := sc.startRequest(ctx).SetResult(&searchRsp).
		Get(fmt.Sprintf("/qualityprofiles/search?qualityProfile=%s",
			strings.ReplaceAll(name, " ", "+")))
	if err = sc.checkError(rsp, err); err != nil {
		return nil, fmt.Errorf("failed to get quality profile: %w", err)
	}

	for _, r := range searchRsp.Profiles {
		if r.Name == name {
			return &r, nil
		}
	}

	return nil, NewHTTPError(http.StatusNotFound, fmt.Sprintf("quality profile %s not found", name))
}
