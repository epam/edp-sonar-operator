package sonar

import (
	"context"
	"fmt"
)

type QualityGates map[string]QualityGateSettings

type QualityGateProperty struct {
	Name        string
	MakeDefault bool
}

type QualityGateSettings struct {
	MakeDefault bool
	Conditions  []QualityGateCondition
}

type QualityGateCondition struct {
	ID     string `json:"id"`
	Error  string `json:"error"`
	Metric string `json:"metric"`
	OP     string `json:"op"`
}

func (q *QualityGateCondition) ToQueryParamMap() map[string]string {
	conditionMap := map[string]string{
		"gateID": q.ID,
		"error":  q.Error,
		"metric": q.Metric,
		"op":     q.OP,
	}

	return conditionMap
}

type QualityGate struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	IsDefault  bool                   `json:"isDefault"`
	IsBuiltIn  bool                   `json:"isBuiltIn"`
	Conditions []QualityGateCondition `json:"conditions"`
}

type Actions struct {
	Rename            bool `json:"rename"`
	SetAsDefault      bool `json:"setAsDefault"`
	Copy              bool `json:"copy"`
	AssociateProjects bool `json:"associateProjects"`
	Delete            bool `json:"delete"`
	ManageConditions  bool `json:"manageConditions"`
}

// CreateQualityGate creates a new quality gate.
// Returns the created quality gate only with ID and name fields filled.
func (sc *Client) CreateQualityGate(ctx context.Context, qualityGateName string) (*QualityGate, error) {
	gate := &QualityGate{}
	resp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{nameField: qualityGateName}).
		SetResult(gate).
		Post("/qualitygates/create")

	if err = sc.checkError(resp, err); err != nil {
		return nil, fmt.Errorf("failed to create quality gate: %w", err)
	}

	return gate, nil
}

// GetQualityGate returns the quality gate with the given name.
func (sc *Client) GetQualityGate(ctx context.Context, name string) (*QualityGate, error) {
	gate := &QualityGate{}
	resp, err := sc.startRequest(ctx).
		SetQueryParam(nameField, name).
		SetResult(gate).
		Get("/qualitygates/show")

	if err = sc.checkError(resp, err); err != nil {
		return nil, fmt.Errorf("failed to get quality gate: %w", err)
	}

	return gate, nil
}

// DeleteQualityGate deletes the quality gate with the given name.
func (sc *Client) DeleteQualityGate(ctx context.Context, name string) error {
	resp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{nameField: name}).
		Post("/qualitygates/destroy")

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to delete quality gate: %w", err)
	}

	return nil
}

// SetAsDefaultQualityGate sets the quality gate with the given name as default.
func (sc *Client) SetAsDefaultQualityGate(ctx context.Context, name string) error {
	resp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{nameField: name}).
		Post("/qualitygates/set_as_default")

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to set default quality gate: %w", err)
	}

	return nil
}

// CreateQualityGateCondition creates a new quality gate condition.
func (sc *Client) CreateQualityGateCondition(ctx context.Context, gate string, condition QualityGateCondition) error {
	resp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			nameField: gate,
			"error":   condition.Error,
			"metric":  condition.Metric,
			"op":      condition.OP,
		}).
		Post("/qualitygates/create_condition")

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to create quality gate condition: %w", err)
	}

	return nil
}

// UpdateQualityGateCondition updates the quality gate condition.
func (sc *Client) UpdateQualityGateCondition(ctx context.Context, condition QualityGateCondition) error {
	resp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"id":     condition.ID,
			"metric": condition.Metric,
			"error":  condition.Error,
			"op":     condition.OP,
		}).
		Post("/qualitygates/update_condition")

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to update quality gate condition: %w", err)
	}

	return nil
}

// DeleteQualityGateCondition deletes the quality gate condition with the given ID.
func (sc *Client) DeleteQualityGateCondition(ctx context.Context, conditionId string) error {
	resp, err := sc.startRequest(ctx).
		SetFormData(map[string]string{
			"id": conditionId,
		}).
		Post("/qualitygates/delete_condition")

	if err = sc.checkError(resp, err); err != nil {
		return fmt.Errorf("failed to delete quality gate condition: %w", err)
	}

	return nil
}
