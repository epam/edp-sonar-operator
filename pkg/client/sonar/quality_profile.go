package sonar

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "unable to get quality profile")
	}

	for _, r := range searchRsp.Profiles {
		if r.Name == name {
			return &r, nil
		}
	}

	return nil, ErrNotFound("quality profile not found")
}
