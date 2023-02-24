package sonar

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
	ID     string `json:"gateID"`
	Error  string `json:"error"`
	Metric string `json:"metric"`
	OP     string `json:"op"`
	Period string `json:"period,omitempty"`
}

func (q *QualityGateCondition) ToQueryParamMap() map[string]string {
	conditionMap := map[string]string{
		"gateID": q.ID,
		"error":  q.Error,
		"metric": q.Metric,
		"op":     q.OP,
	}

	if q.Period != "" {
		conditionMap["period"] = q.Period
	}

	return conditionMap
}
