package sonar

import (
	"context"
	"fmt"
)

type SystemHealth struct {
	// GREEN: SonarQube is fully operational
	// YELLOW: SonarQube is usable, but it needs attention in order to be fully operational
	// RED: SonarQube is not operational
	Health string `json:"health"`
	Causes []any  `json:"causes"`
	Nodes  []any  `json:"nodes"`
}

func (sc *Client) Health(ctx context.Context) (*SystemHealth, error) {
	health := &SystemHealth{}

	rsp, err := sc.startRequest(ctx).
		SetResult(health).
		Get("/system/health")

	if err = sc.checkError(rsp, err); err != nil {
		return nil, fmt.Errorf("failed to check health: %w", err)
	}

	return health, nil
}
